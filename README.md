# image-generation-mcp-server

A Go MCP server and REST API for text-to-image and image-to-image generation using Doubao Seedream.

## Features

- MCP server built with `github.com/modelcontextprotocol/go-sdk`
- Streamable HTTP transport for remote tool access
- REST API for direct service integration
- Doubao Seedream text-to-image support
- Doubao Seedream image-to-image support
- Optional MinIO persistence with public URL rewrite
- Docker deployment support
- Eino integration example

## Quick Start

```bash
cp .env.example .env
go run ./cmd/server
```

The server automatically loads environment variables from the project root `.env` file.

If MinIO is configured, generated images are downloaded or decoded by the service, uploaded to MinIO, and the returned `images` values are rewritten to `MINIO_PUBLIC_BASE_URL/{bucket}/{objectName}`. When an upload fails, the service falls back to the original provider URL or inline image.

When the request `size` is smaller than the provider minimum pixel requirement, the service automatically scales it up while keeping the original aspect ratio.

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
