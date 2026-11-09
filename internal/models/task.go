package models

import (
	"errors"
	"strings"
	"time"
)

// Priority represents the urgency level of a task.
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

// Task represents a single task item in the application.
type Task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Completed   bool      `json:"completed"`
	Priority    Priority  `json:"priority"`
	DueDate     string    `json:"due_date,omitempty"` // format: YYYY-MM-DD
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateTaskRequest defines the payload for creating a new task.
type CreateTaskRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Priority    Priority `json:"priority"`
	DueDate     string   `json:"due_date"`
}

// Validate checks if the create task request has valid inputs.
func (r *CreateTaskRequest) Validate() error {
	r.Title = strings.TrimSpace(r.Title)
	if r.Title == "" {
		return errors.New("title is required and cannot be empty")
	}
	if len(r.Title) > 150 {
		return errors.New("title cannot exceed 150 characters")
	}

	r.Description = strings.TrimSpace(r.Description)
	if len(r.Description) > 1000 {
		return errors.New("description cannot exceed 1000 characters")
	}

	switch r.Priority {
	case PriorityLow, PriorityMedium, PriorityHigh:
		// Valid
	case "":
		r.Priority = PriorityMedium
	default:
		return errors.New("priority must be 'low', 'medium', or 'high'")
	}

	return nil
}

// UpdateTaskRequest defines the payload for updating an existing task.
type UpdateTaskRequest struct {
	Title       *string   `json:"title,omitempty"`
	Description *string   `json:"description,omitempty"`
	Completed   *bool     `json:"completed,omitempty"`
	Priority    *Priority `json:"priority,omitempty"`
	DueDate     *string   `json:"due_date,omitempty"`
}

// Validate checks if the update request contains valid values.
func (r *UpdateTaskRequest) Validate() error {
	if r.Title != nil {
		trimmed := strings.TrimSpace(*r.Title)
		if trimmed == "" {
			return errors.New("title cannot be empty")
		}
		if len(trimmed) > 150 {
			return errors.New("title cannot exceed 150 characters")
		}
		*r.Title = trimmed
	}

	if r.Description != nil {
		trimmed := strings.TrimSpace(*r.Description)
		if len(trimmed) > 1000 {
			return errors.New("description cannot exceed 1000 characters")
		}
		*r.Description = trimmed
	}

	if r.Priority != nil {
		switch *r.Priority {
		case PriorityLow, PriorityMedium, PriorityHigh:
			// Valid
		default:
			return errors.New("priority must be 'low', 'medium', or 'high'")
		}
	}

	return nil
}

// TaskStats holds summary metrics for all tasks.
type TaskStats struct {
	Total          int `json:"total"`
	Completed      int `json:"completed"`
	Pending        int `json:"pending"`
	CompletionRate int `json:"completion_rate"` // percentage 0-100
}

// APIResponse provides a standard JSON envelope for responses.
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}
