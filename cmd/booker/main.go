package main

import (
	"booker/internal/config"
	"booker/internal/server/users/auth"
	"booker/internal/server/users/create"
	"booker/internal/storage/database"
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"log/slog"
	"net/http"
	"os"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg, err := config.LoadConfig(".env")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(cfg)
	logger := setupLogger(cfg.Env)
	logger.Info("Starting application")
	logger.Debug("Debug messages enabled")
	dbConfig := database.DbConfig{
		DbName:     cfg.DbName,
		DbUser:     cfg.DbUser,
		DbPassword: cfg.DbPassword,
		DbHost:     cfg.DbHost,
		DbPort:     cfg.DbPort,
	}

	poll, err := database.CreatePool(context.Background(), &dbConfig, logger)
	err = database.CreateTables(poll, logger)
	if err != nil {
		panic(err)
	}
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post("/user/register", users.CreateUser(logger, poll))
	router.Post("/login", auth.AuthenticationHandler(logger, poll, cfg.JWTSecretKey))
	router.Post("/refresh", auth.RefreshTokenHandler(logger, poll, cfg.JWTSecretKey))

	logger.Info("Starting HTTP server", slog.String("adress", cfg.Address))
	// Run server
	srv := &http.Server{
		Addr:    cfg.Address,
		Handler: router,
		//ReadHeaderTimeout: cfg.HTTPServer.Timeout,
		//WriteTimeout:      cfg.HTTPServer.Timeout,
		//IdleTimeout:       cfg.HTTPServer.IdleTimeout,
	}
	if err := srv.ListenAndServe(); err != nil {
		logger.Error("failed to start server", "error", err.Error())
		os.Exit(1)
	}
	logger.Info("Stopped HTTP server")
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
