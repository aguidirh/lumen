package list

import (
	"errors"
	"testing"

	"github.com/aguidirh/lumen/internal/pkg/list/mock"
	"github.com/aguidirh/lumen/internal/pkg/log"
	"github.com/opencontainers/go-digest"
	"github.com/operator-framework/operator-registry/alpha/declcfg"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCatalogs(t *testing.T) {
	// 1. Setup
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := log.New("error") // Use a quiet logger for tests

	testCases := []struct {
		name          string
		version       string
		setupMocks    func(m *mock.MockImager)
		expected      []string
		expectErr     bool
		expectedError string
	}{
		{
			name:    "Success Case - Catalogs Found",
			version: "4.16",
			setupMocks: func(m *mock.MockImager) {
				m.EXPECT().RemoteInfo(gomock.Any()).Return("name", "tag", digest.FromString("sha256:123"), nil).Times(4)
			},
			expected: []string{
				"registry.redhat.io/redhat/redhat-operator-index:v4.16",
				"registry.redhat.io/redhat/certified-operator-index:v4.16",
				"registry.redhat.io/redhat/community-operator-index:v4.16",
				"registry.redhat.io/redhat/redhat-marketplace-index:v4.16",
			},
			expectErr: false,
		},
		{
			name:    "Failure Case - No Catalogs Found",
			version: "4.16",
			setupMocks: func(m *mock.MockImager) {
				m.EXPECT().RemoteInfo(gomock.Any()).Return("", "", digest.Digest(""), errors.New("not found")).Times(4)
			},
			expected:      nil,
			expectErr:     true,
			expectedError: "no catalogs found for version 4.16",
		},
		{
			name:          "Failure Case - No Version Provided",
			version:       "",
			setupMocks:    func(m *mock.MockImager) {},
			expected:      nil,
			expectErr:     true,
			expectedError: "a version is required when listing catalogs",
		},
	}

	// 2. Execution and Assertion
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockImager := mock.NewMockImager(mockCtrl)
			// We don't need the cataloger for this specific function, so we can pass nil.
			lister := NewCatalogLister(logger, nil, mockImager)

			tc.setupMocks(mockImager)

			result, err := lister.Catalogs(tc.version)

			if tc.expectErr {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tc.expected, result)
			}
		})
	}
}

func TestPackagesByCatalog(t *testing.T) {
	// 1. Setup
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := log.New("error")

	testCases := []struct {
		name          string
		catalogRef    string
		setupMocks    func(m *mock.MockCataloger)
		expected      []Package
		expectErr     bool
		expectedError string
	}{
		{
			name:       "Success Case - Packages Found",
			catalogRef: "test-catalog:latest",
			setupMocks: func(m *mock.MockCataloger) {
				m.EXPECT().CatalogConfig("test-catalog:latest").Return(&declcfg.DeclarativeConfig{
					Packages: []declcfg.Package{
						{Name: "pkg1", DefaultChannel: "stable"},
						{Name: "pkg2", DefaultChannel: "beta"},
					},
				}, nil)
			},
			expected: []Package{
				{Name: "pkg1", DefaultChannel: "stable"},
				{Name: "pkg2", DefaultChannel: "beta"},
			},
			expectErr: false,
		},
		{
			name:       "Success Case - No Packages in Catalog",
			catalogRef: "test-catalog:latest",
			setupMocks: func(m *mock.MockCataloger) {
				m.EXPECT().CatalogConfig("test-catalog:latest").Return(&declcfg.DeclarativeConfig{}, nil)
			},
			expected:  []Package{},
			expectErr: false,
		},
		{
			name:       "Failure Case - CatalogConfig returns error",
			catalogRef: "test-catalog:latest",
			setupMocks: func(m *mock.MockCataloger) {
				m.EXPECT().CatalogConfig("test-catalog:latest").Return(nil, errors.New("some catalog error"))
			},
			expected:      nil,
			expectErr:     true,
			expectedError: "some catalog error",
		},
		{
			name:          "Failure Case - No CatalogRef Provided",
			catalogRef:    "",
			setupMocks:    func(m *mock.MockCataloger) {},
			expected:      nil,
			expectErr:     true,
			expectedError: "catalog reference is required",
		},
	}

	// 2. Execution and Assertion
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockCataloger := mock.NewMockCataloger(mockCtrl)
			// We don't need the imager for this specific function, so we can pass nil.
			lister := NewCatalogLister(logger, mockCataloger, nil)

			tc.setupMocks(mockCataloger)

			result, err := lister.PackagesByCatalog(tc.catalogRef)

			if tc.expectErr {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tc.expected, result)
			}
		})
	}
}

