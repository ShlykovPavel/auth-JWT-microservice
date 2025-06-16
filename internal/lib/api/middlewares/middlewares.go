package middlewares

import (
	"booker/internal/lib/api/authorization"
	resp "booker/internal/lib/api/response"
	"fmt"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"strings"
)

// AuthMiddleware проверяет токен авторизации при выполнении запроса
//
// # При успехе передаёт обработку следующему хендлеру
//
// При ошибке возвращает статус код 401 и ошибку
func AuthMiddleware(secretKey string, log *slog.Logger) func(next http.Handler) http.Handler {
	const op = "internal/lib/api/middlewares/middlewares.go/AuthMiddleware"
	log = log.With(slog.String("op", op))
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				renderUnauthorized(w, r, log, "Authorization header is missing")
				return
			}
			const bearerPrefix = "Bearer "
			if !strings.HasPrefix(authHeader, bearerPrefix) {
				renderUnauthorized(w, r, log, "Authorization header is invalid")
				return
			}

			tokenString := strings.TrimPrefix(authHeader, bearerPrefix)

			_, err := authorization.Authorization(tokenString, secretKey)
			if err != nil {
				renderUnauthorized(w, r, log, fmt.Sprintf("Authorization token is invalid: %v", err))
				return
			}
			log.Debug("Authorization token is valid")
			next.ServeHTTP(w, r)
		})

	}
}

func renderUnauthorized(w http.ResponseWriter, r *http.Request, log *slog.Logger, msg string) {
	log.Error(msg)
	render.Status(r, http.StatusUnauthorized)
	render.JSON(w, r, resp.Error(msg))
}
