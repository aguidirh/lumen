package cli

import (
	"github.com/spf13/cobra"
)

// NewBundlesCmd creates a new bundles command.
func NewBundlesCmd(opts *LumenOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bundles",
		Short: "List all bundle versions in a channel.",
		Long:  "List all available bundle versions for a specific channel of an operator.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			catalog, _ := cmd.Flags().GetString("catalog")
			pkg, _ := cmd.Flags().GetString("package")
			channel, _ := cmd.Flags().GetString("channel")

			bundles, err := opts.lister.BundleVersionsByChannel(catalog, pkg, channel)
			if err != nil {
				return err
			}

			opts.printer.PrintBundles(pkg, channel, bundles)
			return nil
		},
	}

	cmd.Flags().StringP("catalog", "c", "", "The catalog image to list bundles from")
	cmd.Flags().StringP("package", "p", "", "The package to list bundles for")
	cmd.Flags().StringP("channel", "C", "", "The channel to list bundles for")
	cmd.MarkFlagRequired("catalog")
	cmd.MarkFlagRequired("package")
	cmd.MarkFlagRequired("channel")
	return cmd
}
