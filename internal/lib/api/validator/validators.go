package validators

import (
	"errors"
	"github.com/go-playground/validator"
)

// validate Переменная хранящая в себе экземпляр валидатора.
// она нужна что б не инициализировать экземпляр валидатора каждый раз когда нам нужно что-то провалидировать
var validate *validator.Validate

func InitValidator() error {
	if validate != nil {
		return errors.New("validator already initialized")
	}
	validate = validator.New()

	return nil
}

func GetValidator() *validator.Validate {
	return validate
}
