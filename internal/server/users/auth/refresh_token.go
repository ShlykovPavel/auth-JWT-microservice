package auth

import (
	"booker/internal/lib/api/models/tokens"
	resp "booker/internal/lib/api/response"
	"booker/internal/lib/services"
	"booker/internal/storage/database/repositories/auth_db"
	"booker/internal/storage/database/repositories/users_db"
	"errors"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"net/http"
	"time"
)

var ErrSessionNotFound = errors.New("Session not found")

func RefreshTokenHandler(log *slog.Logger, dbPool *pgxpool.Pool, secretKey string, jwtDuration time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "server/users/auth/RefreshTokenHandler"
		log = log.With(slog.String("op", op))
		usersRepository := users_db.NewUsersDB(dbPool, log)
		tokensRepository := auth_db.NewTokensRepositoryImpl(dbPool, log)
		// Инициализируем сервис аутентификации
		authService := services.NewAuthService(usersRepository, tokensRepository, log, secretKey, jwtDuration)

		// Декодируем json в структуру дто
		var refreshDto tokens.RefreshTokensDto
		err := render.DecodeJSON(r.Body, &refreshDto)
		if err != nil {
			log.Error("Error while decoding json to RefreshTokensDto struct", "Error", err)
			resp.RenderResponse(w, r, http.StatusBadRequest, resp.Error("Error while reading request body"))
			return
		}

		// Валидируем полученные поля в структуре
		err = validator.New().Struct(&refreshDto)
		if err != nil {
			validationErrors := err.(validator.ValidationErrors)
			log.Error("Error while validating request body", "err", validationErrors)
			resp.RenderResponse(w, r, http.StatusBadRequest, resp.ValidationError(validationErrors))
			return
		}

		newTokens, err := authService.RefreshTokens(&refreshDto)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				log.Debug("Session not found", "refresh token", refreshDto.RefreshToken)
				resp.RenderResponse(w, r, http.StatusUnauthorized, resp.Error(ErrSessionNotFound.Error()))
				return
			}
			log.Error("Error while updating tokens", "err", err)
			resp.RenderResponse(w, r, http.StatusInternalServerError, resp.Error(err.Error()))
			return
		}
		//Возвращаем новые токены
		resp.RenderResponse(w, r, http.StatusOK, tokens.RefreshTokensDto{AccessToken: newTokens.AccessToken, RefreshToken: newTokens.RefreshToken})
		return
	}
}
