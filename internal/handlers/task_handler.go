package handlers

import (
	"encoding/json"
	"net/http"

	"taskmanager/internal/models"
	"taskmanager/internal/store"
)

// TaskHandler manages HTTP requests for tasks.
type TaskHandler struct {
	store store.TaskStore
}

// NewTaskHandler constructs a new TaskHandler.
func NewTaskHandler(s store.TaskStore) *TaskHandler {
	return &TaskHandler{store: s}
}

// writeJSON serializes data as JSON and sets the appropriate headers.
func (h *TaskHandler) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, `{"success":false,"error":"Failed to encode response"}`, http.StatusInternalServerError)
	}
}

// writeError sends a standardized error response.
func (h *TaskHandler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, models.APIResponse{
		Success: false,
		Error:   message,
	})
}

// HealthCheck returns server status and time.
func (h *TaskHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Task Management API is healthy and operational",
	})
}

// GetTasks returns all tasks matching optional query parameters:
// - status: "all", "active", "completed" (default "all")
// - search: substring query for title or description
// - sort: "newest", "oldest", "priority", "due_date" (default "newest")
func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	search := r.URL.Query().Get("search")
	sortBy := r.URL.Query().Get("sort")

	tasks := h.store.GetAll(status, search, sortBy)

	h.writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    tasks,
	})
}

// GetTaskByID retrieves a single task by its unique ID.
func (h *TaskHandler) GetTaskByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.writeError(w, http.StatusBadRequest, "Task ID is required")
		return
	}

	task, exists := h.store.GetByID(id)
	if !exists {
		h.writeError(w, http.StatusNotFound, "Task not found")
		return
	}

	h.writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    task,
	})
}

// CreateTask handles creating a new task from a JSON body.
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid JSON payload: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	task := h.store.Create(req)

	h.writeJSON(w, http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "Task created successfully",
		Data:    task,
	})
}

// UpdateTask handles updating an existing task.
func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.writeError(w, http.StatusBadRequest, "Task ID is required")
		return
	}

	var req models.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid JSON payload: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		h.writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	task, exists := h.store.Update(id, req)
	if !exists {
		h.writeError(w, http.StatusNotFound, "Task not found")
		return
	}

	h.writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Task updated successfully",
		Data:    task,
	})
}

// ToggleTask toggles a task's completed status.
func (h *TaskHandler) ToggleTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.writeError(w, http.StatusBadRequest, "Task ID is required")
		return
	}

	task, exists := h.store.Toggle(id)
	if !exists {
		h.writeError(w, http.StatusNotFound, "Task not found")
		return
	}

	action := "marked as incomplete"
	if task.Completed {
		action = "marked as completed"
	}

	h.writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Task " + action,
		Data:    task,
	})
}

// DeleteTask removes a task by ID.
func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.writeError(w, http.StatusBadRequest, "Task ID is required")
		return
	}

	if deleted := h.store.Delete(id); !deleted {
		h.writeError(w, http.StatusNotFound, "Task not found")
		return
	}

	h.writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Task deleted successfully",
	})
}

// DeleteCompletedTasks removes all finished tasks.
func (h *TaskHandler) DeleteCompletedTasks(w http.ResponseWriter, r *http.Request) {
	count := h.store.DeleteCompleted()

	h.writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Completed tasks cleared",
		Data:    map[string]int{"deleted_count": count},
	})
}

// GetStats returns summary counts of tasks.
func (h *TaskHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats := h.store.GetStats()

	h.writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    stats,
	})
}

