# GymRat 🏋️‍♂️

**GymRat** is a high-performance, privacy-first fitness data engine, REST API server, interactive CLI, and mobile-friendly web dashboard built in Go. It gives you complete, client-owned control over your exercise library, workout routines, training plans, and live gym workout tracking without third-party cloud apps, subscription paywalls, or database overhead.

---

## 🌟 Key Features

* **3-Tier Fitness Hierarchy**:
  * **1. Exercise Library (`/api/exercises`)**: Independent movement definitions with incremental version history (`Version`) and soft deletion (`IsRemoved`).
  * **2. Workout Routines (`/api/workouts`)**: Standalone workout templates (e.g. *"Push Day A"*, *"Lower Body Heavy"*) built by selecting exercises from your library and setting up target sets.
  * **3. Training Plans & Programs (`/api/plans`)**: High-level macrocycles (e.g. *"12-Week Hypertrophy Program"*) containing a schedule of pre-created Workout Routines with specific planned execution dates (`datePlanned`).
* **🏋️‍♂️ Mobile-First Live Workout Logger**:
  * **Interactive Session Tracker**: Click **"▶ Start Workout"** to launch an interactive live workout logger.
  * **Real-Time Per-Set Input Sync**: Edit weight (lbs) or reps/sec up and down on the fly for any set during a live workout.
  * **Stateful Set Completion Checkmarks**: `✓ Completed` states stay active when adding new sets or running rest timers.
  * **On-the-Fly Set & Exercise Removal (`🗑️`)**: Skip or delete individual sets or entire exercises on the fly mid-workout.
  * **⏱️ Inter-Set Rest Countdown Timer**: Automatic 30-second rest countdown timer triggered when checking off a set.
  * **Fat-Thumb Touch Targets (44px Minimum)**: Optimized for smartphone screens with zero auto-zoom on iOS Safari.
* **📋 Server-Side Duplication (`/api/workouts/copy`, `/api/plans/copy`)**:
  * Instant 1-click **Duplicate Routine** & **Duplicate Plan** with server-generated fresh UUIDs to prevent ID collision.
* **📜 Completed Workout History (`/api/history`)**:
  * Chronological log of past completed workouts displaying completion timestamps, total session duration, total volume lifted (lbs), and set breakdowns.
* **🔍 Real-Time Exercise Search & Muscle Group Filter**:
  * Search bar and dynamic muscle group dropdown filter (*Legs, Chest, Back, Shoulders, Core*) in the Exercise Library.
* **⚡ Stateless In-Memory REST Engine (`STORAGE_MODE=memory`)**:
  * Operates statelessly in server RAM with **0 server disk writes** by default. Users own their JSON data locally and sync/export via the Web UI.
* **🛡 Security & Stability**:
  * Configurable upload body size limit (`MAX_PAYLOAD_MB`) and UUID input validation to prevent path traversal and RAM exhaustion DoS.

---

