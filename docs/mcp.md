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
http://localhost:8080/mcp
```
