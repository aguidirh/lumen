package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aguidirh/lumen/internal/mcphandler"
)

// MCPRequest represents an incoming MCP tool call request
type MCPRequest struct {
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
}

// MCPResponse represents an MCP response
type MCPResponse struct {
	Result interface{} `json:"result,omitempty"`
	Error  *MCPError   `json:"error,omitempty"`
}

// MCPError represents an MCP error
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func main() {
	// Read from stdin (MCP protocol uses stdio)
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for {
		var request MCPRequest
		if err := decoder.Decode(&request); err != nil {
			log.Printf("Error decoding request: %v", err)
			continue
		}

		response := handleRequest(request)
		if err := encoder.Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
	}
}

func handleRequest(request MCPRequest) MCPResponse {
	switch request.Method {
	case "tools/list":
		return MCPResponse{
			Result: map[string]interface{}{
				"tools": []map[string]interface{}{
					{
						"name":        "lumen_list",
						"description": "Introspects an operator-framework catalog image to list its contents. Can list all packages (operators), all channels for a given package, or all bundle versions for a given channel.",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"catalogRef": map[string]interface{}{
									"type":        "string",
									"description": "The full image reference of the catalog to inspect (e.g., 'registry.redhat.io/redhat/community-operator-index:v4.16').",
								},
								"ocpVersion": map[string]interface{}{
									"type":        "string",
									"description": "The OpenShift version (e.g., '4.16') to use when discovering official Red Hat catalogs.",
								},
								"packageName": map[string]interface{}{
									"type":        "string",
									"description": "The name of the operator package to inspect within the catalog.",
								},
								"channelName": map[string]interface{}{
									"type":        "string",
									"description": "The name of the channel to inspect within a package.",
								},
								"listCatalogs": map[string]interface{}{
									"type":        "boolean",
									"description": "Set to true to discover a list of available Red Hat catalogs for a given OpenShift version.",
								},
							},
						},
					},
				},
			},
		}

	case "tools/call":
		return handleToolCall(request.Params)

	default:
		return MCPResponse{
			Error: &MCPError{
				Code:    -32601,
				Message: fmt.Sprintf("Method not found: %s", request.Method),
			},
		}
	}
}

func handleToolCall(params map[string]interface{}) MCPResponse {
	// Extract tool call parameters
	name, ok := params["name"].(string)
	if !ok {
		return MCPResponse{
			Error: &MCPError{
				Code:    -32602,
				Message: "Invalid tool name",
			},
		}
	}

	arguments, ok := params["arguments"].(map[string]interface{})
	if !ok {
		return MCPResponse{
			Error: &MCPError{
				Code:    -32602,
				Message: "Invalid arguments",
			},
		}
	}

	if name != "lumen_list" {
		return MCPResponse{
			Error: &MCPError{
				Code:    -32602,
				Message: fmt.Sprintf("Unknown tool: %s", name),
			},
		}
	}

	// Extract arguments with defaults
	catalogRef := getStringArg(arguments, "catalogRef", "")
	ocpVersion := getStringArg(arguments, "ocpVersion", "")
	packageName := getStringArg(arguments, "packageName", "")
	channelName := getStringArg(arguments, "channelName", "")
	listCatalogs := getBoolArg(arguments, "listCatalogs", false)

	// Call the Lumen tool
	result, err := mcphandler.LumenToolHandler(catalogRef, ocpVersion, packageName, channelName, listCatalogs)
	if err != nil {
		return MCPResponse{
			Error: &MCPError{
				Code:    -32603,
				Message: fmt.Sprintf("Tool execution failed: %v", err),
			},
		}
	}

	return MCPResponse{
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": result,
				},
			},
		},
	}
}

func getStringArg(args map[string]interface{}, key, defaultValue string) string {
	if val, ok := args[key].(string); ok {
		return val
	}
	return defaultValue
}

func getBoolArg(args map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := args[key].(bool); ok {
		return val
	}
	return defaultValue
}
