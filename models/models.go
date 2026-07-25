// Package models defines the primary data models for the GymRat application,
// including decoupled exercise catalogs, versioning, set types, workouts, and plans.
package models

import "time"

// SetType defines whether an exercise set is counting (rep-based) or timed (seconds-based).
type SetType string

const (
	// SetTypeReps represents a set measured by rep count.
	SetTypeReps SetType = "reps"
	// SetTypeTimed represents a set measured by duration in seconds.
	SetTypeTimed SetType = "timed"
)

// AppSession tracks metadata for an active application session.
type AppSession struct {
	Id           string          `json:"id"`
	SessionStart time.Time       `json:"sessionStart"`
	SessionEnd   time.Time       `json:"sessionEnd"`
	Filename     string          `json:"filename"`
	Filelocation string          `json:"filelocation"`
	VaultData    GymRatVaultData `json:"vaultData"`
}

// Exercise defines an independent exercise item stored in the exercise catalog (exercises.json).
// It supports incremental Version numbering for audit history and IsRemoved for soft deletion.
type Exercise struct {
	Id          string    `json:"id"`
	Version     int       `json:"version"`     // Incremental version number (starts at 1)
	Name        string    `json:"name"`        // Display name of the exercise
	Description string    `json:"description"` // Form cues, setup instructions, or notes
	Category    string    `json:"category,omitempty"`
	IsRemoved   bool      `json:"isRemoved"` // Soft delete indicator flag
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ExerciseCatalogData represents the outer wrapper for exercises.json serialization.
type ExerciseCatalogData struct {
	Exercises []Exercise `json:"exercises"`
}

// ExerciseSet represents an individual set performed during an exercise.
// Supports both rep counting (RepCount) and time-based tracking (DurationSeconds).
type ExerciseSet struct {
	Id              string   `json:"id"`
	SetType         SetType  `json:"setType"`                   // "reps" or "timed"
	RepCount        int      `json:"repCount,omitempty"`        // Applicable for counting sets
	DurationSeconds int      `json:"durationSeconds,omitempty"` // Applicable for timed sets
	FrequencyUnit   UnitType `json:"frequencyUnit,omitempty"`   // Legacy unit tag for backward compatibility
	Weight          float64  `json:"weight"`                    // Resistance weight (lbs/kg)
	PerceivedEffort int      `json:"perceivedEffort"`           // Rate of Perceived Exertion (RPE 1-10)
}

// WorkoutExercise links a specific exercise ID and Version from the catalog to a workout,
// alongside the sets planned or completed for that exercise.
type WorkoutExercise struct {
	ExerciseId      string        `json:"exerciseId"`      // Reference ID from exercise catalog
	ExerciseVersion int           `json:"exerciseVersion"` // Specific version of exercise referenced
	NameSnapshot    string        `json:"nameSnapshot,omitempty"`
	Sets            []ExerciseSet `json:"sets"`
	Notes           string        `json:"notes,omitempty"`
}

// Workout represents an individual training session containing planned date, execution status,
// and a map of exercises included in the session.
type Workout struct {
	Id          string                     `json:"id"`
	Name        string                     `json:"name"`
	Description string                     `json:"description"`
	DatePlanned time.Time                  `json:"datePlanned"` // Scheduled workout date
	IsExecuted  bool                       `json:"isExecuted"`  // Indicates whether the workout has been completed
	Exercises   map[string]WorkoutExercise `json:"exercises"`   // Map of exercise ID -> WorkoutExercise
}

// Plan represents an overall workout plan or training block containing a list of Workouts.
type Plan struct {
	Id          string    `json:"id"`
	Name        string    `json:"name"`
	Status      string    `json:"status"` // Operational status e.g., "draft", "active", "completed"
	Description string    `json:"description"`
	DatePlanned time.Time `json:"datePlanned"`
	Workouts    []Workout `json:"workouts"`
}

// GymRatVaultData represents the outer wrapper for gymrat_plans.json serialization.
type GymRatVaultData struct {
	WorkoutPlans     []Plan             `json:"workoutPlans"`
	WorkoutTemplates []Workout          `json:"workoutTemplates,omitempty"` // Standalone workout routines
	Workouts         []HistoricWorkouts `json:"workouts,omitempty"`
}

// HistoricWorkouts tracks historical completed workout logs.
type HistoricWorkouts struct {
	Id              string    `json:"id"`
	WorkoutId       string    `json:"workoutId"`
	PlanId          string    `json:"planId,omitempty"`
	WorkoutName     string    `json:"workoutName"`
	DateCompleted   time.Time `json:"dateCompleted"`
	DurationSeconds int       `json:"durationSeconds"` // Total workout duration in seconds
	TotalVolumeLbs  float64   `json:"totalVolumeLbs"`  // Total volume lifted (weight * reps)
	WorkoutSnapshot Workout   `json:"workoutSnapshot"` // Completed workout state with actual logged sets
}


