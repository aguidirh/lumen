package cli

import (
	"github.com/aguidirh/lumen/internal/pkg/log"
	"github.com/spf13/cobra"
)

// LumenOptions holds the options for the lumen command.
type LumenOptions struct {
	logLevel string
	lister   Lister
	printer  Printer
}

// NewLumenOptions creates a new LumenOptions instance.
func NewLumenOptions(lister Lister, printer Printer) *LumenOptions {
	return &LumenOptions{
		lister:  lister,
		printer: printer,
	}
}

// NewLumenCmd creates a new lumen command.
func NewLumenCmd(lister Lister, printer Printer) *cobra.Command {
	opts := &LumenOptions{
		lister:  lister,
		printer: printer,
	}

	cmd := &cobra.Command{
		Use:   "lumen",
		Short: "A tool to introspect File-Based Catalogs (FBC)",
		Long: `lumen is a command-line tool for introspecting the contents of OCI container images, 
with a special focus on operator-framework File-Based Catalogs (FBC). 
It allows you to pull catalog images, inspect and list their contents, without needing a running Kubernetes cluster.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			log.New(opts.logLevel)
			return nil
		},
	}

	cmd.AddCommand(NewListCmd(opts))
	cmd.PersistentFlags().StringVar(&opts.logLevel, "log-level", "info", "log level (debug, info, warn, error)")
	return cmd
}
