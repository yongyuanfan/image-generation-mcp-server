# API

## Health Check

`GET /healthz`

Response:

```json
{
  "status": "ok"
}
```

## List Models

`GET /api/v1/models`

Response:

```json
{
  "text_to_image_model": "doubao-seedream-3-0-t2i-250415",
  "image_to_image_model": "doubao-seededit-3-0-i2i-250628"
}
```

## Text To Image

`POST /api/v1/images/generations`

Request:

```json
{
  "prompt": "A cinematic cat astronaut on the moon",
  "size": "1024x1024",
  "response_format": "url",
  "num_images": 1
}
```

## Image To Image

`POST /api/v1/images/edits`

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
  "model": "doubao-seedream-3-0-t2i-250415",
  "created_at": 1770000000
}
```
