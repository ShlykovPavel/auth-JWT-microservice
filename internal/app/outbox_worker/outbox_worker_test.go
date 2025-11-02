package outbox_worker

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"
	"time"

	KafkaProducer "github.com/ShlykovPavel/auth-JWT-microservice/internal/kafka/producer"
	"github.com/ShlykovPavel/auth-JWT-microservice/internal/storage/database/repositories/users_db"
	"github.com/ShlykovPavel/auth-JWT-microservice/internal/storage/database/repositories/users_outbox_db"
	"github.com/ShlykovPavel/auth-JWT-microservice/internal/test_helpers"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	kafkaContainer "github.com/testcontainers/testcontainers-go/modules/kafka"
)

const kafkaTopic = "users"

func TestSendUsersToKafka(t *testing.T) {

	ctx := context.Background()
	logger := slog.Default()
	postgresContainer, connStr := test_helpers.CreatePostgresContainer()

	defer postgresContainer.Terminate(ctx)

	dbConn, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("Failed to connect to Postgres: %v", err)
	}
	defer dbConn.Close()

	if err = dbConn.Ping(ctx); err != nil {
		t.Fatalf("Failed to ping Postgres: %v", err)
	}

	// Читаем и выполняем SQL скрипт с тестовыми данными
	sqlBytes, err := os.ReadFile("outbox_worker_test_data.sql")
	if err != nil {
		t.Fatalf("Failed to read test data SQL file: %v", err)
	}

	_, err = dbConn.Exec(ctx, string(sqlBytes))
	if err != nil {
		t.Fatalf("Failed to execute test data SQL script: %v", err)
	}
	t.Log("Test data loaded successfully")

	// Запуск Kafka контейнера
	kafkaCont, err := kafkaContainer.Run(ctx,
		"confluentinc/confluent-local:7.5.0",
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := kafkaCont.Terminate(ctx); err != nil {
			t.Logf("terminate error: %v", err)
		}
	})

	// Brokers: ["localhost:<mapped-port>"]
	brokers, err := kafkaCont.Brokers(ctx)
	assert.NoError(t, err)
	assert.Len(t, brokers, 1)
	t.Logf("Brokers: %v", brokers)

	kafkaProducer := KafkaProducer.InitKafkaProducer(brokers[0], kafkaTopic, logger)

	// Создаем топик вручную перед отправкой
	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		t.Fatalf("Failed to connect to Kafka: %v", err)
	}
	defer conn.Close()

	// Используем тот же broker для создания топиков
	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             kafkaTopic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	err = conn.CreateTopics(topicConfigs...)
	if err != nil {
		t.Logf("Warning: Failed to create topic (may already exist): %v", err)
	}

	usersOutboxRepository := users_outbox_db.NewUsersOutboxDB(dbConn, logger)
	outboxWorker := NewOutboxWorker(dbConn, kafkaProducer, usersOutboxRepository, logger, 2)

	// Проверяем, что данные есть в БД перед отправкой
	unsentUsers, err := usersOutboxRepository.GetUnsentUsers()
	assert.NoError(t, err)
	t.Logf("Unsent users before send: %d", len(unsentUsers))

	// Вызываем функцию
	err = outboxWorker.SendUsersToKafka()
	assert.NoError(t, err)

	// Проверяем изменения в DB
	rows, err := dbConn.Query(ctx, "SELECT user_id, send_to_kafka, attempt_count FROM users_outbox WHERE user_id IN (1,2,5) ORDER BY user_id")
	assert.NoError(t, err)
	defer rows.Close()

	var results []struct {
		userId   int64
		status   string
		attempts int
	}
	for rows.Next() {
		var r struct {
			userId   int64
			status   string
			attempts int
		}
		err := rows.Scan(&r.userId, &r.status, &r.attempts)
		assert.NoError(t, err)
		results = append(results, r)
	}
	assert.Len(t, results, 3)
	assert.Equal(t, int64(1), results[0].userId)
	assert.Equal(t, "sent", results[0].status)
	assert.Equal(t, 0, results[0].attempts)
	assert.Equal(t, int64(2), results[1].userId)
	assert.Equal(t, "sent", results[1].status)
	assert.Equal(t, 0, results[1].attempts)
	assert.Equal(t, int64(5), results[2].userId)
	assert.Equal(t, "sent", results[2].status)
	assert.Equal(t, 0, results[2].attempts)

	// Проверяем сообщения в Kafka
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   kafkaTopic,
		GroupID: "test-group",
	})
	defer reader.Close()

	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var messages []kafka.Message
	for {
		m, err := reader.ReadMessage(ctxTimeout)
		if err != nil {
			break
		}
		messages = append(messages, m)
	}
	assert.Len(t, messages, 3)

	// Проверяем ключи и payload
	expectedUserIds := []int64{1, 2, 5}
	for i, m := range messages {
		expectedId := expectedUserIds[i]
		assert.Equal(t, string(rune(expectedId+'0')), string(m.Key)) // "1", "2", "5"

		var payload users_db.UserInfo
		err := json.Unmarshal(m.Value, &payload)
		assert.NoError(t, err)
		assert.Equal(t, expectedId, payload.ID)
		// Можно добавить проверки на first_name и т.д., но для базового достаточно ID
	}
}
