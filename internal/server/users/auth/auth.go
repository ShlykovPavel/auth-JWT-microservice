package auth

import (
	"booker/internal/lib/api/models"
	resp "booker/internal/lib/api/response"
	"booker/internal/server/users"
	"booker/internal/server/users/users_db"
	"context"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	validator "github.com/go-playground/validator"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"net/http"
)

var ErrWrongPassword = errors.New("Password is incorrect ")
var ErrIncorrectCredentials = errors.New("invalid email or password")

func AuthenticationHandler(log *slog.Logger, dbPool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "server/users/auth/AuthentificationHandler"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
			slog.String("url", r.URL.Path),
		)
		//Берём конекшен к бд и пула
		usrCreate := users_db.NewUsersDB(dbPool)

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
		userInfo, err := Authentication(&user, log, usrCreate)
		if err != nil {
			if errors.Is(err, users_db.ErrUserNotFound) {
				log.Debug("User not found", "user", user)
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, resp.Error(ErrIncorrectCredentials.Error()))
				return
			}
			if errors.Is(err, ErrWrongPassword) {
				log.Debug("Password is incorrect", "user", user)
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, resp.Error(ErrIncorrectCredentials.Error()))
			}
			log.Error("Error while Authentification user: ", "err", err)
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}

	}
}

func Authentication(user *models.AuthUser, log *slog.Logger, dbConn *users_db.UserRepository) (users_db.UserInfo, error) {
	const op = "server/users/auth/Authentification"
	log = log.With(
		slog.String("operation", op),
		slog.String("request email: ", user.Email))
	// Проверяем что пользователь есть в БД
	usr, err := dbConn.GetUser(context.Background(), user.Email)
	if err != nil {
		if errors.Is(err, users_db.ErrUserNotFound) {
			log.Debug("UserInfo not found", "user", user)
			return users_db.UserInfo{}, err
		}
		log.Error("Error while fetching user", "err", err)
		return users_db.UserInfo{}, err
	}
	// Проверяем что нам предоставили правильный пароль
	ok := users.ComparePassword(usr.PasswordHash, user.Password, log)
	if !ok {
		return users_db.UserInfo{}, ErrWrongPassword
	}

}
