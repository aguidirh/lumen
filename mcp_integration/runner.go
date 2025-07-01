package mcp_integration

import (
	"encoding/json"
	"fmt"

	"github.com/aguidirh/lumen/pkg/catalog"
	"github.com/aguidirh/lumen/pkg/fsio"
	"github.com/aguidirh/lumen/pkg/image"
	"github.com/aguidirh/lumen/pkg/list"
	"github.com/aguidirh/lumen/pkg/log"
)

// LumenToolRunner is the function that would be registered with an MCP server or agent tooling platform.
// It acts as a bridge between the agent's tool call and our Go library.
func LumenToolRunner(catalogRef, ocpVersion, packageName, channelName string, listCatalogs bool) (string, error) {
	var (
		result any
		err    error
	)

	// In a real application, the log level might be configurable.
	logger := log.New("error")
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

	// 3. Serialize the concrete result into JSON for the agent to understand.
	jsonResult, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to serialize lumen results: %w", err)
	}

	return string(jsonResult), nil
}
