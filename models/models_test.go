package models_test

import (
	"testing"
	"time"

	"gymrat/models"
)

func TestExerciseVersioningAndSoftDelete(t *testing.T) {
	exV1 := models.Exercise{
		Id:          "ex_squat",
		Version:     1,
		Name:        "Back Squat",
		Description: "High bar back squat",
		Category:    "Legs",
		IsRemoved:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if exV1.Version != 1 {
		t.Errorf("Expected version 1, got %d", exV1.Version)
	}

	// Bump version
	exV2 := exV1
	exV2.Version = 2
	exV2.Name = "Barbell Back Squat"

	if exV2.Version != 2 {
		t.Errorf("Expected version 2, got %d", exV2.Version)
	}

	// Test soft delete
	exV2.IsRemoved = true
	if !exV2.IsRemoved {
		t.Errorf("Expected IsRemoved true")
	}
}

func TestSetTypes(t *testing.T) {
	repSet := models.ExerciseSet{
		Id:              "set_1",
		SetType:         models.SetTypeReps,
		RepCount:        12,
		Weight:          135.0,
		PerceivedEffort: 8,
	}

	timedSet := models.ExerciseSet{
		Id:              "set_2",
		SetType:         models.SetTypeTimed,
		DurationSeconds: 45,
		Weight:          0.0,
		PerceivedEffort: 9,
	}

	if !models.IsValidSetType(string(repSet.SetType)) {
		t.Errorf("Invalid rep set type")
	}
	if !models.IsValidSetType(string(timedSet.SetType)) {
		t.Errorf("Invalid timed set type")
	}
	if repSet.RepCount != 12 {
		t.Errorf("Expected 12 reps, got %d", repSet.RepCount)
	}
	if timedSet.DurationSeconds != 45 {
		t.Errorf("Expected 45 duration seconds, got %d", timedSet.DurationSeconds)
	}
}
