package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/sh6210/go-todo-api/internal/models"
	"github.com/sh6210/go-todo-api/internal/repository"
	"github.com/sh6210/go-todo-api/internal/validation"
)

type TodoHandler struct {
	repo TodoRepositoryInterface
}

func NewTodoHandler(repo TodoRepositoryInterface) *TodoHandler {
	return &TodoHandler{repo: repo}
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func respondValidationError(w http.ResponseWriter, err error) {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		fieldErrors := make(map[string]string)
		for _, fe := range validationErrors {
			fieldErrors[fe.Field()] = validationMessage(fe)
		}
		respondJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":  "validation failed",
			"fields": fieldErrors,
		})
		return
	}
	respondError(w, http.StatusBadRequest, err.Error())
}

func validationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "this field is required"
	case "min":
		return fmt.Sprintf("%s must be %s", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s must be %s", fe.Field(), fe.Param())
	default:
		return "invalid value"
	}
}

func (h *TodoHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := validation.Validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	todo, err := h.repo.Create(r.Context(), req.Title, req.Description)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, todo)
}

func (h *TodoHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	todo, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrTodoNotFound) {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, todo)
}

func (h *TodoHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	todos, err := h.repo.GetAll(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, todos)
}

func (h *TodoHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req models.UpdateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := validation.Validate.Struct(req); err != nil {
		respondValidationError(w, err)
		return
	}

	todo, err := h.repo.Update(r.Context(), id, req.Title, req.Description, req.Completed)
	if err != nil {
		if errors.Is(err, repository.ErrTodoNotFound) {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, todo)
}

func (h *TodoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.repo.Delete(r.Context(), id); err != nil {
		if errors.Is(err, repository.ErrTodoNotFound) {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
