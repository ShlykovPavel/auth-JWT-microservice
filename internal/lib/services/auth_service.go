package services

import (
	getUserDto "booker/internal/lib/api/models/users/get_user"
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

func (a *AuthService) Authentication(user *getUserDto.AuthUser) (getUserDto.UserTokens, error) {
	const op = "server/users/auth/Authentification"
	log := a.log.With(
		slog.String("operation", op),
		slog.String("request email: ", user.Email))

	// Проверяем что пользователь есть в БД
	usr, err := a.userRepo.GetUser(context.Background(), user.Email)
	if err != nil {
		if errors.Is(err, users_db.ErrUserNotFound) {
			log.Debug("UserInfo not found", "user", user)
			return getUserDto.UserTokens{}, err
		}
		log.Error("Error while fetching user", "err", err)
		return getUserDto.UserTokens{}, err
	}
	// Проверяем что нам предоставили правильный пароль
	ok := users.ComparePassword(usr.PasswordHash, user.Password, log)
	if !ok {
		return getUserDto.UserTokens{}, ErrWrongPassword
	}
	accessToken, err := jwt_tokens.CreateAccessToken(usr.ID, a.secretKey, a.log)
	if err != nil {
		log.Error("Error while creating access token", "err", err)
		return getUserDto.UserTokens{}, err
	}
	refreshToken, err := jwt_tokens.CreateRefreshToken(a.log)
	if err != nil {
		log.Error("Error while creating refresh token", "err", err)
		return getUserDto.UserTokens{}, err
	}
	err = a.tokensRepo.DbPutTokens(context.Background(), usr.ID, accessToken, refreshToken)
	if err != nil {
		log.Error("Error while storing tokens", "err", err)
		return getUserDto.UserTokens{}, err
	}
	return getUserDto.UserTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

//func (a *AuthService) RefreshTokens(tokens *refresh_tokens.RefreshTokensDto) (refresh_tokens.RefreshTokensDto, error) {
//	const op = "internal/lib/services/auth_service.go/RefreshTokens"
//	log := a.log.With(
//		slog.String("operation", op),
//	)
//	userId, err := a.tokensRepo.DbGetTokens(context.Background(), tokens.AccessToken, tokens.RefreshToken)
//	if err != nil {
//		log.Error("Error while fetching tokens", "err", err)
//		return tokens.RefreshTokensDto{}, err
//	}
//	accessToken, err := jwt_tokens.CreateAccessToken(userId, a.secretKey, a.log)
//	if err != nil {
//		log.Error("Error while creating access token", "err", err)
//		return tokens.RefreshTokensDto{}, err
//	}
//	refreshToken, err := jwt_tokens.CreateRefreshToken(a.log)
//	if err != nil {
//		log.Error("Error while creating refresh token", "err", err)
//		return tokens.RefreshTokensDto{}, err
//	}
//	//	TODO написать функцию обновления строкчик токена по юзер айди
//}
