package create_user_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	validators "github.com/ShlykovPavel/auth-JWT-microservice/internal/lib/api/validator"
	users "github.com/ShlykovPavel/auth-JWT-microservice/internal/server/auth/register"
	"github.com/ShlykovPavel/auth-JWT-microservice/internal/storage/database/repositories/users_db"
	//"github.com/ShlykovPavel/auth-JWT-microservice/internal/storage/database/repositories/users_db/users_db"
	"github.com/ShlykovPavel/auth-JWT-microservice/models/users/create_user"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	tests := []struct {
		testName       string
		input          create_user.UserCreate
		setupMock      func(*users_db.MockUserRepository)
		expectedStatus int
		expectedBody   string
	}{
		{
			testName: "success creating user",
			input: create_user.UserCreate{
				FirstName: "Ryan",
				LastName:  "Gosling",
				Email:     "ryanGosling@gmail.com",
				Password:  "password",
				Phone:     "+78951235678",
			},
			setupMock: func(mockRepo *users_db.MockUserRepository) {
				mockRepo.On("CreateUser", mock.Anything, mock.MatchedBy(func(u *create_user.UserCreate) bool {
					return u.Email == "ryanGosling@gmail.com" && u.FirstName == "Ryan"
				})).Return(int64(123), nil).Once()
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"status":"OK","id":123}`,
		},
		{
			testName: "wrong type in field",
			input: create_user.UserCreate{
				FirstName: "",
				LastName:  "Gosling",
				Email:     "ryanGosling@gmail.com",
				Password:  "password",
				Phone:     "+78951235678",
			},
			setupMock: func(mockRepo *users_db.MockUserRepository) {
				// Не настраиваем мок так как будет ошибка
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"status":"ERROR","error":"field FirstName is required"}`,
		},
		{
			testName: "email already exists",
			input: create_user.UserCreate{
				FirstName: "Ryan",
				LastName:  "Gosling",
				Email:     "ryanGosling@gmail.com",
				Password:  "password",
				Phone:     "+78951235678",
			},
			setupMock: func(mockRepo *users_db.MockUserRepository) {
				mockRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*create_user.UserCreate")).
					Return(int64(0), users_db.ErrEmailAlreadyExists).Once()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"status":"ERROR","error":"Пользователь с email уже существует. "}`,
		},
		{
			testName: "timeout error",
			input: create_user.UserCreate{
				FirstName: "Ryan",
				LastName:  "Gosling",
				Email:     "ryanGosling@gmail.com",
				Password:  "password",
				Phone:     "+78951235678",
			},
			setupMock: func(mockRepo *users_db.MockUserRepository) {
				mockRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*create_user.UserCreate")).
					Run(func(args mock.Arguments) {
						time.Sleep(6 * time.Second)
					}).
					Return(int64(0), context.DeadlineExceeded).Once()
			},
			expectedStatus: http.StatusGatewayTimeout,
			expectedBody:   `{"status":"ERROR","error":"Request timed out or canceled"}`,
		},
	}

	if err := validators.InitValidator(); err != nil {
		fmt.Println("Failed to initialize validator")
	}
	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			logger := slog.Default()
			timeout := 5 * time.Second
			mockRepo := new(users_db.MockUserRepository)

			handler := users.CreateUser(logger, mockRepo, timeout)

			// Настраиваем мок
			test.setupMock(mockRepo)

			// Создаём запрос
			body, _ := json.Marshal(test.input)
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
			req = req.WithContext(context.WithValue(context.Background(), middleware.RequestIDKey, "test-id"))
			w := httptest.NewRecorder()

			// Вызываем хендлер
			handler.ServeHTTP(w, req)

			// Проверяем статус
			require.Equal(t, test.expectedStatus, w.Code, "unexpected HTTP status")

			// Проверяем тело ответа
			var respBody map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &respBody)
			require.NoError(t, err, "response should be valid JSON")
			require.Contains(t, w.Body.String(), test.expectedBody, "unexpected response body")

			// Проверяем, что все ожидаемые вызовы мока выполнены
			mockRepo.AssertExpectations(t)
		})
	}
}
