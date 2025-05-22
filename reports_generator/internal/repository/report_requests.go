package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/KostySCH/Reports_go/reports_generator/internal/models"
	"github.com/google/uuid"
)

type ReportRequestRepository struct {
	db *sql.DB
}

func NewReportRequestRepository(db *sql.DB) *ReportRequestRepository {
	return &ReportRequestRepository{db: db}
}

func (r *ReportRequestRepository) DB() *sql.DB {
	return r.db
}

func (r *ReportRequestRepository) GetPendingRequests(ctx context.Context, limit int) ([]*models.ReportRequest, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
		SELECT id, user_id, type, params, status, created_at, updated_at, error, retry_count, report_path
		FROM reporting.report_requests
		WHERE status = $1
		ORDER BY created_at ASC
		LIMIT $2
		FOR UPDATE SKIP LOCKED
	`
	rows, err := tx.QueryContext(ctx, query, models.StatusPending, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*models.ReportRequest
	for rows.Next() {
		var req models.ReportRequest
		err := rows.Scan(
			&req.ID,
			&req.UserID,
			&req.Type,
			&req.Params,
			&req.Status,
			&req.CreatedAt,
			&req.UpdatedAt,
			&req.Error,
			&req.RetryCount,
			&req.ReportPath,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, &req)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	if len(requests) > 0 {
		ids := make([]uuid.UUID, len(requests))
		for i, req := range requests {
			ids[i] = req.ID
		}

		placeholders := make([]string, len(ids))
		args := make([]interface{}, len(ids)+1)
		for i := range ids {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
			args[i] = ids[i]
		}
		args[len(ids)] = models.StatusInProgress

		updateQuery := fmt.Sprintf(`
			UPDATE reporting.report_requests
			SET status = $%d, updated_at = $%d
			WHERE id IN (%s)
		`, len(ids)+1, len(ids)+2, strings.Join(placeholders, ","))

		args = append(args, time.Now())

		_, err = tx.ExecContext(ctx, updateQuery, args...)
		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return requests, nil
}

func (r *ReportRequestRepository) UpdateRequestStatus(ctx context.Context, id uuid.UUID, status string, errorMsg *string, reportPath *string) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		SELECT status 
		FROM reporting.report_requests 
		WHERE id = $1 
		FOR UPDATE SKIP LOCKED
	`
	var currentStatus string
	err = tx.QueryRowContext(ctx, query, id).Scan(&currentStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("запрос не найден: %s", id)
		}
		return err
	}

	if currentStatus == models.StatusCompleted || currentStatus == models.StatusFailed {
		return nil
	}

	updateQuery := `
		UPDATE reporting.report_requests
		SET status = $1, error = $2, report_path = $3, updated_at = $4
		WHERE id = $5
	`
	_, err = tx.ExecContext(ctx, updateQuery, status, errorMsg, reportPath, time.Now(), id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *ReportRequestRepository) IncrementRetryCount(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE reporting.report_requests
		SET retry_count = retry_count + 1, updated_at = $1
		WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}
