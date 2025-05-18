package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/KostySCH/Reports_go/reports_generator/internal/logger"
	"github.com/KostySCH/Reports_go/reports_generator/internal/models"
	"github.com/KostySCH/Reports_go/reports_generator/internal/repository"
	"github.com/KostySCH/Reports_go/reports_generator/internal/service"
)

const (
	maxRetries = 5
	workerType = "Повторная обработка"
)

type RetryWorker struct {
	repo        *repository.ReportRequestRepository
	reportSvc   *service.ReportService
	pollPeriod  time.Duration
	workerID    int
	concurrency int
}

func NewRetryWorker(repo *repository.ReportRequestRepository, reportSvc *service.ReportService) *RetryWorker {
	return &RetryWorker{
		repo:        repo,
		reportSvc:   reportSvc,
		pollPeriod:  30 * time.Second,
		workerID:    2,  // ID для воркера повторной обработки
		concurrency: 10, // Добавляем параллельность 10
	}
}

func (w *RetryWorker) Start(ctx context.Context) {
	logger.LogWorkerEvent(workerType, w.workerID, fmt.Sprintf("Запуск воркера с параллельностью %d", w.concurrency))

	// Запускаем горутины для обработки запросов
	for i := 0; i < w.concurrency; i++ {
		go w.processFailedReports(ctx)
	}

	// Мониторим контекст для graceful shutdown
	<-ctx.Done()
	logger.LogWorkerEvent(workerType, w.workerID, "Остановка воркера")
}

func (w *RetryWorker) processFailedReports(ctx context.Context) {
	ticker := time.NewTicker(w.pollPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Получаем список неудачных отчетов
			query := `
				SELECT id, user_id, type, params, status, created_at, updated_at, error, retry_count, report_path
				FROM reporting.report_requests
				WHERE status = $1 AND retry_count < $2
				ORDER BY updated_at ASC
				FOR UPDATE SKIP LOCKED
			`
			rows, err := w.repo.DB().QueryContext(ctx, query, models.StatusFailed, maxRetries)
			if err != nil {
				logger.LogWorkerError(workerType, w.workerID, fmt.Errorf("ошибка получения неудачных отчетов: %v", err))
				continue
			}
			defer rows.Close()

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
					logger.LogWorkerError(workerType, w.workerID, fmt.Errorf("ошибка сканирования строки: %v", err))
					continue
				}

				// Увеличиваем счетчик попыток
				if err := w.repo.IncrementRetryCount(ctx, req.ID); err != nil {
					logger.LogWorkerError(workerType, w.workerID, fmt.Errorf("ошибка обновления счетчика попыток для отчета %s: %v", req.ID, err))
					continue
				}

				// Пытаемся сгенерировать отчет заново
				if err := w.processRequest(ctx, &req); err != nil {
					errorMsg := err.Error()
					if err := w.repo.UpdateRequestStatus(ctx, req.ID, models.StatusFailed, &errorMsg, nil); err != nil {
						logger.LogWorkerError(workerType, w.workerID, fmt.Errorf("ошибка обновления статуса для отчета %s: %v", req.ID, err))
					}
					logger.LogWorkerReport(workerType, w.workerID, req.ID.String(), req.RetryCount+1, maxRetries, false, errorMsg)
				} else {
					logger.LogWorkerReport(workerType, w.workerID, req.ID.String(), req.RetryCount+1, maxRetries, true, "")
				}
			}
			logger.LogWorkerSeparator()
		}
	}
}

func (w *RetryWorker) processRequest(ctx context.Context, req *models.ReportRequest) error {
	// Парсим параметры запроса
	var params map[string]interface{}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return fmt.Errorf("ошибка разбора параметров: %v", err)
	}

	// Генерируем отчет в зависимости от типа
	var reportPath string
	var err error

	switch req.Type {
	case "branch_performance_report":
		// Проверяем наличие всех необходимых параметров
		branchID, ok := params["branch_id"]
		if !ok {
			return fmt.Errorf("отсутствует обязательный параметр branch_id")
		}
		month, ok := params["month"]
		if !ok {
			return fmt.Errorf("отсутствует обязательный параметр month")
		}
		format, ok := params["format"]
		if !ok {
			return fmt.Errorf("отсутствует обязательный параметр format")
		}

		// Преобразуем параметры в нужный формат
		branchParams := &models.BranchPerformanceParams{
			BranchID: int64(branchID.(float64)),
			Month:    month.(string),
			Format:   format.(string),
		}
		reportPath, err = w.reportSvc.GenerateBranchPerformanceReport(ctx, branchParams)
	default:
		return fmt.Errorf("неподдерживаемый тип отчета: %s", req.Type)
	}

	if err != nil {
		return fmt.Errorf("ошибка генерации отчета: %v", err)
	}

	// Обновляем статус запроса
	return w.repo.UpdateRequestStatus(ctx, req.ID, models.StatusCompleted, nil, &reportPath)
}
