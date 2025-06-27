package cli

import "github.com/spf13/cobra"

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List various resources from operator catalogs",
	Long:  `The list command provides subcommands to list different kinds of resources, such as operators, channels, and versions.`,
}

func init() {
	lumenCmd.AddCommand(listCmd)
}
