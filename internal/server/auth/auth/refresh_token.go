package auth

import (
	"context"
	"errors"
	"github.com/ShlykovPavel/auth-JWT-microservice/internal/lib/api/body"
	resp "github.com/ShlykovPavel/auth-JWT-microservice/internal/lib/api/response"
	"github.com/ShlykovPavel/auth-JWT-microservice/internal/lib/services"
	"github.com/ShlykovPavel/auth-JWT-microservice/models/tokens"
	"github.com/jackc/pgx/v5"
	"log/slog"
	"net/http"
	"time"
)

var ErrSessionNotFound = errors.New("Session not found")

// RefreshTokenHandler godoc
// @Summary refresh token
// @Description Обновляет access token и выдаёт новый refresh token
// @Tags Users
// @Param input body tokens.RefreshTokensDto true "Токены"
// @Success 200 {object} tokens.RefreshTokensDto
// @Router /refresh [post]
func RefreshTokenHandler(log *slog.Logger, timeout time.Duration, authService *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "server/auth/auth/RefreshTokenHandler"
		log = log.With(slog.String("op", op))
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		// Декодируем json в структуру дто
		var refreshDto tokens.RefreshTokensDto
		err := body.DecodeAndValidateJson(r, &refreshDto)
		if err != nil {
			log.Error("Error while decoding json to RefreshTokensDto struct", "Error", err)
			resp.RenderResponse(w, r, http.StatusBadRequest, resp.Error("Error while reading request body"))
			return
		}

		newTokens, err := authService.RefreshTokens(&refreshDto, ctx)
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
