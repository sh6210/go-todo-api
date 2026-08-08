package models

import "time"

type Todo struct {
	Id          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateTodoRequest struct {
	Title       string `json:"title" validate:"required,min=3,max=255"`
	Description string `json:"description" validate:"required,min=3,max=255"`
}

type UpdateTodoRequest struct {
	Title       string `json:"title" validate:"required,min=3,max=255"`
	Description string `json:"description" validate:"min=3,max=255"`
	Completed   bool   `json:"completed"`
}
