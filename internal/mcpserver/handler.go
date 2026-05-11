package mcpserver

import (
	"context"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yong/image-generation-mcp-server/internal/config"
	"github.com/yong/image-generation-mcp-server/internal/model"
	imagesvc "github.com/yong/image-generation-mcp-server/internal/service/image"
)

type textToImageInput struct {
	Prompt         string   `json:"prompt" jsonschema:"text prompt used to generate an image"`
	Size           string   `json:"size,omitempty" jsonschema:"output image size, for example 2048x2048"`
	ResponseFormat string   `json:"response_format,omitempty" jsonschema:"response format: url or b64_json"`
	Seed           *int64   `json:"seed,omitempty" jsonschema:"optional random seed"`
	Watermark      *bool    `json:"watermark,omitempty" jsonschema:"whether to keep the provider watermark"`
	GuidanceScale  *float64 `json:"guidance_scale,omitempty" jsonschema:"prompt guidance strength"`
	NumImages      *int     `json:"num_images,omitempty" jsonschema:"number of images to generate"`
}

type imageToImageInput struct {
	Prompt         string   `json:"prompt" jsonschema:"edit prompt used to transform the source image"`
	ImageURL       string   `json:"image_url,omitempty" jsonschema:"publicly accessible source image url"`
	ImageBase64    string   `json:"image_base64,omitempty" jsonschema:"base64 encoded source image content"`
	Size           string   `json:"size,omitempty" jsonschema:"output image size, for example 2048x2048"`
	ResponseFormat string   `json:"response_format,omitempty" jsonschema:"response format: url or b64_json"`
	Seed           *int64   `json:"seed,omitempty" jsonschema:"optional random seed"`
	Watermark      *bool    `json:"watermark,omitempty" jsonschema:"whether to keep the provider watermark"`
	Strength       *float64 `json:"strength,omitempty" jsonschema:"strength of the edit operation"`
	NumImages      *int     `json:"num_images,omitempty" jsonschema:"number of images to generate"`
}

func NewHandler(cfg config.Config, service *imagesvc.Service) http.Handler {
	server := mcp.NewServer(&mcp.Implementation{Name: cfg.MCPServerName, Version: cfg.MCPServerVersion}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "text_to_image",
		Description: "Generate one or more images from a text prompt using Doubao Seedream.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input textToImageInput) (*mcp.CallToolResult, model.GenerateImageResponse, error) {
		response, err := service.TextToImage(ctx, model.GenerateImageRequest{
			Prompt:         input.Prompt,
			Size:           input.Size,
			ResponseFormat: input.ResponseFormat,
			Seed:           input.Seed,
			Watermark:      input.Watermark,
			GuidanceScale:  input.GuidanceScale,
			NumImages:      input.NumImages,
		})
		return nil, response, err
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "image_to_image",
		Description: "Generate one or more edited images from a prompt and an input image using Doubao Seedream.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input imageToImageInput) (*mcp.CallToolResult, model.GenerateImageResponse, error) {
		response, err := service.ImageToImage(ctx, model.GenerateImageRequest{
			Prompt:         input.Prompt,
			ImageURL:       input.ImageURL,
			ImageBase64:    input.ImageBase64,
			Size:           input.Size,
			ResponseFormat: input.ResponseFormat,
			Seed:           input.Seed,
			Watermark:      input.Watermark,
			Strength:       input.Strength,
			NumImages:      input.NumImages,
		})
		return nil, response, err
	})

	return mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, nil)
}
