package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/aguidirh/lumen/internal/mcphandler"
)

// MCPRequest represents an incoming MCP tool call request
type MCPRequest struct {
	ID     json.RawMessage        `json:"id,omitempty"`
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
}

// MCPResponse represents an MCP response
type MCPResponse struct {
	ID      json.RawMessage `json:"id,omitempty"`
	JSONRPC string          `json:"jsonrpc"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *MCPError       `json:"error,omitempty"`
}

// MCPError represents an MCP error
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func main() {
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for {
		var request MCPRequest
		if err := decoder.Decode(&request); err != nil {
			if err == io.EOF {
				return // Clean exit on EOF
			}
			// Don't log here, as it would corrupt stdout
			continue
		}

		response := handleRequest(request)

		// Per JSON-RPC spec, a response must only be sent for requests, not notifications.
		// A notification is a request object without an "id" member.
		if request.ID != nil {
			if err := encoder.Encode(&response); err != nil {
				// Don't log here, as it would corrupt stdout
			}
		}
	}
}

func handleRequest(request MCPRequest) MCPResponse {
	response := MCPResponse{ID: request.ID, JSONRPC: "2.0"}
	switch request.Method {
	case "initialize":
		response.Result = map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "lumen-mcp-server",
				"version": "0.1.0",
			},
		}
	case "tools/list":
		response.Result = map[string]interface{}{
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
		}
	case "tools/call":
		response = handleToolCall(request)
	default:
		response.Error = &MCPError{
			Code:    -32601,
			Message: fmt.Sprintf("Method not found: %s", request.Method),
		}
	}
	return response
}

func handleToolCall(request MCPRequest) MCPResponse {
	response := MCPResponse{ID: request.ID, JSONRPC: "2.0"}
	params := request.Params
	name, ok := params["name"].(string)
	if !ok {
		response.Error = &MCPError{Code: -32602, Message: "Invalid tool name"}
		return response
	}

	if name != "lumen_list" {
		response.Error = &MCPError{Code: -32602, Message: fmt.Sprintf("Unknown tool: %s", name)}
		return response
	}

	var arguments map[string]interface{}
	if args, ok := params["arguments"].(map[string]interface{}); ok {
		arguments = args
	} else {
		arguments = make(map[string]interface{})
	}

	catalogRef := getStringArg(arguments, "catalogRef", "")
	ocpVersion := getStringArg(arguments, "ocpVersion", "")
	packageName := getStringArg(arguments, "packageName", "")
	channelName := getStringArg(arguments, "channelName", "")
	listCatalogs := getBoolArg(arguments, "listCatalogs", false)

	result, err := mcphandler.LumenToolHandler(catalogRef, ocpVersion, packageName, channelName, listCatalogs)
	if err != nil {
		response.Error = &MCPError{Code: -32603, Message: fmt.Sprintf("Tool execution failed: %v", err)}
		return response
	}

	response.Result = map[string]interface{}{
		"content": []map[string]interface{}{
			{"type": "text", "text": result},
		},
	}
	return response
}

func getStringArg(args map[string]interface{}, key, defaultValue string) string {
	if val, ok := args[key]; ok {
		if strVal, isString := val.(string); isString {
			return strVal
		}
	}
	return defaultValue
}

func getBoolArg(args map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := args[key]; ok {
		if boolVal, isBool := val.(bool); isBool {
			return boolVal
		}
	}
	return defaultValue
}
