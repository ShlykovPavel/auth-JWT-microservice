package get_user_tests

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ShlykovPavel/auth-JWT-microservice/internal/lib/api/response"
	"github.com/ShlykovPavel/auth-JWT-microservice/internal/server/users/get_user"
	"github.com/ShlykovPavel/auth-JWT-microservice/internal/storage/database/repositories/users_db"

	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ShlykovPavel/auth-JWT-microservice/models/users/get_user_by_id"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetUser(t *testing.T) {
	tests := []struct {
		name           string
		userId         string
		setupMock      func(*users_db.MockUserRepository)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:   "success get user",
			userId: "1",
			setupMock: func(mockRepo *users_db.MockUserRepository) {
				mockRepo.On("GetUser", mock.Anything, int64(1)).Return(users_db.UserInfo{
					Email:     "ryanGosling@gmail.com",
					FirstName: "Ryan",
					LastName:  "Gosling",
					Phone:     "+1234567890",
				}, nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody: get_user_by_id.UserInfo{
				Email:     "ryanGosling@gmail.com",
				FirstName: "Ryan",
				LastName:  "Gosling",
				Phone:     "+1234567890",
			},
		},
		{
			name:   "empty user id",
			userId: "",
			setupMock: func(mockRepo *users_db.MockUserRepository) {
				// Нет вызова мока, так как хендлер не доходит до обращения к репозиторию
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   response.Error("User ID is required"),
		},
		{
			name:   "invalid user id",
			userId: "invalid",
			setupMock: func(mockRepo *users_db.MockUserRepository) {
				// Нет вызова мока, так как хендлер не доходит до обращения к репозиторию
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   response.Error("Invalid user ID"),
		},
		{
			name:   "user not found",
			userId: "1",
			setupMock: func(mockRepo *users_db.MockUserRepository) {
				mockRepo.On("GetUser", mock.Anything, int64(1)).Return(users_db.UserInfo{}, users_db.ErrUserNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   response.Error("User not found"),
		},
		{
			name:   "internal server error",
			userId: "1",
			setupMock: func(mockRepo *users_db.MockUserRepository) {
				mockRepo.On("GetUser", mock.Anything, int64(1)).Return(users_db.UserInfo{}, errors.New("database error")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   response.Error("Something went wrong, while getting user"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := slog.Default()
			timeout := 5 * time.Second
			mockRepo := new(users_db.MockUserRepository)
			handler := get_user.GetUserById(logger, mockRepo, timeout)

			// Настраиваем мок
			test.setupMock(mockRepo)

			// Создаем запрос
			req := httptest.NewRequest(http.MethodGet, "/auth/"+test.userId, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", test.userId)
			req = req.WithContext(context.WithValue(context.Background(), chi.RouteCtxKey, rctx))
			req = req.WithContext(context.WithValue(req.Context(), middleware.RequestIDKey, "test-id"))
			w := httptest.NewRecorder()

			// Вызываем хендлер
			handler.ServeHTTP(w, req)

			// Проверяем статус
			require.Equal(t, test.expectedStatus, w.Code, "unexpected HTTP status")

			expectedJSON, err := json.Marshal(test.expectedBody)
			require.NoError(t, err, "failed to marshal expected body to JSON")

			// Сравниваем JSON-строки
			require.JSONEq(t, string(expectedJSON), w.Body.String(), "unexpected response body")

			// Проверяем, что все ожидаемые вызовы мока выполнены
			mockRepo.AssertExpectations(t)
		})
	}
}
