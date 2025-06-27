package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var lumenCmd = &cobra.Command{
	Use:   "lumen",
	Short: "lumen is a tool for introspecting operator catalogs",
	Long: `lumen is a command-line tool designed to help users and developers 
explore the contents of OLM (Operator Lifecycle Manager) catalogs.`,
}

func Execute() {
	if err := lumenCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
