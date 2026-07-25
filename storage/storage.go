// Package storage handles persistence operations for exercise catalog JSON files,
// workout plan vault JSON files, and multi-tenant session workspace management.
package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gymrat/models"

	"github.com/google/uuid"
)

const (
	// DefaultExercisesFilename is the standard JSON file name for storing exercise catalogs.
	DefaultExercisesFilename = "exercises.json"
	// DefaultPlansFilename is the standard JSON file name for storing workout plans and workouts.
	DefaultPlansFilename = "gymrat_plans.json"
)

// SaveExercises serializes and writes the exercise catalog to exercises.json in the specified directory.
func SaveExercises(dir string, filename string, catalog models.ExerciseCatalogData) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		return err
	}

	fullPath := filepath.Join(dir, filename)
	return os.WriteFile(fullPath, data, 0644)
}

// LoadExercises reads and unmarshals the exercise catalog from exercises.json.
// Returns an empty catalog struct if the target file does not yet exist.
func LoadExercises(dir string, filename string) (models.ExerciseCatalogData, error) {
	var catalog models.ExerciseCatalogData
	fullPath := filepath.Join(dir, filename)

	bytes, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty catalog if file does not exist yet
			return models.ExerciseCatalogData{Exercises: []models.Exercise{}}, nil
		}
		return catalog, err
	}

	err = json.Unmarshal(bytes, &catalog)
	if err != nil {
		return catalog, fmt.Errorf("failed to parse exercises JSON: %w", err)
	}

	return catalog, nil
}

// SaveVault serializes and writes the workout plans vault to gymrat_plans.json in the specified directory.
func SaveVault(dir string, filename string, vault models.GymRatVaultData) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(vault, "", "  ")
	if err != nil {
		return err
	}

	fullPath := filepath.Join(dir, filename)
	return os.WriteFile(fullPath, data, 0644)
}

// LoadVault reads and unmarshals the plans and workouts vault from gymrat_plans.json.
// Returns an empty vault struct if the file does not yet exist.
func LoadVault(dir string, filename string) (models.GymRatVaultData, error) {
	var vault models.GymRatVaultData
	fullPath := filepath.Join(dir, filename)

	bytes, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty vault if file does not exist yet
			return models.GymRatVaultData{WorkoutPlans: []models.Plan{}}, nil
		}
		return vault, err
	}

	err = json.Unmarshal(bytes, &vault)
	if err != nil {
		return vault, fmt.Errorf("failed to parse plans JSON: %w", err)
	}

	return vault, nil
}

// CombinedSessionBundle represents both exercises catalog and plans vault packaged together for import/export payload.
type CombinedSessionBundle struct {
	SessionId string                     `json:"sessionId"`
	Exercises models.ExerciseCatalogData `json:"exercises"`
	Vault     models.GymRatVaultData     `json:"vault"`
}

// SessionManager manages per-session isolated workspaces under BaseDir using thread-safe synchronization.
type SessionManager struct {
	BaseDir string
	mu      sync.RWMutex
}

// NewSessionManager initializes a new SessionManager with the target root directory.
func NewSessionManager(baseDir string) (*SessionManager, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, err
	}
	return &SessionManager{BaseDir: baseDir}, nil
}

// CreateSession generates a unique Session ID and initializes default session JSON files.
func (sm *SessionManager) CreateSession() (string, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sessionId := uuid.NewString()
	dir := sm.GetSessionDir(sessionId)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	// Initialize default empty files
	emptyCatalog := models.ExerciseCatalogData{Exercises: []models.Exercise{}}
	emptyVault := models.GymRatVaultData{WorkoutPlans: []models.Plan{}}

	if err := SaveExercises(dir, DefaultExercisesFilename, emptyCatalog); err != nil {
		return "", err
	}
	if err := SaveVault(dir, DefaultPlansFilename, emptyVault); err != nil {
		return "", err
	}

	return sessionId, nil
}

// GetSessionDir returns the absolute file system directory path for a session ID.
func (sm *SessionManager) GetSessionDir(sessionId string) string {
	return filepath.Join(sm.BaseDir, sessionId)
}

// SessionExists checks whether a directory exists for the given Session ID.
func (sm *SessionManager) SessionExists(sessionId string) bool {
	dir := sm.GetSessionDir(sessionId)
	info, err := os.Stat(dir)
	return err == nil && info.IsDir()
}

// LoadSession reads and returns the exercise catalog and plans vault for the specified session ID.
func (sm *SessionManager) LoadSession(sessionId string) (models.ExerciseCatalogData, models.GymRatVaultData, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	dir := sm.GetSessionDir(sessionId)
	if !sm.SessionExists(sessionId) {
		return models.ExerciseCatalogData{}, models.GymRatVaultData{}, errors.New("session not found")
	}

	cat, err := LoadExercises(dir, DefaultExercisesFilename)
	if err != nil {
		return models.ExerciseCatalogData{}, models.GymRatVaultData{}, err
	}

	vault, err := LoadVault(dir, DefaultPlansFilename)
	if err != nil {
		return models.ExerciseCatalogData{}, models.GymRatVaultData{}, err
	}

	return cat, vault, nil
}

// SaveSession saves updated catalog and vault data into the session's workspace.
func (sm *SessionManager) SaveSession(sessionId string, cat models.ExerciseCatalogData, vault models.GymRatVaultData) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	dir := sm.GetSessionDir(sessionId)
	if err := SaveExercises(dir, DefaultExercisesFilename, cat); err != nil {
		return err
	}
	return SaveVault(dir, DefaultPlansFilename, vault)
}

// ExportSession serializes a session's exercises and plans into a combined JSON byte slice payload.
func (sm *SessionManager) ExportSession(sessionId string) ([]byte, error) {
	cat, vault, err := sm.LoadSession(sessionId)
	if err != nil {
		return nil, err
	}

	bundle := CombinedSessionBundle{
		SessionId: sessionId,
		Exercises: cat,
		Vault:     vault,
	}

	return json.MarshalIndent(bundle, "", "  ")
}

// ImportSession parses a JSON payload and overwrites the session workspace with the imported data.
func (sm *SessionManager) ImportSession(sessionId string, data []byte) error {
	var bundle CombinedSessionBundle
	if err := json.Unmarshal(data, &bundle); err != nil {
		return fmt.Errorf("invalid bundle JSON: %w", err)
	}

	return sm.SaveSession(sessionId, bundle.Exercises, bundle.Vault)
}

