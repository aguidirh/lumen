package cli

import (
	"github.com/aguidirh/lumen/pkg/list"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// lister defines the interface for all listing operations used by the CLI.
// The CLI package owns this interface, defining what it expects from a lister.
type lister interface {
	Catalogs(version string) ([]string, error)
	PackagesByCatalog(catalogRef string) ([]list.Package, error)
	ChannelsByPackage(catalogRef, pkgName string) ([]list.Channel, error)
	BundleVersionsByChannel(catalogRef, pkgName, channelName string) ([]list.ChannelEntry, error)
}

// NewLumenCmd creates the root "lumen" command and all its subcommands.
func NewLumenCmd(lister lister, logger *logrus.Logger) *cobra.Command {
	var logLevel string
	lumenCmd := &cobra.Command{
		Use:   "lumen",
		Short: "A tool for introspecting Operator catalogs.",
		Long:  `lumen is a command-line tool designed to help users and developers explore the contents of Operator catalogs (File-Based Catalogs).`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			lvl, err := logrus.ParseLevel(logLevel)
			if err != nil {
				return err
			}
			logger.SetLevel(lvl)
			return nil
		},
	}

	lumenCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Set the logging level (e.g., debug, info, warn, error)")
	lumenCmd.AddCommand(NewListCmd(lister))

	return lumenCmd
}
