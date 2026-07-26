// Package server provides HTTP REST API endpoints and static web asset serving
// for managing exercise catalogs, workout plans, and session data synchronization.
package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// handleNewSession handles POST /api/session/new.
// Generates a fresh unique workspace session ID and seeds default exercise/workout starter catalogs into memory.
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

// handleUpload handles POST /api/vault/upload.
// Imports an uploaded gymrat_vault.json bundle into the current workspace session, with strict payload size clamping.
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

// handleExport handles GET /api/vault/export.
// Downloads a complete JSON snapshot of the active workspace session (exercise catalog + workout templates + plans + history).
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

// handleOpenAPI serves the OpenAPI 3.0 YAML specification file for API documentation & SDK generation.
func (s *Server) handleOpenAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
	http.ServeFile(w, r, "docs/openapi.yaml")
}
