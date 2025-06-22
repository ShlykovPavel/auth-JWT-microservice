package auth

import (
	"booker/internal/lib/api/models/tokens"
	resp "booker/internal/lib/api/response"
	"booker/internal/lib/services"
	"booker/internal/storage/database/repositories/auth_db"
	"booker/internal/storage/database/repositories/users_db"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"net/http"
	"time"
)

func LogoutHandler(log *slog.Logger, dbPool *pgxpool.Pool, secretKey string, jwtDuration time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "server/users/auth/LogoutHandler"
		log = log.With(
			slog.String("op", op),
			slog.String("url", r.URL.String()),
			slog.String("requestId", middleware.GetReqID(r.Context())))
		usersRepository := users_db.NewUsersDB(dbPool, log)
		tokensRepository := auth_db.NewTokensRepositoryImpl(dbPool, log)
		// Инициализируем сервис аутентификации
		authService := services.NewAuthService(usersRepository, tokensRepository, log, secretKey, jwtDuration)

		// Декодируем json в структуру дто
		var logoutDto tokens.LogoutRequest
		err := render.DecodeJSON(r.Body, &logoutDto)
		if err != nil {
			log.Error("Error while decoding json to RefreshTokensDto struct", "Error", err)
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error("Error while reading request body"))
			return
		}

		// Валидируем полученные поля в структуре
		err = validator.New().Struct(&logoutDto)
		if err != nil {
			validationErrors := err.(validator.ValidationErrors)
			log.Error("Error while validating request body", "err", validationErrors)
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.ValidationError(validationErrors))
			return
		}

		err = authService.Logout(&logoutDto)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				log.Debug("Session not found", "refresh token", logoutDto.RefreshToken)
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, resp.Error(ErrSessionNotFound.Error()))
				return
			}
			log.Error("Error while delete token", "err", err)
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}
		log.Info("Logout successful, returning 204")
		render.NoContent(w, r)
		return
	}
}
