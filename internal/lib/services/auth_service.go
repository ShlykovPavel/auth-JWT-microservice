package services

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/ShlykovPavel/auth-JWT-microservice/internal/lib/jwt_tokens"
	"github.com/ShlykovPavel/auth-JWT-microservice/internal/server/auth"
	"github.com/ShlykovPavel/auth-JWT-microservice/internal/storage/database/repositories/auth_db"
	"github.com/ShlykovPavel/auth-JWT-microservice/internal/storage/database/repositories/users_db"
	tokens2 "github.com/ShlykovPavel/auth-JWT-microservice/models/tokens"
	getUserDto "github.com/ShlykovPavel/auth-JWT-microservice/models/users/get_user"
)

var ErrWrongPassword = errors.New("Password is incorrect ")

type AuthService struct {
	userRepo    users_db.UserRepository
	tokensRepo  auth_db.TokensRepository
	log         *slog.Logger
	secretKey   string
	JWTDuration time.Duration
}

func NewAuthService(db users_db.UserRepository, tokensRepo auth_db.TokensRepository, log *slog.Logger, secretKey string, jwtDuration time.Duration) *AuthService {
	return &AuthService{
		userRepo:    db,
		tokensRepo:  tokensRepo,
		log:         log,
		secretKey:   secretKey,
		JWTDuration: jwtDuration,
	}
}

func (a *AuthService) Authentication(user *getUserDto.AuthUser, ctx context.Context) (tokens2.RefreshTokensDto, error) {
	const op = "server/auth/auth/Authentification"
	log := a.log.With(
		slog.String("operation", op),
		slog.String("request email: ", user.Email))

	// Проверяем что пользователь есть в БД
	usr, err := a.userRepo.GetUserByEmail(ctx, user.Email)
	if err != nil {
		if errors.Is(err, users_db.ErrUserNotFound) {
			log.Debug("UserInfo not found", "user", user)
			return tokens2.RefreshTokensDto{}, err
		}
		log.Error("Error while fetching user", "err", err)
		return tokens2.RefreshTokensDto{}, err
	}
	// Проверяем что нам предоставили правильный пароль
	ok := auth.ComparePassword(usr.PasswordHash, user.Password, log)
	if !ok {
		return tokens2.RefreshTokensDto{}, ErrWrongPassword
	}
	accessToken, err := jwt_tokens.CreateAccessToken(usr.ID, a.secretKey, usr.Role, a.JWTDuration, a.log)
	if err != nil {
		log.Error("Error while creating access token", "err", err)
		return tokens2.RefreshTokensDto{}, err
	}
	refreshToken, err := jwt_tokens.CreateRefreshToken(a.log)
	if err != nil {
		log.Error("Error while creating refresh token", "err", err)
		return tokens2.RefreshTokensDto{}, err
	}
	err = a.tokensRepo.DbPutTokens(ctx, usr.ID, refreshToken)
	if err != nil {
		log.Error("Error while storing tokens", "err", err)
		return tokens2.RefreshTokensDto{}, err
	}
	return tokens2.RefreshTokensDto{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (a *AuthService) RefreshTokens(tokensForRefresh *tokens2.RefreshTokensDto, ctx context.Context) (tokens2.RefreshTokensDto, error) {
	const op = "internal/lib/services/auth_service.go/RefreshTokens"
	log := a.log.With(
		slog.String("operation", op),
	)
	tokenData, err := a.tokensRepo.DbGetTokens(ctx, tokensForRefresh.RefreshToken)
	if err != nil {
		log.Error("Error while fetching tokensForRefresh", "err", err)
		return tokens2.RefreshTokensDto{}, err
	}
	accessToken, err := jwt_tokens.CreateAccessToken(tokenData.UserId, a.secretKey, tokenData.UserRole, a.JWTDuration, a.log)
	if err != nil {
		log.Error("Error while creating access token", "err", err)
		return tokens2.RefreshTokensDto{}, err
	}
	refreshToken, err := jwt_tokens.CreateRefreshToken(a.log)
	if err != nil {
		log.Error("Error while creating refresh token", "err", err)
		return tokens2.RefreshTokensDto{}, err
	}
	err = a.tokensRepo.DbUpdateTokens(context.Background(), tokenData.UserId, refreshToken, tokensForRefresh.RefreshToken)
	if err != nil {
		log.Error("Error while storing tokensForRefresh", "err", err)
		return tokens2.RefreshTokensDto{}, err
	}
	return tokens2.RefreshTokensDto{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (a *AuthService) Logout(tokens *tokens2.LogoutRequest, ctx context.Context) error {
	const op = "server/auth/auth/Logout"
	log := a.log.With(
		slog.String("operation", op))
	_, err := a.tokensRepo.DbGetTokens(ctx, tokens.RefreshToken)
	if err != nil {
		log.Error("Error while fetching tokens", "err", err)
		return err
	}
	err = a.tokensRepo.DbDeleteToken(context.Background(), tokens.RefreshToken)
	if err != nil {
		log.Error("Error while deleting tokens", "err", err)
		return err
	}
	return nil

}
