# GymRat 🏋️‍♂️

**GymRat** is a high-performance, privacy-first fitness data engine, REST API server, interactive CLI, and web dashboard built in Go. It gives you complete, client-owned control over your exercise library, workout routines, training plans, and live gym workout tracking without third-party cloud apps, subscription paywalls, or database overhead.

---

## 🌟 Key Features

* **3-Tier Fitness Hierarchy**:
  * **1. Exercise Library (`/api/exercises`)**: Independent movement definitions with incremental version history (`Version`) and soft deletion (`IsRemoved`).
  * **2. Workout Routines (`/api/workouts`)**: Standalone workout templates (e.g. *"Push Day A"*, *"Lower Body Heavy"*) built by selecting exercises from your library and setting up target sets.
  * **3. Training Plans & Programs (`/api/plans`)**: High-level macrocycles (e.g. *"12-Week Hypertrophy Program"*) containing a schedule of pre-created Workout Routines with specific planned execution dates (`datePlanned`).
* **🏋️‍♂️ Live Workout Execution Logger**:
  * **Interactive Session Tracker**: Click **"▶ Start Workout"** to launch an interactive live workout overlay with an elapsed session duration timer.
  * **Actual Reps & Weight Logging**: Record actual weight lifted and actual reps completed per set during your gym session.
  * **⏱️ Inter-Set Rest Countdown Timer**: Automatic 30-second rest countdown timer triggered when you check off a completed set (`✓ Set Done`).
  * **Multi-Set Support (`+ Add Set`)**: Add extra sets on-the-fly during a live workout or when building routine templates.
* **📜 Completed Workout History (`/api/history`)**:
  * Chronological log of past completed workouts displaying completion timestamps, total session duration, total volume lifted (lbs), and actual set performance.
* **🔍 Real-Time Exercise Search & Muscle Group Filter**:
  * Search bar and dynamic muscle group dropdown filter (*Legs, Chest, Back, Shoulders, Core*) in the Exercise Library.
* **⚡ Stateless In-Memory REST Engine (`STORAGE_MODE=memory`)**:
  * Operates statelessly in server RAM with **0 server disk writes** by default. Users own their JSON data locally and sync/export via the Web UI.
* **🛡 Security **:
  * Configurable upload body size limit (`MAX_PAYLOAD_MB`) to prevent RAM exhaustion.

---

## 📁 Repository Structure

```
gymrat/
├── cmd/                 # CLI interface handlers and terminal UI
│   ├── cli.go           # Primary CLI menu loop
│   ├── io_helpers.go    # Input reader utilities
│   ├── ui.go            # Terminal banners and styling
│   └── workout_handler.go # Interactive catalog & plan creation routines
├── data/                # Optional persistence storage directory
├── models/              # Core domain models and configuration
│   ├── config.go        # Set types, units, and validation constants
│   ├── models.go        # Structs (Exercise, ExerciseSet, Workout, Plan, HistoricWorkouts, etc.)
│   └── models_test.go   # Unit tests for models and versioning
├── server/              # HTTP REST API server
│   └── server.go        # Handlers for sessions, exercises, workouts, plans, history, upload, export
├── storage/             # Session manager & JSON seed catalog
│   ├── default_exercises.json # Immutable starter exercise catalog checked into Git
│   ├── storage.go       # SessionManager, embedded JSON seed loader, import/export
│   └── storage_test.go  # Unit tests for session CRUD, import/export, and seed copying
├── web/                 # Web Client Frontend
│   ├── index.html       # HTML dashboard layout with 3-tier tabs & modals
│   ├── styles.css       # Glassmorphism dark-mode CSS design system
│   └── app.js           # Live logger controller, timer intervals, search/filter, & REST API client
├── main.go              # Application entry point & CLI/Server flag parser
├── security_assessment.md # Comprehensive security threat model & assessment report
├── Dockerfile           # Multi-stage lightweight Alpine build definition
├── docker-compose.yml   # Docker Compose stack with optional Cloudflare Tunnel service
└── .env.example         # Environment configuration template
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
go build -o gymrat main.go
```

