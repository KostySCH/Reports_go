package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
	"github.com/KostySCH/Reports_go/reports_publisher/internal/config"
)

type KafkaService struct {
	producer sarama.SyncProducer
	topic    string
}

type ReportNotification struct {
	ReportID string `json:"report_id"`
	UserID   string `json:"user_id"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
}

func NewKafkaService(cfg *config.Config) (*KafkaService, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Compression = sarama.CompressionSnappy
	config.Producer.MaxMessageBytes = 1000000

	producer, err := sarama.NewSyncProducer(cfg.Kafka.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания Kafka producer: %w", err)
	}

	return &KafkaService{
		producer: producer,
		topic:    cfg.Kafka.Topic,
	}, nil
}

func (s *KafkaService) SendNotification(ctx context.Context, notification ReportNotification) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("операция отменена: %w", ctx.Err())
	default:
		message, err := json.Marshal(notification)
		if err != nil {
			return fmt.Errorf("ошибка сериализации уведомления: %w", err)
		}

		msg := &sarama.ProducerMessage{
			Topic: s.topic,
			Value: sarama.StringEncoder(message),
		}

		_, _, err = s.producer.SendMessage(msg)
		if err != nil {
			return fmt.Errorf("ошибка отправки сообщения в Kafka: %w", err)
		}

		return nil
	}
}

func (s *KafkaService) Close() error {
	return s.producer.Close()
}
