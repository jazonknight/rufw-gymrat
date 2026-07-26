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

// handlePlans handles GET, POST, and PUT for /api/plans.
// Manages high-level Training Plans / Programs containing scheduled workout routine templates.
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
		// Return list of training plans
		json.NewEncoder(w).Encode(vault.WorkoutPlans)

	case http.MethodPost:
		// Create new training plan
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

		// Ensure scheduled workouts have unique IDs and execution timestamps
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
		// Update execution status or sets for a scheduled workout inside a training plan
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

// handleCopyPlan handles POST /api/plans/copy.
// It deep-copies an existing Training Plan by ID, generates a fresh unique UUID
// for the new plan and all internal workout instances, resets execution status, appends "(Copy)" to the name, and saves the session.
func (s *Server) handleCopyPlan(w http.ResponseWriter, r *http.Request) {
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
		PlanId string `json:"planId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PlanId == "" {
		http.Error(w, "Missing or invalid 'planId'", http.StatusBadRequest)
		return
	}

	var sourcePlan *models.Plan
	for i := range vault.WorkoutPlans {
		if vault.WorkoutPlans[i].Id == req.PlanId {
			sourcePlan = &vault.WorkoutPlans[i]
			break
		}
	}

	if sourcePlan == nil {
		http.Error(w, "Plan ID not found", http.StatusNotFound)
		return
	}

	// Deep copy plan with fresh UUIDs to prevent ID collision across duplicated plans
	bytes, _ := json.Marshal(sourcePlan)
	var newPlan models.Plan
	_ = json.Unmarshal(bytes, &newPlan)

	newPlan.Id = uuid.NewString()
	newPlan.Name = newPlan.Name + " (Copy)"
	newPlan.DatePlanned = time.Now()

	for wIdx := range newPlan.Workouts {
		newPlan.Workouts[wIdx].Id = uuid.NewString()
		newPlan.Workouts[wIdx].IsExecuted = false
		newPlan.Workouts[wIdx].DatePlanned = time.Now()
		for exId, we := range newPlan.Workouts[wIdx].Exercises {
			for sIdx := range we.Sets {
				we.Sets[sIdx].Id = uuid.NewString()
			}
			newPlan.Workouts[wIdx].Exercises[exId] = we
		}
	}

	vault.WorkoutPlans = append(vault.WorkoutPlans, newPlan)
	if err := s.SessionMgr.SaveSession(sessionId, catalog, vault); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Session-ID", sessionId)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newPlan)
}
