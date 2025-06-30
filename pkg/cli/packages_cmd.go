package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func NewPackagesCmd(lister lister) *cobra.Command {
	var packagesOpts struct {
		Catalog string
	}

	listPackagesCmd := &cobra.Command{
		Use:   "packages",
		Short: "List packages (operators) in a catalog",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			packages, err := lister.PackagesByCatalog(packagesOpts.Catalog)
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			defer w.Flush()

			fmt.Fprintln(w, "NAME\tDEFAULT CHANNEL")
			for _, pkg := range packages {
				fmt.Fprintf(w, "%s\t%s\n", pkg.Name, pkg.DefaultChannel)
			}

			return nil
		},
	}

	listPackagesCmd.Flags().StringVar(&packagesOpts.Catalog, "catalog", "", "The catalog to list packages from")
	listPackagesCmd.MarkFlagRequired("catalog")
	return listPackagesCmd
}
