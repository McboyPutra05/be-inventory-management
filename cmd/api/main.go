package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/intern/inventory-api/internal/config"
	"github.com/intern/inventory-api/internal/database"
	"github.com/intern/inventory-api/internal/handler"
	"github.com/intern/inventory-api/internal/pkg/token"
	"github.com/intern/inventory-api/internal/repository"
	"github.com/intern/inventory-api/internal/router"
	"github.com/intern/inventory-api/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err.Error())
		os.Exit(1)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		slog.Error("failed to connect to database", "error", err.Error())
		os.Exit(1)
	}

	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("failed to get sql.DB instance", "error", err.Error())
		os.Exit(1)
	}
	defer sqlDB.Close()

	if err := database.RunMigrations(sqlDB); err != nil {
		slog.Error("failed to run database migrations", "error", err.Error())
		os.Exit(1)
	}

	tokenManager := token.NewManager(cfg.JWTSecret, cfg.JWTExpiresIn)

	userRepo := repository.NewUserRepository(db)
	itemRepo := repository.NewItemRepository(db)

	authService := service.NewAuthService(userRepo, tokenManager)
	itemService := service.NewItemService(itemRepo)

	handlers := &router.Handlers{
		Auth:   handler.NewAuthHandler(authService),
		Item:   handler.NewItemHandler(itemService),
		Health: handler.NewHealthHandler(sqlDB),
	}

	engine := router.SetupRouter(cfg, tokenManager, handlers)

	serverAddr := fmt.Sprintf(":%s", cfg.AppPort)
	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server starting", "port", cfg.AppPort, "env", cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed to start", "error", err.Error())
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err.Error())
		os.Exit(1)
	}

	slog.Info("server gracefully stopped")
}

