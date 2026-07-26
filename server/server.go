// Package server provides HTTP REST API endpoints and static web asset serving
// for managing exercise catalogs, workout plans, and session data synchronization.
package server

import (
	"fmt"
	"net/http"
	"os"
	"strings"

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
		maxPayloadMB = 5 // Default 5 MB payload limit
	}
	return &Server{
		SessionMgr:   sm,
		WebDir:       webDir,
		Port:         port,
		MaxPayloadMB: maxPayloadMB,
	}
}

// Start registers HTTP routes, attaches CORS middleware, and starts listening for requests.
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

	// Registered API Endpoints
	mux.HandleFunc("/api/session/new", wrap(s.handleNewSession))
	mux.HandleFunc("/api/exercises", wrap(s.handleExercises))
	mux.HandleFunc("/api/workouts", wrap(s.handleWorkouts))
	mux.HandleFunc("/api/workouts/copy", wrap(s.handleCopyWorkout))
	mux.HandleFunc("/api/plans", wrap(s.handlePlans))
	mux.HandleFunc("/api/plans/copy", wrap(s.handleCopyPlan))
	mux.HandleFunc("/api/history", wrap(s.handleHistory))
	mux.HandleFunc("/api/vault/upload", wrap(s.handleUpload))
	mux.HandleFunc("/api/vault/export", wrap(s.handleExport))
	mux.HandleFunc("/api/docs/openapi.yaml", wrap(s.handleOpenAPI))

	// Static Web Client File Server (with Cache-Control prevention for active development)
	if s.WebDir != "" {
		fileServer := http.FileServer(http.Dir(s.WebDir))
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, "/api/") {
				w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
				fileServer.ServeHTTP(w, r)
				return
			}
			http.NotFound(w, r)
		})
	}

	fmt.Printf("[GymRat REST Server] Running on http://localhost:%s\n", s.Port)
	return http.ListenAndServe(":"+s.Port, mux)
}

// getSessionId extracts and validates the active session ID from request headers or query params.
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

// EnsureWebDir creates the static web directory if it does not exist.
func EnsureWebDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}
