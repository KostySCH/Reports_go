package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/KostySCH/Reports_go/reports_register/internal/model"
	"github.com/google/uuid"
	_ "github.com/lib/pq" // PostgreSQL драйвер
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func init() {
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetLevel(logrus.DebugLevel)
}

type ReportRequestRepository struct {
	db *sql.DB
}

func NewReportRequestRepository(db *sql.DB) *ReportRequestRepository {
	return &ReportRequestRepository{db: db}
}

// Create создает новый запрос на отчет
func (r *ReportRequestRepository) Create(ctx context.Context, request *model.ReportRequest) error {
	query := `
		INSERT INTO reporting.report_requests (id, user_id, type, params, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	log.WithFields(logrus.Fields{
		"request_id": request.ID,
		"user_id":    request.UserID,
		"type":       request.Type,
	}).Debug("Executing SQL query to create report request")

	_, err := r.db.ExecContext(ctx, query,
		request.ID,
		request.UserID,
		request.Type,
		request.Params,
		request.Status,
		request.CreatedAt,
		request.UpdatedAt,
	)

	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"request_id": request.ID,
			"user_id":    request.UserID,
			"query":      query,
		}).Error("Failed to execute SQL query")
		return err
	}

	log.WithFields(logrus.Fields{
		"request_id": request.ID,
		"user_id":    request.UserID,
	}).Info("Successfully created report request in database")

	return nil
}

// GetPending получает список отложенных отчетов для обработки
func (r *ReportRequestRepository) GetPending(ctx context.Context, limit int) ([]*model.ReportRequest, error) {
	query := `
		SELECT id, user_id, status, type, params, created_at, updated_at
		FROM reporting.report_requests
		WHERE status = $1
		ORDER BY created_at ASC
		LIMIT $2
		FOR UPDATE SKIP LOCKED
	`

	rows, err := r.db.QueryContext(ctx, query, model.StatusPending, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*model.ReportRequest
	for rows.Next() {
		request := &model.ReportRequest{}
		err := rows.Scan(
			&request.ID,
			&request.UserID,
			&request.Status,
			&request.Type,
			&request.Params,
			&request.CreatedAt,
			&request.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}

	return requests, nil
}

// UpdateStatus обновляет статус отчета
func (r *ReportRequestRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.ReportStatus) error {
	query := `
		UPDATE reporting.report_requests
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	return err
}

// GetStuck получает отчеты, которые зависли в статусе PENDING
func (r *ReportRequestRepository) GetStuck(ctx context.Context, timeout time.Duration) ([]*model.ReportRequest, error) {
	query := `
		SELECT id, user_id, status, type, params, created_at, updated_at
		FROM reporting.report_requests
		WHERE status = $1 AND updated_at < $2
		ORDER BY updated_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, model.StatusPending, time.Now().Add(-timeout))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*model.ReportRequest
	for rows.Next() {
		request := &model.ReportRequest{}
		err := rows.Scan(
			&request.ID,
			&request.UserID,
			&request.Status,
			&request.Type,
			&request.Params,
			&request.CreatedAt,
			&request.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}

	return requests, nil
}
