package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/containers/image/v5/transports/alltransports"
	"github.com/containers/image/v5/types"
	"github.com/opencontainers/go-digest"
	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/operator-framework/operator-registry/alpha/declcfg"
)

// loger defines the interface this package expects for logging.
type loger interface {
	Infof(format string, args ...interface{})
	Info(args ...interface{})
	Debugf(format string, args ...interface{})
	Debug(args ...interface{})
}

// imager defines the interface this package expects for image operations.
type imager interface {
	RemoteInfo(imageRef string) (string, string, digest.Digest, error)
	CopyToOci(imageRef, ociDir string) (string, error)
}

// fsioer defines the interface this package expects for filesystem I/O.
type fsioer interface {
	CopyDirectory(src, dst string) error
	UntarFromStream(r io.Reader, dest string) error
}

// Cataloger provides methods for introspecting catalogs.
type Cataloger struct {
	log    loger
	imager imager
	fsio   fsioer
}

// NewCataloger creates a new Cataloger with its dependencies.
func NewCataloger(log loger, imager imager, fsio fsioer) *Cataloger {
	return &Cataloger{
		log:    log,
		imager: imager,
		fsio:   fsio,
	}
}

func (c *Cataloger) CatalogConfig(imageRef string) (*declcfg.DeclarativeConfig, error) {
	name, tag, digest, err := c.imager.RemoteInfo(imageRef)
	if err != nil {
		return nil, fmt.Errorf("failed to get remote info for %s: %w", imageRef, err)
	}

	// TODO not sure if this safeDigest is the correct one to use.
	// from the tests, it seems that the digest is not the correct one to use.
	safeDigest := strings.Replace(digest.String(), ":", "-", 1)
	configsCachePath := filepath.Join("working-dir", "operator-catalogs", name, tag, safeDigest, "configs")
	baseCachePath := filepath.Dir(configsCachePath)

	c.log.Debugf("Checking for cached catalog at %s...", configsCachePath)
	if _, err := os.Stat(configsCachePath); err != nil {
		c.log.Debug("Cache miss. Pulling image and extracting catalog...")
		// Create a temporary directory to pull the full OCI layout
		tmpOciLayoutDir, err := os.MkdirTemp("", "lumen-oci-layout-")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp oci layout dir: %w", err)
		}
		defer os.RemoveAll(tmpOciLayoutDir) // Clean up the full layout after we're done

		// Pull by digest to ensure we get the correct, immutable image version
		imageRefWithDigest := fmt.Sprintf("%s@%s", name, digest)
		if _, err := c.imager.CopyToOci(imageRefWithDigest, tmpOciLayoutDir); err != nil {
			return nil, fmt.Errorf("failed to copy image to oci: %w", err)
		}

		tmpExtractDir, err := os.MkdirTemp("", "lumen-extract-")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp extraction dir: %w", err)
		}
		defer os.RemoveAll(tmpExtractDir)

		declarativeConfigDir, err := extractCatalogConfig(c.fsio, tmpOciLayoutDir, tmpExtractDir)
		if err != nil {
			return nil, fmt.Errorf("failed to find and extract catalog: %w", err)
		}

		// Ensure the final cache directory exists
		if err := os.MkdirAll(baseCachePath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create cache directory %s: %w", baseCachePath, err)
		}

		// Move the extracted 'configs' directory to its permanent cache location
		sourceConfigsDir := filepath.Join(declarativeConfigDir, "configs")
		if err := c.fsio.CopyDirectory(sourceConfigsDir, configsCachePath); err != nil {
			return nil, fmt.Errorf("failed to copy configs to cache: %w", err)
		}
	} else {
		c.log.Debug("Cache hit. Loading catalog from existing directory.")
	}

	fsys := os.DirFS(configsCachePath)

	c.log.Debug("Loading declarative config from filesystem...")
	cfg, err := declcfg.LoadFS(context.Background(), fsys)
	if err != nil {
		return nil, fmt.Errorf("failed to load declarative config: %w", err)
	}
	c.log.Debug("Successfully loaded catalog config.")
	return cfg, nil
}

func extractCatalogConfig(fsSvc fsioer, ociLayoutDir, tmpDir string) (string, error) {
	srcRef, err := alltransports.ParseImageName(fmt.Sprintf("oci:%s", ociLayoutDir))
	if err != nil {
		return "", fmt.Errorf("failed to parse oci layout name: %w", err)
	}
	imgSrc, err := srcRef.NewImageSource(context.Background(), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create image source from oci layout: %w", err)
	}
	defer imgSrc.Close()

	manifestBytes, _, err := imgSrc.GetManifest(context.Background(), nil)
	if err != nil {
		return "", fmt.Errorf("failed to get manifest from oci layout: %w", err)
	}

	var manifest ociv1.Manifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return "", fmt.Errorf("failed to unmarshal oci manifest: %w", err)
	}

	for _, layer := range manifest.Layers {
		blobStream, _, err := imgSrc.GetBlob(context.Background(), types.BlobInfo{Digest: layer.Digest, Size: layer.Size, Annotations: layer.Annotations}, nil)
		if err != nil {
			// Non-fatal, just continue to the next layer.
			continue
		}

		extractDir, err := os.MkdirTemp(tmpDir, "layer-")
		if err != nil {
			blobStream.Close()
			return "", fmt.Errorf("failed to create temp dir for layer extraction: %w", err)
		}

		if err := fsSvc.UntarFromStream(blobStream, extractDir); err != nil {
			os.RemoveAll(extractDir)
			blobStream.Close()
			continue
		}
		blobStream.Close()

		// Check if the extracted directory contains a 'configs' subdirectory,
		// which is the root of the FBC.
		fsys := os.DirFS(extractDir)
		if _, err := fs.Stat(fsys, "configs"); err == nil {
			// We found the catalog, so we can stop searching.
			return extractDir, nil
		}

		// Cleanup the directory for the current layer if it's not the one we want.
		os.RemoveAll(extractDir)
	}

	return "", fmt.Errorf("failed to find a valid FBC in any layer")
}
