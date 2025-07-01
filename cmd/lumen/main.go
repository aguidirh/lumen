package main

import (
	"os"

	"github.com/aguidirh/lumen/pkg/catalog"
	"github.com/aguidirh/lumen/pkg/cli"
	"github.com/aguidirh/lumen/pkg/fsio"
	"github.com/aguidirh/lumen/pkg/image"
	"github.com/aguidirh/lumen/pkg/list"
	"github.com/aguidirh/lumen/pkg/log"
	"github.com/aguidirh/lumen/pkg/printer"
)

func main() {
	logger := log.New("info")
	fs := fsio.NewFsIO()
	imager := image.NewImager(logger)
	cataloger := catalog.NewCataloger(logger, imager, fs)
	lister := list.NewCatalogLister(logger, cataloger, imager)
	printer := printer.NewPrinter(os.Stdout, logger)

	if err := cli.NewLumenCmd(lister, printer).Execute(); err != nil {
		logger.Fatal(err)
	}
}
