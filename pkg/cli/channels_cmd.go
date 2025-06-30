package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func NewChannelsCmd(lister lister) *cobra.Command {
	var channelsOpts struct {
		Catalog string
		Package string
	}

	listChannelsCmd := &cobra.Command{
		Use:   "channels",
		Short: "List channels in a package",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			channels, err := lister.ChannelsByPackage(channelsOpts.Catalog, channelsOpts.Package)
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			defer w.Flush()

			fmt.Fprintln(w, "NAME\tHEAD")
			for _, ch := range channels {
				fmt.Fprintf(w, "%s\t%s\n", ch.Name, ch.Head)
			}

			return nil
		},
	}

	listChannelsCmd.Flags().StringVar(&channelsOpts.Catalog, "catalog", "", "The catalog to list channels from")
	listChannelsCmd.Flags().StringVar(&channelsOpts.Package, "package", "", "The package to list channels for")
	listChannelsCmd.MarkFlagRequired("catalog")
	listChannelsCmd.MarkFlagRequired("package")
	return listChannelsCmd
}
