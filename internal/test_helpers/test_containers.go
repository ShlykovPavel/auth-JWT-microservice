package test_helpers

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	testPostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	postgresImage = "postgres:16-alpine"
	kafkaImage    = "confluentinc/cp-kafka:7.4.1"
)

// CreatePostgresContainer Создаёт и запускает контейнер с Postgres для тестирования.
// Не забудь в тесте руками его остановить и удалить после использования.
func CreatePostgresContainer() (*testPostgres.PostgresContainer, string) {
	ctx := context.Background()

	postgresContainer, err := testPostgres.Run(
		ctx,
		postgresImage,
		testPostgres.WithDatabase("testdb"),
		testPostgres.WithUsername("user"),
		testPostgres.WithPassword("password"),
	)
	if err != nil {
		panic("Failed to start Postgres container: " + err.Error())
	}
	testPostgres.BasicWaitStrategies()
	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic("Failed to get Postgres connection string: " + err.Error())
	}
	fmt.Println("Postgres is running at:", connStr)

	applyMigrations(connStr)

	return postgresContainer, connStr

}

func applyMigrations(connStr string) {
	// Получаем путь к текущему файлу и строим путь к миграциям относительно корня проекта
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "..", "..")
	migrationPath := filepath.Join(projectRoot, "internal", "storage", "database", "migration")

	// Retry логика для подключения к БД
	var m *migrate.Migrate
	var err error
	for i := 0; i < 10; i++ {
		m, err = migrate.New("file://"+migrationPath, connStr)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		panic("Failed to create migrate instance: " + err.Error())
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		panic("Failed to apply migrations: " + err.Error())
	}
}
