package repository

import (
	"context"
	"errors"
	"testing"
)

func TestTodoRepository_Create(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewTodoRepository(pool)
	ctx := context.Background()

	todo, err := repo.Create(ctx, "Buy milk", "2% milk, one gallon")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if todo.Id == 0 {
		t.Error("expected a generated Id, got 0")
	}
	if todo.Title != "Buy milk" {
		t.Errorf("expected title 'Buy milk', got %q", todo.Title)
	}
	if todo.Completed != false {
		t.Error("expected new todo to default to not completed")
	}
	if todo.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestTodoRepository_BulkCreate(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewTodoRepository(pool)
	ctx := context.Background()

	items := []BulkCreateInput{
		{Title: "Task 1", Description: "First"},
		{Title: "Task 2", Description: "Second"},
		{Title: "Task 3", Description: "Third"},
	}

	todos, err := repo.BulkCreate(ctx, items)
	if err != nil {
		t.Fatalf("BulkCreate failed: %v", err)
	}

	if len(todos) != 3 {
		t.Fatalf("expected 3 todos, got %d", len(todos))
	}
	for i, todo := range todos {
		if todo.Id == 0 {
			t.Errorf("todo %d: expected generated ID, got 0", i)
		}
	}
}

func TestTodoRepository_GetByID_NotFound(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewTodoRepository(pool)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, 999)

	if !errors.Is(err, ErrTodoNotFound) {
		t.Errorf("expected ErrTodoNotFound, got %v", err)
	}
}

func TestTodoRepository_UpdateAndDelete(t *testing.T) {
	pool := setupTestDB(t)
	repo := NewTodoRepository(pool)
	ctx := context.Background()

	created, err := repo.Create(ctx, "Original title", "Original description")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	updated, err := repo.Update(ctx, created.Id, "New title", "New description", true)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if updated.Title != "New title" {
		t.Errorf("expected title 'New title', got %q", updated.Title)
	}
	if !updated.Completed {
		t.Error("expected completed to be true after update")
	}
	if !updated.UpdatedAt.After(created.CreatedAt) {
		t.Error("expected UpdatedAt to be updated")
	}

	if err := repo.Delete(ctx, created.Id); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = repo.GetByID(ctx, created.Id)
	if !errors.Is(err, ErrTodoNotFound) {
		t.Errorf("expected ErrTodoNotFound, got %v", err)
	}
}
