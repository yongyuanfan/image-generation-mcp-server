package image

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/yong/image-generation-mcp-server/internal/config"
	"github.com/yong/image-generation-mcp-server/internal/model"
)

type provider interface {
	TextToImage(context.Context, model.GenerateImageRequest) (model.GenerateImageResponse, error)
	ImageToImage(context.Context, model.GenerateImageRequest) (model.GenerateImageResponse, error)
}

type Service struct {
	config   config.Config
	provider provider
}

func NewService(cfg config.Config, provider provider) *Service {
	return &Service{config: cfg, provider: provider}
}

func (s *Service) TextToImage(ctx context.Context, input model.GenerateImageRequest) (model.GenerateImageResponse, error) {
	input = normalizeRequest(input)
	if strings.TrimSpace(input.Prompt) == "" {
		return model.GenerateImageResponse{}, fmt.Errorf("prompt is required")
	}
	return s.withFallbackTimestamp(s.provider.TextToImage(ctx, input))
}

func (s *Service) ImageToImage(ctx context.Context, input model.GenerateImageRequest) (model.GenerateImageResponse, error) {
	input = normalizeRequest(input)
	if strings.TrimSpace(input.Prompt) == "" {
		return model.GenerateImageResponse{}, fmt.Errorf("prompt is required")
	}
	if strings.TrimSpace(input.ImageURL) == "" && strings.TrimSpace(input.ImageBase64) == "" {
		return model.GenerateImageResponse{}, fmt.Errorf("image_url or image_base64 is required")
	}
	return s.withFallbackTimestamp(s.provider.ImageToImage(ctx, input))
}

func (s *Service) Models() model.ModelInfo {
	return model.ModelInfo{
		TextToImageModel:  s.config.ARKTextModel,
		ImageToImageModel: s.config.ARKImageModel,
	}
}

func normalizeRequest(input model.GenerateImageRequest) model.GenerateImageRequest {
	input.Prompt = strings.TrimSpace(input.Prompt)
	input.ImageURL = strings.TrimSpace(input.ImageURL)
	input.ImageBase64 = strings.TrimSpace(input.ImageBase64)
	if input.Size == "" {
		input.Size = "2048x2048"
	}
	if input.ResponseFormat == "" {
		input.ResponseFormat = "url"
	}
	if input.NumImages == nil {
		defaultCount := 1
		input.NumImages = &defaultCount
	}
	return input
}

func (s *Service) withFallbackTimestamp(response model.GenerateImageResponse, err error) (model.GenerateImageResponse, error) {
	if err != nil {
		return model.GenerateImageResponse{}, err
	}
	if response.CreatedAt == 0 {
		response.CreatedAt = time.Now().Unix()
	}
	return response, nil
}
