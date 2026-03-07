package get_user_list

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/ShlykovPavel/auth-JWT-microservice/internal/lib/api/query_params"
	resp "github.com/ShlykovPavel/auth-JWT-microservice/internal/lib/api/response"
	"github.com/ShlykovPavel/auth-JWT-microservice/internal/lib/services/user_service"
	"github.com/ShlykovPavel/auth-JWT-microservice/internal/storage/database/repositories/users_db"
	_ "github.com/ShlykovPavel/auth-JWT-microservice/models/users/get_users_list"
)

// GetUserList godoc
// @Summary Получить список пользователей
// @Description Получить список пользователей
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} get_users_list.UsersList
// @Router /users [get]
func GetUserList(logger *slog.Logger, userDbRepository users_db.UserRepository, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "internal/server/auth/get_user/get_user_list/get_user_list_handler.go/get_user_list"
		log := logger.With(slog.String("op", op))

		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()
		requestQuery := r.URL.Query()

		queryParser := &query_params.DefaultSortParser{
			ValidSortFields: []string{"id", "first_name", "last_name", "email"},
		}
		parsedQuery, err := query_params.ParseStandardQueryParams(requestQuery, log, queryParser)
		if err != nil {
			log.Error("Ошибка парсинга параметров", "error", err, "request", requestQuery)
			resp.RenderResponse(w, r, http.StatusBadRequest, resp.Error("Ошибка параметров запроса"))
			return
		}

		userList, err := user_service.GetUserList(log, userDbRepository, ctx, parsedQuery)
		if err != nil {
			log.Error("Error while getting user list", "error", err)
			resp.RenderResponse(w, r, http.StatusInternalServerError, resp.Error("Something went wrong, while getting user list"))
			return
		}
		log.Debug("Successful get auth list")
		resp.RenderResponse(w, r, http.StatusOK, userList)
		return

	}
}
