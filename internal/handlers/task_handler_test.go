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

	return RecoveryMiddleware(LoggingMiddleware(CORSMiddleware(mux)))
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

func TestGetTasksAndStats(t *testing.T) {
	router := setupTestRouter()

	// 1. Get all tasks
	req := httptest.NewRequest(http.MethodGet, "/api/tasks?status=all&sort=newest", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var res struct {
		Success bool          `json:"success"`
		Data    []models.Task `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(res.Data) != 3 {
		t.Errorf("expected 3 seed tasks, got %d", len(res.Data))
	}

	// 2. Get Stats
	statsReq := httptest.NewRequest(http.MethodGet, "/api/tasks/stats", nil)
	statsRec := httptest.NewRecorder()
	router.ServeHTTP(statsRec, statsReq)

	if statsRec.Code != http.StatusOK {
		t.Fatalf("expected stats status 200, got %d", statsRec.Code)
	}
}

func TestTaskCRUDLifecycle(t *testing.T) {
	router := setupTestRouter()

	// 1. Create task
	payload := []byte(`{"title":"Write documentation","description":"Detailed notes","priority":"high","due_date":"2026-11-15"}`)
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

	// 2. Get task by ID
	getReq := httptest.NewRequest(http.MethodGet, "/api/tasks/"+taskID, nil)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Errorf("expected status 200 for GetTaskByID, got %d", getRec.Code)
	}

	// 3. Update task
	updatePayload := []byte(`{"title":"Updated Title"}`)
	updateReq := httptest.NewRequest(http.MethodPut, "/api/tasks/"+taskID, bytes.NewBuffer(updatePayload))
	updateReq.Header.Set("Content-Type", "application/json")
	updateRec := httptest.NewRecorder()
	router.ServeHTTP(updateRec, updateReq)
	if updateRec.Code != http.StatusOK {
		t.Errorf("expected status 200 for update, got %d", updateRec.Code)
	}

	// 4. Toggle task
	toggleReq := httptest.NewRequest(http.MethodPatch, "/api/tasks/"+taskID+"/toggle", nil)
	toggleRec := httptest.NewRecorder()
	router.ServeHTTP(toggleRec, toggleReq)
	if toggleRec.Code != http.StatusOK {
		t.Errorf("expected status 200 on toggle, got %d", toggleRec.Code)
	}

	// 5. Delete task
	delReq := httptest.NewRequest(http.MethodDelete, "/api/tasks/"+taskID, nil)
	delRec := httptest.NewRecorder()
	router.ServeHTTP(delRec, delReq)
	if delRec.Code != http.StatusOK {
		t.Errorf("expected status 200 on delete, got %d", delRec.Code)
	}

	// 6. Delete already deleted task -> 404
	delNotFoundReq := httptest.NewRequest(http.MethodDelete, "/api/tasks/"+taskID, nil)
	delNotFoundRec := httptest.NewRecorder()
	router.ServeHTTP(delNotFoundRec, delNotFoundReq)
	if delNotFoundRec.Code != http.StatusNotFound {
		t.Errorf("expected status 404 for deleted task, got %d", delNotFoundRec.Code)
	}
}

func TestDeleteCompletedTasks(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodDelete, "/api/tasks/completed", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200 for DeleteCompletedTasks, got %d", rec.Code)
	}
}

func TestValidationErrors(t *testing.T) {
	router := setupTestRouter()

	// Empty title
	payload := []byte(`{"title":"","priority":"low"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for empty title, got %d", rec.Code)
	}

	// Invalid JSON
	badJSONReq := httptest.NewRequest(http.MethodPost, "/api/tasks", bytes.NewBufferString("invalid json"))
	badJSONRec := httptest.NewRecorder()
	router.ServeHTTP(badJSONRec, badJSONReq)
	if badJSONRec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for malformed json, got %d", badJSONRec.Code)
	}

	// Task Not Found for GET
	notFoundReq := httptest.NewRequest(http.MethodGet, "/api/tasks/nonexistent_id", nil)
	notFoundRec := httptest.NewRecorder()
	router.ServeHTTP(notFoundRec, notFoundReq)
	if notFoundRec.Code != http.StatusNotFound {
		t.Errorf("expected status 404 for nonexistent task, got %d", notFoundRec.Code)
	}
}

func TestCORSPreflight(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest(http.MethodOptions, "/api/tasks", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status 204 for OPTIONS, got %d", rec.Code)
	}
}
