package models

import "time"

type AppSession struct {
	Id           string
	SessionStart time.Time
	SessionEnd   time.Time
	Filename     string
	Filelocation string
	VaultData    GymRatVaultData
}

type GymRatVaultData struct {
	WorkoutPlans []Plan
	Workouts     []HistoricWorkouts
}

type Plan struct {
	Id          string
	Name        string
	Status      string // completed, created
	Description string
	DatePlanned time.Time
	Workouts    []Workout
}

type Workout struct {
	Id               string
	Name             string // e.g., "Workout A - Lower Body" or "Session 1"
	Description      string
	PlannedExercises []Exercise
}

type Exercise struct {
	Id   string
	Name string
	Sets []ExerciseSet
}

type ExerciseSet struct {
	Id              string
	RepCount        int      // 10 count or 60 seconds
	FrequencyUnit   UnitType // seconds / count
	Weight          float32  // 34 or 35.6
	PerceivedEffort int      // 1 to 10 with 1 being easy and 10 being extremely difficult
}

// History
type HistoricWorkouts struct {
	Id            string
	DateWorkedOut time.Time
	WorkoutPlan   Plan
}
