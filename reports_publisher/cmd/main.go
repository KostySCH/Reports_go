package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KostySCH/Reports_go/reports_publisher/internal/config"
	handler "github.com/KostySCH/Reports_go/reports_publisher/internal/handler"
	"github.com/KostySCH/Reports_go/reports_publisher/internal/service"
	"github.com/KostySCH/Reports_go/reports_publisher/internal/worker"
	_ "github.com/lib/pq"
)

func main() {
	log.Println("Запуск приложения reports_publisher...")

	// Загружаем конфигурацию
	cfg, err := config.LoadConfig("../config/config.yaml")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}
	log.Println("Конфигурация успешно загружена")

	// Формируем строку подключения к БД
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	// Инициализируем подключение к базе данных
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()
	log.Println("Подключение к базе данных установлено")

	// Инициализируем MinIO сервис
	minioSvc, err := service.NewMinioService(
		cfg.Minio.Endpoint,
		cfg.Minio.AccessKey,
		cfg.Minio.SecretKey,
		cfg.Minio.PDFBucket,
		cfg.Minio.DOCXBucket,
		cfg.Minio.UseSSL,
	)
	if err != nil {
		log.Fatalf("Ошибка инициализации MinIO: %v", err)
	}
	log.Println("MinIO сервис инициализирован")

	// Инициализируем Kafka сервис
	kafkaSvc, err := service.NewKafkaService(cfg)
	if err != nil {
		log.Fatalf("Ошибка инициализации Kafka: %v", err)
	}
	defer kafkaSvc.Close()
	log.Println("Kafka сервис инициализирован")

	// Инициализация сервиса документов
	docSvc := service.New(minioSvc, db)
	log.Println("Сервис документов инициализирован")

	// Инициализация HTTP сервера
	handlers := handler.NewHandler(docSvc)
	router := handlers.InitRoutes()
	log.Println("Маршруты инициализированы")

	// Создаем HTTP сервер
	httpServer := &http.Server{
		Addr:    ":8082",
		Handler: router,
	}

	// Создаем канал для сигналов завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем сервер в отдельной горутине
	log.Printf("Запуск HTTP сервера на порту 8082...")
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка запуска HTTP сервера: %v", err)
		}
	}()

	// Даем серверу время на запуск
	time.Sleep(2 * time.Second)

	// Проверяем доступность сервера
	log.Println("Проверка доступности сервера...")
	resp, err := http.Get("http://localhost:8082/ping")
	if err != nil {
		log.Printf("Ошибка при проверке доступности сервера: %v", err)
	} else {
		defer resp.Body.Close()
		log.Printf("Сервер отвечает, статус: %d", resp.StatusCode)
	}

	log.Println("Сервер запущен и готов к работе")

	// Создаем и запускаем воркеры
	log.Println("Запуск воркеров...")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	notificationWorker := worker.NewNotificationWorker(db, kafkaSvc)
	go func() {
		log.Println("Запуск воркера уведомлений...")
		notificationWorker.Start(ctx)
	}()
	log.Println("Воркер уведомлений запущен")

	log.Println("Приложение полностью инициализировано и готово к работе")
	log.Println("Ожидание сигнала для завершения...")
	<-sigChan

	log.Println("Получен сигнал завершения, останавливаем сервер и воркеры...")

	// Создаем контекст с таймаутом для graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	// Останавливаем сервер
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Ошибка при остановке сервера: %v", err)
	}

	// Останавливаем воркеры
	cancel()
	time.Sleep(time.Second) // Даем время на graceful shutdown
	log.Println("Приложение остановлено")
}
