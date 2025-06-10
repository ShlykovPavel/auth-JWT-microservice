package services

import (
	"booker/internal/lib/api/models"
	"booker/internal/lib/jwt_tokens"
	"booker/internal/server/users"
	"booker/internal/storage/database/repositories/auth_db"
	"booker/internal/storage/database/repositories/users_db"
	"context"
	"errors"
	"log/slog"
)

var ErrWrongPassword = errors.New("Password is incorrect ")

type AuthService struct {
	userRepo   users_db.UserRepository
	tokensRepo auth_db.TokensRepository
	log        *slog.Logger
	secretKey  string
}

func NewAuthService(db users_db.UserRepository, tokensRepo auth_db.TokensRepository, log *slog.Logger, secretKey string) *AuthService {
	return &AuthService{
		userRepo:   db,
		tokensRepo: tokensRepo,
		log:        log,
		secretKey:  secretKey,
	}
}

func (a *AuthService) Authentication(user *models.AuthUser) (models.UserTokens, error) {
	const op = "server/users/auth/Authentification"
	log := a.log.With(
		slog.String("operation", op),
		slog.String("request email: ", user.Email))

	// Проверяем что пользователь есть в БД
	usr, err := a.userRepo.GetUser(context.Background(), user.Email)
	if err != nil {
		if errors.Is(err, users_db.ErrUserNotFound) {
			log.Debug("UserInfo not found", "user", user)
			return models.UserTokens{}, err
		}
		log.Error("Error while fetching user", "err", err)
		return models.UserTokens{}, err
	}
	// Проверяем что нам предоставили правильный пароль
	ok := users.ComparePassword(usr.PasswordHash, user.Password, log)
	if !ok {
		return models.UserTokens{}, ErrWrongPassword
	}
	accessToken, err := jwt_tokens.CreateAccessToken(usr.ID, a.secretKey, a.log)
	if err != nil {
		log.Error("Error while creating access token", "err", err)
		return models.UserTokens{}, err
	}
	refreshToken, err := jwt_tokens.CreateRefreshToken(a.log)
	if err != nil {
		log.Error("Error while creating refresh token", "err", err)
		return models.UserTokens{}, err
	}
	err = a.tokensRepo.DbPutTokens(context.Background(), usr.ID, accessToken, refreshToken)
	if err != nil {
		log.Error("Error while storing tokens", "err", err)
		return models.UserTokens{}, err
	}
	return models.UserTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
