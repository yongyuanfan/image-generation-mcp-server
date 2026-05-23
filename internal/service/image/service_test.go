package image

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"image-generation-mcp-server/internal/config"
	"image-generation-mcp-server/internal/model"
)

type stubProvider struct {
	response model.GenerateImageResponse
	err      error
	request  model.GenerateImageRequest
}

func (s *stubProvider) TextToImage(_ context.Context, input model.GenerateImageRequest) (model.GenerateImageResponse, error) {
	s.request = input
	return s.response, s.err
}

func (s *stubProvider) ImageToImage(_ context.Context, input model.GenerateImageRequest) (model.GenerateImageResponse, error) {
	s.request = input
	return s.response, s.err
}

type stubUploader struct {
	url         string
	err         error
	objectName  string
	contentType string
	data        []byte
}

func (s *stubUploader) Upload(_ context.Context, objectName, contentType string, data []byte) (string, error) {
	s.objectName = objectName
	s.contentType = contentType
	s.data = append([]byte(nil), data...)
	if s.err != nil {
		return "", s.err
	}
	return s.url, nil
}

func TestTextToImageRewritesURLAfterUpload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte("png-data"))
	}))
	defer server.Close()

	uploader := &stubUploader{url: "https://cdn.example.com/generated-images/ai-images/2026/05/14/request-1-0.png"}
	provider := &stubProvider{response: model.GenerateImageResponse{
		Images:    []string{server.URL},
		RequestID: "request-1",
		CreatedAt: 123,
	}}
	service := NewService(config.Config{
		RequestTimeout:    time.Second,
		MinIOObjectPrefix: "ai-images",
	}, provider, uploader)

	response, err := service.TextToImage(context.Background(), model.GenerateImageRequest{Prompt: "test prompt"})
	if err != nil {
		t.Fatalf("TextToImage returned error: %v", err)
	}

	if len(response.Images) != 1 {
		t.Fatalf("expected 1 image, got %d", len(response.Images))
	}
	if response.Images[0] != uploader.url {
		t.Fatalf("expected uploaded url %q, got %q", uploader.url, response.Images[0])
	}
	if uploader.objectName == "" {
		t.Fatal("expected uploader to receive object name")
	}
	if uploader.contentType != "image/png" {
		t.Fatalf("expected content type image/png, got %q", uploader.contentType)
	}
	if string(uploader.data) != "png-data" {
		t.Fatalf("expected uploaded data to match source image, got %q", string(uploader.data))
	}
}

func TestTextToImageFallsBackToOriginalURLWhenUploadFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte("png-data"))
	}))
	defer server.Close()

	uploader := &stubUploader{err: errors.New("upload failed")}
	provider := &stubProvider{response: model.GenerateImageResponse{
		Images:    []string{server.URL},
		RequestID: "request-2",
		CreatedAt: 456,
	}}
	service := NewService(config.Config{
		RequestTimeout:    time.Second,
		MinIOObjectPrefix: "ai-images",
	}, provider, uploader)

	response, err := service.TextToImage(context.Background(), model.GenerateImageRequest{Prompt: "test prompt"})
	if err != nil {
		t.Fatalf("TextToImage returned error: %v", err)
	}

	if len(response.Images) != 1 {
		t.Fatalf("expected 1 image, got %d", len(response.Images))
	}
	if response.Images[0] != server.URL {
		t.Fatalf("expected original url %q after upload failure, got %q", server.URL, response.Images[0])
	}
	if uploader.objectName == "" {
		t.Fatal("expected uploader to be called before fallback")
	}
}

func TestTextToImageRejectsTooSmallSize(t *testing.T) {
	provider := &stubProvider{response: model.GenerateImageResponse{CreatedAt: 123}}
	service := NewService(config.Config{RequestTimeout: time.Second}, provider, nil)
	_, err := service.TextToImage(context.Background(), model.GenerateImageRequest{
		Prompt: "test prompt",
		Size:   "1024x1024",
	})
	if err != nil {
		t.Fatalf("expected small size to be normalized, got %v", err)
	}
	if provider.request.Size != "1920x1920" {
		t.Fatalf("expected normalized size 1920x1920, got %q", provider.request.Size)
	}
}

