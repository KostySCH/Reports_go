package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/KostySCH/Reports_go/reports_publisher/internal/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	services *service.DocumentService
}

func NewHandler(services *service.DocumentService) *Handler {
	if services == nil {
		log.Fatal("DocumentService не может быть nil")
	}
	return &Handler{services: services}
}

func (h *Handler) InitRoutes() *gin.Engine {
	log.Println("Начало инициализации маршрутов...")

	// Устанавливаем режим отладки для Gin
	gin.SetMode(gin.DebugMode)

	router := gin.New()
	if router == nil {
		log.Fatal("Не удалось создать роутер Gin")
	}

	// Добавляем middleware для логирования
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[%s] | %s | %d | %s | %s | %s | %s | %s | %s\n",
			param.TimeStamp.Format(time.RFC3339),
			param.ClientIP,
			param.StatusCode,
			param.Method,
			param.Path,
			param.Request.UserAgent(),
			param.ErrorMessage,
			param.Latency,
			param.Request.Host,
		)
	}))
	router.Use(gin.Recovery())

	// Добавляем тестовые маршруты для проверки
	router.GET("/", func(c *gin.Context) {
		log.Println("Получен запрос на /")
		c.JSON(200, gin.H{
			"message": "Сервер работает",
		})
	})

	router.GET("/ping", func(c *gin.Context) {
		log.Println("Получен запрос на /ping")
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	router.GET("/test", func(c *gin.Context) {
		log.Println("Получен запрос на /test")
		c.JSON(200, gin.H{
			"message": "test endpoint",
		})
	})

	log.Println("Создание групп маршрутов...")

	api := router.Group("/api/v1")
	{
		documents := api.Group("/documents")
		{
			log.Println("Регистрация маршрутов для документов...")
			documents.GET("/pdf", h.getPDFDocuments)
			documents.GET("/docx", h.getDOCXDocuments)
			documents.GET("/:type/:name", h.downloadFile)
		}

		reports := api.Group("/reports")
		{
			log.Println("Регистрация маршрутов для отчетов...")
			reports.GET("/:uuid/download", h.downloadReportByUUID)
		}
	}

	log.Println("Маршруты успешно инициализированы")
	return router
}

func (h *Handler) getPDFDocuments(c *gin.Context) {
	docs, err := h.services.GetAvailablePDFs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, docs)
}

func (h *Handler) getDOCXDocuments(c *gin.Context) {
	docs, err := h.services.GetAvailableDOCXs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, docs)
}

func (h *Handler) downloadFile(c *gin.Context) {
	fileType := c.Param("type")
	fileName := c.Param("name")

	file, err := h.services.DownloadFile(c.Request.Context(), fileType, fileName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}
	defer file.Close()

	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.DataFromReader(http.StatusOK, -1, "application/octet-stream", file, nil)
}

func (h *Handler) downloadReportByUUID(c *gin.Context) {
	uuid := c.Param("uuid")
	if uuid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID отчета не указан"})
		return
	}

	reader, err := h.services.GetReport(c.Request.Context(), uuid)
	if err != nil {
		log.Printf("Ошибка получения отчета: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Отчет не найден"})
		return
	}

	// Если reader реализует io.ReadCloser, закрываем его после использования
	if closer, ok := reader.(io.ReadCloser); ok {
		defer closer.Close()
	}

	// Получаем путь к файлу из базы данных для определения имени файла
	var reportPath string
	err = h.services.GetReportPath(c.Request.Context(), uuid, &reportPath)
	if err != nil {
		log.Printf("Ошибка получения пути к файлу: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения информации о файле"})
		return
	}

	// Извлекаем оригинальное имя файла из пути
	parts := strings.Split(reportPath, "/")
	fileName := parts[len(parts)-1]

	// Определяем Content-Type на основе расширения файла
	contentType := "application/pdf"
	if strings.HasSuffix(strings.ToLower(fileName), ".docx") {
		contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	}

	log.Printf("Скачивание файла: %s с типом контента: %s", fileName, contentType)

	// Устанавливаем заголовки для скачивания
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", fileName, url.QueryEscape(fileName)))
	c.Header("Content-Type", contentType)
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Accept-Ranges", "bytes")
	c.Header("Cache-Control", "private")
	c.Header("Pragma", "private")
	c.Header("Expires", "0")

	// Используем DataFromReader для передачи файла
	c.DataFromReader(http.StatusOK, -1, contentType, reader, nil)
}
