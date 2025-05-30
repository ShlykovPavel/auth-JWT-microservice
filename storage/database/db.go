package database

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5"
	"log/slog"
	"time"
)

type DbConfig struct {
	DbName     string `yaml:"db_name" env:"DB_NAME" `
	DbUser     string `yaml:"db_user" env:"DB_USER" `
	DbPassword string `yaml:"db_password" env:"DB_PASSWORD" `
	DbHost     string `yaml:"db_host" env:"DB_HOST" `
	DbPort     string `yaml:"db_port" env:"DB_PORT"`
}

func DbInit(config *DbConfig, log *slog.Logger) (*pgx.Conn, error) {
	const op = "database/DbInit"
	log = slog.With(
		slog.String("op", op),
		slog.String("host", config.DbHost),
		slog.String("port", config.DbPort),
		slog.String("db_name", config.DbName),
	)
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", config.DbUser, config.DbPassword, config.DbHost, config.DbPort, config.DbName)
	//Ставим таймаут операции, после которого функция завершится с ошибкой
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		log.Error("connect db failed", err)
		return nil, err
	}
	log.Info("Successfully connected with pgx!")
	return conn, nil
}

func CreateTables(conn *pgx.Conn, log *slog.Logger) error {
	const op = "database/CreateTables"
	log = slog.With(
		slog.String("op", op))

	query := `
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(64) NOT NULL,
    last_name VARCHAR(64) NOT NULL,
    email VARCHAR(256) NOT NULL UNIQUE,
    password VARCHAR(128) NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
)
`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := conn.Exec(ctx, query)
	if err != nil {
		log.Error("create table failed", err)
		return fmt.Errorf("failed to create users table: %w", err)
	}
	log.Info("Users table created successfully")
	return nil
}
