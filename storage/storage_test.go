package storage_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"gymrat/models"
	"gymrat/storage"
)

func TestStorageAndSessionManager(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gymrat_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sm, err := storage.NewSessionManager(filepath.Join(tempDir, "sessions"))
	if err != nil {
		t.Fatalf("Failed to create session manager: %v", err)
	}

	sessionId, err := sm.CreateSession()
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	if !sm.SessionExists(sessionId) {
		t.Fatalf("Session %s should exist", sessionId)
	}

	// Prepare test catalog and vault
	cat := models.ExerciseCatalogData{
		Exercises: []models.Exercise{
			{
				Id:          "ex_bench",
				Version:     1,
				Name:        "Bench Press",
				Description: "Flat barbell bench press",
				Category:    "Chest",
				IsRemoved:   false,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		},
	}

	vault := models.GymRatVaultData{
		WorkoutPlans: []models.Plan{
			{
				Id:          "plan_1",
				Name:        "Push Day Plan",
				Status:      "active",
				Description: "Chest and Triceps Focus",
				DatePlanned: time.Now(),
				Workouts: []models.Workout{
					{
						Id:          "w_push_1",
						Name:        "Push Workout A",
						Description: "Chest heavy",
						DatePlanned: time.Now(),
						IsExecuted:  true,
						Exercises: map[string]models.WorkoutExercise{
							"ex_bench": {
								ExerciseId:      "ex_bench",
								ExerciseVersion: 1,
								NameSnapshot:    "Bench Press",
								Sets: []models.ExerciseSet{
									{
										Id:              "set_b1",
										SetType:         models.SetTypeReps,
										RepCount:        8,
										Weight:          185.0,
										PerceivedEffort: 8,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if err := sm.SaveSession(sessionId, cat, vault); err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	loadedCat, loadedVault, err := sm.LoadSession(sessionId)
	if err != nil {
		t.Fatalf("Failed to load session: %v", err)
	}

	if len(loadedCat.Exercises) != 1 || loadedCat.Exercises[0].Name != "Bench Press" {
		t.Errorf("Mismatch in loaded catalog exercises")
	}

	if len(loadedVault.WorkoutPlans) != 1 || !loadedVault.WorkoutPlans[0].Workouts[0].IsExecuted {
		t.Errorf("Mismatch in loaded vault plan execution status")
	}

	// Test Export & Import
	exportedBytes, err := sm.ExportSession(sessionId)
	if err != nil {
		t.Fatalf("Failed to export session: %v", err)
	}

	newSessionId, err := sm.CreateSession()
	if err != nil {
		t.Fatalf("Failed to create second session: %v", err)
	}

	if err := sm.ImportSession(newSessionId, exportedBytes); err != nil {
		t.Fatalf("Failed to import session bundle: %v", err)
	}

	impCat, impVault, err := sm.LoadSession(newSessionId)
	if err != nil {
		t.Fatalf("Failed to load imported session: %v", err)
	}

	if len(impCat.Exercises) != 1 || len(impVault.WorkoutPlans) != 1 {
		t.Errorf("Mismatch in imported data count")
	}
}
