package cli

import (
	"github.com/spf13/cobra"
)

// NewListCmd creates a new list command.
func NewListCmd(opts *LumenOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List resources from an operator catalog.",
		Long:  "List resources from an operator catalog, such as catalogs, packages, channels, and bundles.",
	}

	cmd.AddCommand(NewCatalogsCmd(opts))
	cmd.AddCommand(NewPackagesCmd(opts))
	cmd.AddCommand(NewChannelsCmd(opts))
	cmd.AddCommand(NewBundlesCmd(opts))
	return cmd
}
