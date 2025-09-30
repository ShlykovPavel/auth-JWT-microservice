package kafka

import (
	"context"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	kafkaProducer *kafka.Writer
}

func InitKafkaProducer(address string, topic string, logger *slog.Logger) *KafkaProducer {
	return &KafkaProducer{
		kafkaProducer: &kafka.Writer{
			Addr:                   kafka.TCP(address),
			Topic:                  topic,
			Balancer:               &kafka.Hash{},
			Compression:            kafka.Snappy,
			BatchTimeout:           10 * time.Millisecond,
			Logger:                 kafka.LoggerFunc(logger.Info),
			ErrorLogger:            kafka.LoggerFunc(logger.Error),
			AllowAutoTopicCreation: false,
		},
	}
}

// WriteMessages пишет сообщения в Kafka и логирует ошибки, если они возникают.
func (kw *KafkaProducer) WriteMessages(logger *slog.Logger, ctx context.Context, messages ...kafka.Message) error {
	err := kw.kafkaProducer.WriteMessages(ctx, messages...)
	if err != nil {
		logger.Error("Failed to write messages to Kafka", "error", err)
		return err
	}
	logger.Debug("Successfully written messages to Kafka", "count", len(messages), "topic", kw.kafkaProducer.Topic, "messages", messages)
	return nil
}

// Close закрывает соединение с Kafka.
func (kw *KafkaProducer) Close() error {
	return kw.kafkaProducer.Close()
}
