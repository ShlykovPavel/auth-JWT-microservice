package get_user

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	resp "github.com/ShlykovPavel/auth-JWT-microservice/internal/lib/api/response"
	"github.com/ShlykovPavel/auth-JWT-microservice/internal/lib/services/user_service"
	"github.com/ShlykovPavel/auth-JWT-microservice/internal/storage/database/repositories/users_db"
	"github.com/ShlykovPavel/auth-JWT-microservice/models/users/get_user_by_id"
	"github.com/go-chi/chi/v5"
)

// GetUserById godoc
// @Summary Получить пользователя по ID
// @Description Получить детальную информацию о пользователе
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID пользователя"
// @Success 200 {object} get_user_by_id.UserInfo
// @Router /user/{id} [get]
func GetUserById(logger *slog.Logger, userDbRepository users_db.UserRepository, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "server/auth/get_user/get_user_by_id_handler.go./GetUserById"
		log := logger.With(slog.String("op", op))
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

		userInfo, err := user_service.GetUser(logger, userDbRepository, id, ctx)
		if err != nil {
			if errors.Is(err, users_db.ErrUserNotFound) {
				resp.RenderResponse(w, r, http.StatusNotFound, resp.Error("User not found"))
				return
			}
			log.Error("Error while getting user by id", "err", err)
			resp.RenderResponse(w, r, http.StatusInternalServerError, resp.Error("Something went wrong, while getting user"))
			return
		}
		log.Debug("Successful get user by id", "user", userInfo)
		resp.RenderResponse(w, r, http.StatusOK, get_user_by_id.UserInfo{
			Email:     userInfo.Email,
			Phone:     userInfo.Phone,
			LastName:  userInfo.LastName,
			FirstName: userInfo.FirstName,
		})
		return

	}
}
