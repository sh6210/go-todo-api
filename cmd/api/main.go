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
	"github.com/sh6210/go-todo-api/internal/queue"
	"github.com/sh6210/go-todo-api/internal/repository"
)

func main() {
	// 1. Load configuration
	cfg := config.Load()

	// 2. Set up structured logging
	appLogger := logger.New()

	// 3. Connect to the database
	db, err := database.NewConnection(cfg.DatabaseURL)
	if err != nil {
		appLogger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// 4. Connect to Redis for background job queueing
	queueClient := queue.NewClient(cfg.RedisAddr)
	defer queueClient.Close()

	// 5. Build the repository, injecting the DB pool
	todoRepo := repository.NewTodoRepository(db)

	// 6. Build the handler, injecting the repository, queue client, and logger
	todoHandler := handlers.NewTodoHandler(todoRepo, queueClient, appLogger)

	// 7. Build the router, injecting the handler and logger
	router := handlers.NewRouterWithLogger(todoHandler, appLogger)

	// 8. Construct the HTTP server explicitly (needed for graceful shutdown)
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// 9. Start the server in a background goroutine, so main() stays free
	//    to listen for shutdown signals below.
	go func() {
		appLogger.Info("server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			appLogger.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// 10. Block until we receive SIGINT (Ctrl+C) or SIGTERM (Docker/Kubernetes shutdown).
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	appLogger.Info("shutdown signal received, starting graceful shutdown")

	// 11. Give in-flight requests up to 10 seconds to finish before forcing shutdown.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		appLogger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	appLogger.Info("server stopped cleanly")
}
