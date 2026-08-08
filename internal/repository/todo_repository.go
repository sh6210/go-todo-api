package repository

import (
	//_ "github/sh6210/go-todo-api/internal/models"

	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sh6210/go-todo-api/internal/models"
)

var ErrTodoNotFound = errors.New("todo not found")

type TodoRepository struct {
	db *pgxpool.Pool
}

func NewTodoRepository(db *pgxpool.Pool) *TodoRepository {
	return &TodoRepository{db: db}
}

func (r *TodoRepository) Create(ctx context.Context, title, description string) (*models.Todo, error) {
	query := `
		INSERT INTO todos (title, description)
		VALUES ($1, $2)
		RETURNING id, title, description, completed, created_at, updated_at
	`

	var todo models.Todo
	err := r.db.QueryRow(ctx, query, title, description).Scan(
		&todo.Id,
		&todo.Title,
		&todo.Description,
		&todo.Completed,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create todo %w", err)
	}

	return &todo, nil
}

func (r *TodoRepository) GetById(ctx context.Context, id int) (*models.Todo, error) {
	query := `SELECT id, title, description, completed, created_at, updated_at 
				FROM todos
				WHERE id = $1`

	var todo models.Todo
	err := r.db.QueryRow(ctx, query, id).Scan(
		&todo.Id,
		&todo.Title,
		&todo.Description,
		&todo.Completed,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTodoNotFound
		}
		return nil, fmt.Errorf("failed to get todo: %w", err)
	}
	return &todo, nil
}

func (r *TodoRepository) GetAll(ctx context.Context) ([]models.Todo, error) {
	query := `SELECT id, title, description, completed, created_at, updated_at
				FROM todos
				ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get todos: %w", err)
	}

	// rows holds the connection pool, that's the reason we need to close
	defer rows.Close()

	var todos []models.Todo

	for rows.Next() {
		var todo models.Todo
		if err := rows.Scan(
			&todo.Id,
			&todo.Title,
			&todo.Description,
			&todo.Completed,
			&todo.CreatedAt,
			&todo.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to get todos: %w", err)
		}
		todos = append(todos, todo)
	}

	// need to check rows.Err() explicitly, as rows.Next() could return error instead of done
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to get todos: %w", err)
	}

	return todos, nil
}

func (r *TodoRepository) Update(ctx context.Context, id int, title, description string, completed bool) (*models.Todo, error) {
	query := `Update todos set title=$1, description=$2, completed=$3, updated_at=NOW()
WHERE id=$4
RETURNING id, title, description, completed, created_at, updated_at`

	var todo models.Todo
	err := r.db.QueryRow(ctx, query, title, description, completed, id).Scan(
		&todo.Id,
		&todo.Title,
		&todo.Description,
		&todo.Completed,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTodoNotFound
		}
		return nil, fmt.Errorf("failed to update todo: %w", err)
	}
	return &todo, nil
}

func (r *TodoRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM todos WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete todo: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrTodoNotFound
	}

	return nil
}
