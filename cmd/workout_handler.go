package cmd

import (
	"errors"
	"fmt"
	"gymrat/models"
	"strconv"

	"github.com/google/uuid"
)

func GetSetInput(setInput *models.ExerciseSet) error {

	// ask for reps
	reps, err := GetNumberOfReps(string(setInput.FrequencyUnit))
	if err != nil {
		return err
	}
	setInput.RepCount = reps

	// ask for effort
	effort, err := GetPreceivedEffort()
	if err != nil {
		return err
	}
	setInput.PerceivedEffort = effort

	return nil
}

func GetPreceivedEffort() (int, error) {

	fmt.Printf("Please input the level of Preceived effort for this set (0-%d):>",
		maxUnitValues[models.UnitType("effort")])

	inputChoiceEffort, err := ReadLine()
	if err != nil {
		return 0, err
	}
	effortCount, err := strconv.Atoi(inputChoiceEffort)
	if err != nil {
		return 0, err
	}

	return effortCount, nil
}

func GetNumberOfReps(freqUnit string) (int, error) {

	fmt.Printf("Please input the number of Reps for this set max allowed (1-%d):>",
		maxUnitValues[models.UnitType(freqUnit)])
	inputChoiceReps, err := ReadLine()
	if err != nil {
		return 0, err
	}

	repCount, err := strconv.Atoi(inputChoiceReps)
	if err != nil {
		return 0, err
	}

	if repCount < 1 {
		return 0, errors.New("reps must be at least 1 or more")
	}

	return repCount, nil
}

func GetFreqUnit() (string, error) {
	var setUnit string

	fmt.Printf("\nFor this workout what type of sets do you plan to use? \n")

	for {
		fmt.Printf("Please input one of the options: (secs) or (count) :>")
		inputChoiceFreqUnit, err := ReadLine()
		if err != nil {
			return "", err
		}

		switch inputChoiceFreqUnit {
		case "secs":
			setUnit = "seconds"
		case "count":
			setUnit = "count"
		default:
			fmt.Println("\nError | choice invalid please try again ")
			continue // contiune until valid input is used
		}

		fmt.Printf("\nSelected a valid choice: %s \n", inputChoiceFreqUnit)

		break // break out once valid input has been provided
	}
	return setUnit, nil
}

func GetMaxPerUnit(str string) (int, error) {

	if maxValue, exists := maxUnitValues[models.UnitType(str)]; exists {
		return maxValue, nil
	}

	return 0, errors.New("error: invalid unit provided")

}

// HandleCreateWorkoutPlan guides the user through creating and saving a plan
func HandleCreateWorkoutPlan(vault *models.GymRatVaultData) error {
	fmt.Printf("-- Adding a workout Plan Started -- \n")

	planName, err := promptInput("\nEnter Plan Name:")
	if err != nil {
		return err
	}

	planDescription, err := promptInput("\nEnter Plan Description:")
	if err != nil {
		return err
	}

	countExercises, err := promptInt("\nHow many exercises are in this plan?:>")
	if err != nil {
		return err
	}

	// Delegate the complex nested loop work to a dedicated function
	exercises, err := collectExercisesInput(countExercises)
	if err != nil {
		return err
	}

	// Assemble the domain objects and commit to the vault state
	workout := models.Workout{PlannedExercises: exercises}

	planInDraft := models.Plan{
		Id:          uuid.NewString(),
		Name:        planName,
		Description: planDescription,
		Workouts:    []models.Workout{workout},
	}

	vault.WorkoutPlans = append(vault.WorkoutPlans, planInDraft)
	fmt.Printf("-- Plan successfully created with id: %s\n", planInDraft.Id)
	return nil
}

// collectExercisesInput handles the loop to build out all exercises for the plan
func collectExercisesInput(count int) ([]models.Exercise, error) {
	exercises := make([]models.Exercise, 0, count)

	for e := range count {
		exerciseName, err := promptInput(fmt.Sprintf("Please Input the #%d exercise name:>", e+1))
		if err != nil {
			return nil, err
		}

		setUnit, err := GetFreqUnit()
		if err != nil {
			return nil, err
		}

		maxSets := maxUnitValues[models.UnitType("sets")]
		numOfWorkoutSets, err := promptInput(fmt.Sprintf("Input the number of Sets for exercise (%s) range and max (1-%d) :>", exerciseName, maxSets))
		if err != nil {
			return nil, err
		}

		countOfWorkOutSets, err := GetMeAValidMaxValue(numOfWorkoutSets, false, maxSets)
		if err != nil {
			return nil, err
		}

		workoutSets, err := collectSetsInput(countOfWorkOutSets, models.UnitType(setUnit))
		if err != nil {
			return nil, err
		}

		plannedExercise := models.Exercise{
			Id:   uuid.NewString(),
			Name: exerciseName,
			Sets: workoutSets,
		}
		exercises = append(exercises, plannedExercise)
	}

	return exercises, nil
}

// collectSetsInput handles the inner loop to build out individual exercise sets
func collectSetsInput(count int, unit models.UnitType) ([]models.ExerciseSet, error) {
	workoutSets := make([]models.ExerciseSet, 0, count)

	for s := range count {
		fmt.Printf("About to enter details for set number:%d/%d\n", s+1, count)

		plannedSet := models.ExerciseSet{
			Id:            uuid.NewString(),
			FrequencyUnit: unit,
		}

		if err := GetSetInput(&plannedSet); err != nil {
			return nil, err
		}

		workoutSets = append(workoutSets, plannedSet)
	}

	return workoutSets, nil
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
		fmt.Println("error: please type a valid whole number.")
		return 0, err
	}
	return num, nil
}
