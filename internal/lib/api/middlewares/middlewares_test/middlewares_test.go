package middlewares_test

import (
	"booker/internal/lib/api/middlewares"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAuthMiddleware(t *testing.T) {
	tests := []struct {
		TestName           string
		AuthHeader         string
		expectedStatusCode int
		secretKey          string
		ExpectedBody       string
		NextCalled         bool
	}{
		{"Valid auth header", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIiwiaWF0IjoxNzMzNjAzOTMxLCJleHAiOjE4NDkxODc5MzF9.127K4MnLBcYL78KKrNmzlnDLTlWIwawv5MecxayDkds", 200, "256bitsvalid256bitsvalid256bitsvalid", "{Status: OK}", true},
		{"no auth header", "", 401, "256bitsvalid256bitsvalid256bitsvalid", "Authorization header is missing", false},
		{"expired auth token", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIiwiaWF0IjoxNzMzNjAzOTMxLCJleHAiOjE3MzM2MDM5MzF9.8m0HyNu-MON179DukOYCRHnlDriTU7y_4qhdor8DZJ8", 401, "256bitsvalid256bitsvalid256bitsvalid", "Authorization token is invalid: failed to parse token: token has invalid claims: token is expired", false},
		{"invalid auth token", "Bearer invalid", 401, "256bitsvalid256bitsvalid256bitsvalid", "Authorization token is invalid: failed to parse token: token is malformed: token contains an invalid number of segments", false},
		{"invalid sign in auth token", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIiwiaWF0IjoxNzMzNjAzOTMxLCJleHAiOjE4NDkxODc5MzF9.veHC9BFqcSFow5gOyTlalESHRKgdFEnn23FdFB3Z7dw", 401, "256bitsvalid256bitsvalid256bitsvalid", "Authorization token is invalid: failed to parse token: token signature is invalid: signature is invalid", false},
		{"no bearer prefix in auth token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIiwiaWF0IjoxNzMzNjAzOTMxLCJleHAiOjE4NDkxODc5MzF9.127K4MnLBcYL78KKrNmzlnDLTlWIwawv5MecxayDkds", 401, "256bitsvalid256bitsvalid256bitsvalid", "Authorization header is invalid", false},
		{"HS384 alg in auth token", "Bearer eyJhbGciOiJIUzM4NCIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIiwiaWF0IjoxNzMzNjAzOTMxLCJleHAiOjE3MzM2MDM5MzF9.O1GiEXnpUBebkncsV63FsitcVNzKeIr1LnrbJVLIm6z6wfRHK1AQR08FCoR-XDQz", 401, "256bitsvalid256bitsvalid256bitsvalid", "Authorization token is invalid: failed to parse token: token signature is invalid: signature is invalid", false},
	}
	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			//Инициализируем логгер
			log := slog.Default()
			// Создаём тестовый запрос
			request := httptest.NewRequest(http.MethodGet, "/url", nil)
			if test.AuthHeader != "" {
				request.Header.Set("Authorization", test.AuthHeader)
			}

			// Инициализируем переменную для чтения ответа от сервера
			responseRecorder := httptest.NewRecorder()
			//Следующий обработчик который вызывается при успешной авторизации
			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("{Status: OK}"))
			})
			//Регистрируем обработчик middleware
			middleware := middlewares.AuthMiddleware(test.secretKey, log)
			handler := middleware(next)

			//	Метод для вызова нашего созданного запроса и ответа
			handler.ServeHTTP(responseRecorder, request)

			// Проверяем статус
			if responseRecorder.Code != test.expectedStatusCode {
				t.Errorf("expected status %d, got %d", test.expectedStatusCode, responseRecorder.Code)
			}

			// Проверяем тело ответа
			if test.expectedStatusCode == http.StatusUnauthorized {
				var gotResp struct {
					Error string `json:"error"`
				}
				if err := json.NewDecoder(responseRecorder.Body).Decode(&gotResp); err != nil {
					t.Errorf("failed to decode response: %v", err)
				}
				if gotResp.Error != test.ExpectedBody {
					t.Errorf("expected error message %q, got %q", test.ExpectedBody, gotResp.Error)
				}
			} else {
				if gotBody := strings.TrimSpace(responseRecorder.Body.String()); gotBody != test.ExpectedBody {
					t.Errorf("expected body %q, got %q", test.ExpectedBody, gotBody)
				}
			}

			// Проверяем вызов next
			if test.NextCalled != nextCalled {
				t.Errorf("expected nextCalled=%v, got %v", test.NextCalled, nextCalled)
			}

			// Проверяем Content-Type для ошибок
			if test.expectedStatusCode == http.StatusUnauthorized {
				if contentType := responseRecorder.Header().Get("Content-Type"); contentType != "application/json" {
					t.Errorf("expected Content-Type %q, got %q", "application/json", contentType)
				}
			}

		})
	}
}
