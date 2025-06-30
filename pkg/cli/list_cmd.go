package cli

import "github.com/spf13/cobra"

func NewListCmd(lister lister) *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List various resources from operator catalogs",
		Long:  `The list command provides subcommands to list different kinds of resources, such as catalogs, packages, channels, and bundles.`,
	}

	listCmd.AddCommand(NewCatalogsCmd(lister))
	listCmd.AddCommand(NewPackagesCmd(lister))
	listCmd.AddCommand(NewChannelsCmd(lister))
	listCmd.AddCommand(NewBundlesCmd(lister))

	return listCmd
}
