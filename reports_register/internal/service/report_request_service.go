package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/KostySCH/Reports_go/reports_register/internal/model"
	"github.com/KostySCH/Reports_go/reports_register/internal/repository/postgres"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func init() {
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetLevel(logrus.DebugLevel)
}

type ReportRequestService struct {
	repo *postgres.ReportRequestRepository
}

func NewReportRequestService(repo *postgres.ReportRequestRepository) *ReportRequestService {
	return &ReportRequestService{repo: repo}
}

// Create создает новый запрос на отчет
func (s *ReportRequestService) Create(ctx context.Context, userID int, reportType string, params map[string]interface{}) (*model.ReportRequest, error) {
	log.WithFields(logrus.Fields{
		"user_id": userID,
		"type":    reportType,
		"params":  params,
	}).Debug("Creating new report request")

	// Преобразуем параметры в JSON
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		log.WithError(err).Error("Failed to marshal report parameters")
		return nil, err
	}

	// Создаем новый запрос на отчет
	request := &model.ReportRequest{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      reportType,
		Params:    paramsJSON,
		Status:    model.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Сохраняем запрос в базу данных
	if err := s.repo.Create(ctx, request); err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"user_id": userID,
			"type":    reportType,
		}).Error("Failed to save report request to database")
		return nil, err
	}

	log.WithFields(logrus.Fields{
		"request_id": request.ID,
		"status":     request.Status,
	}).Info("Successfully created report request")

	return request, nil
}

// GetPending получает список отложенных отчетов для обработки
func (s *ReportRequestService) GetPending(ctx context.Context, limit int) ([]*model.ReportRequest, error) {
	return s.repo.GetPending(ctx, limit)
}

// UpdateStatus обновляет статус отчета
func (s *ReportRequestService) UpdateStatus(ctx context.Context, id uuid.UUID, status model.ReportStatus) error {
	return s.repo.UpdateStatus(ctx, id, status)
}

// GetStuck получает отчеты, которые зависли в статусе PENDING
func (s *ReportRequestService) GetStuck(ctx context.Context, timeout time.Duration) ([]*model.ReportRequest, error) {
	return s.repo.GetStuck(ctx, timeout)
}
