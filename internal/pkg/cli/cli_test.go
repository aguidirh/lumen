package cli_test

import (
	"bytes"
	"testing"

	"github.com/aguidirh/lumen/internal/pkg/cli"
	cliMock "github.com/aguidirh/lumen/internal/pkg/cli/mock"
	"github.com/aguidirh/lumen/internal/pkg/list"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewLumenCmd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLister := cliMock.NewMockLister(ctrl)
	mockPrinter := cliMock.NewMockPrinter(ctrl)

	cmd := cli.NewLumenCmd(mockLister, mockPrinter)
	assert.NotNil(t, cmd)
	assert.Equal(t, "lumen", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)

	// Check that subcommands are added
	subcommands := cmd.Commands()
	assert.NotEmpty(t, subcommands)

	// Find the list command
	var listCmd *cobra.Command
	for _, subcmd := range subcommands {
		if subcmd.Use == "list" {
			listCmd = subcmd
			break
		}
	}
	assert.NotNil(t, listCmd, "list command should be present")
}

func TestNewCatalogsCmd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLister := cliMock.NewMockLister(ctrl)
	mockPrinter := cliMock.NewMockPrinter(ctrl)

	// Test successful execution
	catalogs := []string{"catalog1", "catalog2"}
	version := "4.15"

	mockLister.EXPECT().Catalogs(version).Return(catalogs, nil)
	mockPrinter.EXPECT().PrintCatalogs(version, catalogs)

	opts := cli.NewLumenOptions(mockLister, mockPrinter)
	cmd := cli.NewCatalogsCmd(opts)
	cmd.SetArgs([]string{"--ocp-version", version})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestNewCatalogsCmd_MissingVersion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLister := cliMock.NewMockLister(ctrl)
	mockPrinter := cliMock.NewMockPrinter(ctrl)

	opts := cli.NewLumenOptions(mockLister, mockPrinter)
	cmd := cli.NewCatalogsCmd(opts)
	cmd.SetArgs([]string{})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required flag(s)")
}

func TestNewPackagesCmd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLister := cliMock.NewMockLister(ctrl)
	mockPrinter := cliMock.NewMockPrinter(ctrl)

	// Test successful execution
	packages := []list.Package{
		{Name: "package1", DefaultChannel: "stable"},
		{Name: "package2", DefaultChannel: "beta"},
	}
	catalogRef := "registry.redhat.io/redhat/redhat-operator-index:v4.15"

	mockLister.EXPECT().PackagesByCatalog(catalogRef).Return(packages, nil)
	mockPrinter.EXPECT().PrintPackages(packages)

	opts := cli.NewLumenOptions(mockLister, mockPrinter)
	cmd := cli.NewPackagesCmd(opts)
	cmd.SetArgs([]string{"--catalog", catalogRef})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestNewPackagesCmd_MissingCatalog(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLister := cliMock.NewMockLister(ctrl)
	mockPrinter := cliMock.NewMockPrinter(ctrl)

	opts := cli.NewLumenOptions(mockLister, mockPrinter)
	cmd := cli.NewPackagesCmd(opts)
	cmd.SetArgs([]string{})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required flag(s)")
}

func TestNewChannelsCmd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLister := cliMock.NewMockLister(ctrl)
	mockPrinter := cliMock.NewMockPrinter(ctrl)

	// Test successful execution
	channels := []list.Channel{
		{Name: "stable", Head: "package.v1.0.0"},
		{Name: "beta", Head: "package.v1.1.0"},
	}
	catalogRef := "registry.redhat.io/redhat/redhat-operator-index:v4.15"
	packageName := "test-package"

	mockLister.EXPECT().ChannelsByPackage(catalogRef, packageName).Return(channels, nil)
	mockPrinter.EXPECT().PrintChannels(channels)

	opts := cli.NewLumenOptions(mockLister, mockPrinter)
	cmd := cli.NewChannelsCmd(opts)
	cmd.SetArgs([]string{"--catalog", catalogRef, "--package", packageName})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestNewChannelsCmd_MissingFlags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLister := cliMock.NewMockLister(ctrl)
	mockPrinter := cliMock.NewMockPrinter(ctrl)

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "missing catalog",
			args: []string{"--package", "test-package"},
		},
		{
			name: "missing package",
			args: []string{"--catalog", "registry.redhat.io/redhat/redhat-operator-index:v4.15"},
		},
		{
			name: "missing both",
			args: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := cli.NewLumenOptions(mockLister, mockPrinter)
			cmd := cli.NewChannelsCmd(opts)
			cmd.SetArgs(tt.args)

			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			err := cmd.Execute()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "required flag(s)")
		})
	}
}

