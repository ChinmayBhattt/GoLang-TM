package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"taskmanager/internal/models"
	"taskmanager/internal/store"
)

func setupTestRouter() http.Handler {
	s := store.NewMemoryStore()
	h := NewTaskHandler(s)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", h.HealthCheck)
	mux.HandleFunc("GET /api/tasks", h.GetTasks)
	mux.HandleFunc("POST /api/tasks", h.CreateTask)
	mux.HandleFunc("GET /api/tasks/stats", h.GetStats)
	mux.HandleFunc("DELETE /api/tasks/completed", h.DeleteCompletedTasks)
	mux.HandleFunc("GET /api/tasks/{id}", h.GetTaskByID)
	mux.HandleFunc("PUT /api/tasks/{id}", h.UpdateTask)
	mux.HandleFunc("PATCH /api/tasks/{id}/toggle", h.ToggleTask)
	mux.HandleFunc("DELETE /api/tasks/{id}", h.DeleteTask)

	return mux
}

func TestHealthCheck(t *testing.T) {
	router := setupTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var res models.APIResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !res.Success {
		t.Errorf("expected success to be true")
	}
}

func TestTaskCRUDLifecycle(t *testing.T) {
	router := setupTestRouter()

	// 1. Create task
	payload := []byte(`{"title":"Write documentation","priority":"high","due_date":"2026-11-15"}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBuffer(payload))
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()

	router.ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", createRec.Code, createRec.Body.String())
	}

	var createRes struct {
		Success bool        `json:"success"`
		Data    models.Task `json:"data"`
	}
	if err := json.NewDecoder(createRec.Body).Decode(&createRes); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	taskID := createRes.Data.ID
	if taskID == "" {
		t.Fatalf("expected non-empty task ID")
	}

	// 2. Toggle task
	toggleReq := httptest.NewRequest(http.MethodPatch, "/api/tasks/"+taskID+"/toggle", nil)
	toggleRec := httptest.NewRecorder()

	router.ServeHTTP(toggleRec, toggleReq)
	if toggleRec.Code != http.StatusOK {
		t.Errorf("expected status 200 on toggle, got %d", toggleRec.Code)
	}

	// 3. Delete task
	delReq := httptest.NewRequest(http.MethodDelete, "/api/tasks/"+taskID, nil)
	delRec := httptest.NewRecorder()

	router.ServeHTTP(delRec, delReq)
	if delRec.Code != http.StatusOK {
		t.Errorf("expected status 200 on delete, got %d", delRec.Code)
	}
}

func TestCreateTaskValidation(t *testing.T) {
	router := setupTestRouter()

	// Empty title should fail with 400
	payload := []byte(`{"title":"","priority":"low"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for empty title, got %d", rec.Code)
	}
}
