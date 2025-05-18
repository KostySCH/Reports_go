package model

import (
	"time"

	"github.com/google/uuid"
)

// ReportStatus представляет возможные статусы отчета
type ReportStatus string

const (
	StatusPending    ReportStatus = "PENDING"
	StatusInProgress ReportStatus = "IN_PROGRESS"
	StatusCompleted  ReportStatus = "COMPLETED"
	StatusFailed     ReportStatus = "FAILED"
)

// ReportRequest представляет запрос на генерацию отчета
type ReportRequest struct {
	ID        uuid.UUID    `json:"id" db:"id"`
	UserID    int          `json:"user_id" db:"user_id"`
	Type      string       `json:"type" db:"type"`
	Params    []byte       `json:"params" db:"params"`
	Status    ReportStatus `json:"status" db:"status"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at"`
}

// NewReportRequest создает новый запрос на отчет
func NewReportRequest(userID uuid.UUID, reportType string, params []byte) *ReportRequest {
	now := time.Now()
	return &ReportRequest{
		ID:        uuid.New(),
		UserID:    int(userID.ID()),
		Status:    StatusPending,
		Type:      reportType,
		Params:    params,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
