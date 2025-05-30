package users

import (
	resp "booker/lib/api/response"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
	"github.com/jackc/pgx/v5"
	"log/slog"
	"net/http"
)

// User Структура пользователя для создания
type User struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password"  validate:"required,lte=64"`
}

// Response Структура ответа на запрос
type Response struct {
	resp.Response
	UserID int64 `json:"id"`
}

// UserCreate Интерфейс с функцией создания
//type UserCreate interface {
//	Create(ctx context.Context, user *User) (*UserID, error)
//}

func CreateUser(log *slog.Logger, db *pgx.Conn) http.HandlerFunc {
	return (func(w http.ResponseWriter, r *http.Request) {
		const op = "server/users.CreateUser"
		log = log.With(
			slog.String("operation", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
			slog.String("url", r.URL.Path))
		usrCreate := NewUsersDB(db)

		var user User
		err := render.DecodeJSON(r.Body, &user)
		if err != nil {
			log.Error("Error while decoding request body", err)
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}
		//Валидация
		//TODO Посмотреть где ещё создаются валидаторы, и если их много, то нужно вынести инициализацию валидатора глобально для повышения оптимизации
		if err = validator.New().Struct(&user); err != nil {
			validationErrors := err.(validator.ValidationErrors)
			log.Error("Error validating request body", validationErrors)
			render.JSON(w, r, resp.ValidationError(validationErrors))
			return

		}
		//	хешируем пароль
		passwordHash, err := HashUserPassword(user.Password, log)
		if err != nil {
			log.Error("Error while hashing password", err)
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}
		user.Password = passwordHash
		//Записываем в бд
		userId, err := usrCreate.Create(r.Context(), &user)
		if err != nil {
			log.Error("Error while creating user", err)
			render.JSON(w, r, resp.Error(
				err.Error()))
			return
		}
		log.Info("Created user", userId)
		render.JSON(w, r, Response{
			resp.OK(),
			userId.ID,
		})
	})
}
