package minio

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"image-generation-mcp-server/internal/config"

	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	client        *minio.Client
	bucket        string
	publicBaseURL string
	autoCreate    bool
}

func NewClient(ctx context.Context, cfg config.Config) (*Client, error) {
	if strings.TrimSpace(cfg.MinIOEndpoint) == "" || strings.TrimSpace(cfg.MinIOAccessKey) == "" || strings.TrimSpace(cfg.MinIOSecretKey) == "" || strings.TrimSpace(cfg.MinIOBucket) == "" || strings.TrimSpace(cfg.MinIOPublicBaseURL) == "" {
		return nil, nil
	}

	client, err := minio.New(cfg.MinIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""),
		Secure: cfg.MinIOUseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("init minio client: %w", err)
	}

	uploader := &Client{
		client:        client,
		bucket:        cfg.MinIOBucket,
		publicBaseURL: cfg.MinIOPublicBaseURL,
		autoCreate:    cfg.MinIOAutoCreateBucket,
	}

	if err := uploader.ensureBucket(ctx); err != nil {
		return nil, err
	}

	return uploader, nil
}

func (c *Client) Upload(ctx context.Context, objectName, contentType string, data []byte) (string, error) {
	if c == nil {
		return "", fmt.Errorf("minio uploader is not configured")
	}

	_, err := c.client.PutObject(ctx, c.bucket, strings.TrimLeft(objectName, "/"), bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("upload object to minio: %w", err)
	}

	return c.publicBaseURL + "/" + c.bucket + "/" + strings.TrimLeft(objectName, "/"), nil
}

func (c *Client) ensureBucket(ctx context.Context) error {
	exists, err := c.client.BucketExists(ctx, c.bucket)
	if err != nil {
		return fmt.Errorf("check minio bucket: %w", err)
	}
	if exists {
		return nil
	}
	if !c.autoCreate {
		return nil
	}
	if err := c.client.MakeBucket(ctx, c.bucket, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("create minio bucket: %w", err)
	}
	return nil
}
