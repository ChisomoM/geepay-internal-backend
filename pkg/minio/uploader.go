package miniouploader

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinioUploader implements global.FileUploader using MinIO.
type MinioUploader struct {
	client    *minio.Client
	publicURL string
}

// New creates a MinioUploader. endpoint should be host:port without scheme.
func New(endpoint, accessKey, secretKey, publicURL string, useSSL bool) (*MinioUploader, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio client init: %w", err)
	}
	return &MinioUploader{client: client, publicURL: publicURL}, nil
}

func (u *MinioUploader) Upload(bucketName, objectName string, data []byte) (string, error) {
	ctx := context.Background()
	reader := bytes.NewReader(data)
	_, err := u.client.PutObject(ctx, bucketName, objectName, reader, int64(len(data)),
		minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		return "", fmt.Errorf("minio upload: %w", err)
	}
	publicBase := u.publicURL
	if publicBase == "" {
		publicBase = u.client.EndpointURL().String()
	}
	return fmt.Sprintf("%s/%s/%s", publicBase, bucketName, objectName), nil
}

func (u *MinioUploader) Download(bucketName, objectName string) ([]byte, error) {
	ctx := context.Background()
	obj, err := u.client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("minio download: %w", err)
	}
	defer obj.Close()
	return io.ReadAll(obj)
}

func (u *MinioUploader) Delete(bucketName, objectName string) error {
	ctx := context.Background()
	if err := u.client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("minio delete: %w", err)
	}
	return nil
}

func (u *MinioUploader) GeneratePresignedURL(bucketName, objectName string, expirySeconds int) (string, error) {
	ctx := context.Background()
	expiry := time.Duration(expirySeconds) * time.Second
	presignedURL, err := u.client.PresignedGetObject(ctx, bucketName, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("minio presign: %w", err)
	}
	return presignedURL.String(), nil
}
