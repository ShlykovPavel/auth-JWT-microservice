package users

import (
	"context"
	"errors"
	"github.com/ShlykovPavel/auth-JWT-microservice/internal/lib/api/body"
	resp "github.com/ShlykovPavel/auth-JWT-microservice/internal/lib/api/response"
	users "github.com/ShlykovPavel/auth-JWT-microservice/internal/server/users"
	"github.com/ShlykovPavel/auth-JWT-microservice/internal/storage/database/repositories/users_db"
	"github.com/ShlykovPavel/auth-JWT-microservice/models/users/create_user"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"time"
)

// CreateUser godoc
// @Summary Создать пользователя
// @Description Регистрирует пользователя в системе
// @Tags Users
// @Param input body create_user.UserCreate true "Данные пользователя"
// @Success 201 {object} create_user.CreateUserResponse
// @Router /user/register [post]
func CreateUser(log *slog.Logger, userRepo users_db.UserRepository, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "server/users.CreateUser"
		log = log.With(
			slog.String("operation", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
			slog.String("url", r.URL.Path))

		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		var user create_user.UserCreate
		err := body.DecodeAndValidateJson(r, &user)
		if err != nil {
			log.Error("Error while decoding request body", "err", err)
			resp.RenderResponse(w, r, http.StatusBadRequest, resp.Error(err.Error()))
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
		userId, err := userRepo.CreateUser(ctx, &user)
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
		resp.RenderResponse(w, r, http.StatusCreated, create_user.CreateUserResponse{
			resp.OK(),
			userId,
		})
	}
}
