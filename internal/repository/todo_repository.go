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

type BulkCreateInput struct {
	Title       string
	Description string
}

type PaginatedResult struct {
	Todos      []models.Todo `json:"todos"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
	TotalCount int           `json:"totalCount"`
	TotalPages int           `json:"totalPages"`
}

func NewTodoRepository(db *pgxpool.Pool) *TodoRepository {
	return &TodoRepository{db: db}
}

func (r *TodoRepository) BulkCreateCopy(ctx context.Context, items []BulkCreateInput) (int64, error) {
	rows := make([][]any, len(items))
	for i, item := range items {
		rows[i] = []any{item.Title, item.Description}
	}
	count, err := r.db.CopyFrom(
		ctx,
		pgx.Identifier{"todos"},
		[]string{"title", "description"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return 0, fmt.Errorf("bulk create copy: %w", err)
	}
	return int64(count), nil
}

// using pgx batch query
func (r *TodoRepository) BulkCreate(ctx context.Context, items []BulkCreateInput) ([]models.Todo, error) {
	for len(items) == 0 {
		return []models.Todo{}, nil
	}
	batch := &pgx.Batch{}
	query := `
				Insert into todos(title, description)
				values($1, $2)
				Returning id, title, description, completed, created_at, updated_at
			`
	for _, item := range items {
		batch.Queue(query, item.Title, item.Description)
	}

	br := r.db.SendBatch(ctx, batch)
	defer br.Close()

	todos := make([]models.Todo, 0, len(items))
	for range items {
		var todo models.Todo
		err := br.QueryRow().Scan(
			&todo.Id,
			&todo.Title,
			&todo.Description,
			&todo.Completed,
			&todo.CreatedAt,
			&todo.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("BulkCreate: %w", err)
		}
		todos = append(todos, todo)
	}
	return todos, nil
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

func (r *TodoRepository) GetByID(ctx context.Context, id int) (*models.Todo, error) {
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

func (r *TodoRepository) GetAllPaginated(ctx context.Context, page, limit int) (*PaginatedResult, error) {
	offset := (page - 1) * limit

	query := `
		Select id, title, description, completed, created_at, updated_at
FROM todos
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
			`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get todos: %w", err)
	}
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
			return nil, fmt.Errorf("failed to query todos: %w", err)
		}
		todos = append(todos, todo)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to get todos: %w", err)
	}

	var totalCount int
	countQuery := `SELECT COUNT(*) FROM todos`
	if err := r.db.QueryRow(ctx, countQuery).Scan(&totalCount); err != nil {
		return nil, fmt.Errorf("failed to get todos: %w", err)
	}

	totalPages := (totalCount + limit - 1) / limit

	return &PaginatedResult{
		Todos:      todos,
		Page:       page,
		Limit:      limit,
		TotalCount: totalCount,
		TotalPages: totalPages,
	}, nil
}

func (r *TodoRepository) GetAllCursor(ctx context.Context, cursor int, limit int) ([]models.Todo, error) {
	var query string
	var args []any

	if cursor == 0 {
		query = `SELECT id, title, description, completed, created_at, updated_at
					FROM todos
					ORDER BY id DESC
					LIMIT $1`
		args = []any{limit}
	} else {
		query = `Select Id, title, description, completed, created_at, updated_at
					FROM todos
					WHERE id < $1
					ORDER BY created_at DESC
					LIMIT $2`
		args = []any{cursor, limit}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get todos: %w", err)
	}
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
