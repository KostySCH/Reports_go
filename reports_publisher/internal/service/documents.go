package service

import (
	"context"
	"io"

	storage "github.com/KostySCH/Reports_go/reports_publisher/internal/repository/minio"
	"github.com/KostySCH/Reports_go/reports_publisher/pkg/types"
)

type DocumentService struct {
	storage *storage.Client
}

func New(storage *storage.Client) *DocumentService {
	return &DocumentService{storage: storage}
}

func (s *DocumentService) GetAvailablePDFs(ctx context.Context) ([]types.PDFDocument, error) {
	docs, err := s.getAvailableFiles(ctx, storage.PDFBucket)
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
	docs, err := s.getAvailableFiles(ctx, storage.DOCXBucket)
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
	files, err := s.storage.ListFiles(ctx, bucket)
	if err != nil {
		return nil, err
	}

	var docs []types.Document
	for _, file := range files {
		docs = append(docs, types.Document{
			Name:     file.Name,
			SizeMB:   float64(file.Size) / (1024 * 1024),
			Modified: file.LastModified,
		})
	}
	return docs, nil
}

func (s *DocumentService) DownloadFile(ctx context.Context, fileType, fileName string) (io.ReadCloser, error) {
	bucket := storage.PDFBucket
	if fileType == "docx" {
		bucket = storage.DOCXBucket
	}
	return s.storage.DownloadFile(ctx, bucket, fileName)
}
