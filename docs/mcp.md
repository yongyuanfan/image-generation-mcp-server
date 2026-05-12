# MCP

This project exposes a Streamable HTTP MCP server at `/mcp`.

## Tools

### `text_to_image`

Generate one or more images from a text prompt.

Inputs:

- `prompt`
- `size`
- `response_format`
- `seed`
- `watermark`
- `guidance_scale`
- `num_images`

### `image_to_image`

Generate one or more edited images from a prompt and an input image.

Inputs:

- `prompt`
- `image_url`
- `image_base64`
- `size`
- `response_format`
- `seed`
- `watermark`
- `strength`
- `num_images`

Both tools return:

- `images`
- `request_id`
- `model`
- `created_at`

## Local verification

Start the server and connect an MCP client to:

```text
http://localhost:9101/mcp
```

## JSON-RPC examples

Initialize session:

```bash
curl -X POST http://localhost:9101/mcp \
  -H 'Content-Type: application/json' \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2025-03-26",
      "capabilities": {},
      "clientInfo": {
        "name": "curl-client",
        "version": "0.1.0"
      }
    }
  }'
```

List tools:

```bash
curl -X POST http://localhost:9101/mcp \
  -H 'Content-Type: application/json' \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/list",
    "params": {}
  }'
```

Call `text_to_image`:

```bash
curl -X POST http://localhost:9101/mcp \
  -H 'Content-Type: application/json' \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
      "name": "text_to_image",
      "arguments": {
        "prompt": "A cinematic cat astronaut on the moon",
        "size": "2048x2048",
        "response_format": "url"
      }
    }
  }'
```
