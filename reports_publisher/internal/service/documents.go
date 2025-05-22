package service

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"

	"github.com/KostySCH/Reports_go/reports_publisher/pkg/types"
)

type DocumentService struct {
	minioSvc *MinioService
	db       *sql.DB
}

func New(minioSvc *MinioService, db *sql.DB) *DocumentService {
	return &DocumentService{
		minioSvc: minioSvc,
		db:       db,
	}
}

// GetReportPath получает путь к файлу из базы данных
func (s *DocumentService) GetReportPath(ctx context.Context, reportID string, reportPath *string) error {
	err := s.db.QueryRowContext(ctx, `
		SELECT report_path 
		FROM reporting.report_requests
		WHERE id = $1
	`, reportID).Scan(reportPath)
	if err != nil {
		log.Printf("Ошибка получения пути к файлу из БД: %v", err)
		return fmt.Errorf("ошибка получения пути к файлу: %v", err)
	}
	return nil
}

func (s *DocumentService) GetReport(ctx context.Context, reportID string) (io.Reader, error) {
	// Получаем путь к файлу из базы данных
	var reportPath string
	err := s.db.QueryRowContext(ctx, `
		SELECT report_path 
		FROM reporting.report_requests
		WHERE id = $1
	`, reportID).Scan(&reportPath)
	if err != nil {
		log.Printf("Ошибка получения пути к файлу из БД: %v", err)
		return nil, fmt.Errorf("ошибка получения пути к файлу: %v", err)
	}

	log.Printf("Получен путь к файлу из БД: %s", reportPath)

	// Извлекаем путь к файлу из формата minio://bucket/path
	parts := strings.Split(reportPath, "://")
	if len(parts) != 2 {
		return nil, fmt.Errorf("неверный формат пути к файлу: %s", reportPath)
	}

	// Разделяем путь на бакет и путь к файлу
	bucketAndPath := strings.SplitN(parts[1], "/", 2)
	if len(bucketAndPath) != 2 {
		return nil, fmt.Errorf("неверный формат пути к файлу: %s", parts[1])
	}

	bucket := bucketAndPath[0]
	filePath := bucketAndPath[1]
	log.Printf("Извлечен путь к файлу: %s из бакета: %s", filePath, bucket)

	// Определяем тип контента на основе бакета и расширения файла
	contentType := "application/pdf"
	if bucket == s.minioSvc.docxBucket || strings.HasSuffix(strings.ToLower(filePath), ".docx") {
		contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	}
	log.Printf("Определен тип контента: %s для файла с путем: %s из бакета: %s", contentType, filePath, bucket)

	reader, err := s.minioSvc.GetFile(ctx, filePath, contentType)
	if err != nil {
		log.Printf("Ошибка получения файла из MinIO: %v", err)
		return nil, fmt.Errorf("ошибка получения файла: %v", err)
	}

	log.Printf("Файл успешно получен из MinIO")
	return reader, nil
}

func (s *DocumentService) GetReportURL(ctx context.Context, reportID string, format string) (string, error) {
	filePath := filepath.Join(reportID, fmt.Sprintf("report.%s", format))
	contentType := "application/pdf"
	if format == "docx" {
		contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	}

	return s.minioSvc.GetPresignedURL(ctx, filePath, contentType)
}

func (s *DocumentService) GetAvailablePDFs(ctx context.Context) ([]types.PDFDocument, error) {
	docs, err := s.getAvailableFiles(ctx, s.minioSvc.pdfBucket)
	if err != nil {
		return nil, err
	}

	pdfs := make([]types.PDFDocument, len(docs))
	for i, doc := range docs {
		pdfs[i] = types.PDFDocument{Document: doc}
	}
	return pdfs, nil
}

func (s *DocumentService) GetAvailableDOCXs(ctx context.Context) ([]types.DOCXDocument, error) {
	docs, err := s.getAvailableFiles(ctx, s.minioSvc.docxBucket)
	if err != nil {
		return nil, err
	}

	docxs := make([]types.DOCXDocument, len(docs))
	for i, doc := range docs {
		docxs[i] = types.DOCXDocument{
			Document: doc,
			Pages:    0,
		}
	}
	return docxs, nil
}

func (s *DocumentService) getAvailableFiles(ctx context.Context, bucket string) ([]types.Document, error) {
	// TODO: Реализовать получение списка файлов из MinIO
	return nil, fmt.Errorf("метод не реализован")
}

func (s *DocumentService) DownloadFile(ctx context.Context, fileType, fileName string) (io.ReadCloser, error) {
	contentType := "application/pdf"
	if fileType == "docx" {
		contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	}

	reader, err := s.minioSvc.GetFile(ctx, fileName, contentType)
	if err != nil {
		return nil, err
	}

	return io.NopCloser(reader), nil
}
