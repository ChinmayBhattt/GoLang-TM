package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"taskmanager/internal/handlers"
	"taskmanager/internal/store"
)

func main() {
	// Determine server port from environment variable or use default 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize the in-memory task store (thread-safe with sync.RWMutex)
	taskStore := store.NewMemoryStore()

	// Initialize HTTP handlers
	taskHandler := handlers.NewTaskHandler(taskStore)

	// Create root router (Go 1.22+ enhanced ServeMux)
	mux := http.NewServeMux()

	// Register REST API endpoints
	mux.HandleFunc("GET /api/health", taskHandler.HealthCheck)
	mux.HandleFunc("GET /api/tasks", taskHandler.GetTasks)
	mux.HandleFunc("POST /api/tasks", taskHandler.CreateTask)
	mux.HandleFunc("GET /api/tasks/stats", taskHandler.GetStats)
	mux.HandleFunc("DELETE /api/tasks/completed", taskHandler.DeleteCompletedTasks)
	mux.HandleFunc("GET /api/tasks/{id}", taskHandler.GetTaskByID)
	mux.HandleFunc("PUT /api/tasks/{id}", taskHandler.UpdateTask)
	mux.HandleFunc("PATCH /api/tasks/{id}/toggle", taskHandler.ToggleTask)
	mux.HandleFunc("DELETE /api/tasks/{id}", taskHandler.DeleteTask)

	// Serve static assets (HTML, CSS, JS) from the ./static directory
	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("GET /", fileServer)

	// Wrap router with middleware stack (Recovery -> Logging -> CORS)
	handlerWithMiddleware := handlers.RecoveryMiddleware(
		handlers.LoggingMiddleware(
			handlers.CORSMiddleware(mux),
		),
	)

	// Configure HTTP server with robust timeouts
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handlerWithMiddleware,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel to listen for OS interrupt / termination signals
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	// Start server in a background goroutine
	go func() {
		log.Printf("==================================================")
		log.Printf("🚀 Task Management Server starting on port %s", port)
		log.Printf("👉 Web Application: http://localhost:%s", port)
		log.Printf("👉 REST API Health:  http://localhost:%s/api/health", port)
		log.Printf("👉 Tasks Endpoint:   http://localhost:%s/api/tasks", port)
		log.Printf("==================================================")

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server startup failed: %v", err)
		}
	}()

	// Block until an interrupt signal is received
	<-shutdownChan
	log.Println("\nShutdown signal received. Shutting down gracefully...")

	// Create context with 5-second deadline for active connections to finish
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	} else {
		log.Println("Server exited cleanly. Goodbye!")
	}
}
