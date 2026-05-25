package image

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"image-generation-mcp-server/internal/common"
	"image-generation-mcp-server/internal/config"
	"image-generation-mcp-server/internal/model"
	"image-generation-mcp-server/internal/storage"
)

const minimumImagePixels = 3686400

type provider interface {
	TextToImage(context.Context, model.GenerateImageRequest) (model.GenerateImageResponse, error)
	ImageToImage(context.Context, model.GenerateImageRequest) (model.GenerateImageResponse, error)
}

type Service struct {
	config     config.Config
	provider   provider
	uploader   storage.Uploader
	httpClient *http.Client
}

func NewService(cfg config.Config, provider provider, uploader storage.Uploader) *Service {
	return &Service{
		config:     cfg,
		provider:   provider,
		uploader:   uploader,
		httpClient: &http.Client{Timeout: cfg.RequestTimeout},
	}
}

func (s *Service) TextToImage(ctx context.Context, input model.GenerateImageRequest) (model.GenerateImageResponse, error) {
	common.Debugf("text_to_image prompt=%s size=%s", input.Prompt, input.Size)

	var err error
	input, err = normalizeRequest(input)
	if err != nil {
		return model.GenerateImageResponse{}, err
	}
	if strings.TrimSpace(input.Prompt) == "" {
		return model.GenerateImageResponse{}, fmt.Errorf("prompt is required")
	}
	response, err := s.provider.TextToImage(ctx, input)
	return s.finalizeResponse(ctx, response, err)
}

func (s *Service) ImageToImage(ctx context.Context, input model.GenerateImageRequest) (model.GenerateImageResponse, error) {
	common.Debugf("image_to_image prompt=%s size=%s", input.Prompt, input.Size)

	var err error
	input, err = normalizeRequest(input)
	if err != nil {
		return model.GenerateImageResponse{}, err
	}
	if strings.TrimSpace(input.Prompt) == "" {
		return model.GenerateImageResponse{}, fmt.Errorf("prompt is required")
	}
	if strings.TrimSpace(input.ImageURL) == "" && strings.TrimSpace(input.ImageBase64) == "" {
		return model.GenerateImageResponse{}, fmt.Errorf("image_url or image_base64 is required")
	}
	input, err = s.prepareImageToImageInput(ctx, input)
	if err != nil {
		return model.GenerateImageResponse{}, err
	}
	response, err := s.provider.ImageToImage(ctx, input)
	return s.finalizeResponse(ctx, response, err)
}

func (s *Service) Models() model.ModelInfo {
	return model.ModelInfo{
		TextToImageModel:  s.config.ARKTextModel,
		ImageToImageModel: s.config.ARKImageModel,
	}
}

func normalizeRequest(input model.GenerateImageRequest) (model.GenerateImageRequest, error) {
	input.Prompt = strings.TrimSpace(input.Prompt)
	input.ImageURL = strings.TrimSpace(input.ImageURL)
	input.ImageBase64 = strings.TrimSpace(input.ImageBase64)
	input.Size = strings.TrimSpace(input.Size)
	if input.Size == "" {
		input.Size = "2048x2048"
	}
	normalizedSize, err := normalizeImageSize(input.Size)
	if err != nil {
		return model.GenerateImageRequest{}, err
	}
	input.Size = normalizedSize
	if input.ResponseFormat == "" {
		input.ResponseFormat = "url"
	}
	if input.NumImages == nil {
		defaultCount := 1
		input.NumImages = &defaultCount
	}
	return input, nil
}

func parseImageSize(size string) (int, int, error) {
	parts := strings.Split(size, "x")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("size must be in WIDTHxHEIGHT format, for example 2048x2048")
	}

	width, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || width <= 0 {
		return 0, 0, fmt.Errorf("size must be in WIDTHxHEIGHT format, for example 2048x2048")
	}
	height, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil || height <= 0 {
		return 0, 0, fmt.Errorf("size must be in WIDTHxHEIGHT format, for example 2048x2048")
	}

	return width, height, nil
}

func normalizeImageSize(size string) (string, error) {
	width, height, err := parseImageSize(size)
	if err != nil {
		return "", err
	}

	pixels := width * height
	if pixels >= minimumImagePixels {
		return fmt.Sprintf("%dx%d", width, height), nil
	}

	scale := math.Sqrt(float64(minimumImagePixels) / float64(pixels))
	normalizedWidth := int(math.Ceil(float64(width) * scale))
	normalizedHeight := int(math.Ceil(float64(height) * scale))

	for normalizedWidth*normalizedHeight < minimumImagePixels {
		normalizedHeight++
	}

	return fmt.Sprintf("%dx%d", normalizedWidth, normalizedHeight), nil
}

func (s *Service) prepareImageToImageInput(ctx context.Context, input model.GenerateImageRequest) (model.GenerateImageRequest, error) {
	if strings.TrimSpace(input.ImageURL) != "" {
		input.ImageBase64 = ""
		return input, nil
	}

	if s.uploader == nil {
		return model.GenerateImageRequest{}, fmt.Errorf("image_base64 requires configured uploader to obtain a public image url")
	}

	asset, err := decodeInlineImage(input.ImageBase64)
	if err != nil {
		return model.GenerateImageRequest{}, fmt.Errorf("decode image_base64: %w", err)
	}

	objectName := buildObjectName(s.config.MinIOObjectPrefix, "source", 0, asset.ext, time.Now())
	uploadedURL, err := s.uploader.Upload(ctx, objectName, asset.contentType, asset.data)
	if err != nil {
		return model.GenerateImageRequest{}, fmt.Errorf("upload image_base64: %w", err)
	}

	input.ImageURL = uploadedURL
	input.ImageBase64 = ""
	return input, nil
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

func (s *Service) finalizeResponse(ctx context.Context, response model.GenerateImageResponse, err error) (model.GenerateImageResponse, error) {
	response, err = s.withFallbackTimestamp(response, err)
	if err != nil {
		return model.GenerateImageResponse{}, err
	}
	if s.uploader == nil || len(response.Images) == 0 {
		return response, nil
	}

	for index, original := range response.Images {
		common.Debugf("upload image request_id=%s index=%d url=%s", response.RequestID, index, original)

		asset, resolveErr := s.resolveImageAsset(ctx, original)
		if resolveErr != nil {
			log.Printf("skip minio upload for request_id=%s image_index=%d: resolve image failed: %v", response.RequestID, index, resolveErr)
			continue
		}

		objectName := buildObjectName(s.config.MinIOObjectPrefix, response.RequestID, index, asset.ext, time.Now())
		uploadedURL, uploadErr := s.uploader.Upload(ctx, objectName, asset.contentType, asset.data)
		if uploadErr != nil {
			log.Printf("skip minio upload for request_id=%s image_index=%d object_name=%s: %v", response.RequestID, index, objectName, uploadErr)
			continue
		}

		response.Images[index] = uploadedURL
	}

	return response, nil
}
