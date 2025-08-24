package main

import (
	"fmt"
	_ "github.com/ShlykovPavel/auth-JWT-microservice/docs"
	"github.com/ShlykovPavel/auth-JWT-microservice/internal/app"
	"github.com/ShlykovPavel/auth-JWT-microservice/internal/config"
	"log"
	"log/slog"
	"os"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

// @title Auth Microservice API
// @version 1.0
// @description API для управления бронированиями
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
