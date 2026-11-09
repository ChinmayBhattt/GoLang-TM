package store

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"taskmanager/internal/models"
)

// TaskStore defines the interface for task storage operations.
type TaskStore interface {
	GetAll(status, search, sortBy string) []models.Task
	GetByID(id string) (models.Task, bool)
	Create(req models.CreateTaskRequest) models.Task
	Update(id string, req models.UpdateTaskRequest) (models.Task, bool)
	Toggle(id string) (models.Task, bool)
	Delete(id string) bool
	DeleteCompleted() int
	GetStats() models.TaskStats
}

// MemoryStore is an in-memory, thread-safe implementation of TaskStore.
type MemoryStore struct {
	mu    sync.RWMutex
	tasks map[string]models.Task
}

// NewMemoryStore creates and returns an initialized in-memory store with starter data.
func NewMemoryStore() *MemoryStore {
	s := &MemoryStore{
		tasks: make(map[string]models.Task),
	}
	s.seedInitialData()
	return s
}

// generateID generates a unique hex-encoded identifier.
func generateID() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp if crypto/rand fails
		return fmt.Sprintf("task_%d", time.Now().UnixNano())
	}
	return "task_" + hex.EncodeToString(bytes)
}

// seedInitialData populates the store with starter tasks for immediate exploration.
func (s *MemoryStore) seedInitialData() {
	now := time.Now()
	today := now.Format("2006-01-02")
	tomorrow := now.Add(24 * time.Hour).Format("2006-01-02")

	initialTasks := []models.Task{
		{
			ID:          "task_seed_1",
			Title:       "Explore Go REST APIs & In-Memory Store 🚀",
			Description: "The backend is powered by Go 1.22+ standard library with sync.RWMutex for safe concurrent operations.",
			Completed:   true,
			Priority:    models.PriorityHigh,
			DueDate:     today,
			CreatedAt:   now.Add(-2 * time.Hour),
			UpdatedAt:   now.Add(-1 * time.Hour),
		},
		{
			ID:          "task_seed_2",
			Title:       "Test adding, completing, and deleting tasks ✨",
			Description: "Try the form above to add your own tasks, filter them by status, or search keywords.",
			Completed:   false,
			Priority:    models.PriorityMedium,
			DueDate:     tomorrow,
			CreatedAt:   now.Add(-1 * time.Hour),
			UpdatedAt:   now.Add(-1 * time.Hour),
		},
		{
			ID:          "task_seed_3",
			Title:       "Toggle dark/light theme and check responsiveness 🌓",
			Description: "Modern UI built with clean HTML5, CSS variables, and vanilla JavaScript.",
			Completed:   false,
			Priority:    models.PriorityLow,
			DueDate:     "",
			CreatedAt:   now.Add(-30 * time.Minute),
			UpdatedAt:   now.Add(-30 * time.Minute),
		},
	}

	for _, t := range initialTasks {
		s.tasks[t.ID] = t
	}
}

// GetAll returns all tasks filtered by status and search, and sorted accordingly.
func (s *MemoryStore) GetAll(status, search, sortBy string) []models.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	searchLower := strings.ToLower(strings.TrimSpace(search))
	statusLower := strings.ToLower(strings.TrimSpace(status))

	result := make([]models.Task, 0, len(s.tasks))

	for _, task := range s.tasks {
		// Filter by completion status
		if statusLower == "active" && task.Completed {
			continue
		}
		if statusLower == "completed" && !task.Completed {
			continue
		}

		// Filter by search query (title or description)
		if searchLower != "" {
			matchesTitle := strings.Contains(strings.ToLower(task.Title), searchLower)
			matchesDesc := strings.Contains(strings.ToLower(task.Description), searchLower)
			if !matchesTitle && !matchesDesc {
				continue
			}
		}

		result = append(result, task)
	}

	// Sort results
	sortTasks(result, sortBy)

	return result
}

// sortTasks orders tasks based on user preference (default: newest first).
func sortTasks(tasks []models.Task, sortBy string) {
	priorityWeight := func(p models.Priority) int {
		switch p {
		case models.PriorityHigh:
			return 3
		case models.PriorityMedium:
			return 2
		case models.PriorityLow:
			return 1
		default:
			return 0
		}
	}

	switch strings.ToLower(sortBy) {
	case "oldest":
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].CreatedAt.Before(tasks[j].CreatedAt)
		})
	case "priority":
		sort.Slice(tasks, func(i, j int) bool {
			wi, wj := priorityWeight(tasks[i].Priority), priorityWeight(tasks[j].Priority)
			if wi != wj {
				return wi > wj // High priority first
			}
			return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
		})
	case "due_date":
		sort.Slice(tasks, func(i, j int) bool {
			// Tasks with due dates come before tasks without
			if tasks[i].DueDate == "" && tasks[j].DueDate != "" {
				return false
			}
			if tasks[i].DueDate != "" && tasks[j].DueDate == "" {
				return true
			}
			if tasks[i].DueDate != tasks[j].DueDate {
				return tasks[i].DueDate < tasks[j].DueDate
			}
			return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
		})
	default: // "newest" or unspecified
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
		})
	}
}

// GetByID finds a task by its unique ID.
func (s *MemoryStore) GetByID(id string) (models.Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.tasks[id]
	return task, exists
}

// Create adds a new task to the store.
func (s *MemoryStore) Create(req models.CreateTaskRequest) models.Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	task := models.Task{
		ID:          generateID(),
		Title:       req.Title,
		Description: req.Description,
		Completed:   false,
		Priority:    req.Priority,
		DueDate:     req.DueDate,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	s.tasks[task.ID] = task
	return task
}

// Update modifies an existing task by its ID.
func (s *MemoryStore) Update(id string, req models.UpdateTaskRequest) (models.Task, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[id]
	if !exists {
		return models.Task{}, false
	}

	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Completed != nil {
		task.Completed = *req.Completed
	}
	if req.Priority != nil {
		task.Priority = *req.Priority
	}
	if req.DueDate != nil {
		task.DueDate = *req.DueDate
	}

	task.UpdatedAt = time.Now()
	s.tasks[id] = task
	return task, true
}

// Toggle flips the completion status of a task.
func (s *MemoryStore) Toggle(id string) (models.Task, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[id]
	if !exists {
		return models.Task{}, false
	}

	task.Completed = !task.Completed
	task.UpdatedAt = time.Now()
	s.tasks[id] = task
	return task, true
}

// Delete removes a task by its ID.
func (s *MemoryStore) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[id]; !exists {
		return false
	}

	delete(s.tasks, id)
	return true
}

// DeleteCompleted removes all tasks that are marked as completed.
func (s *MemoryStore) DeleteCompleted() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	deletedCount := 0
	for id, task := range s.tasks {
		if task.Completed {
			delete(s.tasks, id)
			deletedCount++
		}
	}
	return deletedCount
}

// GetStats returns summary metrics for the tasks.
func (s *MemoryStore) GetStats() models.TaskStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := len(s.tasks)
	completed := 0
	for _, task := range s.tasks {
		if task.Completed {
			completed++
		}
	}
	pending := total - completed

	rate := 0
	if total > 0 {
		rate = int((float64(completed) / float64(total)) * 100)
	}

	return models.TaskStats{
		Total:          total,
		Completed:      completed,
		Pending:        pending,
		CompletionRate: rate,
	}
}

