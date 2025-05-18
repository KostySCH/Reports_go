package logger

import (
	"fmt"
	"log"
	"runtime"
	"time"
)

const (
	separator = "----------------------------------------"
)

// GetGoroutineID возвращает ID текущей горутины
func GetGoroutineID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	var id uint64
	fmt.Sscanf(string(b), "goroutine %d", &id)
	return id
}

// getGoroutineID возвращает ID текущей горутины
func getGoroutineID() uint64 {
	return GetGoroutineID()
}

// LogWorkerEvent логирует событие воркера с форматированием
func LogWorkerEvent(workerType string, workerID int, message string) {
	goroutineID := getGoroutineID()
	header := fmt.Sprintf("[Воркер %d - %s - Поток %d]", workerID, workerType, goroutineID)
	log.Printf("%s %s", header, message)
}

// LogWorkerError логирует ошибку воркера с форматированием
func LogWorkerError(workerType string, workerID int, err error) {
	goroutineID := getGoroutineID()
	header := fmt.Sprintf("[Воркер %d - %s - Поток %d]", workerID, workerType, goroutineID)
	log.Printf("%s ОШИБКА: %v", header, err)
}

// LogWorkerReport логирует отчет о работе воркера
func LogWorkerReport(workerType string, workerID int, reportID string, attempt int, maxAttempts int, success bool, errorMsg string) {
	goroutineID := getGoroutineID()
	header := fmt.Sprintf("[Воркер %d - %s - Поток %d]", workerID, workerType, goroutineID)
	status := "УСПЕХ"
	if !success {
		status = "НЕУДАЧА"
	}

	report := fmt.Sprintf("Отчет ID: %s | Попытка: %d/%d | Статус: %s", reportID, attempt, maxAttempts, status)
	if !success && errorMsg != "" {
		report += fmt.Sprintf(" | Ошибка: %s", errorMsg)
	}

	log.Printf("%s %s", header, report)
}

// LogWorkerSeparator выводит разделитель
func LogWorkerSeparator() {
	log.Println(separator)
}

// FormatDuration форматирует длительность в читаемый вид
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dмс", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fс", d.Seconds())
	}
	return fmt.Sprintf("%.1fм", d.Minutes())
}
