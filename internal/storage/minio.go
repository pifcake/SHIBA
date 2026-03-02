package storage

import (
	"context"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOClient struct {
	client    *minio.Client
	bucket    string
	publicURL string
}

func NewMinIOClient(endpoint, accessKey, secretKey, bucket, publicURL string, useSSL bool) (*MinIOClient, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio init: %w", err)
	}
	return &MinIOClient{
		client:    client,
		bucket:    bucket,
		publicURL: publicURL,
	}, nil
}

func (m *MinIOClient) EnsureBucket(ctx context.Context) error {
	exists, err := m.client.BucketExists(ctx, m.bucket)
	if err != nil {
		return fmt.Errorf("check bucket: %w", err)
	}
	if !exists {
		if err := m.client.MakeBucket(ctx, m.bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("create bucket: %w", err)
		}
		// Set bucket policy to allow public read
		policy := fmt.Sprintf(`{
			"Version":"2012-10-17",
			"Statement":[{
				"Effect":"Allow",
				"Principal":{"AWS":["*"]},
				"Action":["s3:GetObject"],
				"Resource":["arn:aws:s3:::%s/*"]
			}]
		}`, m.bucket)
		if err := m.client.SetBucketPolicy(ctx, m.bucket, policy); err != nil {
			return fmt.Errorf("set bucket policy: %w", err)
		}
	}
	return nil
}

type UploadResult struct {
	StorageKey string
	URL        string
	SizeBytes  int64
	MimeType   string
}

func (m *MinIOClient) UploadPhoto(ctx context.Context, modelID uuid.UUID, filename string, reader io.Reader, size int64, mimeType string) (*UploadResult, error) {
	ext := path.Ext(filename)
	key := fmt.Sprintf("photos/%s/%s%s", modelID.String(), uuid.New().String(), ext)

	info, err := m.client.PutObject(ctx, m.bucket, key, reader, size, minio.PutObjectOptions{
		ContentType: mimeType,
	})
	if err != nil {
		return nil, fmt.Errorf("upload photo: %w", err)
	}

	return &UploadResult{
		StorageKey: key,
		URL:        m.ObjectURL(key),
		SizeBytes:  info.Size,
		MimeType:   mimeType,
	}, nil
}

func (m *MinIOClient) UploadLogo(ctx context.Context, agencyID uuid.UUID, filename string, reader io.Reader, size int64, mimeType string) (*UploadResult, error) {
	ext := path.Ext(filename)
	key := fmt.Sprintf("logos/%s/%s%s", agencyID.String(), uuid.New().String(), ext)

	info, err := m.client.PutObject(ctx, m.bucket, key, reader, size, minio.PutObjectOptions{
		ContentType: mimeType,
	})
	if err != nil {
		return nil, fmt.Errorf("upload logo: %w", err)
	}

	return &UploadResult{
		StorageKey: key,
		URL:        m.ObjectURL(key),
		SizeBytes:  info.Size,
		MimeType:   mimeType,
	}, nil
}

func (m *MinIOClient) DeleteObject(ctx context.Context, key string) error {
	return m.client.RemoveObject(ctx, m.bucket, key, minio.RemoveObjectOptions{})
}

func (m *MinIOClient) ObjectURL(key string) string {
	return fmt.Sprintf("%s/%s/%s", m.publicURL, m.bucket, key)
}

func (m *MinIOClient) PresignedGetURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	u, err := m.client.PresignedGetObject(ctx, m.bucket, key, expiry, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}
