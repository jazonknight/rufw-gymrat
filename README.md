# GymRat 🏋️‍♂️

**GymRat** is a high-performance, file-based fitness data engine, REST API server, interactive CLI, and web dashboard built in Go. It gives you absolute control over your fitness routines, exercise catalog, and workout tracking without third-party cloud apps, subscription locks, or database overhead.

---

## 🌟 Key Features

* **Decoupled Exercise Catalog (`exercises.json`)**:
  * **Version Control**: Every exercise item maintains an incremental `Version` number. Updating an exercise creates a new version while preserving historical exercise snapshots referenced by past workouts.
  * **Soft Delete (`IsRemoved`)**: Deleting an exercise toggles an `IsRemoved` flag rather than hard-deleting records, maintaining complete auditability.
* **Flexible Set Types**:
  * **Counting Sets (`reps`)**: Track Rep Count, Weight, and Rate of Perceived Exertion (RPE 1-10).
  * **Timed Sets (`timed`)**: Track Duration in Seconds, Weight, and RPE (1-10).
* **Plans & Workouts Vault (`gymrat_plans.json`)**:
  * Track scheduled workout dates (`DatePlanned`) and workout execution status (`IsExecuted`).
  * Attach exercises from the catalog by `(ExerciseId, ExerciseVersion)`.
* **Multi-Tenant Session Manager**:
  * Isolated session workspaces generated under `data/sessions/<sessionId>/`.
  * Support for single-file JSON bundle **Upload** and **Export**.
* **Dual Interface**:
  * **Interactive CLI**: Menu-driven terminal interface for local offline management.
  * **REST API & Web UI**: HTTP server hosting REST endpoints and a modern, dark-mode web dashboard.

---

## 📁 Repository Structure

```
gymrat/
├── cmd/                 # CLI interface handlers and terminal UI
│   ├── cli.go           # Primary CLI menu loop
│   ├── io_helpers.go    # Input reader utilities
│   ├── ui.go            # Terminal banners and styling
│   └── workout_handler.go # Interactive catalog & plan creation routines
├── data/                # Default session workspaces storage directory
├── models/              # Core domain models and configuration
│   ├── config.go        # Set types, units, and validation constants
│   ├── models.go        # Structs (Exercise, ExerciseSet, Workout, Plan, etc.)
│   └── models_test.go   # Unit tests for models and versioning
├── server/              # HTTP REST API server
│   └── server.go        # Handlers for sessions, exercises, plans, upload, export
├── storage/             # File storage and multi-tenant session manager
│   ├── storage.go       # Save/Load functions for JSON catalog & vault
│   └── storage_test.go  # Unit tests for persistence and import/export
├── web/                 # Web Client Frontend
│   ├── index.html       # HTML dashboard layout
│   ├── styles.css       # Glassmorphism dark-mode CSS styling
│   └── app.js           # REST API client & interactive UI controller
├── main.go              # Main application entry point & CLI/Server flag parser
└── storage.go           # Package wrapper delegating root storage calls
```

---

## 🚀 Getting Started

### Prerequisites
* [Go 1.22+](https://golang.org/doc/install) installed on your system.

### Build and Test
Run the unit test suite across all packages:
```bash
go test -v ./...
```

Build the compiled executable binary:
```bash
go build -o gymrat main.go storage.go
```

---

## 💻 Running GymRat

### 1. Interactive Terminal CLI Mode (Default)
Run GymRat directly in your terminal:
```bash
go run main.go storage.go
```
Or run the binary:
```bash
./gymrat
```

**CLI Menu Options:**
1. **View Exercise Catalog**: Displays all active and soft-deleted exercise versions.
2. **Add New Exercise**: Add a new exercise to the catalog (starts at Version 1).
3. **Update Exercise**: Update an exercise, creating a new incremental Version.
4. **Soft Delete Exercise**: Toggle the `IsRemoved` flag on an exercise.
5. **View All Workout Plans**: List all plans and workouts.
6. **Search for a Workout Plan**: Search plans by ID or Name.
7. **Create New Workout Plan**: Interactively build plans with workouts, attach exercise catalog items, and configure Reps or Timed sets.
8. **Mark Workout Executed**: Mark a planned workout as completed (`IsExecuted = true`).
9. **Exit**: Save and close the CLI vault.

---

### 2. REST Server & Web UI Dashboard Mode
Launch the HTTP server and serve the Web Client:
```bash
go run main.go storage.go -server -port 8080
```
Or run the binary with flags:
```bash
./gymrat -server -port 8080
```

Once running, open your web browser to:
👉 **`http://localhost:8080`**

#### Web UI Features:
* **Session Lifecycle**: Automatically manages session IDs. Click "New Session" or switch sessions.
* **Exercise Catalog**: Interactive card grid displaying version tags, category filters, soft-deleted status, and version bump modals.
* **Workout Plans**: Visual builder for creating plans, adding workouts with Reps or Timed sets, and marking workouts executed.
* **Import / Export**: Drag and drop exported `gymrat_vault.json` bundles to import data, or click "Export JSON" to download your current snapshot.

---

## 🌐 REST API Endpoint Reference

All endpoints accept an optional header `X-Session-ID` (or query param `sessionId`). If omitted, a new session workspace is automatically generated.

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `POST` | `/api/session/new` | Creates a new session directory and returns `{ "sessionId": "..." }`. |
| `GET` | `/api/exercises` | List all exercises in the catalog for the active session. |
| `POST` | `/api/exercises` | Create a new exercise (starts at Version 1). |
| `PUT` | `/api/exercises` | Update an existing exercise (bumps version to `Version + 1`). |
| `DELETE` | `/api/exercises?id=<ID>` | Soft-delete an exercise by setting `IsRemoved = true`. |
| `GET` | `/api/plans` | List all workout plans for the active session. |
| `POST` | `/api/plans` | Create a new workout plan with workouts and set configurations. |
| `PUT` | `/api/plans` | Update plan details or toggle a workout's `isExecuted` state. |
| `POST` | `/api/vault/upload` | Import/upload a combined session JSON bundle. |
| `GET` | `/api/vault/export` | Download the active session snapshot as a JSON file. |

---

## 📜 License

Distributed under the MIT License. See `LICENSE` for more information.