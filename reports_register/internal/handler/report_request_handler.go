package handler

import (
	"encoding/json"
	"net/http"

	"github.com/KostySCH/Reports_go/reports_register/internal/service"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func init() {
	// Настройка форматирования логов
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetLevel(logrus.DebugLevel)
}

type ReportRequestHandler struct {
	service *service.ReportRequestService
}

func NewReportRequestHandler(service *service.ReportRequestService) *ReportRequestHandler {
	return &ReportRequestHandler{service: service}
}

// ReportParams представляет параметры отчета
type ReportParams map[string]interface{}

// CreateReportRequest представляет запрос на создание отчета
type CreateReportRequest struct {
	UserID int          `json:"user_id" example:"123" binding:"required"`
	Type   string       `json:"type" example:"daily_report" binding:"required"`
	Params ReportParams `json:"params" binding:"required"`
}

// @Summary Создать новый запрос на отчет
// @Description Создает новый запрос на генерацию отчета
// @Tags reports
// @Accept json
// @Produce json
// @Param request body CreateReportRequest true "Параметры запроса"
// @Success 200 {object} model.ReportRequest
// @Failure 400 {string} string "Invalid request body"
// @Failure 500 {string} string "Failed to create report request"
// @Router /api/reports [post]
func (h *ReportRequestHandler) Create(w http.ResponseWriter, r *http.Request) {
	log.WithFields(logrus.Fields{
		"method": r.Method,
		"path":   r.URL.Path,
	}).Info("Received request to create report")

	var req CreateReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WithError(err).Error("Failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.WithFields(logrus.Fields{
		"user_id": req.UserID,
		"type":    req.Type,
		"params":  req.Params,
	}).Debug("Parsed request parameters")

	report, err := h.service.Create(r.Context(), req.UserID, req.Type, req.Params)
	if err != nil {
		log.WithError(err).WithFields(logrus.Fields{
			"user_id": req.UserID,
			"type":    req.Type,
		}).Error("Failed to create report request")
		http.Error(w, "Failed to create report request", http.StatusInternalServerError)
		return
	}

	log.WithFields(logrus.Fields{
		"report_id": report.ID,
		"status":    report.Status,
	}).Info("Successfully created report request")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// @Summary Получить статус отчета
// @Description Получает текущий статус запроса на отчет
// @Tags reports
// @Accept json
// @Produce json
// @Param id path string true "ID отчета"
// @Success 200 {object} model.ReportRequest
// @Failure 404 {string} string "Report not found"
// @Router /api/reports/{id}/status [get]
func (h *ReportRequestHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement get report request status endpoint
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}
