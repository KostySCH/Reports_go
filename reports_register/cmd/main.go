package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/KostySCH/Reports_go/reports_register/docs" // swagger docs
	"github.com/KostySCH/Reports_go/reports_register/internal/config"
	"github.com/KostySCH/Reports_go/reports_register/internal/handler"
	"github.com/KostySCH/Reports_go/reports_register/internal/repository/postgres"
	"github.com/KostySCH/Reports_go/reports_register/internal/service"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Reports API
// @version 1.0
// @description API для генерации отчетов
// @host localhost:8080
// @BasePath /api
func main() {
	// Загрузка конфигурации
	cfg := config.Load()

	// Подключение к базе данных
	db, err := sql.Open("postgres", cfg.GetDSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Проверка подключения
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Инициализация компонентов
	repo := postgres.NewReportRequestRepository(db)
	svc := service.NewReportRequestService(repo)
	h := handler.NewReportRequestHandler(svc)

	// Настройка маршрутизации
	r := mux.NewRouter()

	// API endpoints
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/reports", h.Create).Methods("POST")
	api.HandleFunc("/reports/{id}/status", h.GetStatus).Methods("GET")

	// Swagger
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	// Запуск сервера
	log.Printf("Starting server on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
