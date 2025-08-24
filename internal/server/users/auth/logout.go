package auth

import (
	"context"
	"errors"
	"github.com/ShlykovPavel/auth-JWT-microservice/internal/lib/api/body"
	resp "github.com/ShlykovPavel/auth-JWT-microservice/internal/lib/api/response"
	"github.com/ShlykovPavel/auth-JWT-microservice/internal/lib/services"
	"github.com/ShlykovPavel/auth-JWT-microservice/models/tokens"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/jackc/pgx/v5"
	"log/slog"
	"net/http"
	"time"
)

// LogoutHandler godoc
// @Summary logout user
// @Description Удаляет сессию пользователя
// @Tags Users
// @Param input body tokens.LogoutRequest true "Токены"
// @Success 204
// @Router /logout [post]
func LogoutHandler(log *slog.Logger, timeout time.Duration, authService *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "server/users/auth/LogoutHandler"
		log = log.With(
			slog.String("op", op),
			slog.String("url", r.URL.String()),
			slog.String("requestId", middleware.GetReqID(r.Context())))
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		// Декодируем json в структуру дто
		var logoutDto tokens.LogoutRequest
		err := body.DecodeAndValidateJson(r, &logoutDto)
		if err != nil {
			log.Error("Error while decoding json to RefreshTokensDto struct", "Error", err)
			resp.RenderResponse(w, r, http.StatusBadRequest, resp.Error("Error while reading request body"))
			return
		}

		err = authService.Logout(&logoutDto, ctx)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				log.Debug("Session not found", "refresh token", logoutDto.RefreshToken)
				resp.RenderResponse(w, r, http.StatusUnauthorized, resp.Error(ErrSessionNotFound.Error()))
				return
			}
			log.Error("Error while delete token", "err", err)
			resp.RenderResponse(w, r, http.StatusInternalServerError, resp.Error(err.Error()))
			return
		}
		log.Info("Logout successful, returning 204")
		render.NoContent(w, r)
		return
	}
}
