package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"testing"
	"time"
)

// Types are defined in main.go

func TestMCPServer(t *testing.T) {
	// Build the server first
	buildCmd := exec.Command("go", "build", "-o", "../bin/mcp-server", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build MCP server: %v", err)
	}

	// Start the MCP server process
	cmd := exec.Command("../bin/mcp-server")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("Failed to create stdin pipe: %v", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to create stdout pipe: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("Failed to create stderr pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start MCP server: %v", err)
	}
	defer func() {
		cmd.Process.Kill()
		cmd.Wait()
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Create readers
	scanner := bufio.NewScanner(stdout)
	stderrScanner := bufio.NewScanner(stderr)

	// Monitor stderr in a goroutine
	go func() {
		for stderrScanner.Scan() {
			t.Logf("Server stderr: %s", stderrScanner.Text())
		}
	}()

	t.Run("ToolsList", func(t *testing.T) {
		t.Log("Testing tools/list...")

		toolsListRequest := MCPRequest{
			Method: "tools/list",
			Params: map[string]interface{}{},
		}

		if err := sendRequest(stdin, toolsListRequest); err != nil {
			t.Fatalf("Failed to send tools/list request: %v", err)
		}

		response, err := readResponse(scanner)
		if err != nil {
			t.Fatalf("Failed to read tools/list response: %v", err)
		}

		t.Logf("Tools list response: %s", formatJSON(response))

		if err := validateToolsList(response); err != nil {
			t.Fatalf("Tools/list validation failed: %v", err)
		}

		t.Log("✅ Tools list test passed")
	})

	t.Run("ToolCall", func(t *testing.T) {
		t.Log("Testing lumen_list tool call (list catalogs for OCP 4.16)...")

		toolCallRequest := MCPRequest{
			Method: "tools/call",
			Params: map[string]interface{}{
				"name": "lumen_list",
				"arguments": map[string]interface{}{
					"ocpVersion":   "4.16",
					"listCatalogs": true,
				},
			},
		}

		if err := sendRequest(stdin, toolCallRequest); err != nil {
			t.Fatalf("Failed to send tool call request: %v", err)
		}

		response, err := readResponse(scanner)
		if err != nil {
			t.Fatalf("Failed to read tool call response: %v", err)
		}

		t.Logf("Tool call response: %s", formatJSON(response))

		if err := validateToolCall(response); err != nil {
			t.Fatalf("Tool call validation failed: %v", err)
		}

		t.Log("✅ Tool call test passed")
	})
}

func sendRequest(stdin io.WriteCloser, request MCPRequest) error {
	encoder := json.NewEncoder(stdin)
	return encoder.Encode(request)
}

func readResponse(scanner *bufio.Scanner) (*MCPResponse, error) {
	if !scanner.Scan() {
		return nil, fmt.Errorf("no response received")
	}

	var response MCPResponse
	if err := json.Unmarshal(scanner.Bytes(), &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

func formatJSON(v interface{}) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("Failed to marshal JSON: %v", err)
	}
	return string(data)
}

func validateToolsList(response *MCPResponse) error {
	if response.Error != nil {
		return fmt.Errorf("received error: %s", response.Error.Message)
	}

	result, ok := response.Result.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected result type")
	}

	tools, ok := result["tools"].([]interface{})
	if !ok {
		return fmt.Errorf("tools not found in result")
	}

	for _, tool := range tools {
		toolMap, ok := tool.(map[string]interface{})
		if !ok {
			continue
		}

		if name, ok := toolMap["name"].(string); ok && name == "lumen_list" {
			return nil
		}
	}

	return fmt.Errorf("lumen_list tool not found")
}

func validateToolCall(response *MCPResponse) error {
	if response.Error != nil {
		return fmt.Errorf("tool call failed: %s", response.Error.Message)
	}

	if response.Result == nil {
		return fmt.Errorf("no result in response")
	}

	return nil
}
