# API

## Health Check

`GET /healthz`

Example:

```bash
curl http://localhost:9101/healthz
```

Response:

```json
{
  "status": "ok"
}
```

## List Models

`GET /api/v1/models`

Example:

```bash
curl http://localhost:9101/api/v1/models
```

Response:

```json
{
  "text_to_image_model": "doubao-seedream-4-5-251128",
  "image_to_image_model": "doubao-seedream-4-5-251128"
}
```

## Text To Image

`POST /api/v1/images/generations`

If MinIO is configured, the server uploads generated images to MinIO and returns `MINIO_PUBLIC_BASE_URL/{bucket}/{objectName}`. If an upload fails, the original upstream image address is returned for that image.

If `size` is smaller than the provider minimum pixel requirement, the server automatically scales it up while preserving the original aspect ratio.

Example:

```bash
curl -X POST http://localhost:9101/api/v1/images/generations \
  -H 'Content-Type: application/json' \
  -d '{
    "prompt": "A cinematic cat astronaut on the moon",
    "size": "2048x2048",
    "response_format": "url",
    "num_images": 1
  }'
```

Request:

```json
{
  "prompt": "A cinematic cat astronaut on the moon",
  "size": "2048x2048",
  "response_format": "url",
  "num_images": 1
}
```

## Image To Image

`POST /api/v1/images/edits`

The same MinIO upload and fallback behavior applies to edited images.

If `size` is smaller than the provider minimum pixel requirement, the server automatically scales it up while preserving the original aspect ratio.

Example with image URL:

```bash
curl -X POST http://localhost:9101/api/v1/images/edits \
  -H 'Content-Type: application/json' \
  -d '{
    "prompt": "Turn this sketch into a watercolor poster",
    "image_url": "https://example.com/input.png",
    "response_format": "url",
    "strength": 0.7
  }'
```

Example with base64 image:

```bash
curl -X POST http://localhost:9101/api/v1/images/edits \
  -H 'Content-Type: application/json' \
  -d '{
    "prompt": "Convert this into a product poster",
    "image_base64": "<BASE64_IMAGE>",
    "response_format": "b64_json"
  }'
```

Request:

```json
{
  "prompt": "Turn this sketch into a watercolor poster",
  "image_url": "https://example.com/input.png",
  "response_format": "url",
  "strength": 0.7
}
```

Common response:

```json
{
  "images": [
    "https://..."
  ],
  "request_id": "202605111234567890",
  "model": "doubao-seedream-4-5-251128",
  "created_at": 1770000000
}
```

Validation error example:

```json
{
  "error": {
    "message": "prompt is required"
  }
}
```
