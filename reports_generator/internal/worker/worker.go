package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/KostySCH/Reports_go/reports_generator/internal/logger"
	"github.com/KostySCH/Reports_go/reports_generator/internal/models"
	"github.com/KostySCH/Reports_go/reports_generator/internal/repository"
	"github.com/KostySCH/Reports_go/reports_generator/internal/service"
)

const (
	mainWorkerType = "Основной"
)

type Worker struct {
	repo        *repository.ReportRequestRepository
	reportSvc   *service.ReportService
	done        chan struct{}
	stopOnce    sync.Once
	concurrency int
	workerID    int
}

func NewWorker(repo *repository.ReportRequestRepository, reportSvc *service.ReportService, concurrency int) *Worker {
	return &Worker{
		repo:        repo,
		reportSvc:   reportSvc,
		done:        make(chan struct{}),
		concurrency: concurrency,
		workerID:    1, // ID для основного воркера
	}
}

func (w *Worker) Start(ctx context.Context) {
	logger.LogWorkerEvent(mainWorkerType, w.workerID, fmt.Sprintf("Запуск воркера с параллельностью %d", w.concurrency))

	// Запускаем горутины для обработки запросов
	for i := 0; i < w.concurrency; i++ {
		go w.processRequests(ctx)
	}

	// Мониторим контекст для graceful shutdown
	go func() {
		<-ctx.Done()
		w.Stop()
	}()
}

func (w *Worker) Stop() {
	w.stopOnce.Do(func() {
		logger.LogWorkerEvent(mainWorkerType, w.workerID, "Остановка воркера")
		close(w.done)
	})
}

func (w *Worker) processRequests(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-w.done:
			return
		default:
			// Получаем запросы на обработку
			requests, err := w.repo.GetPendingRequests(ctx, w.concurrency)
			if err != nil {
				logger.LogWorkerError(mainWorkerType, w.workerID, fmt.Errorf("ошибка получения запросов: %v", err))
				time.Sleep(time.Second)
				continue
			}

			if len(requests) == 0 {
				time.Sleep(time.Second)
				continue
			}

			// Обрабатываем каждый запрос
			for _, req := range requests {
				if err := w.processRequest(ctx, req); err != nil {
					errorMsg := err.Error()
					if err := w.repo.UpdateRequestStatus(ctx, req.ID, models.StatusFailed, &errorMsg, nil); err != nil {
						logger.LogWorkerError(mainWorkerType, w.workerID, fmt.Errorf("ошибка обновления статуса запроса %s: %v", req.ID, err))
					}
					logger.LogWorkerReport(mainWorkerType, w.workerID, req.ID.String(), 1, 1, false, errorMsg)
				} else {
					logger.LogWorkerReport(mainWorkerType, w.workerID, req.ID.String(), 1, 1, true, "")
				}
			}
			logger.LogWorkerSeparator()
		}
	}
}

func (w *Worker) processRequest(ctx context.Context, req *models.ReportRequest) error {
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
