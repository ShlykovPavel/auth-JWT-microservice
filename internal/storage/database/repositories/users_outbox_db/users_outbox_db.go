package users_outbox_db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ShlykovPavel/auth-JWT-microservice/internal/storage/database"
	"github.com/ShlykovPavel/auth-JWT-microservice/internal/storage/database/repositories/users_db"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UsersOutboxRepository interface {
	GetUnsentUsers() ([]User, error)
	UpdateAttemptCount(usersId []int64) (int64, error)
	MarkAsSentToKafka(usersId []int64) (int64, error)
	AddUserToOutbox(userId int64, eventType string) error
	GetAttemptCountLimitList(attemptLimit int8) ([]int64, error)
	MarkAsFailed(usersId []int64) error
}

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
	UserId        int64             `json:"user_id"`
	SendToKafka   bool              `json:"send_to_kafka"`
	Payload       users_db.UserInfo `json:"payload"`
	EventType     string            `json:"event_type"`
	AttemptCount  int               `json:"attempt_count"`
	LastAttemptAt *time.Time        `json:"last_attempt_at"`
}

func (or *UsersOutboxRepositoryImpl) GetUnsentUsers() ([]User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query := `SELECT 
    outbox.user_id, 
    outbox.event_type, 
    outbox.attempt_count, 
    outbox.last_attempt_at,
    users.id,
    users.first_name,
    users.last_name,
    users.email,
    users.role,
    users.phone
FROM users_outbox outbox
JOIN users ON outbox.user_id = users.id
WHERE outbox.send_to_kafka = 'pending'
ORDER BY outbox.id
LIMIT 100`
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
		err = rows.Scan(
			&user.UserId,
			&user.EventType,
			&user.AttemptCount,
			&user.LastAttemptAt,
			&user.Payload.ID,
			&user.Payload.FirstName,
			&user.Payload.LastName,
			&user.Payload.Email,
			&user.Payload.Role,
			&user.Payload.Phone)

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
func (or *UsersOutboxRepositoryImpl) GetAttemptCountLimitList(attemptLimit int8) ([]int64, error) {
	query := `SELECT user_id FROM users_outbox WHERE attempt_count >= $1 AND send_to_kafka != 'failed' AND send_to_kafka != 'sent'`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var usersExceedingLimit []int64
	rows, err := or.dbPool.Query(ctx, query, attemptLimit)

	if err != nil {
		or.logger.Error("Failed to fetch users exceeding attempt limit", slog.String("error", err.Error()))
		dbErr := database.PsqlErrorHandler(err)
		return nil, dbErr
	}
	defer rows.Close()
	for rows.Next() {
		var userId int64
		err = rows.Scan(&userId)
		if err != nil {
			or.logger.Error("Failed to scan user_id", slog.String("error", err.Error()))
			dbErr := database.PsqlErrorHandler(err)
			return nil, dbErr
		}
		usersExceedingLimit = append(usersExceedingLimit, userId)
	}
	if err = rows.Err(); err != nil {
		or.logger.Error("Error reading rows", slog.Any("error", err))
		return nil, fmt.Errorf("error reading rows: %w", err)
	}

	return usersExceedingLimit, nil
}

func (or *UsersOutboxRepositoryImpl) UpdateAttemptCount(usersId []int64) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	baseQuery := `UPDATE users_outbox SET attempt_count = attempt_count + 1, last_attempt_at = $1 WHERE user_id = $2`
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

func (or *UsersOutboxRepositoryImpl) MarkAsSentToKafka(usersId []int64) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	baseQuery := `UPDATE users_outbox SET send_to_kafka = 'sent' WHERE user_id = $1`
	var totalUpdated int64
	tx, err := or.dbPool.Begin(ctx)
	if err != nil {
		or.logger.Error("Failed to begin transaction for marking users as sent to Kafka", slog.String("error", err.Error()))
		dbErr := database.PsqlErrorHandler(err)
		return 0, dbErr
	}
	defer tx.Rollback(ctx)
	for _, userId := range usersId {
		cmdTag, err := tx.Exec(ctx, baseQuery, userId)
		if err != nil {
			or.logger.Error("Failed to mark user as sent to Kafka", slog.Int64("user_id", userId), slog.String("error", err.Error()))
			dbErr := database.PsqlErrorHandler(err)
			return 0, dbErr
		}
		totalUpdated += cmdTag.RowsAffected()
	}
	if err = tx.Commit(ctx); err != nil {
		or.logger.Error("Failed to commit transaction for marking users as sent to Kafka", slog.String("error", err.Error()))
		dbErr := database.PsqlErrorHandler(err)
		return 0, dbErr
	}
	return totalUpdated, nil
}

// AddUserToOutbox добавляет пользователя в outbox для последующей отправки в Kafka
func (or *UsersOutboxRepositoryImpl) AddUserToOutbox(userId int64, eventType string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	query := `INSERT INTO users_outbox (user_id, event_type, attempt_count, last_attempt_at, send_to_kafka)
		VALUES ($1, $2, 0, $3, 'pending')`
	now := time.Now()
	_, err := or.dbPool.Exec(ctx, query,
		userId,
		eventType,
		now,
	)
	if err != nil {
		or.logger.Error("Failed to insert user into outbox", slog.String("error", err.Error()))
		return database.PsqlErrorHandler(err)
	}
	return nil
}

func (or *UsersOutboxRepositoryImpl) MarkAsFailed(usersId []int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if len(usersId) == 0 {
		return fmt.Errorf("no users ids provided")
	}

	baseQuery := `UPDATE users_outbox SET send_to_kafka = 'failed' WHERE user_id = ANY($1)`
	tx, err := or.dbPool.Begin(ctx)
	if err != nil {
		or.logger.Error("Failed to begin transaction for marking users as failed in outbox", slog.String("error", err.Error()))
		dbErr := database.PsqlErrorHandler(err)
		return dbErr
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, baseQuery, usersId)
	if err != nil {
		or.logger.Error("Failed to mark users as failed in outbox", slog.String("error", err.Error()))
		dbErr := database.PsqlErrorHandler(err)
		return dbErr
	}
	if err = tx.Commit(ctx); err != nil {
		or.logger.Error("Failed to commit transaction for marking users as failed in outbox", slog.String("error", err.Error()))
		dbErr := database.PsqlErrorHandler(err)
		return dbErr
	}
	return nil
}