## 📁 Repository & Modular Architecture

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
├── scripts/             # Executable runner and deployment scripts
│   ├── build_run_local.sh # Local Go compilation & server launcher
│   └── publish_docker.sh  # Multi-arch Docker build & push script
├── server/              # Modular HTTP REST API server
│   ├── server.go        # Core server setup, middleware, and route registration
│   ├── exercises.go     # Exercise catalog handlers (GET/POST/PUT/DELETE /api/exercises)
│   ├── workouts.go      # Workout routine handlers (GET/POST/PUT/DELETE /api/workouts & /copy)
│   ├── plans.go         # Training plan handlers (GET/POST/PUT /api/plans & /copy)
│   ├── history.go       # Workout history handlers (GET/POST /api/history)
│   └── vault.go         # Session initialization & JSON vault upload/export
├── storage/             # Session manager & JSON seed catalog
│   ├── default_exercises.json # Starter exercise catalog
│   ├── default_workouts.json  # Starter workout routine templates
│   ├── default_plans.json     # Starter training plans
│   ├── storage.go       # SessionManager, embedded JSON seed loader, import/export
│   └── storage_test.go  # Unit tests for session CRUD, import/export, and seed copying
├── web/                 # Modular Web Client Frontend
│   ├── js/              # ES6 JavaScript Component Modules
│   │   ├── app.js       # Main SPA entry point & modal/tab initializers
│   │   ├── state.js     # Central state management & app constants
│   │   ├── api.js       # Dedicated REST API client functions
│   │   └── components/  # Focused UI Component Modules
│   │       ├── catalog.js    # Exercise library rendering & category filtering
│   │       ├── routines.js   # Workout routine cards & routine builder
│   │       ├── plans.js      # Training plan cards & schedule builder
│   │       ├── history.js    # Completed workout history tab
│   │       └── liveLogger.js # Live workout session logger & timer
│   ├── index.html       # Single-page dashboard HTML layout
│   └── styles.css       # Glassmorphism dark-mode CSS & mobile media queries
├── .github/workflows/ci.yml # Automated GitHub Actions CI pipeline
├── .gitlab-ci.yml       # GitLab CI/CD pipeline definition
├── .gitea/workflows/ci.yml  # Gitea Actions CI/CD workflow
├── main.go              # Application entry point & CLI/Server flag parser
├── Dockerfile           # Multi-stage Alpine build definition
├── docker-compose.yml   # Docker Compose stack definition
└── .env.example         # Environment configuration template
```

---

## 🚀 Commands & Quickstart

### Prerequisites
* [Go 1.23+](https://golang.org/doc/install) installed on your system.
* [Docker Desktop](https://www.docker.com/) (Optional, for container deployment).

---

### 1. Run Unit Test Suite
Verify that all unit tests pass cleanly across models and storage packages:
```bash
go test -v ./...
```

---

### 2. Run Locally via Executable Script (Recommended)
Use the included helper script to build and launch GymRat locally:
```bash
./scripts/build_run_local.sh
```
*(Or set a custom port: `PORT=9876 ./scripts/build_run_local.sh`)*

---

### 3. Run Locally via Go CLI
Run directly via Go without building an explicit binary:
```bash
go run main.go -server -port 8080
```

Or compile and run the Go binary:
```bash
go build -o gymrat main.go
./gymrat -server -port 8080
```

Once running, open your web browser to:
👉 **`http://localhost:8080`**

---

### 4. Interactive Terminal CLI Mode
Run GymRat in terminal menu mode (no web server required):
```bash
go run main.go
```

---

### 5. Deploy via Docker / Synology Container Manager 🐳

#### Option A: Docker Compose Command
```bash
docker compose up -d --build
```

#### Option B: Synology Container Manager (DSM 7.2+)
In Synology Container Manager ➔ **Project** ➔ **Create**:

```yaml
version: '3.8'

services:
  gymrat:
    image: jazonknight/rufw-gymrat:latest
    container_name: gymrat-app
    pull_policy: always
    restart: unless-stopped
    ports:
      - "9876:8080" # Host Port 9876 -> Container Port 8080
    environment:
      - SERVER_MODE=true
      - PORT=8080
      - STORAGE_MODE=memory
      - MAX_PAYLOAD_MB=5
      - SESSIONS_DIR=/app/data/sessions
      - WEB_DIR=/app/web
    volumes:
      - /volume1/docker/gymrat/data:/app/data
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:8080/api/exercises"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 5s
```

Local Synology Access: 👉 **`http://<YOUR-SYNOLOGY-NAS-IP>:9876`**

---

### 6. Publish Multi-Arch Image to Docker Hub
Build and publish multi-platform Docker images (`linux/amd64`, `linux/arm64`) to Docker Hub:
```bash
./scripts/publish_docker.sh
```

---

## 🌐 REST API Endpoint Reference

All REST endpoints accept an optional header `X-Session-ID` (or query param `sessionId`). If omitted, a new session workspace is automatically generated and returned in the `X-Session-ID` response header.

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `POST` | `/api/session/new` | Generates a new session and deep-copies starter exercise/workout catalog into memory. |
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