package handlers

import (
	"context"

	"github.com/sh6210/go-todo-api/internal/models"
	"github.com/sh6210/go-todo-api/internal/repository"
)

type mockTodoRepository struct {
	CreateFunc  func(ctx context.Context, title, description string) (*models.Todo, error)
	GetByIDFunc func(ctx context.Context, id int) (*models.Todo, error)
	GetAllFunc  func(ctx context.Context) ([]models.Todo, error)
	UpdateFunc  func(ctx context.Context, id int, title, description string, completed bool) (*models.Todo, error)
	DeleteFunc  func(ctx context.Context, id int) error
}

func (m *mockTodoRepository) Create(ctx context.Context, title, description string) (*models.Todo, error) {
	return m.CreateFunc(ctx, title, description)
}

func (m *mockTodoRepository) GetByID(ctx context.Context, id int) (*models.Todo, error) {
	return m.GetByIDFunc(ctx, id)
}

func (m *mockTodoRepository) GetAll(ctx context.Context) ([]models.Todo, error) {
	return m.GetAllFunc(ctx)
}

func (m *mockTodoRepository) Update(ctx context.Context, id int, title, description string, completed bool) (*models.Todo, error) {
	return m.UpdateFunc(ctx, id, title, description, completed)
}

func (m *mockTodoRepository) Delete(ctx context.Context, id int) error {
	return m.DeleteFunc(ctx, id)
}

// keep the repository import used in real tests later (e.g. repository.ErrTodoNotFound)
var _ = repository.ErrTodoNotFound
