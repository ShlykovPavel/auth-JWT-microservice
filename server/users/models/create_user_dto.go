package models

import resp "booker/lib/api/response"

type UserCreate struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password"  validate:"required,lte=64"`
}

// CreateUserResponse Структура ответа на запрос
type CreateUserResponse struct {
	resp.Response
	UserID int64 `json:"id"`
}
