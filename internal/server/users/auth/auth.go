package auth

import (
	"booker/internal/lib/api/models/tokens"
	"booker/internal/lib/api/models/users/get_user"
	resp "booker/internal/lib/api/response"
	"booker/internal/lib/services"
	"booker/internal/storage/database/repositories/auth_db"
	"booker/internal/storage/database/repositories/users_db"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	validator "github.com/go-playground/validator"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"net/http"
	"time"
)

var ErrIncorrectCredentials = errors.New("invalid email or password")

func AuthenticationHandler(log *slog.Logger, dbPool *pgxpool.Pool, secretKey string, jwtDuration time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "server/users/auth/AuthentificationHandler"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
			slog.String("url", r.URL.Path),
		)

		userRepository := users_db.NewUsersDB(dbPool, log)
		tokensRepository := auth_db.NewTokensRepositoryImpl(dbPool, log)
		// Инициализируем сервис аутентификации
		authService := services.NewAuthService(userRepository, tokensRepository, log, secretKey, jwtDuration)

		var user get_user.AuthUser
		//Парсим тело запроса из json
		if err := render.DecodeJSON(r.Body, &user); err != nil {
			log.Error("Error while decoding request body", "err", err)
			resp.RenderResponse(w, r, http.StatusBadRequest, resp.Error(err.Error()))
			return
		}
		//Валидируем полученное тело запроса
		if err := validator.New().Struct(user); err != nil {
			validationErrors := err.(validator.ValidationErrors)
			log.Error("Error while validating request body", "err", validationErrors)
			resp.RenderResponse(w, r, http.StatusBadRequest, resp.ValidationError(validationErrors))
		}
		//TODO Посмотреть что можно сделать с телом ответа при валидации полей (приходит 2 json)
		authTokens, err := authService.Authentication(&user)
		if err != nil {
			if errors.Is(err, users_db.ErrUserNotFound) {
				log.Debug("User not found", "user", user)
				resp.RenderResponse(w, r, http.StatusUnauthorized, resp.Error(ErrIncorrectCredentials.Error()))
				return
			} else if errors.Is(err, services.ErrWrongPassword) {
				log.Debug("Password is incorrect", "user", user)
				resp.RenderResponse(w, r, http.StatusUnauthorized, resp.Error(ErrIncorrectCredentials.Error()))
			}
			log.Error("Error while Authentification user: ", "err", err)
			resp.RenderResponse(w, r, http.StatusInternalServerError, resp.Error(err.Error()))
			return
		}
		log.Debug("User authenticated", "user", user)
		resp.RenderResponse(w, r, http.StatusOK, tokens.RefreshTokensDto{
			AccessToken:  authTokens.AccessToken,
			RefreshToken: authTokens.RefreshToken,
		})

	}
}
