package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewBundlesCmd(lister lister) *cobra.Command {
	var bundlesOpts struct {
		Catalog string
		Package string
		Channel string
	}

	listBundlesCmd := &cobra.Command{
		Use:   "bundles",
		Short: "List bundle versions in a channel",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			bundles, err := lister.BundleVersionsByChannel(bundlesOpts.Catalog, bundlesOpts.Package, bundlesOpts.Channel)
			if err != nil {
				return err
			}

			fmt.Printf("Bundle versions for package %q, channel %q:\n", bundlesOpts.Package, bundlesOpts.Channel)
			for _, b := range bundles {
				fmt.Println(b.Name)
			}

			return nil
		},
	}

	listBundlesCmd.Flags().StringVar(&bundlesOpts.Catalog, "catalog", "", "The catalog to list bundles from")
	listBundlesCmd.Flags().StringVar(&bundlesOpts.Package, "package", "", "The package to list bundles for")
	listBundlesCmd.Flags().StringVar(&bundlesOpts.Channel, "channel", "", "The channel to list bundles for")
	listBundlesCmd.MarkFlagRequired("catalog")
	listBundlesCmd.MarkFlagRequired("package")
	listBundlesCmd.MarkFlagRequired("channel")
	return listBundlesCmd
}
