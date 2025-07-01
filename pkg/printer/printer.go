package printer

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/aguidirh/lumen/pkg/list"
)

// Printer handles formatting and printing command output.
type Printer struct {
	w   *tabwriter.Writer
	out io.Writer
	log Logger
}

// NewPrinter creates a new Printer that writes to the given io.Writer.
func NewPrinter(out io.Writer, log Logger) *Printer {
	return &Printer{
		// initialize a tabwriter with 2-space padding
		w:   tabwriter.NewWriter(out, 0, 0, 2, ' ', 0),
		out: out,
		log: log,
	}
}

// PrintCatalogs formats and prints the list of available catalogs.
func (p *Printer) PrintCatalogs(ocpVersion string, catalogs []string) {
	p.log.Debugf("Printing %d catalogs for OCP version %s", len(catalogs), ocpVersion)
	fmt.Fprintf(p.out, "OpenShift %s Operator Catalogs:\n\n", ocpVersion)
	for _, catalog := range catalogs {
		fmt.Fprintln(p.out, catalog)
	}
}

// PrintPackages formats and prints the list of packages in a table.
func (p *Printer) PrintPackages(packages []list.Package) {
	p.log.Debugf("Printing %d packages", len(packages))
	fmt.Fprintln(p.w, "NAME\tDEFAULT CHANNEL")
	for _, pkg := range packages {
		fmt.Fprintf(p.w, "%s\t%s\n", pkg.Name, pkg.DefaultChannel)
	}
	p.w.Flush()
}

// PrintChannels formats and prints the list of channels for a package.
func (p *Printer) PrintChannels(channels []list.Channel) {
	p.log.Debugf("Printing %d channels", len(channels))
	fmt.Fprintln(p.w, "NAME\tHEAD")
	for _, ch := range channels {
		fmt.Fprintf(p.w, "%s\t%s\n", ch.Name, ch.Head)
	}
	p.w.Flush()
}

// PrintBundles formats and prints the list of bundle versions in a channel.
func (p *Printer) PrintBundles(pkgName, channelName string, bundles []list.ChannelEntry) {
	p.log.Debugf("Printing %d bundles for package %s, channel %s", len(bundles), pkgName, channelName)
	fmt.Fprintln(p.w, "BUNDLE_VERSION")
	for _, bundle := range bundles {
		fmt.Fprintf(p.w, "%s\n", bundle.Name)
	}
	p.w.Flush()
}
