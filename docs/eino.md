# Eino Integration

This project is designed to be consumed by `cloudwego/eino` through MCP.

## Run the server

```bash
cp .env.example .env
go run ./cmd/server
```

## Run the Eino example

```bash
MCP_SERVER_URL=http://localhost:9101/mcp \
ARK_API_KEY=your_ark_api_key \
go run ./examples/eino
```

The example shows how to:

1. Connect to the Streamable HTTP MCP endpoint.
2. Discover `text_to_image` and `image_to_image` tools.
3. Attach the tools for later use in an Eino agent.

## Typical agent wiring

Use the discovered tools together with an Ark chat model in a `ChatModelAgent`.

The exact agent prompt depends on your application, but the MCP integration point is the tool list returned from `mcptool.GetTools(...)`.
