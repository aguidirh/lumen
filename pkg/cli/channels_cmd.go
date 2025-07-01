package cli

import (
	"github.com/spf13/cobra"
)

// NewChannelsCmd creates a new channels command.
func NewChannelsCmd(opts *LumenOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "channels",
		Short: "List all channels for a package in a catalog.",
		Long:  "List all available channels for a single operator package within a catalog.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			catalog, _ := cmd.Flags().GetString("catalog")
			pkg, _ := cmd.Flags().GetString("package")

			channels, err := opts.lister.ChannelsByPackage(catalog, pkg)
			if err != nil {
				return err
			}

			opts.printer.PrintChannels(channels)
			return nil
		},
	}

	cmd.Flags().StringP("catalog", "c", "", "The catalog image to list channels from")
	cmd.Flags().StringP("package", "p", "", "The package to list channels for")
	cmd.MarkFlagRequired("catalog")
	cmd.MarkFlagRequired("package")
	return cmd
}
