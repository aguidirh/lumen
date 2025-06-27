package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/aguidirh/lumen/pkg/list"
)

// ListOperatorsOptions holds the flag values for the 'list operators' command.
type ListOperatorsOptions struct {
	Catalogs    bool
	OCPVersion  string
	CatalogRef  string
	PackageName string
	ChannelName string
}

func PrintResults(opts ListOperatorsOptions, results list.Results) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	defer w.Flush()

	if len(results.Catalogs) > 0 {
		fmt.Println("Available OpenShift OperatorHub catalogs:")
		fmt.Printf("OpenShift %s:\n", opts.OCPVersion)
		for _, cat := range results.Catalogs {
			fmt.Println(cat)
		}
	} else if len(results.Packages) > 0 {
		fmt.Fprintln(w, "NAME\tDEFAULT CHANNEL")
		for _, pkg := range results.Packages {
			fmt.Fprintf(w, "%s\t%s\n", pkg.Name, pkg.DefaultChannel)
		}
	} else if len(results.Channels) > 0 {
		fmt.Fprintln(w, "PACKAGE\tCHANNEL\tHEAD")
		for _, ch := range results.Channels {
			fmt.Fprintf(w, "%s\t%s\t%s\n", opts.PackageName, ch.Name, ch.Head)
		}
	} else if len(results.Versions) > 0 {
		fmt.Fprintln(w, "NAME")
		for _, ver := range results.Versions {
			fmt.Fprintln(w, ver.Name)
		}
	}
}