func TestNewBundlesCmd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLister := cliMock.NewMockLister(ctrl)
	mockPrinter := cliMock.NewMockPrinter(ctrl)

	// Test successful execution
	bundles := []list.ChannelEntry{
		{Name: "package.v1.0.0"},
		{Name: "package.v1.1.0"},
	}
	catalogRef := "registry.redhat.io/redhat/redhat-operator-index:v4.15"
	packageName := "test-package"
	channelName := "stable"

	mockLister.EXPECT().BundleVersionsByChannel(catalogRef, packageName, channelName).Return(bundles, nil)
	mockPrinter.EXPECT().PrintBundles(packageName, channelName, bundles)

	opts := cli.NewLumenOptions(mockLister, mockPrinter)
	cmd := cli.NewBundlesCmd(opts)
	cmd.SetArgs([]string{"--catalog", catalogRef, "--package", packageName, "--channel", channelName})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestNewBundlesCmd_MissingFlags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLister := cliMock.NewMockLister(ctrl)
	mockPrinter := cliMock.NewMockPrinter(ctrl)

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "missing catalog",
			args: []string{"--package", "test-package", "--channel", "stable"},
		},
		{
			name: "missing package",
			args: []string{"--catalog", "registry.redhat.io/redhat/redhat-operator-index:v4.15", "--channel", "stable"},
		},
		{
			name: "missing channel",
			args: []string{"--catalog", "registry.redhat.io/redhat/redhat-operator-index:v4.15", "--package", "test-package"},
		},
		{
			name: "missing all",
			args: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := cli.NewLumenOptions(mockLister, mockPrinter)
			cmd := cli.NewBundlesCmd(opts)
			cmd.SetArgs(tt.args)

			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			err := cmd.Execute()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "required flag(s)")
		})
	}
}

func TestCLICommandStructure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	lister := cliMock.NewMockLister(ctrl)
	printer := cliMock.NewMockPrinter(ctrl)

	cmd := cli.NewLumenCmd(lister, printer)
	listCmd, _, err := cmd.Find([]string{"list"})
	assert.NoError(t, err)

	assert.Equal(t, "list", listCmd.Use)
	assert.True(t, listCmd.HasSubCommands())

	// Test catalogs command flags
	catalogsCmd, _, err := cmd.Find([]string{"list", "catalogs"})
	assert.NoError(t, err)

	ocpVersionFlag := catalogsCmd.Flags().Lookup("ocp-version")
	assert.NotNil(t, ocpVersionFlag)
	assert.Equal(t, "string", ocpVersionFlag.Value.Type())

	// Test packages command flags
	packagesCmd, _, err := cmd.Find([]string{"list", "packages"})
	assert.NoError(t, err)

	catalogFlag := packagesCmd.Flags().Lookup("catalog")
	assert.NotNil(t, catalogFlag)
	assert.Equal(t, "string", catalogFlag.Value.Type())

	// Test channels command flags
	channelsCmd, _, err := cmd.Find([]string{"list", "channels"})
	assert.NoError(t, err)

	catalogFlag = channelsCmd.Flags().Lookup("catalog")
	assert.NotNil(t, catalogFlag)
	packageFlag := channelsCmd.Flags().Lookup("package")
	assert.NotNil(t, packageFlag)

	// Test bundles command flags
	bundlesCmd, _, err := cmd.Find([]string{"list", "bundles"})
	assert.NoError(t, err)

	catalogFlag = bundlesCmd.Flags().Lookup("catalog")
	assert.NotNil(t, catalogFlag)
	packageFlag = bundlesCmd.Flags().Lookup("package")
	assert.NotNil(t, packageFlag)
	channelFlag := bundlesCmd.Flags().Lookup("channel")
	assert.NotNil(t, channelFlag)
}
