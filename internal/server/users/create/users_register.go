package users

import (
	"booker/internal/lib/api/models"
	resp "booker/internal/lib/api/response"
	users2 "booker/internal/server/users"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"net/http"
)

// User Структура пользователя для создания

func CreateUser(log *slog.Logger, dbPoll *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "server/users.CreateUser"
		log = log.With(
			slog.String("operation", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
			slog.String("url", r.URL.Path))
		usrCreate := users2.NewUsersDB(dbPoll)

		var user models.UserCreate
		err := render.DecodeJSON(r.Body, &user)
		if err != nil {
			log.Error("Error while decoding request body", "err", err)
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}

		//Валидация
		//TODO Посмотреть где ещё создаются валидаторы, и если их много, то нужно вынести инициализацию валидатора глобально для повышения оптимизации
		if err = validator.New().Struct(&user); err != nil {
			validationErrors := err.(validator.ValidationErrors)
			log.Error("Error validating request body", "err", validationErrors)
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, resp.ValidationError(validationErrors))
			return
		}

		//	Хешируем пароль
		passwordHash, err := users2.HashUserPassword(user.Password, log)
		if err != nil {
			log.Error("Error while hashing password", "err", err)
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}

		user.Password = passwordHash
		//Записываем в бд
		userId, err := usrCreate.Create(r.Context(), &user)
		if err != nil {
			log.Error("Error while creating user", "err", err)
			if errors.Is(err, users2.ErrEmailAlreadyExists) {
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r, resp.Error(
					err.Error()))
				return
			}
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, resp.Error(
				err.Error()))
			return
		}

		log.Info("Created user", "user id", userId)
		render.Status(r, http.StatusCreated)
		render.JSON(w, r, models.CreateUserResponse{
			resp.OK(),
			userId,
		})
	}
}
