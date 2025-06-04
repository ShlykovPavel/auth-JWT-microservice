package users

import (
	"golang.org/x/crypto/bcrypt"
	"log/slog"
)

// HashUserPassword хеширует пароль
// при хешировании автоматически генерируется соль
// ограничение длинны пароля 72 байта (можно выставить 64 символа)
func HashUserPassword(password string, log *slog.Logger) (string, error) {
	log = log.With(
		slog.String("operation", "server/users.HashUserPassword"))
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("Hashing password is failed. Error: ", "err", err)
		return "", err
	}
	return string(hashedPassword), nil
}

// ComparePassword Проверяет пароль на соответствие
//
// Из переданного хеша пароля он сам достаёт соль, и хеширует пароль с этой солью
// Если пароли совпадают, то ничего не возвращается
func ComparePassword(hashedPassword string, password string, log *slog.Logger) bool {
	log = log.With(
		slog.String("operation", "server/users.ComparePassword"))
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		log.Error("Password is incorrect. Error: ", "err", err)
		return false
	}
	return true
}
