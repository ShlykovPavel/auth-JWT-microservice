package services

import (
	"booker/internal/lib/api/models"
	"booker/internal/lib/jwt_tokens"
	"booker/internal/server/users"
	"booker/internal/server/users/users_db"
	"context"
	"errors"
	"log/slog"
)

var ErrWrongPassword = errors.New("Password is incorrect ")

type AuthService struct {
	db        users_db.UserRepository
	log       *slog.Logger
	secretKey string
}

func NewAuthService(db users_db.UserRepository, log *slog.Logger, secretKey string) *AuthService {
	return &AuthService{
		db:        db,
		log:       log,
		secretKey: secretKey,
	}
}

func (a *AuthService) Authentication(user *models.AuthUser) (models.UserTokens, error) {
	const op = "server/users/auth/Authentification"
	log := a.log.With(
		slog.String("operation", op),
		slog.String("request email: ", user.Email))

	// Проверяем что пользователь есть в БД
	usr, err := a.db.GetUser(context.Background(), user.Email)
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
	access_token, err = jwt_tokens.CreateAccessToken(usr.ID, a.secretKey, a.log)
	if err != nil {
		log.Error("Error while creating access token", "err", err)
		return models.UserTokens{}, err
	}
	resresh_token, err = jwt_tokens.CreateRefreshToken(a.log)
	if err != nil {
		log.Error("Error while creating refresh token", "err", err)
		return models.UserTokens{}, err
	}
	//	TODO Сделать функцию записи токенов в БД
}
