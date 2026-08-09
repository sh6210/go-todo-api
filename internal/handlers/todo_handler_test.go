package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/sh6210/go-todo-api/internal/models"
	"github.com/sh6210/go-todo-api/internal/repository"
)

func TestCreate_Success(t *testing.T) {
	// Arrange: set up a mock repository that always succeeds
	mock := &mockTodoRepository{
		CreateFunc: func(ctx context.Context, title, description string) (*models.Todo, error) {
			return &models.Todo{Title: title, Description: description}, nil
		},
	}
	handler := NewTodoHandler(mock)

	body := strings.NewReader(`{"title": "Test Todo", "description": "Testing"}`)
	req := httptest.NewRequest(http.MethodPost, "/todos", body)
	res := httptest.NewRecorder()

	// Act
	handler.Create(res, req)

	// Assert
	if res.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", res.Code)
	}

	var got models.Todo
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got.Title != "Test Todo" {
		t.Errorf("expected title 'Test Todo', got %q", got.Title)
	}
}

func TestCreate_ValidationFailure(t *testing.T) {
	mock := &mockTodoRepository{} // no CreateFunc needed — should never be called
	handler := NewTodoHandler(mock)

	body := strings.NewReader(`{"title": ""}`) // empty title, fails "required"
	req := httptest.NewRequest(http.MethodPost, "/todos", body)
	res := httptest.NewRecorder()

	handler.Create(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", res.Code)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	mock := &mockTodoRepository{
		GetByIDFunc: func(ctx context.Context, id int) (*models.Todo, error) {
			return nil, repository.ErrTodoNotFound
		},
	}
	handler := NewTodoHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/todos/999", nil)
	res := httptest.NewRecorder()

	// chi URL params aren't parsed automatically outside a real router,
	// so we manually inject a route context — this is the standard way
	// to unit test chi handlers that use chi.URLParam.
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.GetByID(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", res.Code)
	}
}

func TestGetByID_UnexpectedError(t *testing.T) {
	mock := &mockTodoRepository{
		GetByIDFunc: func(ctx context.Context, id int) (*models.Todo, error) {
			return nil, errors.New("database exploded")
		},
	}
	handler := NewTodoHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/todos/1", nil)
	res := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.GetByID(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", res.Code)
	}
}
