package image

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"
)

type imageAsset struct {
	data        []byte
	contentType string
	ext         string
}

func (s *Service) resolveImageAsset(ctx context.Context, value string) (imageAsset, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return imageAsset{}, fmt.Errorf("image content is empty")
	}

	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		return s.downloadImageAsset(ctx, trimmed)
	}

	return decodeInlineImage(trimmed)
}

func (s *Service) downloadImageAsset(ctx context.Context, imageURL string) (imageAsset, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return imageAsset{}, err
	}

	response, err := s.httpClient.Do(request)
	if err != nil {
		return imageAsset{}, err
	}
	defer response.Body.Close()

	if response.StatusCode >= http.StatusBadRequest {
		return imageAsset{}, fmt.Errorf("download image: status %d", response.StatusCode)
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return imageAsset{}, err
	}

	contentType := strings.TrimSpace(response.Header.Get("Content-Type"))
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}

	return imageAsset{
		data:        data,
		contentType: normalizeMediaType(contentType),
		ext:         extensionFromContentType(contentType),
	}, nil
}

func decodeInlineImage(value string) (imageAsset, error) {
	contentType := ""
	encoded := value

	if strings.HasPrefix(value, "data:") {
		header, payload, ok := strings.Cut(value, ",")
		if !ok {
			return imageAsset{}, fmt.Errorf("invalid data url")
		}
		encoded = payload
		mediaType, _, ok := strings.Cut(header[5:], ";")
		if ok {
			contentType = mediaType
		}
	}

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(encoded)
	}
	if err != nil {
		return imageAsset{}, fmt.Errorf("decode image: %w", err)
	}

	if contentType == "" {
		contentType = http.DetectContentType(decoded)
	}

	return imageAsset{
		data:        decoded,
		contentType: normalizeMediaType(contentType),
		ext:         extensionFromContentType(contentType),
	}, nil
}

func normalizeMediaType(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "image/png"
	}
	mediaType, _, ok := strings.Cut(trimmed, ";")
	if !ok {
		return trimmed
	}
	return strings.TrimSpace(mediaType)
}

func extensionFromContentType(contentType string) string {
	switch normalizeMediaType(contentType) {
	case "image/jpeg":
		return ".jpg"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ".png"
	}
}

func buildObjectName(prefix, requestID string, index int, ext string, now time.Time) string {
	if ext == "" {
		ext = ".png"
	}
	objectID := strings.TrimSpace(requestID)
	if objectID == "" {
		objectID = fmt.Sprintf("generated-%d", now.UnixNano())
	}
	datePath := now.UTC().Format("2006/01/02")
	base := fmt.Sprintf("%s-%d%s", objectID, index, ext)
	if prefix == "" {
		return path.Join(datePath, base)
	}
	return path.Join(prefix, datePath, base)
}
