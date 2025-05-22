package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KostySCH/Reports_go/reports_generator/internal/config"
	"github.com/KostySCH/Reports_go/reports_generator/internal/repository"
	"github.com/KostySCH/Reports_go/reports_generator/internal/service"
	"github.com/KostySCH/Reports_go/reports_generator/internal/worker"
	_ "github.com/lib/pq" // PostgreSQL драйвер
)

func main() {
	// Загружаем конфигурацию
	cfg := config.Load()

	// Инициализируем подключение к базе данных
	db, err := sql.Open("postgres", cfg.GetDSN())
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()

	// Инициализируем MinIO сервис
	minioSvc, err := service.NewMinioService(
		cfg.MinIO.Endpoint,
		cfg.MinIO.AccessKey,
		cfg.MinIO.SecretKey,
		cfg.MinIO.PDFBucket,
		cfg.MinIO.DOCXBucket,
		cfg.MinIO.UseSSL,
	)
	if err != nil {
		log.Fatalf("Ошибка инициализации MinIO: %v", err)
	}

	// Инициализируем репозиторий и сервисы
	repo := repository.NewReportRequestRepository(db)
	reportSvc := service.NewReportService(db, "output", minioSvc)

	// Создаем и запускаем воркеры
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mainWorker := worker.NewWorker(repo, reportSvc, 10)
	retryWorker := worker.NewRetryWorker(repo, reportSvc)

	mainWorker.Start(ctx)
	retryWorker.Start(ctx)

	// Ждем сигнала для завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Получен сигнал завершения, останавливаем воркеры...")
	cancel()
	time.Sleep(time.Second) // Даем время на graceful shutdown
}
