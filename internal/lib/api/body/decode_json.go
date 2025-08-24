package body

import (
	"errors"
	validators "github.com/ShlykovPavel/auth-JWT-microservice/internal/lib/api/validator"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
)

// Ошибки
var ErrDecodeJSON = errors.New("failed to decode JSON")

// DecodeAndValidateJson декодирует JSON и валидирует структуру
func DecodeAndValidateJson(r *http.Request, v interface{}) error {
	// Декодируем JSON
	if err := render.DecodeJSON(r.Body, v); err != nil {
		slog.Default().Error("DecodeAndValidateJson: error decoding body or validating", "error", err)
		return ErrDecodeJSON
	}

	// Валидируем структуру
	if err := validators.GetValidator().Struct(v); err != nil {
		return err
	}

	return nil
}
