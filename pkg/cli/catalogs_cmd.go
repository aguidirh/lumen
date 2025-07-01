package cli

import (
	"github.com/spf13/cobra"
)

// NewCatalogsCmd creates a new catalogs command.
func NewCatalogsCmd(opts *LumenOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "catalogs",
		Short: "List available OpenShift OperatorHub catalogs for a specific version.",
		Long:  "List available OpenShift OperatorHub catalogs for a specific version.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ocpVersion, _ := cmd.Flags().GetString("ocp-version")

			catalogs, err := opts.lister.Catalogs(ocpVersion)
			if err != nil {
				return err
			}

			opts.printer.PrintCatalogs(ocpVersion, catalogs)
			return nil
		},
	}
	cmd.Flags().StringP("ocp-version", "v", "", "The OpenShift version to list catalogs for (e.g., 4.16)")
	cmd.MarkFlagRequired("ocp-version")
	return cmd
}
