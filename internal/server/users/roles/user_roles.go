package roles

import (
	"booker/internal/storage/database/repositories/users_db"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
)

func CheckAdminInDB(poll *pgxpool.Pool, log *slog.Logger) error {
	userRepository := users_db.NewUsersDB(poll, log)

	_, err := userRepository.CheckAdminInDB(context.Background())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Error("no admin role found", "error", err)
			err = userRepository.AddFirstAdmin(context.Background())
			if err != nil {
				log.Error("error adding admin role", "error", err)
				return err
			}
		}
		log.Error("error checking admin role", "error", err)
		return err
	}
	log.Info("admin role check ok. no need to create admin role")
	return nil
}

//TODO Добавить после проверки на роль админа (поинт может прожать только с ролью админ
//func SetAdminRole(poll  *pgxpool.Pool, log *slog.Logger) http.HandlerFunc  {
//	return func(w http.ResponseWriter, r *http.Request) {
//		userRepository := users_db.NewUsersDB(poll, log)
//
//		err := userRepository.SetAdminRole(context.Background(), id)
//
//	}
//
//
//}
