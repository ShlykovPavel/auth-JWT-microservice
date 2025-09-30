package outbox_worker

import (
	kafka "github.com/ShlykovPavel/auth-JWT-microservice/internal/kafka/producer"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OutboxWorker struct {
	dbPoll        *pgxpool.Pool
	kafkaProducer *kafka.KafkaProducer
}

func NewOutboxWorker(dbPoll *pgxpool.Pool, kafkaProducer *kafka.KafkaProducer) *OutboxWorker {
	return &OutboxWorker{
		dbPoll:        dbPoll,
		kafkaProducer: kafkaProducer,
	}
}

func (ow *OutboxWorker) SendUsersToKafka() error {
	//	Выполняем поиск новых записей в outbox
	//	TODO Реализовать репозиторий для outbox и использовать его здесь
}
