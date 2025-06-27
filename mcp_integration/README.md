# Lumen MCP/Agent Tool Integration

This directory contains the necessary components and instructions to expose the `lumen` tool's capabilities to a Large Language Model (LLM) agent via a Model-Context-Protocol (MCP) server or a similar agent tooling platform.

The core principle is that our `lumen` project is already designed as a clean Go library (`pkg/catalog`), which can be easily imported and used by other Go programs. This `runner` is one such program.

## Steps for Integration

### 1. Define the Tool Schema

First, you need to define a schema that tells the LLM agent how to use your tool. This is typically done in a format like JSON or YAML on the agent tooling platform. The schema describes the tool's name, its purpose, and the parameters it accepts.

**Example Tool Schema (in JSON):**
```json
{
  "name": "lumen.List",
  "description": "Introspects an operator-framework catalog image to list its contents. Can list all packages (operators), all channels for a given package, or all bundle versions for a given channel.",
  "parameters": {
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

### 2. Implement the Tool Runner

The tooling platform needs a Go function to execute when the agent calls the tool. The `runner.go` file in this directory contains a template for this function. This runner code acts as an adapter: it receives the simple string parameters from the agent, calls our `lumen` library using the proper Go structs, and then serializes the results back into a format (like JSON) that the agent can understand.

### 3. Register the Tool on the Platform

The final step is to register the tool on your chosen platform. This typically involves:
1.  Uploading or referencing your tool's schema from Step 1.
2.  Providing the compiled Go code (the runner) that the platform will execute.

Once registered, the LLM agent will be aware of the `lumen.List` tool and can call it to help answer user questions about operator catalogs. 