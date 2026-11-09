package store

import (
	"testing"
	"time"

	"taskmanager/internal/models"
)

func TestMemoryStore_CRUD(t *testing.T) {
	s := NewMemoryStore()

	// Initial seed data check
	stats := s.GetStats()
	if stats.Total != 3 {
		t.Fatalf("expected 3 seed tasks, got %d", stats.Total)
	}

	// 1. Create Task
	req := models.CreateTaskRequest{
		Title:       "Test task",
		Description: "A unit test task",
		Priority:    models.PriorityHigh,
		DueDate:     "2026-12-31",
	}
	created := s.Create(req)
	if created.ID == "" {
		t.Fatalf("expected non-empty ID for created task")
	}
	if created.Title != req.Title {
		t.Errorf("expected title '%s', got '%s'", req.Title, created.Title)
	}
	if created.Completed {
		t.Errorf("expected new task to not be completed")
	}

	// 2. Get By ID
	fetched, found := s.GetByID(created.ID)
	if !found {
		t.Fatalf("expected task with ID %s to be found", created.ID)
	}
	if fetched.ID != created.ID {
		t.Errorf("expected fetched ID %s, got %s", created.ID, fetched.ID)
	}

	// 3. Toggle Status
	toggled, ok := s.Toggle(created.ID)
	if !ok || !toggled.Completed {
		t.Fatalf("expected task to be marked completed")
	}

	// 4. Update Task
	newTitle := "Updated Title"
	newPri := models.PriorityLow
	updated, ok := s.Update(created.ID, models.UpdateTaskRequest{
		Title:    &newTitle,
		Priority: &newPri,
	})
	if !ok {
		t.Fatalf("expected update to succeed")
	}
	if updated.Title != newTitle {
		t.Errorf("expected title '%s', got '%s'", newTitle, updated.Title)
	}
	if updated.Priority != newPri {
		t.Errorf("expected priority '%s', got '%s'", newPri, updated.Priority)
	}

	// 5. Delete Task
	deleted := s.Delete(created.ID)
	if !deleted {
		t.Fatalf("expected delete to succeed")
	}
	_, foundAfterDelete := s.GetByID(created.ID)
	if foundAfterDelete {
		t.Errorf("expected task to be deleted")
	}
}

func TestMemoryStore_FilterAndSearch(t *testing.T) {
	s := &MemoryStore{
		tasks: make(map[string]models.Task),
	}

	now := time.Now()
	s.tasks["1"] = models.Task{
		ID:        "1",
		Title:     "Buy Groceries",
		Completed: false,
		Priority:  models.PriorityLow,
		CreatedAt: now.Add(-3 * time.Minute),
	}
	s.tasks["2"] = models.Task{
		ID:        "2",
		Title:     "Deploy Go Service",
		Completed: true,
		Priority:  models.PriorityHigh,
		CreatedAt: now.Add(-2 * time.Minute),
	}
	s.tasks["3"] = models.Task{
		ID:        "3",
		Title:     "Write Go unit tests",
		Completed: false,
		Priority:  models.PriorityMedium,
		CreatedAt: now.Add(-1 * time.Minute),
	}

	// Filter: active
	activeTasks := s.GetAll("active", "", "newest")
	if len(activeTasks) != 2 {
		t.Errorf("expected 2 active tasks, got %d", len(activeTasks))
	}

	// Filter: completed
	completedTasks := s.GetAll("completed", "", "newest")
	if len(completedTasks) != 1 {
		t.Errorf("expected 1 completed task, got %d", len(completedTasks))
	}

	// Search keyword "go"
	searchResults := s.GetAll("all", "go", "newest")
	if len(searchResults) != 2 {
		t.Errorf("expected 2 search results for 'go', got %d", len(searchResults))
	}

	// Sort by priority
	prioritySorted := s.GetAll("all", "", "priority")
	if len(prioritySorted) != 3 || prioritySorted[0].Priority != models.PriorityHigh {
		t.Errorf("expected High priority task first, got %v", prioritySorted[0].Priority)
	}
}