func TestChannelsByPackage(t *testing.T) {
	// 1. Setup
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := log.New("error")

	testCases := []struct {
		name          string
		catalogRef    string
		packageName   string
		setupMocks    func(m *mock.MockCataloger)
		expected      []Channel
		expectErr     bool
		expectedError string
	}{
		{
			name:        "Success Case - Channels Found",
			catalogRef:  "test-catalog:latest",
			packageName: "pkg1",
			setupMocks: func(m *mock.MockCataloger) {
				m.EXPECT().CatalogConfig("test-catalog:latest").Return(&declcfg.DeclarativeConfig{
					Packages: []declcfg.Package{{Name: "pkg1"}},
					Channels: []declcfg.Channel{
						{Name: "stable", Package: "pkg1"},
						{Name: "beta", Package: "pkg1"},
						{Name: "alpha", Package: "pkg2"}, // This should be ignored
					},
				}, nil)
			},
			expected: []Channel{
				{Name: "stable"},
				{Name: "beta"},
			},
			expectErr: false,
		},
		{
			name:        "Success Case - No Channels Found",
			catalogRef:  "test-catalog:latest",
			packageName: "pkg1",
			setupMocks: func(m *mock.MockCataloger) {
				m.EXPECT().CatalogConfig("test-catalog:latest").Return(&declcfg.DeclarativeConfig{
					Packages: []declcfg.Package{{Name: "pkg1"}},
				}, nil)
			},
			expected:  []Channel{},
			expectErr: false,
		},
		{
			name:        "Failure Case - Package Not Found",
			catalogRef:  "test-catalog:latest",
			packageName: "nonexistent",
			setupMocks: func(m *mock.MockCataloger) {
				m.EXPECT().CatalogConfig("test-catalog:latest").Return(&declcfg.DeclarativeConfig{
					Packages: []declcfg.Package{{Name: "pkg1"}},
				}, nil)
			},
			expected:      nil,
			expectErr:     true,
			expectedError: `package "nonexistent" not found in catalog "test-catalog:latest"`,
		},
		{
			name:        "Failure Case - CatalogConfig returns error",
			catalogRef:  "test-catalog:latest",
			packageName: "pkg1",
			setupMocks: func(m *mock.MockCataloger) {
				m.EXPECT().CatalogConfig("test-catalog:latest").Return(nil, errors.New("some catalog error"))
			},
			expected:      nil,
			expectErr:     true,
			expectedError: "some catalog error",
		},
		{
			name:          "Failure Case - Missing CatalogRef",
			catalogRef:    "",
			packageName:   "pkg1",
			setupMocks:    func(m *mock.MockCataloger) {},
			expected:      nil,
			expectErr:     true,
			expectedError: "catalog reference and package name are required",
		},
		{
			name:          "Failure Case - Missing PackageName",
			catalogRef:    "test-catalog:latest",
			packageName:   "",
			setupMocks:    func(m *mock.MockCataloger) {},
			expected:      nil,
			expectErr:     true,
			expectedError: "catalog reference and package name are required",
		},
	}

	// 2. Execution and Assertion
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockCataloger := mock.NewMockCataloger(mockCtrl)
			// We don't need the imager for this specific function, so we can pass nil.
			lister := NewCatalogLister(logger, mockCataloger, nil)

			tc.setupMocks(mockCataloger)

			result, err := lister.ChannelsByPackage(tc.catalogRef, tc.packageName)

			if tc.expectErr {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tc.expected, result)
			}
		})
	}
}

