package create_user

import (
	resp "github.com/ShlykovPavel/auth-JWT-microservice/internal/lib/api/response"
)

// CreateUserResponse Структура ответа на запрос
type CreateUserResponse struct {
	resp.Response
	UserID int64 `json:"id"`
}
