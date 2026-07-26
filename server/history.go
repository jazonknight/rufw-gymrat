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

// handleHistory handles GET and POST for /api/history.
// Manages completed workout logs, recording completion duration, total volume lifted (lbs), and exercise snapshots.
func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
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
		// Return completed workout history logs
		if vault.Workouts == nil {
			vault.Workouts = []models.HistoricWorkouts{}
		}
		json.NewEncoder(w).Encode(vault.Workouts)

	case http.MethodPost:
		// Record completed workout log entry
		var logEntry models.HistoricWorkouts
		if err := json.NewDecoder(r.Body).Decode(&logEntry); err != nil {
			http.Error(w, "Invalid JSON input", http.StatusBadRequest)
			return
		}

		if logEntry.Id == "" {
			logEntry.Id = uuid.NewString()
		}
		if logEntry.DateCompleted.IsZero() {
			logEntry.DateCompleted = time.Now()
		}

		vault.Workouts = append(vault.Workouts, logEntry)
		if err := s.SessionMgr.SaveSession(sessionId, catalog, vault); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(logEntry)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
