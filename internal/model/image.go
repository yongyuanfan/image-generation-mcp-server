package model

type GenerateImageRequest struct {
	Prompt         string   `json:"prompt"`
	ImageURL       string   `json:"image_url,omitempty"`
	ImageBase64    string   `json:"image_base64,omitempty"`
	Size           string   `json:"size,omitempty"`
	ResponseFormat string   `json:"response_format,omitempty"`
	Seed           *int64   `json:"seed,omitempty"`
	Watermark      *bool    `json:"watermark,omitempty"`
	GuidanceScale  *float64 `json:"guidance_scale,omitempty"`
	Strength       *float64 `json:"strength,omitempty"`
	NumImages      *int     `json:"num_images,omitempty"`
}

type GenerateImageResponse struct {
	Images    []string `json:"images"`
	RequestID string   `json:"request_id"`
	Model     string   `json:"model"`
	CreatedAt int64    `json:"created_at"`
}

type ModelInfo struct {
	TextToImageModel  string `json:"text_to_image_model"`
	ImageToImageModel string `json:"image_to_image_model"`
}
