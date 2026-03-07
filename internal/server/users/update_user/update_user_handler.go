package update_user

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/ShlykovPavel/auth-JWT-microservice/internal/lib/api/body"
	resp "github.com/ShlykovPavel/auth-JWT-microservice/internal/lib/api/response"
	"github.com/ShlykovPavel/auth-JWT-microservice/internal/lib/services/user_service"
	"github.com/ShlykovPavel/auth-JWT-microservice/internal/storage/database/repositories/users_db"
	"github.com/ShlykovPavel/auth-JWT-microservice/models/users/create_user"
	"github.com/ShlykovPavel/auth-JWT-microservice/models/users/update_user"
	"github.com/go-chi/chi/v5"
)

// UpdateUserHandler godoc
// @Summary Обновить пользователя по ID
// @Description Обновить детальную информацию о пользователе
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID пользователя"
// @Param input body update_user.UpdateUserDto true "Данные пользователя"
// @Success 200 {object} create_user.CreateUserResponse
// @Router /users/{id} [put]
func UpdateUserHandler(log *slog.Logger, userRepository users_db.UserRepository, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "server/auth.UpdateUser"
		log = log.With(slog.String("op", op))

		userID := chi.URLParam(r, "id")
		if userID == "" {
			log.Error("User ID is empty")
			resp.RenderResponse(w, r, http.StatusBadRequest, resp.Error("User ID is required"))
			return
		}
		id, err := strconv.ParseInt(userID, 10, 64)
		if err != nil {
			log.Error("User ID is invalid", "error", err)
			resp.RenderResponse(w, r, http.StatusBadRequest, resp.Error("Invalid user ID"))
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		var UpdateUserDto update_user.UpdateUserDto
		err = body.DecodeAndValidateJson(r, &UpdateUserDto)
		if err != nil {
			log.Error("Failed decoding body", "err", err, "body", r.Body)
			resp.RenderResponse(w, r, http.StatusInternalServerError, resp.Error("Failed reading body"))
			return
		}

		err = user_service.UpdateUser(log, userRepository, ctx, UpdateUserDto, id)
		if err != nil {
			if errors.Is(err, users_db.ErrUserNotFound) {
				resp.RenderResponse(w, r, http.StatusNotFound, resp.Error(err.Error()))
				return
			}
			log.Error("Failed to update user", "err", err)
			resp.RenderResponse(w, r, http.StatusInternalServerError, resp.Error("Failed updating user"))
			return
		}
		log.Debug("Successfully updated user", "id", id)
		resp.RenderResponse(w, r, http.StatusOK, create_user.CreateUserResponse{UserID: id, Response: resp.OK()})

	}
}
