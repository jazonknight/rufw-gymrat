package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"gymrat/models"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// WorkoutSessionsPlans
func ShowWorkoutSessionsPlansCLI(workoutPlans []models.Plan) {

	fmt.Println("Workout Session Plans : ")
	for i, wp := range workoutPlans {

		fmt.Printf("Session Sequence: %d | Session Data: + %+v\n", i, wp)
	}
}

func ShowWorkoutPlanCLI(workoutPlan models.Plan) {

	fmt.Println("Workout Session Plan : ")
	fmt.Printf("Plan Data: + %+v\n", workoutPlan)

}

func ShowMenuOptionsCLI() {

	fmt.Println("---------------------------------")
	fmt.Println("---       Menu Options        ---")
	fmt.Println("---------------------------------")
	fmt.Println("1: Show all work session plans")
	fmt.Println("2: Show a work session plan")
	fmt.Println("3: Add a work session plan")
	fmt.Println("4: Add a workout")
	fmt.Println("5: End session and close vault")
	fmt.Println("---------------------------------")
	fmt.Println("---                           ---")
	fmt.Println("---------------------------------")
}

func ReadLine() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	//Read string until you see the user hit enter
	option, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	option = strings.TrimSpace(option) // trim the \n

	return option, nil
}

func GetMeAValidMaxValue(str string, isZeroAllowed bool, maxAllowed int) (int, error) {

	defaultZero := 0
	newIntValue, err := strconv.Atoi(str)
	if err != nil {
		return defaultZero, err
	}

	if !isZeroAllowed && newIntValue == 0 {
		return defaultZero, errors.New("error: zero is not allowed")
	}

	if newIntValue > maxAllowed {
		return defaultZero, errors.New("error: value is over the max allowed")
	}

	return newIntValue, nil
}

// define map of max units
var maxUnitValues = map[string]int{
	"seconds": 360,
	"count":   30,
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
	//maxValue := 0

	/*
		switch str {
		case "seconds":
			maxValue = 360
		case "count":
			maxValue = 30
		default:
			return 0, errors.New("error: invalid unit provided")
		}
	*/

	if maxValue, exists := maxUnitValues[str]; exists {
		return maxValue, nil
	}

	return 0, errors.New("error: invalid unit provided")

}

func AddWorkoutToPlanCLI(vault *models.GymRatVaultData) error {

	planInDraft := models.Plan{
		Id:       uuid.NewString(),
		Workouts: []models.Workout{},
	}
	fmt.Printf("-- Plan created with id: %s\n", planInDraft.Id)
	fmt.Printf("-- Adding a workout Plan Started -- \n")

	// ask for the name of the plan
	fmt.Printf("\nEnter Plan Name:")
	planName, err := ReadLine()
	if err != nil {
		return err
	}

	planInDraft.Name = planName
	fmt.Printf("Plan name Added successfully: %s\n", planInDraft.Name)

	//ask for descripton of the plan
	fmt.Printf("\nEnter Plan Description:")
	planDescription, err := ReadLine()
	if err != nil {
		return err
	}

	planInDraft.Description = planDescription
	fmt.Printf("Plan Description Added successfully: %s\n", planInDraft.Description)

	// Get the number of excercises in this workout
	fmt.Printf("\n How many exercises are in this plan?: ")
	countOfExercises, err := ReadLine()
	if err != nil {
		return err
	}

	countExercises, err := strconv.Atoi(countOfExercises)
	if err != nil {
		fmt.Println("Error | Please type a valid whole number.")
		return err
	}

	// set max capacity starting with 0 excercises added
	exercises := make([]models.Exercise, 0, countExercises)

	for e := range countExercises {

		// temp exercise
		plannedExercise := models.Exercise{
			Id: uuid.NewString(), // add a unique UUID
		}

		// ask for each exercise details
		fmt.Printf("Lets add details for exercise #%d\n", e)

		//ask for the exercise name
		exerciseName, err := ReadLine()
		if err != nil {
			return err
		}

		plannedExercise.Name = exerciseName // set name

		// get the Unit for the sets
		setUnit, err := GetFreqUnit()
		if err != nil {
			return err
		}

		//Ask how many work out sets for this exercise
		fmt.Println("How many sets do you want to do for this exercise?")
		fmt.Printf("Input the number of Sets for exercise (%s) range (1-%d) :>", exerciseName, maxUnitValues[setUnit])
		numOfWorkoutSets, err := ReadLine()
		if err != nil {
			return err
		}

		//max number of sets
		countOfWorkOutSets, err := GetMeAValidMaxValue(numOfWorkoutSets, false, 5) // no more then 5 sets allowed
		if err != nil {
			return err
		}

		workoutSets := make([]models.ExerciseSet, 0, countOfWorkOutSets)

		//then loop through each work out set
		for s := range countOfWorkOutSets {
			plannedSet := models.ExerciseSet{
				Id:            uuid.NewString(),
				FrequencyUnit: setUnit, // each set is the same unit
			}

			// loop through the set lets say set 0 = s+1
			fmt.Printf("About to enter details for set number:%d/%d", s, countOfWorkOutSets)
			fmt.Printf("\nEnter the number of reps for set number:%d/%d max allowed: %d rep count:>",
				s+1, countOfWorkOutSets, maxUnitValues[plannedSet.FrequencyUnit])
			// ask for the number of reps in this set

			// add the planned set to the main workout sets
			workoutSets = append(workoutSets, plannedSet)
		}

		plannedExercise.Sets = workoutSets

		// add the exercise from the loop
		exercises = append(exercises, plannedExercise)
	}

	//PlannedExercise -> Workout
	workout := models.Workout{
		PlannedExercises: exercises,
	}

	planInDraft.Workouts = append(planInDraft.Workouts, workout) // add workout to plan

	vault.WorkoutPlans = append(vault.WorkoutPlans, planInDraft) // add the plan to the vault

	return nil
}

func MenuCLI(vault *models.GymRatVaultData, dir string, filename string) error {
	// output the menu

	var defaultErrorWriter io.Writer = os.Stderr

	for {

		ShowMenuOptionsCLI()

		choice, err := ReadLine()
		if err != nil {
			fmt.Printf("Terminal error reading input: %v\n", err)
			return err
		}

		fmt.Printf("User Successfully selected an option: %s\n", choice)

		switch choice {
		case "1":
			ShowWorkoutSessionsPlansCLI(vault.WorkoutPlans)
		case "2":

			fmt.Printf("Enter Plan search criteria i.e Id or Name:>")
			choosePlan, err := ReadLine()
			if err != nil {
				return err
			}
			fmt.Printf("Searching for plan...\n")

			foundPlan := false
			for _, p := range vault.WorkoutPlans {

				if p.Id == choosePlan || p.Name == choosePlan {
					ShowWorkoutPlanCLI(p)
					foundPlan = true
					break
				}
			}

			if !foundPlan {
				fmt.Println("info:  No matching plan found.")
			}
		case "3":
			// Add a plan
			// which can have X number of excercises
			// so we will need to know how many excercises they will be adding to the plan
			// high level details
			// then start into the loop of exercises to be added
			//
			err := AddWorkoutToPlanCLI(vault)
			if err != nil {
				fmt.Fprintf(defaultErrorWriter, "error: failed to add workout plan: %v\n", err)
				return err
			}

		case "5":
			fmt.Println("Closing the vault. Session Over!")
			return nil // kills the loop
		default:
			fmt.Printf("\ninfo: You typed: %s, that is not a option yet!\n", choice)
		}

	}

	//return nil
}
