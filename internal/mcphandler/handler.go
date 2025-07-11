package mcphandler

import (
	"encoding/json"
	"fmt"

	"github.com/aguidirh/lumen/internal/pkg/catalog"
	"github.com/aguidirh/lumen/internal/pkg/fsio"
	"github.com/aguidirh/lumen/internal/pkg/image"
	"github.com/aguidirh/lumen/internal/pkg/list"
	"github.com/aguidirh/lumen/internal/pkg/log"
)

// LumenToolHandler is the function that would be registered with an MCP server or agent tooling platform.
// It acts as a handler between the agent's tool call and our Go library.
func LumenToolHandler(catalogRef, ocpVersion, packageName, channelName string, listCatalogs bool) (string, error) {
	var (
		result any
		err    error
	)

	logger := log.New("panic")
	fs := fsio.NewFsIO()
	imager := image.NewImager(logger)
	cataloger := catalog.NewCataloger(logger, imager, fs)
	lister := list.NewCatalogLister(logger, cataloger, imager)

	switch {
	case listCatalogs:
		result, err = lister.Catalogs(ocpVersion)
	case packageName != "":
		if channelName != "" {
			result, err = lister.BundleVersionsByChannel(catalogRef, packageName, channelName)
		} else {
			result, err = lister.ChannelsByPackage(catalogRef, packageName)
		}
	case catalogRef != "":
		result, err = lister.PackagesByCatalog(catalogRef)
	default:
		return "", fmt.Errorf("invalid set of options provided to lumen tool")
	}

	if err != nil {
		return "", fmt.Errorf("lumen tool failed: %w", err)
	}

	// Serialize the concrete result into JSON for the agent to understand.
	jsonResult, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to serialize lumen results: %w", err)
	}

	return string(jsonResult), nil
}
