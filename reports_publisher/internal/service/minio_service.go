package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioService struct {
	client     *minio.Client
	pdfBucket  string
	docxBucket string
}

func NewMinioService(endpoint, accessKey, secretKey, pdfBucket, docxBucket string, useSSL bool) (*MinioService, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка создания MinIO клиента: %v", err)
	}

	// Проверяем существование бакетов
	ctx := context.Background()
	buckets := []string{pdfBucket, docxBucket}
	for _, bucket := range buckets {
		exists, err := client.BucketExists(ctx, bucket)
		if err != nil {
			return nil, fmt.Errorf("ошибка проверки бакета %s: %v", bucket, err)
		}
		if !exists {
			err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
			if err != nil {
				return nil, fmt.Errorf("ошибка создания бакета %s: %v", bucket, err)
			}
		}
	}

	return &MinioService{
		client:     client,
		pdfBucket:  pdfBucket,
		docxBucket: docxBucket,
	}, nil
}

func (s *MinioService) UploadFile(ctx context.Context, filePath string, contentType string, reader io.Reader) error {
	bucket := s.pdfBucket
	if contentType == "application/vnd.openxmlformats-officedocument.wordprocessingml.document" {
		bucket = s.docxBucket
	}

	_, err := s.client.PutObject(ctx, bucket, filePath, reader, -1, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("ошибка загрузки файла в MinIO: %v", err)
	}

	return nil
}

func (s *MinioService) GetFile(ctx context.Context, filePath string, contentType string) (io.Reader, error) {
	// Определяем бакет на основе типа контента
	bucket := s.pdfBucket
	if contentType == "application/vnd.openxmlformats-officedocument.wordprocessingml.document" {
		bucket = s.docxBucket
		log.Printf("Используем бакет для DOCX: %s", bucket)
	} else {
		log.Printf("Используем бакет для PDF: %s", bucket)
	}

	log.Printf("Попытка получить файл из бакета %s по пути %s", bucket, filePath)

	// Проверяем существование файла
	stat, err := s.client.StatObject(ctx, bucket, filePath, minio.StatObjectOptions{})
	if err != nil {
		log.Printf("Ошибка при проверке существования файла: %v", err)
		return nil, fmt.Errorf("файл не найден в MinIO: %v", err)
	}
	log.Printf("Файл найден в MinIO, размер: %d байт, тип контента: %s", stat.Size, stat.ContentType)

	// Получаем объект с правильными опциями
	opts := minio.GetObjectOptions{}

	// Устанавливаем тип контента
	opts.Set("Content-Type", contentType)
	log.Printf("Установлен тип контента: %s", contentType)

	object, err := s.client.GetObject(ctx, bucket, filePath, opts)
	if err != nil {
		log.Printf("Ошибка при получении файла: %v", err)
		return nil, fmt.Errorf("ошибка получения файла из MinIO: %v", err)
	}
	log.Printf("Файл успешно получен из MinIO")

	return object, nil
}

func (s *MinioService) GetPresignedURL(ctx context.Context, filePath string, contentType string) (string, error) {
	bucket := s.pdfBucket
	if contentType == "application/vnd.openxmlformats-officedocument.wordprocessingml.document" {
		bucket = s.docxBucket
	}

	url, err := s.client.PresignedGetObject(ctx, bucket, filePath, time.Hour*24, nil)
	if err != nil {
		return "", fmt.Errorf("ошибка получения presigned URL: %v", err)
	}

	return url.String(), nil
}
