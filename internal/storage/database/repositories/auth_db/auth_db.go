package auth_db

import (
	"booker/internal/storage/database"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"strconv"
)

type TokensRepository interface {
	DbPutTokens(ctx context.Context, userId int64, refreshToken string) error
	DbGetTokens(ctx context.Context, refreshToken string) (int64, error)
	DbUpdateTokens(ctx context.Context, userId int64, refreshToken string) error
	DbDeleteToken(ctx context.Context, refreshToken string) error
}
type TokensRepositoryImpl struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

func NewTokensRepositoryImpl(db *pgxpool.Pool, log *slog.Logger) *TokensRepositoryImpl {
	return &TokensRepositoryImpl{
		db:  db,
		log: log,
	}
}

func (r *TokensRepositoryImpl) DbPutTokens(ctx context.Context, userId int64, refreshToken string) error {
	const op = "internal/storage/database/repositories/auth_db/auth_db.go/db.PutTokens"
	log := r.log.With(
		slog.String("operation", op),
		slog.String("User_id", strconv.FormatInt(userId, 10)),
		slog.String("refresh_token", refreshToken))

	query := `INSERT INTO tokens(user_id, refresh_token) VALUES($1, $2)`
	_, err := r.db.Exec(ctx, query, userId, refreshToken)
	if err != nil {
		log.Error("Error while put tokens in db", "err", err.Error())
		return database.PsqlErrorHandler(err)
	}
	return nil
}

func (r *TokensRepositoryImpl) DbGetTokens(ctx context.Context, refreshToken string) (int64, error) {
	const op = "internal/storage/database/repositories/auth_db/auth_db.go/db.DbGetTokens"
	log := r.log.With(
		slog.String("operation", op),
		slog.String("refresh_token", refreshToken))
	query := `SELECT user_id FROM tokens WHERE refresh_token = $1`
	var userId int64
	err := r.db.QueryRow(ctx, query, refreshToken).Scan(&userId)
	if err != nil {
		log.Error("Error while get tokens", "err", err.Error())
		return 0, database.PsqlErrorHandler(err)
	}
	return userId, nil
}

func (r *TokensRepositoryImpl) DbUpdateTokens(ctx context.Context, userId int64, refreshToken string) error {
	const op = "internal/storage/database/repositories/auth_db/auth_db.go/DbUpdateTokens"
	log := r.log.With(
		slog.String("operation", op),
	)
	query := `UPDATE tokens SET refresh_token = $1, updated_at = CURRENT_TIMESTAMP WHERE user_id = $2`
	_, err := r.db.Exec(ctx, query, refreshToken, userId)
	if err != nil {
		log.Error("Error while update tokens", "err", err.Error())
		return database.PsqlErrorHandler(err)
	}
	return nil
}

func (r *TokensRepositoryImpl) DbDeleteToken(ctx context.Context, refreshToken string) error {
	const op = "internal/storage/database/repositories/auth_db/auth_db.go/DbDeleteToken"
	log := r.log.With(
		slog.String("operation", op),
	)
	query := `DELETE FROM tokens WHERE refresh_token = $1`
	_, err := r.db.Exec(ctx, query, refreshToken)
	if err != nil {
		log.Error("Error while delete token", "err", err.Error())
		return database.PsqlErrorHandler(err)
	}
	return nil
}
