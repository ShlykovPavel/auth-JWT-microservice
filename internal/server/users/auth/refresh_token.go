package auth

import (
	"booker/internal/lib/api/models/tokens/refresh_tokens"
	resp "booker/internal/lib/api/response"
	"booker/internal/lib/services"
	"booker/internal/storage/database/repositories/auth_db"
	"booker/internal/storage/database/repositories/users_db"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"net/http"
)

func RefreshTokenHandler(log *slog.Logger, dbPool *pgxpool.Pool, secretKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "server/users/auth/RefreshTokenHandler"
		log = log.With(
			slog.String("op", op))
		usersRepository := users_db.NewUsersDB(dbPool, log)
		tokensRepository := auth_db.NewTokensRepositoryImpl(dbPool, log)
		// Инициализируем сервис аутентификации
		authService := services.NewAuthService(usersRepository, tokensRepository, log, secretKey)

		// Декодируем json в структуру дто
		var refreshDto refresh_tokens.RefreshTokensDto
		err := render.DecodeJSON(r.Body, &refreshDto)
		if err != nil {
			log.Error("Error while decoding json to RefreshTokensDto struct", "Error", err)
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("Error while reading request body"))
			return
		}

		// Валидируем полученные поля в структуре
		err = validator.New().Struct(&refreshDto)
		if err != nil {
			validationErrors := err.(validator.ValidationErrors)
			log.Error("Error while validating request body", "err", validationErrors)
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.ValidationError(validationErrors))
			return
		}

		newTokens, err := authService.RefreshTokens(&refreshDto)
		if err != nil {
			log.Error("Error while updating tokens", "err", err)
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}
		//Возвращаем новые токены
		render.Status(r, http.StatusOK)
		render.JSON(w, r, refresh_tokens.RefreshTokensDto{AccessToken: newTokens.AccessToken, RefreshToken: newTokens.RefreshToken})
		return
	}
}
