package auth_db

import (
	"booker/internal/storage/database"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"strconv"
)

type TokensRepository interface {
	DbPutTokens(ctx context.Context, userId int64, accessToken string, refreshToken string) error
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

func (r *TokensRepositoryImpl) DbPutTokens(ctx context.Context, userId int64, accessToken string, refreshToken string) error {
	const op = "internal/storage/database/repositories/auth_db/auth_db.go/db.PutTokens"
	log := r.log.With(
		slog.String("operation", op),
		slog.String("User_id", strconv.FormatInt(userId, 10)),
		slog.String("access_token", accessToken),
		slog.String("refresh_token", refreshToken))

	query := `INSERT INTO tokens(user_id, access_token, refresh_token) VALUES($1, $2, $3)`
	_, err := r.db.Exec(ctx, query, userId, accessToken, refreshToken)
	if err != nil {
		log.Error("Error while put tokens in db", "err", err.Error())
		return database.PsqlErrorHandler(err)
	}
	return nil
}
