package users_db

import (
	"booker/internal/lib/api/models"
	"booker/internal/storage/database"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
)

var ErrEmailAlreadyExists = errors.New("Пользователь с email уже существует. ")
var ErrUserNotFound = errors.New("Пользователь не найден ")

type UserRepository interface {
	CreateUser(ctx context.Context, userinfo *models.UserCreate) (int64, error)
	GetUser(ctx context.Context, userEmail string) (UserInfo, error)
}

type UserRepositoryImpl struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

// UserInfo Структура с информацие о пользователе
type UserInfo struct {
	ID           int64
	FirstName    string
	LastName     string
	Email        string
	PasswordHash string
}

func NewUsersDB(dbPoll *pgxpool.Pool, log *slog.Logger) *UserRepositoryImpl {
	return &UserRepositoryImpl{
		db:  dbPoll,
		log: log,
	}
}

// CreateUser Создание пользователя
// Принимает:
// ctx - внешний контекст, что б вызывающая сторона могла контролировать запрос (например выставить таймаут)
// userinfo - структуру UserInfo с необходимыми полями для добавления
//
// После запроса возвращается Id созданного пользователя
func (us *UserRepositoryImpl) CreateUser(ctx context.Context, userinfo *models.UserCreate) (int64, error) {
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

func (us *UserRepositoryImpl) GetUser(ctx context.Context, userEmail string) (UserInfo, error) {
	query := `SELECT id, first_name, last_name, email, password FROM users WHERE email = $1`

	var user UserInfo
	err := us.db.QueryRow(ctx, query, userEmail).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return UserInfo{}, ErrUserNotFound
	}
	if err != nil {
		dbErr := database.PsqlErrorHandler(err)
		return UserInfo{}, dbErr
	}
	return user, nil
}
