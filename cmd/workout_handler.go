package cmd

import (
	"errors"
	"fmt"
	"gymrat/models"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// HandleShowCatalog prints exercises in the catalog
func HandleShowCatalog(catalog *models.ExerciseCatalogData) {
	fmt.Println("\n--- Exercise Catalog ---")
	if len(catalog.Exercises) == 0 {
		fmt.Println("Catalog is empty.")
		return
	}

	for _, ex := range catalog.Exercises {
		status := "Active"
		if ex.IsRemoved {
			status = "Removed (Soft-deleted)"
		}
		fmt.Printf("ID: %s | Ver: %d | Name: %s | Category: %s | Status: %s | Description: %s\n",
			ex.Id, ex.Version, ex.Name, ex.Category, status, ex.Description)
	}
}

// HandleAddExercise adds a new exercise to the catalog (version 1)
func HandleAddExercise(catalog *models.ExerciseCatalogData) error {
	fmt.Println("\n-- Add New Exercise to Catalog --")
	name, err := promptInput("Enter Exercise Name: ")
	if err != nil {
		return err
	}

	description, err := promptInput("Enter Exercise Description: ")
	if err != nil {
		return err
	}

	category, err := promptInput("Enter Category (e.g. Legs, Chest, Cardio): ")
	if err != nil {
		return err
	}

	newEx := models.Exercise{
		Id:          uuid.NewString(),
		Version:     1,
		Name:        name,
		Description: description,
		Category:    category,
		IsRemoved:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	catalog.Exercises = append(catalog.Exercises, newEx)
	PrintSuccess(fmt.Sprintf("Exercise added successfully! ID: %s (Version 1)", newEx.Id))
	return nil
}

// HandleUpdateExercise creates a new version of an exercise in the catalog
func HandleUpdateExercise(catalog *models.ExerciseCatalogData) error {
	fmt.Println("\n-- Update Exercise (Creates New Version) --")
	HandleShowCatalog(catalog)

	exId, err := promptInput("\nEnter Exercise ID to update: ")
	if err != nil {
		return err
	}

	var latestEx *models.Exercise
	for i := range catalog.Exercises {
		if catalog.Exercises[i].Id == exId {
			if latestEx == nil || catalog.Exercises[i].Version > latestEx.Version {
				latestEx = &catalog.Exercises[i]
			}
		}
	}

	if latestEx == nil {
		return errors.New("exercise ID not found")
	}

	newName, err := promptInput(fmt.Sprintf("New Name [%s]: ", latestEx.Name))
	if err != nil {
		return err
	}
	if newName == "" {
		newName = latestEx.Name
	}

	newDesc, err := promptInput(fmt.Sprintf("New Description [%s]: ", latestEx.Description))
	if err != nil {
		return err
	}
	if newDesc == "" {
		newDesc = latestEx.Description
	}

	newCat, err := promptInput(fmt.Sprintf("New Category [%s]: ", latestEx.Category))
	if err != nil {
		return err
	}
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
	PrintSuccess(fmt.Sprintf("Exercise '%s' updated to Version %d!", updatedEx.Name, updatedEx.Version))
	return nil
}

// HandleSoftDeleteExercise toggles the IsRemoved soft delete status
func HandleSoftDeleteExercise(catalog *models.ExerciseCatalogData) error {
	fmt.Println("\n-- Toggle Exercise Soft Delete (IsRemoved) --")
	exId, err := promptInput("Enter Exercise ID: ")
	if err != nil {
		return err
	}

	found := false
	for i := range catalog.Exercises {
		if catalog.Exercises[i].Id == exId {
			catalog.Exercises[i].IsRemoved = !catalog.Exercises[i].IsRemoved
			catalog.Exercises[i].UpdatedAt = time.Now()
			found = true
			PrintSuccess(fmt.Sprintf("Exercise %s (Version %d) IsRemoved set to %v",
				catalog.Exercises[i].Name, catalog.Exercises[i].Version, catalog.Exercises[i].IsRemoved))
		}
	}

	if !found {
		return errors.New("no exercise found with provided ID")
	}
	return nil
}

// HandleCreateWorkoutPlan guides user through building a Plan with Workouts and Sets
func HandleCreateWorkoutPlan(catalog *models.ExerciseCatalogData, vault *models.GymRatVaultData) error {
	fmt.Println("\n-- Create Workout Plan --")

	planName, err := promptInput("Enter Plan Name: ")
	if err != nil {
		return err
	}

	planDescription, err := promptInput("Enter Plan Description: ")
	if err != nil {
		return err
	}

	countWorkouts, err := promptInt("How many workouts in this plan?: ")
	if err != nil {
		return err
	}

	workouts := make([]models.Workout, 0, countWorkouts)

	for i := 0; i < countWorkouts; i++ {
		fmt.Printf("\n--- Workout #%d Config ---\n", i+1)
		wName, err := promptInput("Enter Workout Name (e.g., Lower Body A): ")
		if err != nil {
			return err
		}

		wDesc, err := promptInput("Enter Workout Description: ")
		if err != nil {
			return err
		}

		workoutExercises, err := collectWorkoutExercises(catalog)
		if err != nil {
			return err
		}

		workout := models.Workout{
			Id:          uuid.NewString(),
			Name:        wName,
			Description: wDesc,
			DatePlanned: time.Now(),
			IsExecuted:  false,
			Exercises:   workoutExercises,
		}

		workouts = append(workouts, workout)
	}

	plan := models.Plan{
		Id:          uuid.NewString(),
		Name:        planName,
		Status:      "active",
		Description: planDescription,
		DatePlanned: time.Now(),
		Workouts:    workouts,
	}

	vault.WorkoutPlans = append(vault.WorkoutPlans, plan)
	PrintSuccess(fmt.Sprintf("Plan '%s' created with %d workouts! ID: %s", plan.Name, len(plan.Workouts), plan.Id))
	return nil
}

func collectWorkoutExercises(catalog *models.ExerciseCatalogData) (map[string]models.WorkoutExercise, error) {
	result := make(map[string]models.WorkoutExercise)

	HandleShowCatalog(catalog)

	count, err := promptInt("\nHow many exercises to include in this workout?: ")
	if err != nil {
		return nil, err
	}

	for i := 0; i < count; i++ {
		exId, err := promptInput(fmt.Sprintf("Enter Exercise #%d ID from catalog: ", i+1))
		if err != nil {
			return nil, err
		}

		// Find latest version of exercise
		var selectedEx *models.Exercise
		for idx := range catalog.Exercises {
			if catalog.Exercises[idx].Id == exId {
				if selectedEx == nil || catalog.Exercises[idx].Version > selectedEx.Version {
					selectedEx = &catalog.Exercises[idx]
				}
			}
		}

		if selectedEx == nil {
			fmt.Println("Warning: Exercise ID not found in catalog. Using raw ID.")
			selectedEx = &models.Exercise{Id: exId, Version: 1, Name: "Unknown"}
		}

		setCount, err := promptInt(fmt.Sprintf("How many sets for %s?: ", selectedEx.Name))
		if err != nil {
			return nil, err
		}

		sets, err := collectSetsInput(setCount)
		if err != nil {
			return nil, err
		}

		workoutEx := models.WorkoutExercise{
			ExerciseId:      selectedEx.Id,
			ExerciseVersion: selectedEx.Version,
			NameSnapshot:    selectedEx.Name,
			Sets:            sets,
		}

		result[selectedEx.Id] = workoutEx
	}

	return result, nil
}

func collectSetsInput(count int) ([]models.ExerciseSet, error) {
	sets := make([]models.ExerciseSet, 0, count)

	for s := 0; s < count; s++ {
		fmt.Printf(" Set #%d/%d:\n", s+1, count)
		setTypeInput, err := promptInput("   Set Type ('reps' for counting, 'timed' for time-based): ")
		if err != nil {
			return nil, err
		}

		setType := models.SetTypeReps
		if setTypeInput == "timed" {
			setType = models.SetTypeTimed
		}

		set := models.ExerciseSet{
			Id:      uuid.NewString(),
			SetType: setType,
		}

		if setType == models.SetTypeReps {
			reps, err := promptInt("   Enter Rep Count: ")
			if err != nil {
				return nil, err
			}
			set.RepCount = reps
		} else {
			duration, err := promptInt("   Enter Duration in Seconds: ")
			if err != nil {
				return nil, err
			}
			set.DurationSeconds = duration
		}

		weightStr, err := promptInput("   Enter Weight (lbs/kg): ")
		if err != nil {
			return nil, err
		}
		weight, _ := strconv.ParseFloat(weightStr, 64)
		set.Weight = weight

		effort, err := promptInt("   Enter Perceived Effort (1-10): ")
		if err != nil {
			return nil, err
		}
		set.PerceivedEffort = effort

		sets = append(sets, set)
	}

	return sets, nil
}

// HandleMarkWorkoutExecuted toggles IsExecuted for a workout inside a plan
func HandleMarkWorkoutExecuted(vault *models.GymRatVaultData) error {
	fmt.Println("\n-- Mark Workout as Executed --")
	if len(vault.WorkoutPlans) == 0 {
		return errors.New("no workout plans available")
	}

	for pIdx, p := range vault.WorkoutPlans {
		fmt.Printf("Plan #%d: %s (ID: %s)\n", pIdx+1, p.Name, p.Id)
		for wIdx, w := range p.Workouts {
			fmt.Printf("  Workout #%d: %s | Executed: %v | DatePlanned: %s\n",
				wIdx+1, w.Name, w.IsExecuted, w.DatePlanned.Format(time.RFC3339))
		}
	}

	planId, err := promptInput("\nEnter Plan ID: ")
	if err != nil {
		return err
	}

	wName, err := promptInput("Enter Workout Name to mark executed: ")
	if err != nil {
		return err
	}

	for pIdx := range vault.WorkoutPlans {
		if vault.WorkoutPlans[pIdx].Id == planId {
			for wIdx := range vault.WorkoutPlans[pIdx].Workouts {
				if vault.WorkoutPlans[pIdx].Workouts[wIdx].Name == wName {
					vault.WorkoutPlans[pIdx].Workouts[wIdx].IsExecuted = true
					PrintSuccess(fmt.Sprintf("Workout '%s' marked as Executed!", wName))
					return nil
				}
			}
		}
	}

	return errors.New("matching Plan ID and Workout Name not found")
}

func promptInput(prompt string) (string, error) {
	fmt.Print(prompt)
	return ReadLine()
}

func promptInt(prompt string) (int, error) {
	val, err := promptInput(prompt)
	if err != nil {
		return 0, err
	}

	num, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("invalid integer: %w", err)
	}
	return num, nil
}

