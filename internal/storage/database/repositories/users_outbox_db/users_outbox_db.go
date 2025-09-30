package users_outbox_db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ShlykovPavel/auth-JWT-microservice/internal/storage/database"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UsersOutboxRepositoryImpl struct {
	logger *slog.Logger
	dbPool *pgxpool.Pool
}

func NewUsersOutboxDB(dbPool *pgxpool.Pool, logger *slog.Logger) *UsersOutboxRepositoryImpl {
	return &UsersOutboxRepositoryImpl{
		logger: logger,
		dbPool: dbPool,
	}
}

type User struct {
	id            int64
	userId        int64
	sendToKafka   bool
	payload       string
	eventType     string
	attemptCount  int
	lastAttemptAt time.Time
}

func (or *UsersOutboxRepositoryImpl) GetUnsentUsers() ([]User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query := `SELECT id, user_id, payload, event_type, attempt_count, last_attempt_time FROM users_outbox WHERE sent_to_kafka = false ORDER BY id LIMIT 100`
	rows, err := or.dbPool.Query(ctx, query)
	if err != nil {
		or.logger.Error("Failed to fetch unsent users from outbox", slog.String("error", err.Error()))
		dbErr := database.PsqlErrorHandler(err)
		return nil, dbErr
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err = rows.Scan(&user.id, &user.userId, &user.payload, &user.eventType, &user.attemptCount, &user.lastAttemptAt)
		if err != nil {
			or.logger.Error("Failed to scan user row", slog.String("error", err.Error()))
			dbErr := database.PsqlErrorHandler(err)
			return nil, dbErr
		}
		users = append(users, user)

	}

	if err = rows.Err(); err != nil {
		or.logger.Error("Error reading rows", slog.Any("error", err))
		return nil, fmt.Errorf("error reading rows: %w", err)
	}
	return users, nil
}

func (or *UsersOutboxRepositoryImpl) UpdateAttemptCount(usersId []int64) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	baseQuery := `UPDATE users_outbox SET attempt_count = attempt_count + 1, last_attempt_time = $1 WHERE user_id = $2`
	attemptTime := time.Now()
	var totalUpdated int64
	tx, err := or.dbPool.Begin(ctx)
	if err != nil {
		or.logger.Error("Failed to begin transaction for updating attempt counts", slog.String("error", err.Error()))
		dbErr := database.PsqlErrorHandler(err)
		return 0, dbErr
	}
	defer tx.Rollback(ctx)
	for _, userId := range usersId {
		cmdTag, err := tx.Exec(ctx, baseQuery, attemptTime, userId)
		if err != nil {
			or.logger.Error("Failed to update attempt count for user", slog.Int64("user_id", userId), slog.String("error", err.Error()))
			dbErr := database.PsqlErrorHandler(err)
			return 0, dbErr
		}
		totalUpdated += cmdTag.RowsAffected()
	}
	if err = tx.Commit(ctx); err != nil {
		or.logger.Error("Failed to commit transaction for updating attempt counts", slog.String("error", err.Error()))
		dbErr := database.PsqlErrorHandler(err)
		return 0, dbErr
	}
	return totalUpdated, nil
}

//TODO добавить функцию установки send_to_kafka в true после успешной отправки в кафку
