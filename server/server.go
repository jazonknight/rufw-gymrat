// Package server provides HTTP REST API endpoints and static web asset serving
// for managing exercise catalogs, workout plans, and session data synchronization.
package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"gymrat/models"
	"gymrat/storage"

	"github.com/google/uuid"
)

// Server configures and runs the HTTP server instance.
type Server struct {
	SessionMgr   *storage.SessionManager // Workspace session manager
	WebDir       string                  // Path to static web UI files
	Port         string                  // TCP port to listen on
	MaxPayloadMB int                     // Max upload request payload size in MB
}

// NewServer initializes a new REST API Server.
func NewServer(sm *storage.SessionManager, webDir string, port string, maxPayloadMB int) *Server {
	if maxPayloadMB <= 0 {
		maxPayloadMB = 5 // Default 5 MB
	}
	return &Server{
		SessionMgr:   sm,
		WebDir:       webDir,
		Port:         port,
		MaxPayloadMB: maxPayloadMB,
	}
}

// Start registers HTTP routes and starts listening for requests.
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// CORS & Helper middleware wrapper
	wrap := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Session-ID")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			h(w, r)
		}
	}

	// API Endpoints
	mux.HandleFunc("/api/session/new", wrap(s.handleNewSession))
	mux.HandleFunc("/api/exercises", wrap(s.handleExercises))
	mux.HandleFunc("/api/workouts", wrap(s.handleWorkouts))
	mux.HandleFunc("/api/plans", wrap(s.handlePlans))
	mux.HandleFunc("/api/vault/upload", wrap(s.handleUpload))
	mux.HandleFunc("/api/vault/export", wrap(s.handleExport))

	// Static Web Client File Server
	if s.WebDir != "" {
		fileServer := http.FileServer(http.Dir(s.WebDir))
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, "/api/") {
				fileServer.ServeHTTP(w, r)
				return
			}
			http.NotFound(w, r)
		})
	}

	fmt.Printf("[GymRat REST Server] Running on http://localhost:%s\n", s.Port)
	return http.ListenAndServe(":"+s.Port, mux)
}

func (s *Server) getSessionId(r *http.Request) (string, error) {
	sessionId := r.Header.Get("X-Session-ID")
	if sessionId == "" {
		sessionId = r.URL.Query().Get("sessionId")
	}

	if sessionId != "" {
		// Validate UUID format to prevent path traversal
		if _, err := uuid.Parse(sessionId); err != nil {
			return "", fmt.Errorf("invalid session ID format: %w", err)
		}
	}

	if sessionId == "" {
		// Default session or auto-create
		var err error
		sessionId, err = s.SessionMgr.CreateSession()
		if err != nil {
			return "", err
		}
	} else if !s.SessionMgr.SessionExists(sessionId) {
		// Create session directory if requested ID doesn't exist yet
		dir := s.SessionMgr.GetSessionDir(sessionId)
		if s.SessionMgr.StorageMode == "disk" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return "", err
			}
		}
	}

	return sessionId, nil
}

func (s *Server) handleNewSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionId, err := s.SessionMgr.CreateSession()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"sessionId": sessionId,
		"message":   "Session initialized successfully",
	})
}

