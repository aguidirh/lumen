package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewCatalogsCmd(lister lister) *cobra.Command {
	var catalogsOpts struct {
		OCPVersion string
	}

	listCatalogsCmd := &cobra.Command{
		Use:   "catalogs",
		Short: "List available operator catalogs for a specific OpenShift version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			catalogs, err := lister.Catalogs(catalogsOpts.OCPVersion)
			if err != nil {
				return err
			}

			// This should be handled by a proper printer/view
			fmt.Println("Available OpenShift OperatorHub catalogs:")
			fmt.Printf("OpenShift %s:\n", catalogsOpts.OCPVersion)
			for _, cat := range catalogs {
				fmt.Println(cat)
			}

			return nil
		},
	}
	listCatalogsCmd.Flags().StringVar(&catalogsOpts.OCPVersion, "ocp-version", "", "The OpenShift version to list catalogs for (e.g., 4.16)")
	listCatalogsCmd.MarkFlagRequired("ocp-version")
	return listCatalogsCmd
}
