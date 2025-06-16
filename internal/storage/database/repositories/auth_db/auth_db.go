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
	DbGetTokens(ctx context.Context, accessToken string, refreshToken string) (int64, error)
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

func (r *TokensRepositoryImpl) DbGetTokens(ctx context.Context, accessToken string, refreshToken string) (int64, error) {
	const op = "internal/storage/database/repositories/auth_db/auth_db.go/db.DbGetTokens"
	log := r.log.With(
		slog.String("operation", op),
		slog.String("access_token", accessToken),
		slog.String("refresh_token", refreshToken))
	query := `SELECT user_id FROM tokens WHERE access_token  = $1 AND refresh_token = $2`
	var userId int64
	err := r.db.QueryRow(ctx, query, accessToken, refreshToken).Scan(&userId)
	if err != nil {
		log.Error("Error while get tokens", "err", err.Error())
		return 0, database.PsqlErrorHandler(err)
	}
	return userId, nil
}

//func (r *TokensRepositoryImpl) DbUpdateTokens(ctx context.Context, accessToken string, refreshToken string) error {
//	//	TODO написать функцию обновления строкчик токена по юзер айди
//}
