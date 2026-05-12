# Deployment

## Docker

```bash
cp .env.example .env
docker compose up --build
```

The service listens on port `9101` by default.

## Exposed endpoints

- `GET /healthz`
- `GET /api/v1/models`
- `POST /api/v1/images/generations`
- `POST /api/v1/images/edits`
- `POST /mcp`

## Required environment variables

- `ARK_API_KEY`

## Optional environment variables

- `HTTP_ADDR`
- `API_PREFIX`
- `MCP_PATH`
- `ARK_BASE_URL`
- `ARK_IMAGE_ENDPOINT_PATH`
- `ARK_MODEL_TEXT2IMAGE`
- `ARK_MODEL_IMAGE2IMAGE`
- `REQUEST_TIMEOUT_SECONDS`
