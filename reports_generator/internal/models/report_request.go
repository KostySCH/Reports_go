package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	StatusPending    = "PENDING"
	StatusInProgress = "IN_PROGRESS"
	StatusCompleted  = "COMPLETED"
	StatusFailed     = "FAILED"
)

type JSON json.RawMessage

func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.RawMessage(j).MarshalJSON()
}

func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	s, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("invalid scan source")
	}
	*j = append((*j)[0:0], s...)
	return nil
}

func (j JSON) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return json.RawMessage(j).MarshalJSON()
}

func (j *JSON) UnmarshalJSON(data []byte) error {
	if j == nil {
		return fmt.Errorf("null point exception")
	}
	*j = append((*j)[0:0], data...)
	return nil
}

type BranchPerformanceParams struct {
	BranchID int64  `json:"branch_id"`
	Month    string `json:"month"`
	Format   string `json:"format"`
}

type ReportRequest struct {
	ID         uuid.UUID      `json:"id"`
	UserID     int64          `json:"user_id"`
	Type       string         `json:"type"`
	Params     []byte         `json:"params"`
	Status     string         `json:"status"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	Error      sql.NullString `json:"error"`
	RetryCount int            `json:"retry_count"`
	ReportPath sql.NullString `json:"report_path"`
}
