// Package storage handles persistence operations for exercise catalog JSON files,
// workout plan vault JSON files, and multi-tenant session workspace management.
package storage

import (
	_ "embed"
	"encoding/json"
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
	// DefaultSeedFilename is the immutable starter exercise catalog file checked into Git.
	DefaultSeedFilename = "default_exercises.json"
)

//go:embed default_exercises.json
var embeddedDefaultExercises []byte

//go:embed default_workouts.json
var embeddedDefaultWorkouts []byte

//go:embed default_plans.json
var embeddedDefaultPlans []byte

// LoadDefaultSeedCatalog loads the default exercise catalog from a specified JSON file path
// or falls back to the embedded default_exercises.json file.
func LoadDefaultSeedCatalog(seedFilePath string) models.ExerciseCatalogData {
	var catalog models.ExerciseCatalogData

	if seedFilePath != "" {
		bytes, err := os.ReadFile(seedFilePath)
		if err == nil {
			if err := json.Unmarshal(bytes, &catalog); err == nil && len(catalog.Exercises) > 0 {
				return catalog
			}
		}
	}

	// Load from embedded standalone default_exercises.json file
	_ = json.Unmarshal(embeddedDefaultExercises, &catalog)
	return catalog
}

// LoadDefaultSeedVault loads starter workout routines and training plans from default_workouts.json and default_plans.json.
func LoadDefaultSeedVault() models.GymRatVaultData {
	var vault models.GymRatVaultData
	_ = json.Unmarshal(embeddedDefaultWorkouts, &vault)

	var plansVault models.GymRatVaultData
	if err := json.Unmarshal(embeddedDefaultPlans, &plansVault); err == nil && len(plansVault.WorkoutPlans) > 0 {
		vault.WorkoutPlans = plansVault.WorkoutPlans
	}

	return vault
}

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
func LoadExercises(dir string, filename string) (models.ExerciseCatalogData, error) {
	var catalog models.ExerciseCatalogData
	fullPath := filepath.Join(dir, filename)

	bytes, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
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
func LoadVault(dir string, filename string) (models.GymRatVaultData, error) {
	var vault models.GymRatVaultData
	fullPath := filepath.Join(dir, filename)

	bytes, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
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

// SessionManager manages per-session isolated workspaces using thread-safe in-memory storage.
type SessionManager struct {
	BaseDir        string
	DefaultCatalog models.ExerciseCatalogData
	DefaultVault   models.GymRatVaultData
	StorageMode    string // "memory" (default, 0 disk writes) or "disk"
	sessions       map[string]*CombinedSessionBundle
	mu             sync.RWMutex
}

// NewSessionManager initializes a new SessionManager with target base directory, seed file path, and storage mode.
func NewSessionManager(baseDir string, seedFilePath string, storageMode string) (*SessionManager, error) {
	if storageMode == "" {
		storageMode = "memory"
	}

	if storageMode == "disk" {
		if err := os.MkdirAll(baseDir, 0755); err != nil {
			return nil, err
		}
	}

	seedCatalog := LoadDefaultSeedCatalog(seedFilePath)
	seedVault := LoadDefaultSeedVault()

	return &SessionManager{
		BaseDir:        baseDir,
		DefaultCatalog: seedCatalog,
		DefaultVault:   seedVault,
		StorageMode:    storageMode,
		sessions:       make(map[string]*CombinedSessionBundle),
	}, nil
}

func cloneVault(v models.GymRatVaultData) models.GymRatVaultData {
	bytes, _ := json.Marshal(v)
	var clone models.GymRatVaultData
	_ = json.Unmarshal(bytes, &clone)
	return clone
}

// CreateSession generates a unique Session ID and seeds a copy of the default exercise catalog and starter routines into RAM.
func (sm *SessionManager) CreateSession() (string, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sessionId := uuid.NewString()

	// Deep copy default exercise catalog so session edits do not alter the master default seed
	seedExercisesCopy := make([]models.Exercise, len(sm.DefaultCatalog.Exercises))
	copy(seedExercisesCopy, sm.DefaultCatalog.Exercises)

	bundle := &CombinedSessionBundle{
		SessionId: sessionId,
		Exercises: models.ExerciseCatalogData{Exercises: seedExercisesCopy},
		Vault:     cloneVault(sm.DefaultVault),
	}

	sm.sessions[sessionId] = bundle

	// If storageMode is "disk", also persist session files to disk
	if sm.StorageMode == "disk" {
		dir := sm.GetSessionDir(sessionId)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", err
		}
		if err := SaveExercises(dir, DefaultExercisesFilename, bundle.Exercises); err != nil {
			return "", err
		}
		if err := SaveVault(dir, DefaultPlansFilename, bundle.Vault); err != nil {
			return "", err
		}
	}

	return sessionId, nil
}

// GetSessionDir returns the absolute file system directory path for a session ID.
func (sm *SessionManager) GetSessionDir(sessionId string) string {
	return filepath.Join(sm.BaseDir, sessionId)
}

// SessionExists checks whether a session exists in memory (or on disk if storageMode == "disk").
func (sm *SessionManager) SessionExists(sessionId string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if _, exists := sm.sessions[sessionId]; exists {
		return true
	}

	if sm.StorageMode == "disk" {
		dir := sm.GetSessionDir(sessionId)
		info, err := os.Stat(dir)
		return err == nil && info.IsDir()
	}

	return false
}

// LoadSession reads and returns the exercise catalog and plans vault for the specified session ID from RAM.
// If the session does not exist in memory, it auto-initializes it seeded with the default exercise catalog and starter routines.
func (sm *SessionManager) LoadSession(sessionId string) (models.ExerciseCatalogData, models.GymRatVaultData, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	bundle, exists := sm.sessions[sessionId]
	if !exists {
		if sm.StorageMode == "disk" {
			dir := sm.GetSessionDir(sessionId)
			cat, err := LoadExercises(dir, DefaultExercisesFilename)
			if err == nil {
				vault, _ := LoadVault(dir, DefaultPlansFilename)
				b := &CombinedSessionBundle{SessionId: sessionId, Exercises: cat, Vault: vault}
				sm.sessions[sessionId] = b
				return cat, vault, nil
			}
		}

		// Auto-initialize session in memory seeded with default starter exercises and routines
		seedExercisesCopy := make([]models.Exercise, len(sm.DefaultCatalog.Exercises))
		copy(seedExercisesCopy, sm.DefaultCatalog.Exercises)

		bundle = &CombinedSessionBundle{
			SessionId: sessionId,
			Exercises: models.ExerciseCatalogData{Exercises: seedExercisesCopy},
			Vault:     cloneVault(sm.DefaultVault),
		}
		sm.sessions[sessionId] = bundle
	}

	return bundle.Exercises, bundle.Vault, nil
}

// SaveSession updates the catalog and vault data for a session in memory (and disk if storageMode == "disk").
func (sm *SessionManager) SaveSession(sessionId string, cat models.ExerciseCatalogData, vault models.GymRatVaultData) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	bundle, exists := sm.sessions[sessionId]
	if !exists {
		bundle = &CombinedSessionBundle{SessionId: sessionId}
		sm.sessions[sessionId] = bundle
	}

	bundle.Exercises = cat
	bundle.Vault = vault

	if sm.StorageMode == "disk" {
		dir := sm.GetSessionDir(sessionId)
		if err := SaveExercises(dir, DefaultExercisesFilename, cat); err != nil {
			return err
		}
		return SaveVault(dir, DefaultPlansFilename, vault)
	}

	return nil
}

// ExportSession serializes a session's in-memory exercises and plans into a combined JSON byte slice payload.
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

// ImportSession parses a JSON payload and overwrites the session's in-memory state with the imported data.
func (sm *SessionManager) ImportSession(sessionId string, data []byte) error {
	var bundle CombinedSessionBundle
	if err := json.Unmarshal(data, &bundle); err != nil {
		return fmt.Errorf("invalid bundle JSON: %w", err)
	}

	return sm.SaveSession(sessionId, bundle.Exercises, bundle.Vault)
}
