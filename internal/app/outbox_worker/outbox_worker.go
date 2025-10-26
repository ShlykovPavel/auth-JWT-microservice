package outbox_worker

import (
	"context"
	"encoding/json"
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
	attemptLimit    int8
}

func NewOutboxWorker(dbPoll *pgxpool.Pool, kafkaProducer *kafkaProducer.KafkaProducer, usersOutboxDBRepo users_outbox_db.UsersOutboxRepository, logger *slog.Logger, attemptLimit int8) *OutboxWorker {
	return &OutboxWorker{
		dbPoll:          dbPoll,
		kafkaProducer:   kafkaProducer,
		usersOutboxRepo: usersOutboxDBRepo,
		logger:          logger,
		attemptLimit:    attemptLimit,
	}
}

func (ow *OutboxWorker) SendUsersToKafka() error {
	//	Выполняем поиск новых записей в outbox
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
	ow.logger.Debug("Found unsent users in outbox", "count", len(unsentUsers))

	var messages []kafka.Message
	var userIds []int64
	for _, user := range unsentUsers {
		kafkaPayload, err := json.Marshal(user.Payload)
		if err != nil {
			ow.logger.Error("Failed to marshal user payload", "user_id", user.UserId, "error", err)
			continue
		}
		message := kafka.Message{
			Key:   []byte(strconv.FormatInt(user.UserId, 10)),
			Value: kafkaPayload,
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
	ow.logger.Debug("Successfully sent users to Kafka", "count", len(messages))
	//	Обновляем записи в outbox, помечая их как отправленные
	updatedCount, err := ow.usersOutboxRepo.MarkAsSentToKafka(userIds)
	if err != nil {
		ow.logger.Error("Failed to mark users as sent in outbox", "error", err)
		return err
	}
	ow.logger.Debug("Marked users as sent in outbox", "count", updatedCount)
	return nil
}

func (ow *OutboxWorker) MarkFailedUsers() error {
	//	Ищем записи, у которых количество попыток превышает лимит
	failedUsers, err := ow.usersOutboxRepo.GetAttemptCountLimitList(ow.attemptLimit)
	if err != nil {
		ow.logger.Error("Failed to get users exceeding attempt limit from outbox", "error", err)
		return err
	}
	if len(failedUsers) == 0 {
		ow.logger.Debug("No users exceeding attempt limit found in outbox")
		return nil
	}
	ow.logger.Debug("Found users exceeding attempt limit in outbox", "count", len(failedUsers))
	ow.logger.Debug("Start making users as failed to send in kafka", "Users", failedUsers)
	//	Помечаем эти записи как ошибочные
	err = ow.usersOutboxRepo.MarkAsFailed(failedUsers)
	if err != nil {
		ow.logger.Error("Failed to mark users as failed in outbox", "error", err)
		return err
	}
	ow.logger.Debug("Successfully marked users as failed in outbox")
	return nil
}
