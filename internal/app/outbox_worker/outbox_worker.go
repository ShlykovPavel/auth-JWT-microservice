package outbox_worker

import (
	"context"
	"log/slog"
	"strconv"

	kafkaProducer "github.com/ShlykovPavel/auth-JWT-microservice/internal/kafka/producer"
	"github.com/ShlykovPavel/auth-JWT-microservice/internal/storage/database/repositories/users_outbox_db"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
)

type OutboxWorker struct {
	dbPoll          *pgxpool.Pool
	kafkaProducer   *kafkaProducer.KafkaProducer
	usersOutboxRepo users_outbox_db.UsersOutboxRepository
	logger          *slog.Logger
}

func NewOutboxWorker(dbPoll *pgxpool.Pool, kafkaProducer *kafkaProducer.KafkaProducer, usersOutboxDBRepo users_outbox_db.UsersOutboxRepository, logger *slog.Logger) *OutboxWorker {
	return &OutboxWorker{
		dbPoll:          dbPoll,
		kafkaProducer:   kafkaProducer,
		usersOutboxRepo: usersOutboxDBRepo,
		logger:          logger,
	}
}

func (ow *OutboxWorker) SendUsersToKafka() error {
	//	Выполняем поиск новых записей в outbox
	//	TODO Реализовать репозиторий для outbox и использовать его здесь
	ctx := context.Background()
	var unsentUsers []users_outbox_db.User
	unsentUsers, err := ow.usersOutboxRepo.GetUnsentUsers()
	if err != nil {
		ow.logger.Error("Failed to get unsent users from outbox", "error", err)
		return err
	}
	if len(unsentUsers) == 0 {
		ow.logger.Debug("No unsent users found in outbox")
		return nil
	}
	ow.logger.Info("Found unsent users in outbox", "count", len(unsentUsers))

	var messages []kafka.Message
	var userIds []int64
	for _, user := range unsentUsers {
		message := kafka.Message{
			Key:   []byte(strconv.FormatInt(user.UserId, 10)),
			Value: []byte(user.Payload.Email + "|" + user.Payload.FirstName + "|" + user.Payload.LastName + "|" + user.Payload.Role + "|" + user.Payload.Phone),
		}
		messages = append(messages, message)
		userIds = append(userIds, user.UserId)
	}
	err = ow.kafkaProducer.WriteMessages(ow.logger, ctx, messages...)
	if err != nil {
		ow.logger.Error("Failed to send users to Kafka", "error", err)
		//	Обновляем счётчик попыток для всех пользователей, которых пытались отправить
		_, updateErr := ow.usersOutboxRepo.UpdateAttemptCount(userIds)
		if updateErr != nil {
			ow.logger.Error("Failed to update attempt count for users in outbox", "error", updateErr)
		}
		return err

	}
	ow.logger.Info("Successfully sent users to Kafka", "count", len(messages))
	//	Обновляем записи в outbox, помечая их как отправленные
	updatedCount, err := ow.usersOutboxRepo.MarkAsSentToKafka(userIds)
	if err != nil {
		ow.logger.Error("Failed to mark users as sent in outbox", "error", err)
		return err
	}
	ow.logger.Info("Marked users as sent in outbox", "count", updatedCount)
	return nil

}
