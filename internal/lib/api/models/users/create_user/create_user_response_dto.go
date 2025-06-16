package create_user

import (
	resp "booker/internal/lib/api/response"
)

// CreateUserResponse Структура ответа на запрос
type CreateUserResponse struct {
	resp.Response
	UserID int64 `json:"id"`
}
