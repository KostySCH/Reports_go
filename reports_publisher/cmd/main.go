package main

import (
	"context"
	"log"

	server "github.com/KostySCH/Reports_go/reports_publisher"
	"github.com/KostySCH/Reports_go/reports_publisher/internal/config"
	handler "github.com/KostySCH/Reports_go/reports_publisher/internal/handler"
	storage "github.com/KostySCH/Reports_go/reports_publisher/internal/repository/minio"
	"github.com/KostySCH/Reports_go/reports_publisher/internal/service"
)

func main() {
	ctx := context.Background()
	cfg := config.Load()
	srv := new(server.Server)

	minioClient, err := storage.New(ctx, cfg)
	if err != nil {
		log.Fatalf("fail init: %v", err)
	}

	docService := service.New(minioClient)

	handlers := handler.NewHandler(docService)

	if err := srv.Run("8080", handlers.InitRoutes()); err != nil {
		log.Fatalf("error runing %s", err.Error())
	}
}
