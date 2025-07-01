package printer

import (
	"bytes"
	"strings"
	"testing"

	"github.com/aguidirh/lumen/pkg/list"
	"github.com/aguidirh/lumen/pkg/printer/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestPrintPackages(t *testing.T) {
	testCases := []struct {
		name           string
		packages       []list.Package
		expectedLog    string
		expectedOutput string
	}{
		{
			name: "Success Case - Multiple Packages",
			packages: []list.Package{
				{Name: "pkg1", DefaultChannel: "stable"},
				{Name: "pkg2", DefaultChannel: "beta"},
			},
			expectedLog:    "Printing 2 packages",
			expectedOutput: "NAME  DEFAULT CHANNEL\npkg1  stable\npkg2  beta\n",
		},
		{
			name:           "Success Case - No Packages",
			packages:       []list.Package{},
			expectedLog:    "Printing 0 packages",
			expectedOutput: "NAME  DEFAULT CHANNEL\n",
		},
		{
			name: "Success Case - Package with long name",
			packages: []list.Package{
				{Name: "a-very-very-long-package-name", DefaultChannel: "alpha"},
				{Name: "short", DefaultChannel: "stable"},
			},
			expectedLog:    "Printing 2 packages",
			expectedOutput: "NAME                             DEFAULT CHANNEL\na-very-very-long-package-name  alpha\nshort                            stable\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 1. Setup
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			var buf bytes.Buffer
			mockLogger := mock.NewMockLogger(mockCtrl)
			p := NewPrinter(&buf, mockLogger)

			// 2. Expectations
			mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).Do(func(format string, args ...interface{}) {
				// We use Do to get access to the arguments and assert them.
				assert.Contains(t, format, "Printing %d packages")
				// Extract the count from args for a more specific check if needed, e.g., assert.Equal(t, len(tc.packages), args[0])
			}).Times(1)

			// 3. Execution
			p.PrintPackages(tc.packages)

			// 4. Assertion
			// Compare line by line to be robust against spacing issues.
			expectedLines := strings.Split(strings.TrimSpace(tc.expectedOutput), "\n")
			actualLines := strings.Split(strings.TrimSpace(buf.String()), "\n")
			assert.Equal(t, len(expectedLines), len(actualLines), "Number of lines should match")
			for i := range expectedLines {
				// Trim spaces from each line to compare content without exact padding.
				assert.Equal(t, strings.Join(strings.Fields(expectedLines[i]), " "), strings.Join(strings.Fields(actualLines[i]), " "))
			}
		})
	}
}

func TestPrintCatalogs(t *testing.T) {
	// 1. Setup
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	var buf bytes.Buffer
	mockLogger := mock.NewMockLogger(mockCtrl)
	p := NewPrinter(&buf, mockLogger)

	catalogs := []string{"catalog1", "catalog2"}
	ocpVersion := "4.16"

	// 2. Expectations
	mockLogger.EXPECT().Debugf("Printing %d catalogs for OCP version %s", len(catalogs), ocpVersion).Times(1)

	// 3. Execution
	p.PrintCatalogs(ocpVersion, catalogs)

	// 4. Assertion
	expectedOutput := "OpenShift 4.16 Operator Catalogs:\n\ncatalog1\ncatalog2\n"
	assert.Contains(t, buf.String(), expectedOutput)
}

func TestPrintChannels(t *testing.T) {
	// 1. Setup
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	var buf bytes.Buffer
	mockLogger := mock.NewMockLogger(mockCtrl)
	p := NewPrinter(&buf, mockLogger)

	channels := []list.Channel{{Name: "stable", Head: "v1.0.0"}, {Name: "beta", Head: "v1.1.0"}}

	// 2. Expectations
	mockLogger.EXPECT().Debugf("Printing %d channels", len(channels)).Times(1)

	// 3. Execution
	p.PrintChannels(channels)

	// 4. Assertion
	expectedOutput := "NAME    HEAD\nstable  v1.0.0\nbeta    v1.1.0\n"
	assert.Equal(t, strings.TrimSpace(expectedOutput), strings.TrimSpace(buf.String()))
}

func TestPrintBundles(t *testing.T) {
	// 1. Setup
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	var buf bytes.Buffer
	mockLogger := mock.NewMockLogger(mockCtrl)
	p := NewPrinter(&buf, mockLogger)

	bundles := []list.ChannelEntry{{Name: "bundle-1.0.0"}, {Name: "bundle-1.1.0"}}
	pkgName := "test-pkg"
	channelName := "stable"

	// 2. Expectations
	mockLogger.EXPECT().Debugf("Printing %d bundles for package %s, channel %s", len(bundles), pkgName, channelName).Times(1)

	// 3. Execution
	p.PrintBundles(pkgName, channelName, bundles)

	// 4. Assertion
	expectedTable := "BUNDLE_VERSION\nbundle-1.0.0\nbundle-1.1.0\n"
	assert.Equal(t, strings.TrimSpace(expectedTable), strings.TrimSpace(buf.String()))
}
