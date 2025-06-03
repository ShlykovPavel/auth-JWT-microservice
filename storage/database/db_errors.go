package database

import (
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
)

// Коды ошибок PosgreSQL
const PSQLUniqueError = "23505"                    // Нарушение уникальности поля
const PSQLForeignKeyError = "23503"                // Нарушение внешнего ключа
const PSQLNotNullError = "23502"                   // Нельзя вставить NULL
const PSQLStringDataRightTruncationError = "22001" // Данные слишком длинные
const PSQLSyntaxError = "42601"                    // Синтаксическая ошибка в SQL

func PsqlErrorHandler(err error) error {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		switch pgErr.Code {
		case PSQLUniqueError: // unique_violation
			return fmt.Errorf("Нарушение уникальности: %w", err)
		case PSQLForeignKeyError: // foreign_key_violation
			return fmt.Errorf("нарушение внешнего ключа: %w", err)
		case PSQLNotNullError: // not_null_violation
			return fmt.Errorf("нельзя вставить NULL: %w", err)
		case PSQLStringDataRightTruncationError: // string_data_right_truncation
			return fmt.Errorf("данные слишком длинные: %w", err)
		case PSQLSyntaxError: // syntax_error
			return fmt.Errorf("синтаксическая ошибка в SQL: %w", err)
		default:
			return fmt.Errorf("ошибка PostgreSQL (%s): %w", pgErr.Code, err)
		}
	}
	return err
}
