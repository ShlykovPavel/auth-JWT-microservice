package auth_db

import (
	"context"
	"github.com/ShlykovPavel/auth-JWT-microservice/internal/storage/database"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"strconv"
)

type JWTTokenData struct {
	UserId   int64
	UserRole string
}
type TokensRepository interface {
	DbPutTokens(ctx context.Context, userId int64, refreshToken string) error
	DbGetTokens(ctx context.Context, refreshToken string) (JWTTokenData, error)
	DbUpdateTokens(ctx context.Context, userId int64, refreshToken string, oldRefreshToken string) error
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

func (r *TokensRepositoryImpl) DbGetTokens(ctx context.Context, refreshToken string) (JWTTokenData, error) {
	const op = "internal/storage/database/repositories/auth_db/auth_db.go/db.DbGetTokens"
	log := r.log.With(
		slog.String("operation", op),
		slog.String("refresh_token", refreshToken))
	query := `SELECT public.tokens.user_id, users.role FROM tokens JOIN public.users ON tokens.user_id = users.id WHERE refresh_token = $1`
	var tokenData JWTTokenData
	err := r.db.QueryRow(ctx, query, refreshToken).Scan(&tokenData.UserId, &tokenData.UserRole)
	if err != nil {
		log.Error("Error while get tokens", "err", err.Error())
		return tokenData, database.PsqlErrorHandler(err)
	}
	return tokenData, nil
}

func (r *TokensRepositoryImpl) DbUpdateTokens(ctx context.Context, userId int64, refreshToken string, oldRefreshToken string) error {
	const op = "internal/storage/database/repositories/auth_db/auth_db.go/DbUpdateTokens"
	log := r.log.With(
		slog.String("operation", op),
	)
	query := `UPDATE tokens SET refresh_token = $1, updated_at = CURRENT_TIMESTAMP WHERE user_id = $2 AND refresh_token = $3`
	_, err := r.db.Exec(ctx, query, refreshToken, userId, oldRefreshToken)
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
