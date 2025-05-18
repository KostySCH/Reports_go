package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Database struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
		SSLMode  string `yaml:"sslmode"`
	} `yaml:"database"`
	MinIO struct {
		Endpoint   string `yaml:"endpoint"`
		AccessKey  string `yaml:"access_key"`
		SecretKey  string `yaml:"secret_key"`
		PDFBucket  string `yaml:"pdf_bucket"`
		DOCXBucket string `yaml:"docx_bucket"`
		UseSSL     bool   `yaml:"use_ssl"`
	} `yaml:"minio"`
}

func Load() *Config {
	config := &Config{}

	// Получаем текущую рабочую директорию
	workDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting working directory: %v\n", err)
		return getDefaultConfig()
	}

	// Пробуем найти конфигурационный файл в разных местах
	configPaths := []string{
		filepath.Join(workDir, "config.yaml"),                       // В текущей директории
		filepath.Join(workDir, "config", "config.yaml"),             // В поддиректории config
		filepath.Join(workDir, "..", "config", "config.yaml"),       // В родительской директории/config
		filepath.Join(workDir, "..", "..", "config", "config.yaml"), // В корне проекта
	}

	var configData []byte
	for _, path := range configPaths {
		data, err := os.ReadFile(path)
		if err == nil {
			configData = data
			fmt.Printf("Found config file at: %s\n", path)
			break
		}
	}

	if configData == nil {
		fmt.Printf("Config file not found in any of the expected locations\n")
		return getDefaultConfig()
	}

	// Парсим YAML
	if err := yaml.Unmarshal(configData, config); err != nil {
		fmt.Printf("Error parsing config file: %v\n", err)
		return getDefaultConfig()
	}

	return config
}

func getDefaultConfig() *Config {
	return &Config{
		Database: struct {
			Host     string `yaml:"host"`
			Port     string `yaml:"port"`
			User     string `yaml:"user"`
			Password string `yaml:"password"`
			DBName   string `yaml:"dbname"`
			SSLMode  string `yaml:"sslmode"`
		}{
			Host:     "localhost",
			Port:     "5432",
			User:     "postgres",
			Password: "postgres",
			DBName:   "reports_db",
			SSLMode:  "disable",
		},
		MinIO: struct {
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
	}
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode)
}
