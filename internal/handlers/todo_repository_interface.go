package handlers

import (
	"context"

	"github.com/sh6210/go-todo-api/internal/models"
)

type TodoRepositoryInterface interface {
	Create(ctx context.Context, title, description string) (*models.Todo, error)
	GetByID(ctx context.Context, id int) (*models.Todo, error)
	GetAll(ctx context.Context) ([]models.Todo, error)
	Update(ctx context.Context, id int, title, description string, completed bool) (*models.Todo, error)
	Delete(ctx context.Context, id int) error
}
