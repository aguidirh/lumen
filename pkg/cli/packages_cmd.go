package cli

import (
	"github.com/spf13/cobra"
)

// NewPackagesCmd creates a new packages command.
func NewPackagesCmd(opts *LumenOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "packages",
		Short: "List all packages in a catalog.",
		Long:  "List all packages (operators) in a given catalog image.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			catalog, _ := cmd.Flags().GetString("catalog")

			packages, err := opts.lister.PackagesByCatalog(catalog)
			if err != nil {
				return err
			}

			opts.printer.PrintPackages(packages)
			return nil
		},
	}
	cmd.Flags().StringP("catalog", "c", "", "The catalog image to list packages from")
	cmd.MarkFlagRequired("catalog")
	return cmd
}
