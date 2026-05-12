# image-generation-mcp-server

A Go MCP server and REST API for text-to-image and image-to-image generation using Doubao Seedream.

## Features

- MCP server built with `github.com/modelcontextprotocol/go-sdk`
- Streamable HTTP transport for remote tool access
- REST API for direct service integration
- Doubao Seedream text-to-image support
- Doubao Seedream image-to-image support
- Docker deployment support
- Eino integration example

## Quick Start

```bash
cp .env.example .env
go run ./cmd/server
```

The server automatically loads environment variables from the project root `.env` file.

Default endpoints:

- MCP: `http://localhost:9101/mcp`
- API: `http://localhost:9101/api/v1`
- Health: `http://localhost:9101/healthz`

## Docker

```bash
docker compose up --build
```

## Quick Verification

Run the local smoke test after the server starts:

```bash
bash scripts/smoke_test.sh
```

## Documents

- `docs/api.md`
- `docs/mcp.md`
- `docs/eino.md`
- `docs/deploy.md`

## Notes

This project uses a thin Ark HTTP client so the MCP layer and REST layer share the same service contract.

The exact upstream request schema for some Seedream image editing options can vary by model revision. The integration keeps the business interface stable and forwards image-edit specific options through the provider layer.

Current default model ID:

- `doubao-seedream-4-5-251128`

Current default image size:

- `2048x2048`
