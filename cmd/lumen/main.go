package main

import (
	"fmt"
	"os"

	"github.com/aguidirh/lumen/pkg/catalog"
	"github.com/aguidirh/lumen/pkg/cli"
	"github.com/aguidirh/lumen/pkg/fsio"
	"github.com/aguidirh/lumen/pkg/image"
	"github.com/aguidirh/lumen/pkg/list"
	"github.com/aguidirh/lumen/pkg/log"
)

func main() {
	logger := log.New("info")

	// Instantiate dependencies
	fsioSvc := fsio.NewFsIO()
	imageSvc := image.NewImager(logger)
	catalogSvc := catalog.NewCataloger(logger, imageSvc, fsioSvc)
	listSvc := list.NewCatalogLister(logger, catalogSvc, imageSvc)

	rootCmd := cli.NewLumenCmd(listSvc, logger)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
