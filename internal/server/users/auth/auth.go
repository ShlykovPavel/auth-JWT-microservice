package auth

import (
	"booker/internal/lib/api/models"
	resp "booker/internal/lib/api/response"
	"booker/internal/lib/services"
	"booker/internal/server/users/users_db"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	validator "github.com/go-playground/validator"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"net/http"
)

var ErrIncorrectCredentials = errors.New("invalid email or password")

func AuthenticationHandler(log *slog.Logger, dbPool *pgxpool.Pool, secretKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "server/users/auth/AuthentificationHandler"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
			slog.String("url", r.URL.Path),
		)
		// Инициализируем сервис аутентификации
		authService := services.NewAuthService(users_db.NewUsersDB(dbPool, log), log, secretKey)

		var user models.AuthUser
		//Парсим тело запроса из json
		if err := render.DecodeJSON(r.Body, &user); err != nil {
			log.Error("Error while decoding request body", "err", err)
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}
		//Валидируем полученное тело запроса
		if err := validator.New().Struct(user); err != nil {
			validationErrors := err.(validator.ValidationErrors)
			log.Error("Error while validating request body", "err", validationErrors)
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.ValidationError(validationErrors))
		}
		userInfo, err := authService.Authentication(&user)
		if err != nil {
			if errors.Is(err, users_db.ErrUserNotFound) {
				log.Debug("User not found", "user", user)
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, resp.Error(ErrIncorrectCredentials.Error()))
				return
			}
			if errors.Is(err, services.ErrWrongPassword) {
				log.Debug("Password is incorrect", "user", user)
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, resp.Error(ErrIncorrectCredentials.Error()))
			}
			log.Error("Error while Authentification user: ", "err", err)
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}

		//	TODO Добавить вывод ответа с токенами

	}
}