func TestBundleVersionsByChannel(t *testing.T) {
	// 1. Setup
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	logger := log.New("error")

	testCases := []struct {
		name          string
		catalogRef    string
		packageName   string
		channelName   string
		setupMocks    func(m *mock.MockCataloger)
		expected      []ChannelEntry
		expectErr     bool
		expectedError string
	}{
		{
			name:        "Success Case - Bundle Versions Found",
			catalogRef:  "test-catalog:latest",
			packageName: "pkg1",
			channelName: "stable",
			setupMocks: func(m *mock.MockCataloger) {
				m.EXPECT().CatalogConfig("test-catalog:latest").Return(&declcfg.DeclarativeConfig{
					Channels: []declcfg.Channel{
						{
							Name:    "stable",
							Package: "pkg1",
							Entries: []declcfg.ChannelEntry{
								{Name: "pkg1.v1.0.0"},
								{Name: "pkg1.v1.1.0"},
							},
						},
						{
							Name:    "beta",
							Package: "pkg1",
							Entries: []declcfg.ChannelEntry{
								{Name: "pkg1.v2.0.0-beta"},
							},
						},
					},
				}, nil)
			},
			expected: []ChannelEntry{
				{Name: "pkg1.v1.0.0"},
				{Name: "pkg1.v1.1.0"},
			},
			expectErr: false,
		},
		{
			name:        "Success Case - No Bundle Versions Found",
			catalogRef:  "test-catalog:latest",
			packageName: "pkg1",
			channelName: "stable",
			setupMocks: func(m *mock.MockCataloger) {
				m.EXPECT().CatalogConfig("test-catalog:latest").Return(&declcfg.DeclarativeConfig{
					Channels: []declcfg.Channel{
						{
							Name:    "stable",
							Package: "pkg1",
							Entries: []declcfg.ChannelEntry{},
						},
					},
				}, nil)
			},
			expected:  []ChannelEntry{},
			expectErr: false,
		},
		{
			name:        "Failure Case - Channel Not Found",
			catalogRef:  "test-catalog:latest",
			packageName: "pkg1",
			channelName: "nonexistent",
			setupMocks: func(m *mock.MockCataloger) {
				m.EXPECT().CatalogConfig("test-catalog:latest").Return(&declcfg.DeclarativeConfig{
					Channels: []declcfg.Channel{
						{
							Name:    "stable",
							Package: "pkg1",
						},
					},
				}, nil)
			},
			expected:      nil,
			expectErr:     true,
			expectedError: `channel "nonexistent" for package "pkg1" not found`,
		},
		{
			name:        "Failure Case - CatalogConfig returns error",
			catalogRef:  "test-catalog:latest",
			packageName: "pkg1",
			channelName: "stable",
			setupMocks: func(m *mock.MockCataloger) {
				m.EXPECT().CatalogConfig("test-catalog:latest").Return(nil, errors.New("some catalog error"))
			},
			expected:      nil,
			expectErr:     true,
			expectedError: "some catalog error",
		},
		{
			name:          "Failure Case - Missing CatalogRef",
			catalogRef:    "",
			packageName:   "pkg1",
			channelName:   "stable",
			setupMocks:    func(m *mock.MockCataloger) {},
			expected:      nil,
			expectErr:     true,
			expectedError: "catalog reference, package name, and channel name are required",
		},
		{
			name:          "Failure Case - Missing PackageName",
			catalogRef:    "test-catalog:latest",
			packageName:   "",
			channelName:   "stable",
			setupMocks:    func(m *mock.MockCataloger) {},
			expected:      nil,
			expectErr:     true,
			expectedError: "catalog reference, package name, and channel name are required",
		},
		{
			name:          "Failure Case - Missing ChannelName",
			catalogRef:    "test-catalog:latest",
			packageName:   "pkg1",
			channelName:   "",
			setupMocks:    func(m *mock.MockCataloger) {},
			expected:      nil,
			expectErr:     true,
			expectedError: "catalog reference, package name, and channel name are required",
		},
	}

	// 2. Execution and Assertion
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockCataloger := mock.NewMockCataloger(mockCtrl)
			// We don't need the imager for this specific function, so we can pass nil.
			lister := NewCatalogLister(logger, mockCataloger, nil)

			tc.setupMocks(mockCataloger)

			result, err := lister.BundleVersionsByChannel(tc.catalogRef, tc.packageName, tc.channelName)

			if tc.expectErr {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tc.expected, result)
			}
		})
	}
}
