package catalog_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aguidirh/lumen/pkg/catalog"
	catalogMock "github.com/aguidirh/lumen/pkg/catalog/mock"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewCataloger(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := catalogMock.NewMockLogger(ctrl)
	imager := catalogMock.NewMockImager(ctrl)
	fsio := catalogMock.NewMockFsIO(ctrl)

	cataloger := catalog.NewCataloger(logger, imager, fsio)
	assert.NotNil(t, cataloger)
}

func TestCataloger_CatalogConfig_RemoteInfoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := catalogMock.NewMockLogger(ctrl)
	imager := catalogMock.NewMockImager(ctrl)
	fsio := catalogMock.NewMockFsIO(ctrl)

	imageRef := "registry.redhat.io/redhat/redhat-operator-index:v4.15"
	expectedError := fmt.Errorf("remote info failed")

	imager.EXPECT().RemoteInfo(imageRef).Return("", "", digest.Digest(""), expectedError)

	cataloger := catalog.NewCataloger(logger, imager, fsio)
	config, err := cataloger.CatalogConfig(imageRef)

	assert.Nil(t, config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get remote info")
	assert.Contains(t, err.Error(), expectedError.Error())
}

func TestCataloger_CatalogConfig_CacheHit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := catalogMock.NewMockLogger(ctrl)
	imager := catalogMock.NewMockImager(ctrl)
	fsio := catalogMock.NewMockFsIO(ctrl)

	// Create a temporary directory structure that simulates a cache hit
	tempDir := t.TempDir()

	// Create the cache directory structure
	name := "redhat-operator-index"
	tag := "v4.15"
	testDigest := digest.FromString("test-content")
	safeDigest := strings.Replace(testDigest.String(), ":", "-", 1)

	configsCachePath := filepath.Join(tempDir, "working-dir", "operator-catalogs", name, tag, safeDigest, "configs")
	err := os.MkdirAll(configsCachePath, 0755)
	require.NoError(t, err)

	// Create a simple FBC structure
	err = os.WriteFile(filepath.Join(configsCachePath, "catalog.yaml"), []byte(`
schema: olm.package
name: test-package
`), 0644)
	require.NoError(t, err)

	// Change to temp directory so the working-dir path resolves correctly
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	imageRef := "registry.redhat.io/redhat/redhat-operator-index:v4.15"

	imager.EXPECT().RemoteInfo(imageRef).Return(name, tag, testDigest, nil)
	logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Debug(gomock.Any()).AnyTimes()

	cataloger := catalog.NewCataloger(logger, imager, fsio)
	config, err := cataloger.CatalogConfig(imageRef)

	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.NotEmpty(t, config.Packages)
}

func TestCataloger_CatalogConfig_CacheMiss_CopyToOciError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := catalogMock.NewMockLogger(ctrl)
	imager := catalogMock.NewMockImager(ctrl)
	fsio := catalogMock.NewMockFsIO(ctrl)

	// Use a non-existent cache path to simulate cache miss
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	imageRef := "registry.redhat.io/redhat/redhat-operator-index:v4.15"
	name := "redhat-operator-index"
	tag := "v4.15"
	testDigest := digest.FromString("test-content")
	expectedError := fmt.Errorf("copy to oci failed")

	imager.EXPECT().RemoteInfo(imageRef).Return(name, tag, testDigest, nil)
	imager.EXPECT().CopyToOci(fmt.Sprintf("%s@%s", name, testDigest), gomock.Any()).Return("", expectedError)
	logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Debug(gomock.Any()).AnyTimes()

	cataloger := catalog.NewCataloger(logger, imager, fsio)
	config, err := cataloger.CatalogConfig(imageRef)

	assert.Nil(t, config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to copy image to oci")
	assert.Contains(t, err.Error(), expectedError.Error())
}

func TestCataloger_CatalogConfig_CacheMiss_ExtractCatalogError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := catalogMock.NewMockLogger(ctrl)
	imager := catalogMock.NewMockImager(ctrl)
	fsio := catalogMock.NewMockFsIO(ctrl)

	// Use a non-existent cache path to simulate cache miss
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	imageRef := "registry.redhat.io/redhat/redhat-operator-index:v4.15"
	name := "redhat-operator-index"
	tag := "v4.15"
	testDigest := digest.FromString("test-content")

	imager.EXPECT().RemoteInfo(imageRef).Return(name, tag, testDigest, nil)
	imager.EXPECT().CopyToOci(fmt.Sprintf("%s@%s", name, testDigest), gomock.Any()).Return("invalid-oci-path", nil)
	logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Debug(gomock.Any()).AnyTimes()

	cataloger := catalog.NewCataloger(logger, imager, fsio)
	config, err := cataloger.CatalogConfig(imageRef)

	assert.Nil(t, config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to find and extract catalog")
}

func TestCataloger_CatalogConfig_InvalidDeclarativeConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	logger := catalogMock.NewMockLogger(ctrl)
	imager := catalogMock.NewMockImager(ctrl)
	fsio := catalogMock.NewMockFsIO(ctrl)

	// Create a temporary directory with invalid FBC content
	tempDir := t.TempDir()
	name := "redhat-operator-index"
	tag := "v4.15"
	testDigest := digest.FromString("test-content")
	safeDigest := strings.Replace(testDigest.String(), ":", "-", 1)

	configsCachePath := filepath.Join(tempDir, "working-dir", "operator-catalogs", name, tag, safeDigest, "configs")
	err := os.MkdirAll(configsCachePath, 0755)
	require.NoError(t, err)

	// Create invalid YAML content
	err = os.WriteFile(filepath.Join(configsCachePath, "catalog.yaml"), []byte(`
invalid: yaml: content: [
`), 0644)
	require.NoError(t, err)

	// Change to temp directory so the working-dir path resolves correctly
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)
	err = os.Chdir(tempDir)
	require.NoError(t, err)

	imageRef := "registry.redhat.io/redhat/redhat-operator-index:v4.15"

	imager.EXPECT().RemoteInfo(imageRef).Return(name, tag, testDigest, nil)
	logger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	logger.EXPECT().Debug(gomock.Any()).AnyTimes()

	cataloger := catalog.NewCataloger(logger, imager, fsio)
	config, err := cataloger.CatalogConfig(imageRef)

	assert.Nil(t, config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load declarative config")
}
