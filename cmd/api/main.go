package main

import (
	"log"
	"net/http"

	"github.com/sh6210/go-todo-api/internal/config"
	"github.com/sh6210/go-todo-api/internal/database"
	"github.com/sh6210/go-todo-api/internal/handlers"
	"github.com/sh6210/go-todo-api/internal/repository"
)

func main() {
	cfg := config.Load()

	db, err := database.NewConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	todoRepo := repository.NewTodoRepository(db)

	todoHandler := handlers.NewTodoHandler(todoRepo)

	router := handlers.NewRouter(todoHandler)

	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
