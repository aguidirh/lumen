package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/aguidirh/lumen/pkg/fsio"
	"github.com/aguidirh/lumen/pkg/image"
	"github.com/containers/image/v5/transports/alltransports"
	"github.com/containers/image/v5/types"
	ociv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/operator-framework/operator-registry/alpha/declcfg"
)

// TODO improve teh log mechanism to something more robust

func CatalogConfig(imageRef string) (*declcfg.DeclarativeConfig, error) {
	name, tag, digest, err := image.RemoteInfo(imageRef)
	if err != nil {
		return nil, fmt.Errorf("failed to get remote info for %s: %w", imageRef, err)
	}

	// TODO not sure if this safeDigest is the correct one to use.
	// from the tests, it seems that the digest is not the correct one to use.
	safeDigest := strings.Replace(digest.String(), ":", "-", 1)
	configsCachePath := filepath.Join("working-dir", "operator-catalogs", name, tag, safeDigest, "configs")
	baseCachePath := filepath.Dir(configsCachePath)

	fmt.Printf("INFO: Checking for cached catalog at %s...\n", configsCachePath)
	if _, err := os.Stat(configsCachePath); err != nil {
		fmt.Println("INFO: Cache miss. Pulling image and extracting catalog...")
		// Create a temporary directory to pull the full OCI layout
		tmpOciLayoutDir, err := os.MkdirTemp("", "lumen-oci-layout-")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp oci layout dir: %w", err)
		}
		defer os.RemoveAll(tmpOciLayoutDir) // Clean up the full layout after we're done

		// Pull by digest to ensure we get the correct, immutable image version
		imageRefWithDigest := fmt.Sprintf("%s@%s", name, digest)
		if _, err := image.CopyToOci(imageRefWithDigest, tmpOciLayoutDir); err != nil {
			return nil, fmt.Errorf("failed to copy image to oci: %w", err)
		}

		tmpExtractDir, err := os.MkdirTemp("", "lumen-extract-")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp extraction dir: %w", err)
		}
		defer os.RemoveAll(tmpExtractDir)

		declarativeConfigDir, err := extractCatalogConfig(tmpOciLayoutDir, tmpExtractDir)
		if err != nil {
			return nil, fmt.Errorf("failed to find and extract catalog: %w", err)
		}

		// Ensure the final cache directory exists
		if err := os.MkdirAll(baseCachePath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create cache directory %s: %w", baseCachePath, err)
		}

		// Move the extracted 'configs' directory to its permanent cache location
		sourceConfigsDir := filepath.Join(declarativeConfigDir, "configs")
		if err := fsio.CopyDirectory(sourceConfigsDir, configsCachePath); err != nil {
			return nil, fmt.Errorf("failed to copy configs to cache: %w", err)
		}
	} else {
		fmt.Println("INFO: Cache hit. Loading catalog from existing directory.")
	}

	fsys := os.DirFS(configsCachePath)

	fmt.Println("INFO: Loading declarative config from filesystem...")
	cfg, err := declcfg.LoadFS(context.Background(), fsys)
	if err != nil {
		return nil, fmt.Errorf("failed to load declarative config: %w", err)
	}
	fmt.Println("INFO: Successfully loaded catalog config.")
	return cfg, nil
}

func extractCatalogConfig(ociLayoutDir, tmpDir string) (string, error) {
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

		if err := fsio.UntarFromStream(blobStream, extractDir); err != nil {
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
