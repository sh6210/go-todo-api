package tasks

import (
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

const TypeNotifyTodoCreated = "todo:notify_created"

type NotifyTodoCreatedPayload struct {
	TodoId int    `json:"todo_id"`
	Title  string `json:"title"`
}

func NewNotifyTodoCreatedTask(todoId int, title string) (*asynq.Task, error) {
	payload, err := json.Marshal(NotifyTodoCreatedPayload{TodoId: todoId, Title: title})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	return asynq.NewTask(TypeNotifyTodoCreated, payload), nil
}