---

## 💻 Running GymRat

### 1. Interactive Terminal CLI Mode (Default)
Run GymRat directly in your terminal:
```bash
go run main.go
```
Or run the compiled binary:
```bash
./gymrat
```

---

### 2. REST Server & Web UI Dashboard Mode
Launch the HTTP REST server and serve the Web Client:
```bash
go run main.go -server -port 8080
```
Or run the binary with flags:
```bash
./gymrat -server -port 8080 -storage-mode memory -max-payload-mb 5
```

Once running, open your web browser to:
👉 **`http://localhost:8080`**

---

### 3. Docker Deployment 🐳

GymRat includes a multi-stage `Dockerfile` and `docker-compose.yml` pre-configured for Cloudflare Tunnel (`gymrat.rufw.io`).

#### Quick Launch:
```bash
docker compose up -d --build
```


#### Environment Variables (`.env`):

| Variable | Default | Description |
| :--- | :--- | :--- |
| `PORT` | `8080` | Server HTTP port mapping. |
| `DOMAIN_NAME` | `gymrat.rufw.io` | Target domain name for reverse proxy routing. |
| `STORAGE_MODE` | `memory` | Session storage mode (`memory` for 0 disk writes, `disk` for file persistence). |
| `MAX_PAYLOAD_MB` | `5` | Maximum JSON upload payload size limit in MB. |
| `SESSIONS_DIR` | `/app/data/sessions` | Container directory path for session JSON data. |
| `WEB_DIR` | `/app/web` | Container directory path to static web assets. |
| `SERVER_MODE` | `true` | Auto-starts REST API server on container boot. |

---

## 🌐 REST API Endpoint Reference

All REST endpoints accept an optional header `X-Session-ID` (or query param `sessionId`). If omitted, a new session workspace is automatically generated and returned in the `X-Session-ID` response header.

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `POST` | `/api/session/new` | Generates a new session and deep-copies the default starter exercise catalog into memory. |
| `GET` | `/api/exercises` | List all exercises in the catalog for the active session. |
| `POST` | `/api/exercises` | Create a new exercise in the catalog (starts at Version 1). |
| `PUT` | `/api/exercises` | Update an existing exercise (bumps version to `Version + 1`). |
| `DELETE` | `/api/exercises?id=<ID>` | Soft-delete or restore an exercise by toggling `IsRemoved`. |
| `GET` | `/api/workouts` | List all standalone Workout Routine templates in the active session. |
| `POST` | `/api/workouts` | Create a new standalone Workout Routine template. |
| `POST` | `/api/workouts/copy` | Duplicates an existing Workout Routine template with a fresh UUID. |
| `PUT` | `/api/workouts` | Update an existing Workout Routine template. |
| `DELETE` | `/api/workouts?id=<ID>` | Delete a Workout Routine template by ID. |
| `GET` | `/api/plans` | List all Training Plans in the active session. |
| `POST` | `/api/plans` | Create a new Training Plan with scheduled workout instances (`datePlanned`). |
| `POST` | `/api/plans/copy` | Duplicates an existing Training Plan with fresh UUIDs for all workout instances. |
| `PUT` | `/api/plans` | Update plan details or toggle a scheduled workout's `isExecuted` state. |
| `GET` | `/api/history` | List all completed workout history logs chronologically. |
| `POST` | `/api/history` | Record a completed live workout session log (duration, actual reps/weights, total volume). |
| `POST` | `/api/vault/upload` | Import/upload a combined session JSON vault file. |
| `GET` | `/api/vault/export` | Export and download the active session snapshot as a JSON bundle file. |

---

## 📜 License

Distributed under the GNU Affero General Public License v3.0 (AGPL-3.0). See `LICENSE` for more information.