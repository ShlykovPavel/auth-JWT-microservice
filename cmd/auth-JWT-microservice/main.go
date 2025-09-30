package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	_ "github.com/ShlykovPavel/auth-JWT-microservice/docs"
	"github.com/ShlykovPavel/auth-JWT-microservice/internal/app"
	"github.com/ShlykovPavel/auth-JWT-microservice/internal/config"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

// TODO Решено использовать паттерн transaction outbox для надёжной отправки сообщений в Kafka
// Необходимо реализовать воркер, который будет читать из таблицы outbox и отправлять сообщения в Kafka
// @description В таблице outbox будут храниться события, которые необходимо отправить в Kafka
// @description После успешной отправки события в Kafka, необходимо отметить запись в таблице outbox как отправленную
// @description В случае ошибки при отправке события в Kafka, необходимо увеличить счётчик попыток и обновить время последней попытки
// @description Если количество попыток превышает максимальное значение, необходимо пометить запись как ошибочную и не пытаться отправлять её снова
// @description Максимальное количество попыток и интервал между попытками должны быть настраиваемыми параметрами
// @description Воркер должен работать в фоне и периодически проверять таблицу outbox на наличие новых записей
// @description Воркер должен быть устойчив к сбоям и перезапускам приложения
// @description Воркер должен логировать свою работу и ошибки
// @description Воркер должен быть протестирован
// @description Воркер должен быть задокументирован
// @description Воркер должен быть интегрирован в основное приложение
// @description Воркер должен быть запущен при старте приложения
// @description Воркер должен быть остановлен при завершении работы приложения

// @title Auth Microservice API
// @version 1.0
// @description API для регистрации и авторизации
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Add "Bearer" before token
func main() {
	cfg, err := config.LoadConfig("secret_config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(cfg)
	logger := setupLogger(cfg.Env)
	logger.Info("Starting application")
	logger.Debug("Debug messages enabled")

	application := app.NewApp(logger, cfg)
	application.Run()
}

func setupLogger(env string) *slog.Logger {
	var logger *slog.Logger
	switch env {
	case envLocal:
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	case envDev:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	case envProd:
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	}
	return logger
}
