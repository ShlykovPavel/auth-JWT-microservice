package users

import (
	usersDto "booker/internal/lib/api/models/users/create_user"
	resp "booker/internal/lib/api/response"
	users "booker/internal/server/users"
	"booker/internal/storage/database/repositories/users_db"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"net/http"
)

func CreateUser(log *slog.Logger, dbPoll *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "server/users.CreateUser"
		log = log.With(
			slog.String("operation", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
			slog.String("url", r.URL.Path))
		usrCreate := users_db.NewUsersDB(dbPoll, log)

		var user usersDto.UserCreate
		err := render.DecodeJSON(r.Body, &user)
		if err != nil {
			log.Error("Error while decoding request body", "err", err)
			resp.RenderResponse(w, r, http.StatusBadRequest, resp.Error(err.Error()))
			return
		}

		//Валидация
		//TODO Посмотреть где ещё создаются валидаторы, и если их много, то нужно вынести инициализацию валидатора глобально для повышения оптимизации
		if err = validator.New().Struct(&user); err != nil {
			validationErrors := err.(validator.ValidationErrors)
			log.Error("Error validating request body", "err", validationErrors)
			resp.RenderResponse(w, r, http.StatusBadRequest, resp.ValidationError(validationErrors))
			return
		}

		//	Хешируем пароль
		passwordHash, err := users.HashUserPassword(user.Password, log)
		if err != nil {
			log.Error("Error while hashing password", "err", err)
			resp.RenderResponse(w, r, http.StatusInternalServerError, resp.Error(err.Error()))
			return
		}

		user.Password = passwordHash
		//Записываем в бд
		userId, err := usrCreate.CreateUser(r.Context(), &user)
		if err != nil {
			log.Error("Error while creating user", "err", err)
			if errors.Is(err, users_db.ErrEmailAlreadyExists) {
				resp.RenderResponse(w, r, http.StatusBadRequest, resp.Error(
					err.Error()))
				return
			}
			resp.RenderResponse(w, r, http.StatusInternalServerError, resp.Error(err.Error()))
			return
		}

		log.Info("Created user", "user id", userId)
		resp.RenderResponse(w, r, http.StatusCreated, usersDto.CreateUserResponse{
			resp.OK(),
			userId,
		})
	}
}
