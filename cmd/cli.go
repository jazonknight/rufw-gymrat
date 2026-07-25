// Package cmd implements the interactive terminal command-line interface (CLI) for GymRat.
package cmd

import (
	"fmt"
	"gymrat/models"
	"gymrat/storage"
	"io"
	"os"
)

// MenuCLI runs the primary interactive menu loop for managing exercise catalogs and workout plans.
func MenuCLI(catalog *models.ExerciseCatalogData, vault *models.GymRatVaultData, dir string, exercisesFile string, plansFile string) error {
	var defaultErrorWriter io.Writer = os.Stderr

	for {
		ShowMenuOptionsCLI()

		choice, err := ReadLine()
		if err != nil {
			fmt.Printf("Terminal error reading input: %v\n", err)
			return err
		}

		switch choice {
		case "1":
			HandleShowCatalog(catalog)
		case "2":
			if err := HandleAddExercise(catalog); err != nil {
				fmt.Fprintf(defaultErrorWriter, "error adding exercise: %v\n", err)
			} else {
				_ = storage.SaveExercises(dir, exercisesFile, *catalog)
			}
		case "3":
			if err := HandleUpdateExercise(catalog); err != nil {
				fmt.Fprintf(defaultErrorWriter, "error updating exercise: %v\n", err)
			} else {
				_ = storage.SaveExercises(dir, exercisesFile, *catalog)
			}
		case "4":
			if err := HandleSoftDeleteExercise(catalog); err != nil {
				fmt.Fprintf(defaultErrorWriter, "error toggling soft delete: %v\n", err)
			} else {
				_ = storage.SaveExercises(dir, exercisesFile, *catalog)
			}
		case "5":
			ShowWorkoutPlansCLI(vault.WorkoutPlans)
		case "6":
			fmt.Printf("Enter Plan search criteria (ID or Name): ")
			choosePlan, err := ReadLine()
			if err != nil {
				return err
			}
			foundPlan := false
			for _, p := range vault.WorkoutPlans {
				if p.Id == choosePlan || p.Name == choosePlan {
					ShowWorkoutPlanCLI(p)
					foundPlan = true
					break
				}
			}
			if !foundPlan {
				fmt.Println("info: No matching plan found.")
			}
		case "7":
			if err := HandleCreateWorkoutPlan(catalog, vault); err != nil {
				fmt.Fprintf(defaultErrorWriter, "error creating plan: %v\n", err)
			} else {
				_ = storage.SaveVault(dir, plansFile, *vault)
			}
		case "8":
			if err := HandleMarkWorkoutExecuted(vault); err != nil {
				fmt.Fprintf(defaultErrorWriter, "error marking workout executed: %v\n", err)
			} else {
				_ = storage.SaveVault(dir, plansFile, *vault)
			}
		case "9":
			fmt.Println("Closing GymRat CLI. Goodbye!")
			return nil
		default:
			fmt.Printf("\nInvalid option: '%s'. Please select a valid number (1-9).\n", choice)
		}
	}
}

