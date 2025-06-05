package users

import (
	"booker/internal/lib/api/models"
	"booker/internal/storage/database"
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrEmailAlreadyExists = errors.New("Пользователь с email уже существует. ")

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUsersDB(dbPoll *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db: dbPoll,
	}
}

// Create Создание пользователя
// Принимает:
// ctx - внешний контекст, что б вызывающая сторона могла контролировать запрос (например выставить таймаут)
// userinfo - структуру User с необходимыми полями для добавления
//
// После запроса возвращается Id созданного пользователя
func (us *UserRepository) Create(ctx context.Context, userinfo *models.UserCreate) (int64, error) {
	query := `
INSERT INTO users (first_name, last_name, email, password)
VALUES ($1, $2, $3, $4)
RETURNING id`

	var id int64
	err := us.db.QueryRow(ctx, query, userinfo.FirstName, userinfo.LastName, userinfo.Email, userinfo.Password).Scan(&id)
	if err != nil {
		dbErr := database.PsqlErrorHandler(err)
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == database.PSQLUniqueError {
			return 0, ErrEmailAlreadyExists
		}
		return 0, dbErr
	}

	return id, nil
}
