package ark

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/yong/image-generation-mcp-server/internal/config"
	"github.com/yong/image-generation-mcp-server/internal/model"
)

type Client struct {
	baseURL    string
	endpoint   string
	apiKey     string
	httpClient *http.Client
	textModel  string
	imageModel string
}

type imageGenerationRequest struct {
	Model          string         `json:"model"`
	Prompt         string         `json:"prompt"`
	Size           string         `json:"size,omitempty"`
	ResponseFormat string         `json:"response_format,omitempty"`
	Seed           *int64         `json:"seed,omitempty"`
	Watermark      *bool          `json:"watermark,omitempty"`
	GuidanceScale  *float64       `json:"guidance_scale,omitempty"`
	NumImages      *int           `json:"n,omitempty"`
	Image          string         `json:"image,omitempty"`
	Extra          map[string]any `json:"extra_body,omitempty"`
}

type imageGenerationResponse struct {
	Created int64 `json:"created"`
	Data    []struct {
		URL           string `json:"url"`
		B64JSON       string `json:"b64_json"`
		RevisedPrompt string `json:"revised_prompt"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func NewClient(cfg config.Config) *Client {
	return &Client{
		baseURL:    cfg.ARKBaseURL,
		endpoint:   cfg.ARKImageEndpointPath,
		apiKey:     cfg.ARKAPIKey,
		httpClient: &http.Client{Timeout: cfg.RequestTimeout},
		textModel:  cfg.ARKTextModel,
		imageModel: cfg.ARKImageModel,
	}
}

func (c *Client) TextToImage(ctx context.Context, input model.GenerateImageRequest) (model.GenerateImageResponse, error) {
	return c.generate(ctx, c.textModel, input)
}

func (c *Client) ImageToImage(ctx context.Context, input model.GenerateImageRequest) (model.GenerateImageResponse, error) {
	return c.generate(ctx, c.imageModel, input)
}

func (c *Client) generate(ctx context.Context, modelName string, input model.GenerateImageRequest) (model.GenerateImageResponse, error) {
	requestBody := imageGenerationRequest{
		Model:          modelName,
		Prompt:         input.Prompt,
		Size:           input.Size,
		ResponseFormat: input.ResponseFormat,
		Seed:           input.Seed,
		Watermark:      input.Watermark,
		GuidanceScale:  input.GuidanceScale,
		NumImages:      input.NumImages,
	}

	if input.ImageBase64 != "" {
		requestBody.Image = input.ImageBase64
	}
	if input.ImageURL != "" {
		if requestBody.Extra == nil {
			requestBody.Extra = map[string]any{}
		}
		requestBody.Extra["image_url"] = input.ImageURL
	}
	if input.Strength != nil {
		if requestBody.Extra == nil {
			requestBody.Extra = map[string]any{}
		}
		requestBody.Extra["strength"] = input.Strength
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return model.GenerateImageResponse{}, err
	}

	endpointURL := c.baseURL + c.endpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpointURL, bytes.NewReader(body))
	if err != nil {
		return model.GenerateImageResponse{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return model.GenerateImageResponse{}, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.GenerateImageResponse{}, err
	}

	var parsed imageGenerationResponse
	if err := json.Unmarshal(responseBody, &parsed); err != nil {
		return model.GenerateImageResponse{}, fmt.Errorf("decode ark response: %w", err)
	}

	if resp.StatusCode >= 400 {
		if parsed.Error != nil && parsed.Error.Message != "" {
			return model.GenerateImageResponse{}, fmt.Errorf("ark api error: %s", parsed.Error.Message)
		}
		return model.GenerateImageResponse{}, fmt.Errorf("ark api error: status %d", resp.StatusCode)
	}

	images := make([]string, 0, len(parsed.Data))
	for _, item := range parsed.Data {
		switch {
		case item.URL != "":
			images = append(images, item.URL)
		case item.B64JSON != "":
			images = append(images, normalizeBase64Image(item.B64JSON))
		}
	}

	return model.GenerateImageResponse{
		Images:    images,
		RequestID: resp.Header.Get("X-Tt-Logid"),
		Model:     modelName,
		CreatedAt: parsed.Created,
	}, nil
}

func normalizeBase64Image(value string) string {
	if strings.HasPrefix(value, "data:") {
		return value
	}

	_, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		_, err = base64.RawStdEncoding.DecodeString(value)
	}
	if err != nil {
		return value
	}

	contentType := http.DetectContentType(decodedBytes(value))
	mediaType, _, parseErr := mime.ParseMediaType(contentType)
	if parseErr != nil || mediaType == "" {
		mediaType = "image/png"
	}
	return "data:" + mediaType + ";base64," + value
}

func decodedBytes(value string) []byte {
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err == nil {
		return decoded
	}
	decoded, err = base64.RawStdEncoding.DecodeString(value)
	if err == nil {
		return decoded
	}
	return []byte(value)
}
