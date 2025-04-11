package rest

import (
	"net/http"

	"github.com/KostySCH/Reports_go/reports_publisher/internal/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	services *service.DocumentService
}

func NewHandler(services *service.DocumentService) *Handler {
	return &Handler{services: services}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	api := router.Group("/api/v1")
	{
		documents := api.Group("/documents")
		{
			documents.GET("/pdf", h.getPDFDocuments)
			documents.GET("/docx", h.getDOCXDocuments)
			documents.GET("/:type/:name", h.downloadFile)
		}
	}

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
