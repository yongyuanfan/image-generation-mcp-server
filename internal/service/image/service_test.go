package image

import (
	"context"
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
}

func (s stubProvider) TextToImage(context.Context, model.GenerateImageRequest) (model.GenerateImageResponse, error) {
	return s.response, s.err
}

func (s stubProvider) ImageToImage(context.Context, model.GenerateImageRequest) (model.GenerateImageResponse, error) {
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
	service := NewService(config.Config{
		RequestTimeout:    time.Second,
		MinIOObjectPrefix: "ai-images",
	}, stubProvider{response: model.GenerateImageResponse{
		Images:    []string{server.URL},
		RequestID: "request-1",
		CreatedAt: 123,
	}}, uploader)

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
	service := NewService(config.Config{
		RequestTimeout:    time.Second,
		MinIOObjectPrefix: "ai-images",
	}, stubProvider{response: model.GenerateImageResponse{
		Images:    []string{server.URL},
		RequestID: "request-2",
		CreatedAt: 456,
	}}, uploader)

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
	service := NewService(config.Config{RequestTimeout: time.Second}, stubProvider{}, nil)
	_, err := service.TextToImage(context.Background(), model.GenerateImageRequest{
		Prompt: "test prompt",
		Size:   "1024x1024",
	})
	if err == nil || !strings.Contains(err.Error(), "size must be at least 3686400 pixels") {
		t.Fatalf("expected size validation error, got %v", err)
	}
}

func TestTextToImageRejectsInvalidSizeFormat(t *testing.T) {
	service := NewService(config.Config{RequestTimeout: time.Second}, stubProvider{}, nil)
	_, err := service.TextToImage(context.Background(), model.GenerateImageRequest{
		Prompt: "test prompt",
		Size:   "1024*1024",
	})
	if err == nil || !strings.Contains(err.Error(), "size must be in WIDTHxHEIGHT format") {
		t.Fatalf("expected invalid size format error, got %v", err)
	}
}

func TestTextToImageUsesDefaultValidSize(t *testing.T) {
	service := NewService(config.Config{RequestTimeout: time.Second}, stubProvider{response: model.GenerateImageResponse{}}, nil)
	_, err := service.TextToImage(context.Background(), model.GenerateImageRequest{Prompt: "test prompt"})
	if err != nil {
		t.Fatalf("expected default size to pass validation, got %v", err)
	}
}
