package handlers

import (
	"context"

	"github.com/sh6210/go-todo-api/internal/models"
	"github.com/sh6210/go-todo-api/internal/repository"
)

type TodoRepositoryInterface interface {
	Create(ctx context.Context, title, description string) (*models.Todo, error)
	GetByID(ctx context.Context, id int) (*models.Todo, error)
	GetAll(ctx context.Context) ([]models.Todo, error)
	GetAllPaginated(ctx context.Context, page int, limit int) (*repository.PaginatedResult, error)
	GetAllCursor(ctx context.Context, cursor int, limit int) ([]models.Todo, error)
	Update(ctx context.Context, id int, title, description string, completed bool) (*models.Todo, error)
	Delete(ctx context.Context, id int) error
	Archive(ctx context.Context, id int) error
}
