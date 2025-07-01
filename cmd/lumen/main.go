package main

import (
	"os"

	"github.com/aguidirh/lumen/internal/pkg/catalog"
	"github.com/aguidirh/lumen/internal/pkg/cli"
	"github.com/aguidirh/lumen/internal/pkg/fsio"
	"github.com/aguidirh/lumen/internal/pkg/image"
	"github.com/aguidirh/lumen/internal/pkg/list"
	"github.com/aguidirh/lumen/internal/pkg/log"
	"github.com/aguidirh/lumen/internal/pkg/printer"
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
