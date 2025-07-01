//go:generate mockgen -source=interface.go -destination=mock/interface_generated.go -package=mock

package cli

import "github.com/aguidirh/lumen/internal/pkg/list"

// Lister defines the interface for all listing operations used by the CLI.
type Lister interface {
	Catalogs(version string) ([]string, error)
	PackagesByCatalog(catalogRef string) ([]list.Package, error)
	ChannelsByPackage(catalogRef, pkgName string) ([]list.Channel, error)
	BundleVersionsByChannel(catalogRef, pkgName, channelName string) ([]list.ChannelEntry, error)
}

// Printer defines the interface for printing operations used by the CLI.
type Printer interface {
	PrintCatalogs(ocpVersion string, catalogs []string)
	PrintPackages(packages []list.Package)
	PrintChannels(channels []list.Channel)
	PrintBundles(pkgName, channelName string, bundles []list.ChannelEntry)
}
