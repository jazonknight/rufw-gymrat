// Package server provides HTTP REST API endpoints and static web asset serving
// for managing exercise catalogs, workout plans, and session data synchronization.
package server

import (
	"encoding/json"
	"net/http"
	"time"

	"gymrat/models"

	"github.com/google/uuid"
)

// handleWorkouts handles GET, POST, PUT, and DELETE for /api/workouts.
// Manages standalone Workout Routine templates constructed from exercises in the catalog.
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
		// Return list of workout routine templates
		if vault.WorkoutTemplates == nil {
			vault.WorkoutTemplates = []models.Workout{}
		}
		json.NewEncoder(w).Encode(vault.WorkoutTemplates)

	case http.MethodPost:
		// Create new workout routine template
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
		// Update existing workout routine template
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
		// Delete workout routine template by ID
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

// handleCopyWorkout handles POST /api/workouts/copy.
// It deep-copies an existing Workout Routine template by ID, generates a fresh unique UUID
// for the new workout and all nested set instances, appends "(Copy)" to the name, and saves the session.
func (s *Server) handleCopyWorkout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	var req struct {
		WorkoutId string `json:"workoutId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.WorkoutId == "" {
		http.Error(w, "Missing or invalid 'workoutId'", http.StatusBadRequest)
		return
	}

	var sourceWorkout *models.Workout
	for i := range vault.WorkoutTemplates {
		if vault.WorkoutTemplates[i].Id == req.WorkoutId {
			sourceWorkout = &vault.WorkoutTemplates[i]
			break
		}
	}

	if sourceWorkout == nil {
		http.Error(w, "Workout template ID not found", http.StatusNotFound)
		return
	}

	// Deep copy workout template with fresh UUIDs to prevent ID collision across duplicated routines
	bytes, _ := json.Marshal(sourceWorkout)
	var newWorkout models.Workout
	_ = json.Unmarshal(bytes, &newWorkout)

	newWorkout.Id = uuid.NewString()
	newWorkout.Name = newWorkout.Name + " (Copy)"
	newWorkout.DatePlanned = time.Now()

	// Assign fresh UUIDs to all individual set instances inside the copied routine
	for exId, we := range newWorkout.Exercises {
		for sIdx := range we.Sets {
			we.Sets[sIdx].Id = uuid.NewString()
		}
		newWorkout.Exercises[exId] = we
	}

	vault.WorkoutTemplates = append(vault.WorkoutTemplates, newWorkout)
	if err := s.SessionMgr.SaveSession(sessionId, catalog, vault); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Session-ID", sessionId)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newWorkout)
}
