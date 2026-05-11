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

Default endpoints:

- MCP: `http://localhost:8080/mcp`
- API: `http://localhost:8080/api/v1`
- Health: `http://localhost:8080/healthz`

## Docker

```bash
docker compose up --build
```

## Documents

- `docs/api.md`
- `docs/mcp.md`
- `docs/eino.md`
- `docs/deploy.md`

## Notes

This project uses a thin Ark HTTP client so the MCP layer and REST layer share the same service contract.

The exact upstream request schema for some Seedream image editing options can vary by model revision. The integration keeps the business interface stable and forwards image-edit specific options through the provider layer.
