package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sh6210/go-todo-api/internal/config"
	"github.com/sh6210/go-todo-api/internal/database"
	"github.com/sh6210/go-todo-api/internal/handlers"
	"github.com/sh6210/go-todo-api/internal/logger"
	"github.com/sh6210/go-todo-api/internal/repository"
)

func main() {
	cfg := config.Load()
	appLogger := logger.New()

	db, err := database.NewConnection(cfg.DatabaseURL)
	if err != nil {
		appLogger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	todoRepo := repository.NewTodoRepository(db)
	todoHandler := handlers.NewTodoHandler(todoRepo)
	//router := handlers.NewRouter(todoHandler)
	router := handlers.NewRouterWithLogger(todoHandler, appLogger)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		appLogger.Info("server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			appLogger.Error("server shutdown", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	appLogger.Info("shutting signal received, starting graceful shutdown")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		appLogger.Error("failed to shutdown server", "error", err)
		os.Exit(1)
	}

	appLogger.Info("shutting down graceful shutdown")
}
