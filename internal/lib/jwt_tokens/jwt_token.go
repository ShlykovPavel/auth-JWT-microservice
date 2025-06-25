package jwt_tokens

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"log/slog"
	"strconv"
	"time"
)

func CreateAccessToken(userID int64, secretKey string, userRole string, duration time.Duration, log *slog.Logger) (string, error) {
	const op = "internal/lib/jwt_tokens/jwt_token.go/CreateAccessToken"
	log = log.With(
		slog.String("op", op),
		slog.String("user_id", strconv.FormatInt(userID, 10)))

	if len([]byte(secretKey)) < 32 {
		log.Error("secret key too short")
		return "", errors.New("secret key too short")
	}
	// Создаем claims
	claims := jwt.MapClaims{
		"user_role": userRole,
		"sub":       userID,                          // Идентификатор пользователя
		"iat":       time.Now().Unix(),               // Время выпуска токена
		"exp":       time.Now().Add(duration).Unix(), // Время истечения (1 час)
	}
	// Создаем токен с алгоритмом HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Подписываем токен секретным ключом
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		log.Error("token generate failed", op, err)
		return "", err
	}
	return tokenString, nil
}

// VerifyToken verifies a JWT token and returns its claims
func VerifyToken(tokenString string, secretKey string) (jwt.MapClaims, error) {
	// Парсим токен
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверяем алгоритм подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}
	// Проверяем, валиден ли токен
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	// Извлекаем claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	return claims, nil
}

func CreateRefreshToken(log *slog.Logger) (string, error) {
	log = log.With(
		slog.String("op", "internal/lib/jwt_tokens/jwt_token.go/CreateRefreshToken"),
	)
	byteArray := make([]byte, 32)
	_, err := rand.Read(byteArray)
	if err != nil {
		return "", err
	}
	// Кодируем в base64 и обрезаем до нужной длины
	return base64.URLEncoding.EncodeToString(byteArray)[:32], nil
}
