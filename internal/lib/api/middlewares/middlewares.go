package middlewares

import (
	"booker/internal/lib/api/authorization"
	resp "booker/internal/lib/api/response"
	"context"
	"fmt"
	"github.com/go-chi/render"
	"github.com/golang-jwt/jwt/v5"
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

			claims, err := authorization.Authorization(tokenString, secretKey)
			if err != nil {
				renderUnauthorized(w, r, log, fmt.Sprintf("Authorization token is invalid: %v", err))
				return
			}
			log.Debug("Authorization token is valid", slog.Any("claims", claims))
			ctx := context.WithValue(r.Context(), "tokenClaims", claims)

			next.ServeHTTP(w, r.WithContext(ctx))

		})

	}
}

func renderUnauthorized(w http.ResponseWriter, r *http.Request, log *slog.Logger, msg string) {
	log.Error(msg)
	render.Status(r, http.StatusUnauthorized)
	render.JSON(w, r, resp.Error(msg))
}

func AuthAdminMiddleware(secretKey string, log *slog.Logger) func(next http.Handler) http.Handler {
	const op = "internal/lib/api/middlewares/middlewares.go/AuthAdminMiddleware"
	log = log.With(slog.String("op", op))

	return func(next http.Handler) http.Handler {
		// Используем AuthMiddleware для проверки авторизации
		return AuthMiddleware(secretKey, log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Извлекаем claims из контекста
			claims, ok := r.Context().Value("tokenClaims").(jwt.MapClaims)
			if !ok {
				log.Error("Failed to retrieve claims from context")
				render.Status(r, http.StatusInternalServerError)
				render.JSON(w, r, resp.Error("Internal server error"))
				return
			}

			value, ok := claims["user_role"].(string)
			if !ok {
				log.Error("Failed to retrieve user role from context", "user_role", claims["user_role"])
				resp.RenderResponse(w, r, 403, nil)
			}
			// Проверяем, является ли пользователь администратором
			if value == "admin" {
				log.Debug("User is authorized and has admin privileges")
				next.ServeHTTP(w, r)
			} else {
				log.Debug("User is NOT authorized and has admin privileges")
				resp.RenderResponse(w, r, 403, nil)
			}
		}))
	}
}
