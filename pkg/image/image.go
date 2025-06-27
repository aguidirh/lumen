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

// RemoteInfoFunc is a function variable that can be swapped for testing.
var RemoteInfoFunc = RemoteInfo

// PolicyContext returns a default policy context for container image operations.
func PolicyContext() (*signature.PolicyContext, error) {
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
func CopyToOci(imageRef, ociDir string) (string, error) {
	fmt.Printf("INFO: Copying image %s to OCI layout at %s...\n", imageRef, ociDir)
	srcRef, err := alltransports.ParseImageName("docker://" + imageRef)
	if err != nil {
		return "", fmt.Errorf("failed to parse source image name: %w", err)
	}

	destRef, err := alltransports.ParseImageName("oci:" + ociDir)
	if err != nil {
		return "", fmt.Errorf("failed to parse destination image name: %w", err)
	}

	policyCtx, err := PolicyContext()
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
	fmt.Printf("INFO: Successfully copied image. Digest: %s\n", d.String())
	return d.String(), nil
}

// RemoteInfo retrieves the name, tag, and digest of a remote image.
func RemoteInfo(imageRef string) (string, string, digest.Digest, error) {
	fmt.Printf("INFO: Retrieving remote information for %s...\n", imageRef)
	srcRef, err := alltransports.ParseImageName("docker://" + imageRef)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to parse image name: %w", err)
	}

	policyCtx, err := PolicyContext()
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

	fmt.Printf("INFO: Successfully retrieved remote information for %s\n", imageRef)
	return repoName, tag, d, nil
}
