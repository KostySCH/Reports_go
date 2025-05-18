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
	// Инициализация клиента MinIO
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка создания клиента MinIO: %v", err)
	}

	// Проверяем и создаем бакеты если нужно
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
		defaultBucket: pdfBucket, // По умолчанию используем PDF бакет
	}, nil
}

// getBucketForFile определяет бакет на основе расширения файла
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

// UploadReport загружает отчет в MinIO и возвращает путь к файлу
func (s *MinioService) UploadReport(ctx context.Context, localPath string, reportID string) (string, error) {
	// Определяем бакет на основе расширения файла
	bucketName := s.getBucketForFile(localPath)

	// Генерируем имя файла в MinIO
	fileName := fmt.Sprintf("reports/%s/%s", time.Now().Format("2006/01/02"), filepath.Base(localPath))

	// Загружаем файл в MinIO
	_, err := s.client.FPutObject(ctx, bucketName, fileName, localPath, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return "", fmt.Errorf("ошибка загрузки файла в MinIO: %v", err)
	}

	// Возвращаем путь к файлу в формате minio://bucket/path
	return fmt.Sprintf("minio://%s/%s", bucketName, fileName), nil
}

// GetReportURL возвращает URL для скачивания отчета
func (s *MinioService) GetReportURL(ctx context.Context, reportPath string) (string, error) {
	// Извлекаем имя файла из пути
	fileName := reportPath[8:] // Убираем "minio://"

	// Определяем бакет из пути
	bucketName := strings.Split(fileName, "/")[0]

	// Генерируем URL для скачивания
	url, err := s.client.PresignedGetObject(ctx, bucketName, strings.TrimPrefix(fileName, bucketName+"/"), time.Hour*24, nil)
	if err != nil {
		return "", fmt.Errorf("ошибка генерации URL для скачивания: %v", err)
	}

	return url.String(), nil
}
