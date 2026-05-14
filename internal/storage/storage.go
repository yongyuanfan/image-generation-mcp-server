package storage

import "context"

type Uploader interface {
	Upload(ctx context.Context, objectName, contentType string, data []byte) (string, error)
}
