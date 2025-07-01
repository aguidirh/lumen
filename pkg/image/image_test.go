// Package image_test contains tests for the image package.
package image_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aguidirh/lumen/pkg/image"
	mock_image "github.com/aguidirh/lumen/pkg/image/mock"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewImager(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mock_image.NewMockLogger(ctrl)
	imager := image.NewImager(mockLogger)
	assert.NotNil(t, imager)
}

func TestImager_PolicyContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mock_image.NewMockLogger(ctrl)
	imager := image.NewImager(mockLogger)

	policyContext, err := imager.PolicyContext()
	assert.NoError(t, err)
	assert.NotNil(t, policyContext)
}

func TestImager_CopyToOci(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping test in CI environment due to network dependency")
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mock_image.NewMockLogger(ctrl)
	imager := image.NewImager(mockLogger)

	tempDir := t.TempDir()
	ociDir := filepath.Join(tempDir, "oci")

	mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

	d, err := imager.CopyToOci("hello-world:latest", ociDir)
	require.NoError(t, err)
	assert.NotEmpty(t, d)

	// Verify that the OCI directory is not empty and contains the expected files
	_, err = os.Stat(filepath.Join(ociDir, "oci-layout"))
	assert.NoError(t, err, "oci-layout file should exist")

	_, err = os.Stat(filepath.Join(ociDir, "index.json"))
	assert.NoError(t, err, "index.json file should exist")
}

func TestImager_RemoteInfo(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping test in CI environment due to network dependency")
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mock_image.NewMockLogger(ctrl)
	imager := image.NewImager(mockLogger)

	mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()

	repoName, tag, d, err := imager.RemoteInfo("hello-world:latest")
	require.NoError(t, err)

	assert.Equal(t, "docker.io/library/hello-world", repoName)
	assert.Equal(t, "latest", tag)

	// The digest can change, so we just verify that it's a valid digest.
	_, err = digest.Parse(d.String())
	assert.NoError(t, err, "should be a valid digest")
}
