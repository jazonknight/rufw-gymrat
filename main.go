package main

import (
	"fmt"
	"os"

	//"time"
	"gymrat/cmd"
	"gymrat/models" // Double check that this matches your go.mod module name
)

func main() {
	targetDir := "."
	targetFile := "gymrat_vault.json"
	fmt.Println("---- RupertFrameworks: GymRat Boot Sequence ----")

	gymVault, err := LoadVault(targetDir, targetFile)

	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("File Not Found, We need to create a new one here")
			// This is where we save vault
			gymVault = models.GymRatVaultData{
				WorkoutPlans: []models.Plan{},
				Workouts:     []models.HistoricWorkouts{},
			}

			err = SaveVault(targetDir, targetFile, gymVault)
			if err != nil {
				fmt.Printf("Save Error Found:%v \n", err)
				return
			}

			fmt.Println("File Actually Written")
			return
		} else {
			fmt.Printf("Fatal Error reading file: %v\n", err)
			return // dont contiune
		}
	} else {
		fmt.Println("GymRat Vault Loaded Successfully")
		fmt.Printf("Loaded Struct Memory Address: %+v\n", gymVault)
		cmd.MenuCLI(&gymVault, targetDir, targetFile)
		return
	}

}
