// Package image provides functions for container image operations.
package image

import (
	"context"
	"fmt"

	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/docker/reference"
	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/transports/alltransports"
	"github.com/opencontainers/go-digest"
)

// Imager provides methods for container image operations.
type Imager struct {
	log Logger
}

// NewImager creates a new Imager instance.
func NewImager(log Logger) *Imager {
	return &Imager{log: log}
}

// PolicyContext returns a default policy context for container image operations.
func (i *Imager) PolicyContext() (*signature.PolicyContext, error) {
	policy, err := signature.DefaultPolicy(nil)
	if err != nil {
		return nil, fmt.Errorf("error getting default policy: %w", err)
	}
	policyCtx, err := signature.NewPolicyContext(policy)
	if err != nil {
		return nil, fmt.Errorf("error creating new policy context: %w", err)
	}
	return policyCtx, nil
}

// CopyToOci copies an image from a Docker registry to a local OCI layout.
func (i *Imager) CopyToOci(imageRef, ociDir string) (string, error) {
	// TODO: add a progress bar and improve logging
	i.log.Infof("Pulling image %s from registry...", imageRef)
	i.log.Debugf("Copying image %s to OCI layout at %s...", imageRef, ociDir)
	srcRef, err := alltransports.ParseImageName("docker://" + imageRef)
	if err != nil {
		return "", fmt.Errorf("failed to parse source image name: %w", err)
	}

	destRef, err := alltransports.ParseImageName("oci:" + ociDir)
	if err != nil {
		return "", fmt.Errorf("failed to parse destination image name: %w", err)
	}

	policyCtx, err := i.PolicyContext()
	if err != nil {
		return "", err
	}
	defer policyCtx.Destroy()

	manifestBytes, err := copy.Image(context.Background(), policyCtx, destRef, srcRef, &copy.Options{
		RemoveSignatures: true,
	})
	if err != nil {
		return "", fmt.Errorf("failed to copy image: %w", err)
	}

	d := digest.FromBytes(manifestBytes)
	i.log.Infof("Successfully pulled image %s\n\n", imageRef)
	i.log.Debugf("Successfully copied image. Digest: %s", d.String())
	return d.String(), nil
}

// RemoteInfo retrieves the name, tag, and digest of a remote image.
func (i *Imager) RemoteInfo(imageRef string) (string, string, digest.Digest, error) {
	i.log.Debugf("Retrieving remote information for %s...", imageRef)
	srcRef, err := alltransports.ParseImageName("docker://" + imageRef)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to parse image name: %w", err)
	}

	policyCtx, err := i.PolicyContext()
	if err != nil {
		return "", "", "", err
	}
	defer policyCtx.Destroy()

	imgSrc, err := srcRef.NewImageSource(context.Background(), nil)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create image source: %w", err)
	}
	defer imgSrc.Close()

	manifestBytes, _, err := imgSrc.GetManifest(context.Background(), nil)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get manifest: %w", err)
	}

	d := digest.FromBytes(manifestBytes)

	dockerRef := srcRef.DockerReference()
	if dockerRef == nil {
		return "", "", "", fmt.Errorf("reference is not a Docker reference")
	}
	repoName := dockerRef.Name()

	// Use a type assertion to get the tag, which is more robust than string parsing.
	var tag string
	if tagged, ok := dockerRef.(reference.NamedTagged); ok {
		tag = tagged.Tag()
	}

	i.log.Debugf("Successfully retrieved remote information for %s", imageRef)
	return repoName, tag, d, nil
}
