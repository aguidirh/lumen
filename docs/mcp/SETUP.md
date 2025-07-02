# Lumen MCP Server Setup

This directory contains a complete MCP (Model Context Protocol) server implementation that allows AI assistants to interact with the Lumen tool directly.

The core principle is that our `lumen` project is designed to be a tool, which can be used by other programs. This MCP handler is one such program that exposes Lumen's capabilities to AI assistants via the Model Context Protocol.

## What We've Built

1. **MCP Server** (`server/main.go`) - A complete MCP server that exposes Lumen functionality
2. **MCP Tests** (`server/main_test.go`) - Go tests to verify the MCP server functionality
3. **MCP Handler** (`internal/mcphandler/handler.go`) - The bridge between MCP calls and Lumen library functions

## How It Works

The MCP integration follows this architecture:

```
AI Assistant (Cursor/Claude/etc.)
        ↓ (MCP protocol over stdio)
    MCP Server (server/main.go)
        ↓ (function calls)
    MCP Handler (internal/mcphandler/handler.go)
        ↓ (library calls)
    Lumen Library (internal/pkg/*)
        ↓ (container registry calls)
    Red Hat Container Registries
```

## IDE Integration Setup

The MCP server works with any MCP-compatible client. Here are the setup instructions for VSCode.

### VSCode

1.  Make sure the MCP server binary is built:
    ```bash
    make build-mcp
    ```
2.  Open your VSCode `.vscode/mcp.json` file.
3.  Add the following configuration, ensuring the `command` path is an **absolute path** to the `mcp-server` binary on your system.
```json
{
  "mcpServers": {
    "lumen": {
      "type": "stdio",
      "command": "/home/aguidi/go/src/github.com/aguidirh/lumen/bin/mcp-server",
      "args": []
    }
  }
} 
```
4.  Restart VSCode. The `lumen_list` tool should now be available to any MCP-compatible extension.

## Testing the MCP Server

### Build the MCP Server
```bash
make build-mcp
```

This will build the MCP server binary to `bin/mcp-server`. 

To build everything (including the main application), use `make all` as described in the main [README.md](../../README.md).

### Run MCP Server Tests
```bash
make test-mcp
```

This will:
1. Build the MCP server binary
2. Run Go tests that start the server
3. Test the `tools/list` endpoint
4. Test calling the `lumen_list` tool with OCP version 4.16
5. Verify all responses

### Manual Testing

The MCP server is designed for interactive, back-and-forth communication. The `echo ... | ./bin/mcp-server` command will not work because the connection is closed immediately after the command is sent, causing the server to encounter an error.

Follow these steps to test manually:

**1. Start the Server Interactively**

Run the server from your project root. It will wait for input.

```bash
./bin/mcp-server
```

**2. Send a Request**

Once the server is running, copy one of the JSON requests below, paste it into the same terminal, and press **Enter**.

**3. View the Response**

The server will print the JSON response directly to your terminal and then continue listening for more requests.

**4. Stop the Server**

Press `Ctrl+C` to terminate the server process when you are finished.

### Example Scenarios

Here are some requests you can use for testing:

#### List Catalogs for OCP 4.16
Copy and paste this JSON into the running server:
```json
{"method":"tools/call","params":{"name":"lumen_list","arguments":{"ocpVersion":"4.16","listCatalogs":true}}}
```

#### List Packages in a Catalog
Copy and paste this JSON into the running server:
```json
{"method":"tools/call","params":{"name":"lumen_list","arguments":{"catalogRef":"registry.redhat.io/redhat/community-operator-index:v4.16"}}}
```

#### List Channels for a Package
Copy and paste this JSON into the running server:
```json
{"method":"tools/call","params":{"name":"lumen_list","arguments":{"catalogRef":"registry.redhat.io/redhat/community-operator-index:v4.16","packageName":"prometheus"}}}
```

## Available Tool

The MCP server exposes one tool:

### `lumen_list`
Introspects operator-framework catalog images to list their contents.

**Tool Schema:**
```json
{
  "name": "lumen_list",
  "description": "Introspects an operator-framework catalog image to list its contents. Can list all packages (operators), all channels for a given package, or all bundle versions for a given channel.",
  "inputSchema": {
    "type": "object",
    "properties": {
      "catalogRef": {
        "type": "string",
        "description": "The full image reference of the catalog to inspect (e.g., 'registry.redhat.io/redhat/community-operator-index:v4.16')."
      },
      "ocpVersion": {
        "type": "string",
        "description": "The OpenShift version (e.g., '4.16') to use when discovering official Red Hat catalogs."
      },
      "packageName": {
        "type": "string",
        "description": "The name of the operator package to inspect within the catalog."
      },
      "channelName": {
        "type": "string",
        "description": "The name of the channel to inspect within a package."
      },
      "listCatalogs": {
        "type": "boolean",
        "description": "Set to true to discover a list of available Red Hat catalogs for a given OpenShift version."
      }
    },
    "required": []
  }
}
```

**Parameters:**
- `catalogRef` (string): Full image reference of the catalog to inspect
- `ocpVersion` (string): OpenShift version for discovering Red Hat catalogs
- `packageName` (string): Name of the operator package to inspect
- `channelName` (string): Name of the channel to inspect within a package
- `listCatalogs` (boolean): Set to true to discover available Red Hat catalogs

**Example Usage:**
```json
{
  "name": "lumen_list",
  "arguments": {
    "ocpVersion": "4.16",
    "listCatalogs": true
  }
}
```

## Available Make Targets

For MCP-specific operations:
- `make build-mcp` - Build MCP server only
- `make test-mcp` - Test MCP server functionality
- `make run-mcp` - Start the MCP server interactively

See the main [README.md](../../README.md) for all available make targets.

## What's Next

Once the MCP server is configured in your IDE, you can:
1. Ask AI assistants to help you explore OpenShift operator catalogs
2. Get information about available operators and their versions
3. Get information about available operator channels

The AI assistant will be able to use the `lumen_list` tool automatically to answer questions about OpenShift operators and catalogs. 