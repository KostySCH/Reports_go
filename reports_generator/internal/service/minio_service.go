package service

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioService struct {
	client        *minio.Client
	pdfBucket     string
	docxBucket    string
	defaultBucket string
}

func NewMinioService(endpoint, accessKey, secretKey string, pdfBucket, docxBucket string, useSSL bool) (*MinioService, error) {

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка создания клиента MinIO: %v", err)
	}

	buckets := []string{pdfBucket, docxBucket}
	for _, bucket := range buckets {
		exists, err := minioClient.BucketExists(context.Background(), bucket)
		if err != nil {
			return nil, fmt.Errorf("ошибка проверки существования бакета %s: %v", bucket, err)
		}

		if !exists {
			err = minioClient.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{})
			if err != nil {
				return nil, fmt.Errorf("ошибка создания бакета %s: %v", bucket, err)
			}
		}
	}

	return &MinioService{
		client:        minioClient,
		pdfBucket:     pdfBucket,
		docxBucket:    docxBucket,
		defaultBucket: pdfBucket,
	}, nil
}

func (s *MinioService) getBucketForFile(fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".pdf":
		return s.pdfBucket
	case ".docx":
		return s.docxBucket
	default:
		return s.defaultBucket
	}
}

func (s *MinioService) UploadReport(ctx context.Context, localPath string, reportID string) (string, error) {

	bucketName := s.getBucketForFile(localPath)

	fileName := fmt.Sprintf("reports/%s/%s", time.Now().Format("2006/01/02"), filepath.Base(localPath))

	_, err := s.client.FPutObject(ctx, bucketName, fileName, localPath, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return "", fmt.Errorf("ошибка загрузки файла в MinIO: %v", err)
	}

	return fmt.Sprintf("minio://%s/%s", bucketName, fileName), nil
}

func (s *MinioService) GetReportURL(ctx context.Context, reportPath string) (string, error) {

	fileName := reportPath[8:]

	bucketName := strings.Split(fileName, "/")[0]

	url, err := s.client.PresignedGetObject(ctx, bucketName, strings.TrimPrefix(fileName, bucketName+"/"), time.Hour*24, nil)
	if err != nil {
		return "", fmt.Errorf("ошибка генерации URL для скачивания: %v", err)
	}

	return url.String(), nil
}
