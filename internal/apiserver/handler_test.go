package apiserver

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"image-generation-mcp-server/internal/config"
	"image-generation-mcp-server/internal/model"
	imagesvc "image-generation-mcp-server/internal/service/image"
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

func TestHandleImageToImageNormalizesSmallSize(t *testing.T) {
	provider := &stubProvider{response: model.GenerateImageResponse{Images: []string{"https://example.com/generated.png"}, CreatedAt: 123}}
	service := imagesvc.NewService(config.Config{RequestTimeout: time.Second}, provider, nil)
	handler := NewHandler(config.Config{}, service)

	body, err := json.Marshal(model.GenerateImageRequest{
		Prompt:   "edit prompt",
		ImageURL: "https://example.com/source.png",
		Size:     "1024x1024",
	})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/images/edits", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if provider.request.Size != "1920x1920" {
		t.Fatalf("expected provider to receive normalized size 1920x1920, got %q", provider.request.Size)
	}
}
