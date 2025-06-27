package cli

import (
	"log"

	"github.com/aguidirh/lumen/pkg/list"
	"github.com/spf13/cobra"
)

var operatorsOptions ListOperatorsOptions

var operatorsCmd = &cobra.Command{
	Use:   "operators",
	Short: "List operators in a catalog",
	Long:  `Lists all the operators (packages) available in a specified OLM catalog image.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Cobra's 'MarkFlagRequired' is used for validation, so the manual Validate() is no longer needed.
		listOpts := list.ListOptions{
			Catalog:     operatorsOptions.CatalogRef,
			Catalogs:    operatorsOptions.Catalogs,
			OCPVersion:  operatorsOptions.OCPVersion,
			PackageName: operatorsOptions.PackageName,
			ChannelName: operatorsOptions.ChannelName,
		}

		listImpl := list.NewListImpl(listOpts)
		results, err := listImpl.List()
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		PrintResults(operatorsOptions, results)
	},
}

func init() {
	listCmd.AddCommand(operatorsCmd)

	operatorsCmd.Flags().BoolVar(&operatorsOptions.Catalogs, "catalogs", false, "List available catalogs for a specific OpenShift version")
	operatorsCmd.Flags().StringVar(&operatorsOptions.OCPVersion, "version", "", "The OpenShift version (e.g., 4.19) to find catalogs for")
	operatorsCmd.Flags().StringVar(&operatorsOptions.CatalogRef, "catalog", "", "The catalog image reference to introspect")
	operatorsCmd.Flags().StringVar(&operatorsOptions.PackageName, "package", "", "The name of the operator package to inspect")
	operatorsCmd.Flags().StringVar(&operatorsOptions.ChannelName, "channel", "", "The name of the channel to inspect")

	// If --catalogs is not used, --catalog is required.
	operatorsCmd.MarkFlagsMutuallyExclusive("catalogs", "catalog")
}
