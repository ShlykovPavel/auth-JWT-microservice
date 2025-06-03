package database

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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

func DbConnect(config *DbConfig, log *slog.Logger) (*pgx.Conn, error) {
	const op = "database/DbConnect"
	log = slog.With(
		slog.String("op", op),
		slog.String("host", config.DbHost),
		slog.String("port", config.DbPort),
		slog.String("db_name", config.DbName),
	)
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", config.DbUser, config.DbPassword, config.DbHost, config.DbPort, config.DbName)

	const retryCount = 5
	const retryDelay = 5 * time.Second

	var conn *pgx.Conn
	var err error

	for i := 1; i <= retryCount; i++ {
		//Ставим таймаут операции, после которого функция завершится с ошибкой
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		//Попытка соединения
		conn, err = pgx.Connect(ctx, connStr)
		//Закрываем наш контекст, что б освободить ресурсы
		cancel()

		if err == nil {
			log.Info("Successfully connected with pgx!")
			return conn, nil
		}

		log.Error("connect db failed", err)
		if i < retryCount {
			log.Info(fmt.Sprintf("Retrying in %v... (attempt %d/%d)", retryDelay, i+1, retryCount))
			time.Sleep(retryDelay)
		}

	}
	return nil, err
}

func CreatePool(ctx context.Context, config *DbConfig, logger *slog.Logger) (*pgxpool.Pool, error) {
	const op = "database/CreatePool"
	logger = logger.With(
		slog.String("op", op),
		slog.String("host", config.DbHost),
		slog.String("port", config.DbPort),
		slog.String("db_name", config.DbName),
	)
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", config.DbUser, config.DbPassword, config.DbHost, config.DbPort, config.DbName)

	connConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pgx config: %w", err)
	}

	// Настройка параметров пула
	connConfig.MaxConns = 10                      // Максимальное количество соединений
	connConfig.MinConns = 0                       // Минимальное количество соединений
	connConfig.MaxConnLifetime = time.Hour        // Максимальное время жизни соединения
	connConfig.MaxConnIdleTime = 30 * time.Minute // Время бездействия перед закрытием
	connConfig.HealthCheckPeriod = time.Minute    // Период проверки жизни соединения с БД

	pool, err := pgxpool.NewWithConfig(ctx, connConfig)
	if err != nil {
		return nil, fmt.Errorf("create pool failed: %w", err)
	}
	logger.Info("Successfully created pool")
	// Проверка соединения
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping failed: %w", err)
	}
	stats := pool.Stat()
	logger.Debug("current pool state",
		slog.Int("max_conns", int(stats.MaxConns())),
		slog.Int("total_conns", int(stats.TotalConns())),
		slog.Int("idle_conns", int(stats.IdleConns())),
		slog.Int("acquired_conns", int(stats.AcquiredConns())),
	)
	return pool, nil
}

func CreateTables(poll *pgxpool.Pool, logger *slog.Logger) error {
	const op = "database/CreateTables"
	logger = logger.With(
		slog.String("op", op))
	//Берём соединение с БД и пула
	connection, err := poll.Acquire(context.Background())
	if err != nil {
		return fmt.Errorf("acquire failed: %w", err)
	}
	defer connection.Release()

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
	_, err = connection.Exec(ctx, query)
	if err != nil {
		logger.Error("create table failed", err)
		return fmt.Errorf("failed to create users table: %w", err)
	}
	logger.Info("Users table created successfully")
	return nil
}
