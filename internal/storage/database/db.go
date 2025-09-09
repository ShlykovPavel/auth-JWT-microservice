package database

import (
	"context"
	"fmt"
	"github.com/ShlykovPavel/auth-JWT-microservice/metrics"
	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"time"
)

type DbConfig struct {
	DbName              string        `yaml:"db_name" env:"DB_NAME" `
	DbUser              string        `yaml:"db_user" env:"DB_USER" `
	DbPassword          string        `yaml:"db_password" env:"DB_PASSWORD" `
	DbHost              string        `yaml:"db_host" env:"DB_HOST" `
	DbPort              string        `yaml:"db_port" env:"DB_PORT"`
	DbMaxConnections    int32         `yaml:"db_max_connections" env:"DB_MAX_CONNECTIONS"`
	DbMinConnections    int32         `yaml:"db_min_connections" env:"DB_MIN_CONNECTIONS"`
	DbMaxConnLifetime   time.Duration `yaml:"db_max_conn_lifetime" env:"DB_MAX_CONN_LIFETIME"`
	DbMaxConnIdleTime   time.Duration `yaml:"db_max_conn_idle_time" env:"DB_MAX_CONN_IDLE_TIME"`
	DbHealthCheckPeriod time.Duration `yaml:"db_health_check_period" env:"DB_HEALTH_CHECK_PERIOD"`
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

		log.Error("connect users_db failed", "err", err.Error())
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
	connConfig.MaxConns = config.DbMaxConnections             // Максимальное количество соединений
	connConfig.MinConns = config.DbMinConnections             // Минимальное количество соединений
	connConfig.MaxConnLifetime = config.DbMaxConnLifetime     // Максимальное время жизни соединения
	connConfig.MaxConnIdleTime = config.DbMaxConnIdleTime     // Время бездействия перед закрытием
	connConfig.HealthCheckPeriod = config.DbHealthCheckPeriod // Период проверки жизни соединения с БД
	connConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		_, err = conn.Exec(ctx, "SET TIME ZONE 'UTC'")
		return err
	}

	pool, err := pgxpool.NewWithConfig(ctx, connConfig)
	if err != nil {
		return nil, fmt.Errorf("register pool failed: %w", err)
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
CREATE TABLE IF NOT EXISTS auth (
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
		logger.Error("register table failed", "err", err)
		return fmt.Errorf("failed to register auth table: %w", err)
	}
	logger.Info("Users table created successfully")
	return nil
}

func MonitorPool(ctx context.Context, pool *pgxpool.Pool, metrics *metrics.Metrics) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				stats := pool.Stat()
				metrics.PgxPoolMaxConns.Set(float64(stats.MaxConns()))
				metrics.PgxPoolUsedConns.Set(float64(stats.AcquiredConns()))
				metrics.PgxPoolIdleConns.Set(float64(stats.IdleConns()))
				time.Sleep(5 * time.Second)
			}
		}
	}()
}