func TestTextToImageRejectsInvalidSizeFormat(t *testing.T) {
	service := NewService(config.Config{RequestTimeout: time.Second}, &stubProvider{}, nil)
	_, err := service.TextToImage(context.Background(), model.GenerateImageRequest{
		Prompt: "test prompt",
		Size:   "1024*1024",
	})
	if err == nil || !strings.Contains(err.Error(), "size must be in WIDTHxHEIGHT format") {
		t.Fatalf("expected invalid size format error, got %v", err)
	}
}

func TestTextToImageUsesDefaultValidSize(t *testing.T) {
	provider := &stubProvider{response: model.GenerateImageResponse{}}
	service := NewService(config.Config{RequestTimeout: time.Second}, provider, nil)
	_, err := service.TextToImage(context.Background(), model.GenerateImageRequest{Prompt: "test prompt"})
	if err != nil {
		t.Fatalf("expected default size to pass validation, got %v", err)
	}
	if provider.request.Size != "2048x2048" {
		t.Fatalf("expected default size 2048x2048, got %q", provider.request.Size)
	}
}

func TestImageToImageNormalizesTooSmallSize(t *testing.T) {
	provider := &stubProvider{response: model.GenerateImageResponse{CreatedAt: 123}}
	service := NewService(config.Config{RequestTimeout: time.Second}, provider, nil)

	_, err := service.ImageToImage(context.Background(), model.GenerateImageRequest{
		Prompt:   "edit prompt",
		ImageURL: "https://example.com/source.png",
		Size:     "1024x1024",
	})
	if err != nil {
		t.Fatalf("ImageToImage returned error: %v", err)
	}
	if provider.request.Size != "1920x1920" {
		t.Fatalf("expected normalized size 1920x1920, got %q", provider.request.Size)
	}
}

func TestImageToImageKeepsImageURL(t *testing.T) {
	provider := &stubProvider{response: model.GenerateImageResponse{CreatedAt: 123}}
	service := NewService(config.Config{RequestTimeout: time.Second}, provider, nil)

	_, err := service.ImageToImage(context.Background(), model.GenerateImageRequest{
		Prompt:   "edit prompt",
		ImageURL: "https://example.com/source.png",
	})
	if err != nil {
		t.Fatalf("ImageToImage returned error: %v", err)
	}

	if provider.request.ImageURL != "https://example.com/source.png" {
		t.Fatalf("expected image_url to be preserved, got %q", provider.request.ImageURL)
	}
	if provider.request.ImageBase64 != "" {
		t.Fatalf("expected image_base64 to be cleared, got %q", provider.request.ImageBase64)
	}
}

func TestImageToImageUploadsInlineBase64(t *testing.T) {
	provider := &stubProvider{response: model.GenerateImageResponse{CreatedAt: 123}}
	uploader := &stubUploader{url: "https://cdn.example.com/source/source-0.png"}
	service := NewService(config.Config{RequestTimeout: time.Second, MinIOObjectPrefix: "source"}, provider, uploader)

	inline := "data:image/png;base64," + base64.StdEncoding.EncodeToString([]byte("inline-image"))
	_, err := service.ImageToImage(context.Background(), model.GenerateImageRequest{
		Prompt:      "edit prompt",
		ImageBase64: inline,
	})
	if err != nil {
		t.Fatalf("ImageToImage returned error: %v", err)
	}

	if string(uploader.data) != "inline-image" {
		t.Fatalf("expected uploaded inline image data, got %q", string(uploader.data))
	}
	if provider.request.ImageURL != uploader.url {
		t.Fatalf("expected uploaded image url %q, got %q", uploader.url, provider.request.ImageURL)
	}
	if provider.request.ImageBase64 != "" {
		t.Fatalf("expected image_base64 to be cleared, got %q", provider.request.ImageBase64)
	}
}

func TestImageToImageRejectsInlineBase64WithoutUploader(t *testing.T) {
	provider := &stubProvider{response: model.GenerateImageResponse{CreatedAt: 123}}
	service := NewService(config.Config{RequestTimeout: time.Second}, provider, nil)

	inline := "data:image/png;base64," + base64.StdEncoding.EncodeToString([]byte("inline-image"))
	_, err := service.ImageToImage(context.Background(), model.GenerateImageRequest{
		Prompt:      "edit prompt",
		ImageBase64: inline,
	})
	if err == nil || !strings.Contains(err.Error(), "image_base64 requires configured uploader") {
		t.Fatalf("expected uploader configuration error, got %v", err)
	}
}
