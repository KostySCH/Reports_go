package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
		SSLMode  string `yaml:"sslmode"`
	} `yaml:"database"`

	Minio struct {
		Endpoint   string `yaml:"endpoint"`
		AccessKey  string `yaml:"access_key"`
		SecretKey  string `yaml:"secret_key"`
		PDFBucket  string `yaml:"pdf_bucket"`
		DOCXBucket string `yaml:"docx_bucket"`
		UseSSL     bool   `yaml:"use_ssl"`
	} `yaml:"minio"`

	Kafka struct {
		Brokers []string `yaml:"brokers"`
		Topic   string   `yaml:"topic"`
	} `yaml:"kafka"`
}

func LoadConfig(configPath string) (*Config, error) {
	config := &Config{}

	file, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения файла конфигурации: %w", err)
	}

	err = yaml.Unmarshal(file, config)
	if err != nil {
		return nil, fmt.Errorf("ошибка разбора файла конфигурации: %w", err)
	}

	return config, nil
}

func getDefaultConfig() *Config {
	return &Config{
		Database: struct {
			Host     string `yaml:"host"`
			Port     int    `yaml:"port"`
			User     string `yaml:"user"`
			Password string `yaml:"password"`
			DBName   string `yaml:"dbname"`
			SSLMode  string `yaml:"sslmode"`
		}{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "postgres",
			DBName:   "reports_db",
			SSLMode:  "disable",
		},
		Minio: struct {
			Endpoint   string `yaml:"endpoint"`
			AccessKey  string `yaml:"access_key"`
			SecretKey  string `yaml:"secret_key"`
			PDFBucket  string `yaml:"pdf_bucket"`
			DOCXBucket string `yaml:"docx_bucket"`
			UseSSL     bool   `yaml:"use_ssl"`
		}{
			Endpoint:   "localhost:9000",
			AccessKey:  "minioadmin",
			SecretKey:  "minioadmin",
			PDFBucket:  "reports-pdf",
			DOCXBucket: "reports-docx",
			UseSSL:     false,
		},
		Kafka: struct {
			Brokers []string `yaml:"brokers"`
			Topic   string   `yaml:"topic"`
		}{
			Brokers: []string{"localhost:9092"},
			Topic:   "report-notifications",
		},
	}
}
