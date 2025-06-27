package mcp_integration

import (
	"encoding/json"
	"fmt"

	"github.com/aguidirh/lumen/pkg/list"
)

// LumenToolRunner is the function that would be registered with an MCP server or agent tooling platform.
// It acts as a bridge between the agent's tool call and our Go library.
func LumenToolRunner(catalogRef, ocpVersion, packageName, channelName string, listCatalogs bool) (string, error) {

	// 1. Create the options struct from the parameters provided by the agent.
	opts := list.ListOptions{
		Catalog:     catalogRef,
		Catalogs:    listCatalogs,
		OCPVersion:  ocpVersion,
		PackageName: packageName,
		ChannelName: channelName,
	}

	// 2. Call the library's main function.
	listImpl := list.NewListImpl(opts)
	results, err := listImpl.List()
	if err != nil {
		return "", fmt.Errorf("lumen tool failed: %w", err)
	}

	// 3. Serialize the concrete Results struct into JSON for the agent to understand.
	jsonResult, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to serialize lumen results: %w", err)
	}

	return string(jsonResult), nil
}
