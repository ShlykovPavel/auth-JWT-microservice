package jwt_tokens

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

// TODO Добавить sectretkey в конфиг
// Сделать созадние access и refresh токена
func CreateToken(userID int64, secretKey string) (string, error) {
	// Создаем claims
	claims := jwt.MapClaims{
		"sub": userID,                           // Идентификатор пользователя
		"iat": time.Now().Unix(),                // Время выпуска токена
		"exp": time.Now().Add(time.Hour).Unix(), // Время истечения (1 час)
	}
	// Создаем токен с алгоритмом HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Подписываем токен секретным ключом
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func VerifyToken(tokenString string, secretKey string) (jwt.MapClaims, error) {
	// Парсим токен
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверяем алгоритм подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, err
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
