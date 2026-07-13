package cmd

import (
	"fmt"
	"gymrat/models"
	"io"
	"os"
)

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
			ShowWorkoutPlansCLI(vault.WorkoutPlans)
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
			err := HandleCreateWorkoutPlan(vault)
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
