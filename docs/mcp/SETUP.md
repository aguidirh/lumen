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
```bash
# Start the MCP server manually
make run-mcp

# Or run the binary directly
./bin/mcp-server
```

## Starting the MCP Server

**IMPORTANT**: Before configuring your IDE, you must start the MCP server as a background process.

### Start the Server
```bash
# Start the MCP server in the background
make run-mcp &

# Or run the binary directly in the background
./bin/mcp-server &
```

The server will run continuously in the background, listening for MCP requests from your IDE. You can verify it's running by checking the process:

```bash
# Check if the server is running
ps aux | grep mcp-server
```

### Stop the Server
When you're done, you can stop the server:

```bash
# Find and kill the MCP server process
pkill -f mcp-server
```

## IDE Integration Setup

The MCP server works with any MCP-compatible client. Here are setup instructions for different IDEs:

### Cursor IDE

**Option 1: Using Cursor's MCP Settings**
1. Open Cursor IDE
2. Go to Settings → Extensions → MCP
3. Add a new MCP server with these settings:
   - **Name**: `lumen`
   - **Command**: `./bin/mcp-server`
   - **Working Directory**: `/home/aguidi/go/src/github.com/aguidirh/lumen`

**Option 2: Using Global Configuration File**
1. Create or edit the global MCP configuration file:
   - **Linux**: `~/.cursor/mcp.json`
   - **macOS**: `~/.cursor/mcp.json` 
   - **Windows**: `%USERPROFILE%\.cursor\mcp.json`

2. Copy the contents of `cursor-config.json` into this file

3. Restart Cursor IDE

**Option 3: Using Project-Level Configuration**
1. Create an `mcp.json` file in your project root directory
2. Copy the contents of `cursor-config.json` into this file
3. Restart Cursor IDE

**Note**: Recent versions of Cursor use JSON configuration files instead of the UI-based setup. The project-level configuration (`mcp.json` in project root) tends to be more reliable than global configuration on some systems.

### Claude Desktop App

1. Locate your Claude Desktop configuration file:
   - **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
   - **Windows**: `%APPDATA%/Claude/claude_desktop_config.json`
   - **Linux**: `~/.config/Claude/claude_desktop_config.json`

2. Copy the contents of `claude-desktop-config.json` into your configuration file
3. Restart Claude Desktop

### VSCode

1. Install an MCP extension (like "MCP Client" or "Continue.dev")
2. Add the configuration from `vscode-settings.json` to your VSCode settings
3. Restart VSCode

### Other MCP Clients

The server implements the standard MCP protocol, so it should work with any MCP client. Use these connection details:
- **Command**: `./bin/mcp-server`
- **Working Directory**: `/home/aguidi/go/src/github.com/aguidirh/lumen`
- **Protocol**: stdio (reads from stdin, writes to stdout)

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

## Testing Different Scenarios

### List Catalogs for OCP 4.16
```bash
echo '{"method":"tools/call","params":{"name":"lumen_list","arguments":{"ocpVersion":"4.16","listCatalogs":true}}}' | ./bin/mcp-server
```

### List Packages in a Catalog
```bash
echo '{"method":"tools/call","params":{"name":"lumen_list","arguments":{"catalogRef":"registry.redhat.io/redhat/community-operator-index:v4.16"}}}' | ./bin/mcp-server
```

### List Channels for a Package
```bash
echo '{"method":"tools/call","params":{"name":"lumen_list","arguments":{"catalogRef":"registry.redhat.io/redhat/community-operator-index:v4.16","packageName":"prometheus"}}}' | ./bin/mcp-server
```

## Available Make Targets

For MCP-specific operations:
- `make build-mcp` - Build MCP server only
- `make test-mcp` - Test MCP server functionality
- `make run-mcp` - Start the MCP server interactively

See the main [README.md](../../README.md) for all available make targets.

## Troubleshooting

### IDE Cannot Connect to MCP Server
- **First, ensure the MCP server is running**: `ps aux | grep mcp-server`
- If not running, start it: `./bin/mcp-server &`
- Restart your IDE after starting the server

### Server Not Starting
- Ensure binaries are built: `make build-mcp`
- Check that all dependencies are available: `make tidy`

### Network Issues
- The tool will show progress messages when pulling images from registries
- Ensure you have internet connectivity for accessing container registries
- Check that you have permissions to access Red Hat registries

### Tool Calls Failing
- Run `make test-mcp` to verify basic functionality
- Check the server logs for detailed error messages
- Verify the parameters match the expected schema
- Ensure the catalog references are valid and accessible

## What's Next

Once the MCP server is configured in your IDE, you can:
1. Ask AI assistants to help you explore OpenShift operator catalogs
2. Get information about available operators and their versions
3. Analyze operator upgrade paths and channel structures
4. All with real-time data from official Red Hat registries!

The AI assistant will be able to use the `lumen_list` tool automatically to answer questions about OpenShift operators and catalogs. 

## Step-by-Step Troubleshooting

### 3. **Restart Cursor Completely**
1. **Quit Cursor entirely** (not just close the window)
2. **Kill any lingering processes**:
   ```bash
   pkill -f cursor
   ```
3. **Start Cursor fresh**

### 4. **Check Cursor's MCP Support**
Some versions of Cursor have limited or experimental MCP support. Try these approaches:

**Option A: Enable MCP in Settings**
1. Open Cursor Settings
2. Search for "MCP" or "Model Context Protocol"
3. Enable MCP support if there's a toggle

**Option B: Try Alternative Configuration**
Some Cursor versions expect different configuration formats. Create a `.cursor-server` directory approach: 