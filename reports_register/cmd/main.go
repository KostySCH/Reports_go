package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/KostySCH/Reports_go/reports_register/docs"
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

	cfg := config.Load()

	db, _ := sql.Open("postgres", cfg.GetDSN())
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	repo := postgres.NewReportRequestRepository(db)
	svc := service.NewReportRequestService(repo)
	h := handler.NewReportRequestHandler(svc)

	r := mux.NewRouter()

	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/reports", h.Create).Methods("POST")
	api.HandleFunc("/reports/{id}/status", h.GetStatus).Methods("GET")

	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	log.Printf("Starting server on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
