package users

import (
	"booker/storage/database"
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type UserRepository struct {
	db *pgx.Conn
}

type UserID struct {
	ID int64
}

func NewUsersDB(db *pgx.Conn) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// Create Создание пользователя
// Принимает:
// ctx - внешний контекст, что б вызывающая сторона могла контролировать запрос (например выставить таймаут)
// userinfo - структуру User с необходимыми полями для добавления
//
// После запроса возвращается Id созданного пользователя
func (us *UserRepository) Create(ctx context.Context, userinfo *User) (*UserID, error) {
	query := `
INSERT INTO users (first_name, last_name, email, password)
VALUES ($1, $2, $3, $4)
RETURNING id`

	var id int64
	err := us.db.QueryRow(ctx, query, userinfo.FirstName, userinfo.LastName, userinfo.Email, userinfo.Password).Scan(&id)
	if err != nil {
		dbErr := database.PsqlErrorHandler(err)
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return &UserID{}, ErrEmailAlreadyExists
		}
		return &UserID{}, dbErr
	}

	return &UserID{
		ID: id,
	}, nil
}
