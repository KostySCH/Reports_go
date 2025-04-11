package storage

import (
	"context"

	"github.com/KostySCH/Reports_go/reports_publisher/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	PDFBucket  = "pdf-documents"
	DOCXBucket = "docx-documents"
)

type Client struct {
	client *minio.Client
}

func New(ctx context.Context, cfg *config.MinioConfig) (*Client, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}
	return &Client{client: client}, nil
}

func (c *Client) SetupBuckets(ctx context.Context) error {
	buckets := []string{PDFBucket, DOCXBucket}
	for _, bucket := range buckets {
		exists, err := c.client.BucketExists(ctx, bucket)
		if err != nil {
			return err
		}
		if !exists {
			if err := c.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Client) DownloadFile(ctx context.Context, bucket, name string) (*minio.Object, error) {
	return c.client.GetObject(ctx, bucket, name, minio.GetObjectOptions{})
}

func (c *Client) ListFiles(ctx context.Context, bucket string) ([]FileMeta, error) {
	var files []FileMeta
	for obj := range c.client.ListObjects(ctx, bucket, minio.ListObjectsOptions{}) {
		if obj.Err != nil {
			return nil, obj.Err
		}
		files = append(files, FileMeta{
			Name:         obj.Key,
			Size:         obj.Size,
			LastModified: obj.LastModified,
			ContentType:  obj.ContentType,
		})
	}
	return files, nil
}