func (s *Server) handleExercises(w http.ResponseWriter, r *http.Request) {
	sessionId, err := s.getSessionId(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	catalog, vault, err := s.SessionMgr.LoadSession(sessionId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Session-ID", sessionId)

	switch r.Method {
	case http.MethodGet:
		json.NewEncoder(w).Encode(catalog.Exercises)

	case http.MethodPost:
		var input struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Category    string `json:"category"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Invalid JSON input", http.StatusBadRequest)
			return
		}

		newEx := models.Exercise{
			Id:          uuid.NewString(),
			Version:     1,
			Name:        input.Name,
			Description: input.Description,
			Category:    input.Category,
			IsRemoved:   false,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		catalog.Exercises = append(catalog.Exercises, newEx)
		if err := s.SessionMgr.SaveSession(sessionId, catalog, vault); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newEx)

	case http.MethodPut:
		var input struct {
			Id          string `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
			Category    string `json:"category"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Invalid JSON input", http.StatusBadRequest)
			return
		}

		var latestEx *models.Exercise
		for i := range catalog.Exercises {
			if catalog.Exercises[i].Id == input.Id {
				if latestEx == nil || catalog.Exercises[i].Version > latestEx.Version {
					latestEx = &catalog.Exercises[i]
				}
			}
		}

		if latestEx == nil {
			http.Error(w, "Exercise ID not found", http.StatusNotFound)
			return
		}

		newName := input.Name
		if newName == "" {
			newName = latestEx.Name
		}
		newDesc := input.Description
		if newDesc == "" {
			newDesc = latestEx.Description
		}
		newCat := input.Category
		if newCat == "" {
			newCat = latestEx.Category
		}

		updatedEx := models.Exercise{
			Id:          latestEx.Id,
			Version:     latestEx.Version + 1,
			Name:        newName,
			Description: newDesc,
			Category:    newCat,
			IsRemoved:   latestEx.IsRemoved,
			CreatedAt:   latestEx.CreatedAt,
			UpdatedAt:   time.Now(),
		}

		catalog.Exercises = append(catalog.Exercises, updatedEx)
		if err := s.SessionMgr.SaveSession(sessionId, catalog, vault); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(updatedEx)

	case http.MethodDelete:
		exId := r.URL.Query().Get("id")
		if exId == "" {
			http.Error(w, "Missing exercise 'id' parameter", http.StatusBadRequest)
			return
		}

		found := false
		for i := range catalog.Exercises {
			if catalog.Exercises[i].Id == exId {
				catalog.Exercises[i].IsRemoved = true
				catalog.Exercises[i].UpdatedAt = time.Now()
				found = true
			}
		}

		if !found {
			http.Error(w, "Exercise ID not found", http.StatusNotFound)
			return
		}

		if err := s.SessionMgr.SaveSession(sessionId, catalog, vault); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"message": "Exercise soft-deleted successfully"})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handlePlans(w http.ResponseWriter, r *http.Request) {
	sessionId, err := s.getSessionId(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	catalog, vault, err := s.SessionMgr.LoadSession(sessionId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Session-ID", sessionId)

	switch r.Method {
	case http.MethodGet:
		json.NewEncoder(w).Encode(vault.WorkoutPlans)

	case http.MethodPost:
		var newPlan models.Plan
		if err := json.NewDecoder(r.Body).Decode(&newPlan); err != nil {
			http.Error(w, "Invalid JSON input", http.StatusBadRequest)
			return
		}

		if newPlan.Id == "" {
			newPlan.Id = uuid.NewString()
		}
		if newPlan.DatePlanned.IsZero() {
			newPlan.DatePlanned = time.Now()
		}

		// Ensure workouts have IDs and datePlanned
		for i := range newPlan.Workouts {
			if newPlan.Workouts[i].Id == "" {
				newPlan.Workouts[i].Id = uuid.NewString()
			}
			if newPlan.Workouts[i].DatePlanned.IsZero() {
				newPlan.Workouts[i].DatePlanned = time.Now()
			}
		}

		vault.WorkoutPlans = append(vault.WorkoutPlans, newPlan)
		if err := s.SessionMgr.SaveSession(sessionId, catalog, vault); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newPlan)

	case http.MethodPut:
		// Mark workout executed or update plan
		var req struct {
			PlanId     string `json:"planId"`
			WorkoutId  string `json:"workoutId"`
			IsExecuted bool   `json:"isExecuted"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON input", http.StatusBadRequest)
			return
		}

		updated := false
		for pIdx := range vault.WorkoutPlans {
			if vault.WorkoutPlans[pIdx].Id == req.PlanId {
				for wIdx := range vault.WorkoutPlans[pIdx].Workouts {
					if vault.WorkoutPlans[pIdx].Workouts[wIdx].Id == req.WorkoutId {
						vault.WorkoutPlans[pIdx].Workouts[wIdx].IsExecuted = req.IsExecuted
						updated = true
					}
				}
			}
		}

		if !updated {
			http.Error(w, "Plan or Workout ID not found", http.StatusNotFound)
			return
		}

		if err := s.SessionMgr.SaveSession(sessionId, catalog, vault); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"message": "Workout status updated successfully"})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionId, err := s.getSessionId(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Restrict upload body size to configured MaxPayloadMB to prevent RAM exhaustion DoS
	maxBytes := int64(s.MaxPayloadMB) << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read request body or payload exceeds %dMB limit", s.MaxPayloadMB), http.StatusBadRequest)
		return
	}

	if err := s.SessionMgr.ImportSession(sessionId, body); err != nil {
		http.Error(w, fmt.Sprintf("Import error: %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":   "Session data uploaded and imported successfully",
		"sessionId": sessionId,
	})
}

func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionId, err := s.getSessionId(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data, err := s.SessionMgr.ExportSession(sessionId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=gymrat_vault_%s.json", sessionId))
	w.Write(data)
}

func (s *Server) handleWorkouts(w http.ResponseWriter, r *http.Request) {
	sessionId, err := s.getSessionId(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	catalog, vault, err := s.SessionMgr.LoadSession(sessionId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Session-ID", sessionId)

	switch r.Method {
	case http.MethodGet:
		if vault.WorkoutTemplates == nil {
			vault.WorkoutTemplates = []models.Workout{}
		}
		json.NewEncoder(w).Encode(vault.WorkoutTemplates)

	case http.MethodPost:
		var newWorkout models.Workout
		if err := json.NewDecoder(r.Body).Decode(&newWorkout); err != nil {
			http.Error(w, "Invalid JSON input", http.StatusBadRequest)
			return
		}

		if newWorkout.Id == "" {
			newWorkout.Id = uuid.NewString()
		}
		if newWorkout.DatePlanned.IsZero() {
			newWorkout.DatePlanned = time.Now()
		}

		vault.WorkoutTemplates = append(vault.WorkoutTemplates, newWorkout)
		if err := s.SessionMgr.SaveSession(sessionId, catalog, vault); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newWorkout)

	case http.MethodPut:
		var updatedWorkout models.Workout
		if err := json.NewDecoder(r.Body).Decode(&updatedWorkout); err != nil {
			http.Error(w, "Invalid JSON input", http.StatusBadRequest)
			return
		}

		found := false
		for i := range vault.WorkoutTemplates {
			if vault.WorkoutTemplates[i].Id == updatedWorkout.Id {
				vault.WorkoutTemplates[i] = updatedWorkout
				found = true
				break
			}
		}

		if !found {
			http.Error(w, "Workout template ID not found", http.StatusNotFound)
			return
		}

		if err := s.SessionMgr.SaveSession(sessionId, catalog, vault); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(updatedWorkout)

	case http.MethodDelete:
		workoutId := r.URL.Query().Get("id")
		if workoutId == "" {
			http.Error(w, "Missing workout 'id' parameter", http.StatusBadRequest)
			return
		}

		filtered := make([]models.Workout, 0)
		found := false
		for _, w := range vault.WorkoutTemplates {
			if w.Id == workoutId {
				found = true
			} else {
				filtered = append(filtered, w)
			}
		}

		if !found {
			http.Error(w, "Workout ID not found", http.StatusNotFound)
			return
		}

		vault.WorkoutTemplates = filtered
		if err := s.SessionMgr.SaveSession(sessionId, catalog, vault); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"message": "Workout template deleted successfully"})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ServeStaticWeb handles static web directory creation
func EnsureWebDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}
