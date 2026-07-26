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

// handleExercises handles GET, POST, PUT, and DELETE for /api/exercises.
// Supports exercise catalog listing, versioned exercise creation, updating (which bumps versions), and soft deletion.
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
		// Return exercise library catalog
		json.NewEncoder(w).Encode(catalog.Exercises)

	case http.MethodPost:
		// Create new exercise definition (v1)
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
		// Update existing exercise definition (creates new immutable version v+1)
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
		// Soft-delete exercise definition by setting IsRemoved = true
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
