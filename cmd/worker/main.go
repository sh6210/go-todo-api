package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"github.com/hibiken/asynq"
	"github.com/sh6210/go-todo-api/internal/config"
	"github.com/sh6210/go-todo-api/internal/logger"
	"github.com/sh6210/go-todo-api/internal/tasks"
)

func main() {
	cfg := config.Load()
	appLogger := logger.New()

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.RedisAddr},
		asynq.Config{
			Concurrency: 5, // process up to 5 tasks concurrently
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeNotifyTodoCreated, handleNotifyTodoCreated(appLogger))

	appLogger.Info("worker starting")
	if err := srv.Run(mux); err != nil {
		appLogger.Error("worker failed", "error", err)
		os.Exit(1)
	}
}

// handleNotifyTodoCreated returns a task handler function, with the logger
// captured via closure — a common Go pattern for injecting dependencies
// into functions that must match a specific signature (here, asynq's handler type).
func handleNotifyTodoCreated(logger *slog.Logger) func(ctx context.Context, t *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		var payload tasks.NotifyTodoCreatedPayload
		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return err // asynq will treat this as a failed task and may retry it
		}

		// In a real app: send an actual email/push notification here.
		// For now, we simulate it with a log line.
		logger.Info("notification sent",
			"todo_id", payload.TodoId,
			"title", payload.Title,
		)

		return nil // nil = task succeeded, asynq marks it done
	}
}
