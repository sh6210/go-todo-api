package main

import (
	"log"
	"net/http"
	"os"

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

	log.Printf("Listening on port %s", cfg.Port)

	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
