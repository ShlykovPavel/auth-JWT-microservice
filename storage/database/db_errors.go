package database

import (
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
)

func PsqlErrorHandler(err error) error {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		switch pgErr.Code {
		case "23505": // unique_violation
			return fmt.Errorf("запись уже существует: %w", err)
		case "23503": // foreign_key_violation
			return fmt.Errorf("нарушение внешнего ключа: %w", err)
		case "23502": // not_null_violation
			return fmt.Errorf("нельзя вставить NULL: %w", err)
		case "22001": // string_data_right_truncation
			return fmt.Errorf("данные слишком длинные: %w", err)
		case "42601": // syntax_error
			return fmt.Errorf("синтаксическая ошибка в SQL: %w", err)
		default:
			return fmt.Errorf("ошибка PostgreSQL (%s): %w", pgErr.Code, err)
		}
	}
	return err
}
