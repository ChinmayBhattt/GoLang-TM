# TaskFlow — Modern Go Task Management Application

A simple, fast, and modern task management web application built with **Go (Golang) as the backend** and **vanilla HTML, CSS, and JavaScript** for the frontend.

Designed with clean architecture, beginner-friendly Go patterns, standard library routing, and a responsive UI.

---

## 🌟 Key Features

### Backend (Go)
- **Zero External Dependencies**: Built entirely with Go's standard library (`net/http`, `encoding/json`, `sync`, `time`).
- **RESTful JSON API**: Endpoints for full CRUD (Create, Read, Update, Delete), status toggling, filtering, and summary statistics.
- **Thread-Safe In-Memory Storage**: Uses `sync.RWMutex` to guarantee safe concurrent reads and writes across goroutines.
- **Go 1.22+ Standard Routing**: Clean path parameter extraction (`r.PathValue("id")`) and method matching.
- **Middleware Pipeline**: Request logging with execution duration, CORS preflight handling, and panic recovery.
- **Graceful Shutdown**: Listens for interrupt signals (`SIGINT`, `SIGTERM`) and allows active requests to complete cleanly.

### Frontend (HTML, CSS, JavaScript)
- **Modern Responsive Design**: Clean cards, subtle blur glows, and adaptive layout that looks great on mobile, tablet, and desktop.
- **Light & Dark Mode**: Instant theme switching persisted in `localStorage`.
- **Live Stats & Progress Bar**: Real-time counter of total, active, and completed tasks with an animated percentage completion bar.
- **Filtering, Search & Sort**: Filter by *All*, *Active*, or *Completed*; live substring search; sort by *Newest*, *Oldest*, *Priority*, or *Due Date*.
- **Task Management**:
  - Add tasks with title, optional description, priority (*Low*, *Medium*, *High*), and due date.
  - Complete / uncomplete tasks with a single click.
  - Edit task details in an accessible modal dialog.
  - Delete individual tasks or batch clear all completed tasks.
- **Toast Notifications**: Floating feedback alerts for all actions.
- **Live Backend Status**: Visual indicator displaying API connectivity.

---

## 📂 Project Structure

```text
golang/
├── go.mod                     # Go module definition
├── main.go                    # Application entry point & server setup
├── internal/
│   ├── models/
│   │   └── task.go            # Task, Request/Response payloads, & validation
│   ├── store/
│   │   ├── memory_store.go    # Thread-safe in-memory store (sync.RWMutex)
│   │   └── memory_store_test.go # Store unit tests
│   └── handlers/
│       ├── task_handler.go    # REST API HTTP handlers
│       ├── middleware.go      # Logging, CORS, and Recovery middleware
│       └── task_handler_test.go # Comprehensive HTTP endpoint tests
├── static/
│   ├── index.html             # Semantic, accessible HTML5 layout
│   ├── css/
│   │   └── styles.css         # Modern styling, CSS variables, dark/light themes
│   └── js/
│       └── app.js             # Vanilla JS state, API client, & DOM interactions
├── .gitignore
└── README.md                  # Project documentation
```

---

## 🚀 Getting Started

### Prerequisites
- [Go](https://go.dev/dl/) installed (version 1.22 or newer recommended).

### Running the Application

1. Open your terminal in this directory:
   ```bash
   cd /Users/chinmaybhatt/golang
   ```

2. Run the application:
   ```bash
   go run .
   ```

3. Open your browser and visit:
   ```
   http://localhost:8080
   ```

To specify a custom port:
```bash
PORT=3000 go run .
```

---

## 🧪 Running Unit Tests

Run all unit tests with full coverage:

```bash
go test -v -cover ./...
```

---

## 📡 REST API Reference

All requests and responses use JSON format (`application/json`).

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/api/health` | Service health check |
| `GET` | `/api/tasks` | Get all tasks (supports `status`, `search`, `sort`) |
| `POST` | `/api/tasks` | Create a new task |
| `GET` | `/api/tasks/{id}` | Get single task by ID |
| `PUT` | `/api/tasks/{id}` | Update task details |
| `PATCH` | `/api/tasks/{id}/toggle` | Toggle task completion |
| `DELETE` | `/api/tasks/{id}` | Delete task by ID |
| `DELETE` | `/api/tasks/completed` | Clear all completed tasks |
| `GET` | `/api/tasks/stats` | Get task counts & completion percentage |

### Request & Response Examples

#### 1. Create Task
- **`POST /api/tasks`**
- **Request Body**:
  ```json
  {
    "title": "Build landing page",
    "description": "Design modern layout with CSS Grid",
    "priority": "high",
    "due_date": "2026-11-15"
  }
  ```
- **Response `201 Created`**:
  ```json
  {
    "success": true,
    "message": "Task created successfully",
    "data": {
      "id": "task_8135f6acc0080ddd",
      "title": "Build landing page",
      "description": "Design modern layout with CSS Grid",
      "completed": false,
      "priority": "high",
      "due_date": "2026-11-15",
      "created_at": "2026-11-09T10:05:37Z",
      "updated_at": "2026-11-09T10:05:37Z"
    }
  }
  ```

#### 2. Toggle Task Completion
- **`PATCH /api/tasks/{id}/toggle`**
- **Response `200 OK`**:
  ```json
  {
    "success": true,
    "message": "Task marked as completed",
    "data": {
      "id": "task_8135f6acc0080ddd",
      "title": "Build landing page",
      "completed": true,
      ...
    }
  }
  ```

---

## 🔒 Concurrency & Thread Safety

Go's HTTP server handles incoming requests across concurrent goroutines. To prevent race conditions, the in-memory store uses `sync.RWMutex`:
- Multiple read operations (e.g. `GetAll`, `GetByID`, `GetStats`) acquire `RLock()` simultaneously for high read throughput.
- Write operations (e.g. `Create`, `Update`, `Toggle`, `Delete`) acquire an exclusive `Lock()`.

---

## 📄 License
MIT
