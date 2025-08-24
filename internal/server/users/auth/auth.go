package auth

import (
	"context"
	"errors"
	"github.com/ShlykovPavel/auth-JWT-microservice/internal/lib/api/body"
	resp "github.com/ShlykovPavel/auth-JWT-microservice/internal/lib/api/response"
	"github.com/ShlykovPavel/auth-JWT-microservice/internal/lib/services"
	"github.com/ShlykovPavel/auth-JWT-microservice/internal/storage/database/repositories/users_db"
	"github.com/ShlykovPavel/auth-JWT-microservice/models/tokens"
	"github.com/ShlykovPavel/auth-JWT-microservice/models/users/get_user"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"time"
)

var ErrIncorrectCredentials = errors.New("invalid email or password")

// AuthenticationHandler godoc
// @Summary Логин
// @Description Логинит пользователя. Выдаёт access и refresh токены
// @Tags Users
// @Param input body get_user.AuthUser true "Данные пользователя"
// @Success 200 {object} tokens.RefreshTokensDto
// @Router /login [post]
func AuthenticationHandler(log *slog.Logger, timeout time.Duration, authService *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "server/users/auth/AuthentificationHandler"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
			slog.String("url", r.URL.Path),
		)

		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		var user get_user.AuthUser
		//Парсим тело запроса из json
		if err := body.DecodeAndValidateJson(r, &user); err != nil {
			log.Error("Error while decoding request body", "err", err)
			resp.RenderResponse(w, r, http.StatusBadRequest, resp.Error(err.Error()))
			return
		}

		//TODO Посмотреть что можно сделать с телом ответа при валидации полей (приходит 2 json)
		authTokens, err := authService.Authentication(&user, ctx)
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
