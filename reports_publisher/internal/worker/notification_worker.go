package worker

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/KostySCH/Reports_go/reports_publisher/internal/service"
)

const (
	workerType = "Уведомления"
)

type NotificationWorker struct {
	db          *sql.DB
	kafkaSvc    *service.KafkaService
	pollPeriod  time.Duration
	workerID    int
	concurrency int
}

func NewNotificationWorker(db *sql.DB, kafkaSvc *service.KafkaService) *NotificationWorker {
	return &NotificationWorker{
		db:          db,
		kafkaSvc:    kafkaSvc,
		pollPeriod:  5 * time.Second,
		workerID:    3,
		concurrency: 10,
	}
}

func (w *NotificationWorker) Start(ctx context.Context) {
	fmt.Printf("[Воркер %d - %s] Запуск\n", w.workerID, workerType)

	for i := 0; i < w.concurrency; i++ {
		go w.processNotifications(ctx)
	}

	<-ctx.Done()
	fmt.Printf("[Воркер %d - %s] Остановка\n", w.workerID, workerType)
}

func (w *NotificationWorker) processNotifications(ctx context.Context) {
	ticker := time.NewTicker(w.pollPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.db.PingContext(ctx); err != nil {
				fmt.Printf("[Воркер %d - %s] Ошибка БД: %v\n", w.workerID, workerType, err)
				time.Sleep(time.Second)
				continue
			}

			tx, err := w.db.BeginTx(ctx, &sql.TxOptions{
				Isolation: sql.LevelReadCommitted,
			})
			if err != nil {
				continue
			}

			query := `
				SELECT id, status, error, user_id
				FROM reporting.report_requests
				WHERE status IN ($1, $2)
				AND notification_sent = false
				ORDER BY updated_at ASC
				LIMIT 10
				FOR UPDATE SKIP LOCKED
			`
			rows, err := tx.QueryContext(ctx, query, "COMPLETED", "FAILED")
			if err != nil {
				tx.Rollback()
				continue
			}

			var notifications []struct {
				id     string
				userID string
				status string
				error  sql.NullString
			}

			for rows.Next() {
				var n struct {
					id     string
					userID string
					status string
					error  sql.NullString
				}
				if err := rows.Scan(&n.id, &n.status, &n.error, &n.userID); err != nil {
					continue
				}
				notifications = append(notifications, n)
			}

			if err = rows.Err(); err != nil {
				rows.Close()
				tx.Rollback()
				continue
			}
			rows.Close()

			if len(notifications) == 0 {
				tx.Rollback()
				continue
			}

			for _, n := range notifications {
				notification := service.ReportNotification{
					ReportID: n.id,
					UserID:   n.userID,
					Status:   n.status,
				}
				if n.error.Valid {
					notification.Error = n.error.String
				}

				if err := w.kafkaSvc.SendNotification(ctx, notification); err != nil {
					fmt.Printf("[Воркер %d - %s] Ошибка отправки уведомления: %v\n", w.workerID, workerType, err)
					continue
				}

				updateQuery := `
					UPDATE reporting.report_requests
					SET notification_sent = true
					WHERE id = $1
				`
				if _, err := tx.ExecContext(ctx, updateQuery, n.id); err != nil {
					fmt.Printf("[Воркер %d - %s] Ошибка обновления статуса: %v\n", w.workerID, workerType, err)
					continue
				}
			}

			if err := tx.Commit(); err != nil {
				fmt.Printf("[Воркер %d - %s] Ошибка коммита: %v\n", w.workerID, workerType, err)
				continue
			}
		}
	}
}
