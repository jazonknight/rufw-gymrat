package cmd

import (
	"fmt"
	"gymrat/models"
	"os"
	"os/exec"
	"runtime"
)

const Banner = `
   ______               ____       __ 
  / ____/_  ______ ___ / __ \____ _/ /_
 / / __/ / / / __ '____ / /_/ / __ '/ __/
/ /_/ / /_/ / / / / / / / _, _/ /_/ / /_  
\____/\__, /_/ /_/ /_/_/_/ |_|\__,_/\__/  
     /____/                               
`

func ShowMenuOptionsCLI() {

	fmt.Print(Banner)
	fmt.Println("---------------------------------")
	fmt.Println("---       Menu Options        ---")
	fmt.Println("---------------------------------")
	fmt.Println("1: Show all workouts")
	fmt.Println("2: Search for a workout")
	fmt.Println("3: Start a new workout")
	fmt.Println("5: Exit")
	fmt.Println("---------------------------------")
	fmt.Println("---                           ---")
	fmt.Println("---------------------------------")
}

// ClearScreen wipes the terminal view port based on the OS
func ClearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// PrintMainMenu displays the primary choices for the gymrat application
func PrintMainMenu() {
	fmt.Print(Banner)
	fmt.Println("=== MAIN MENU ===")
	fmt.Println("1. Start a New Workout")
	fmt.Println("2. View History")
	fmt.Println("3. Check Progressive Overload Stats")
	fmt.Println("4: Load a workout")
	fmt.Println("5. Exit")
	fmt.Print("\nChoose an option: ")
}

func ShowWorkoutPlansCLI(workoutPlans []models.Plan) {

	fmt.Println("Workout Plans : ")
	for i, wp := range workoutPlans {

		fmt.Printf("Workout Sequence: %d | Workout Data: + %+v\n", i+1, wp)
	}
}

func ShowWorkoutPlanCLI(workoutPlan models.Plan) {

	fmt.Println("Workout Plan : ")
	fmt.Printf("Plan Data: + %+v\n", workoutPlan)

}

const (
	colorReset = "\033[0m"
	colorRed   = "\033[31m"
	colorGreen = "\033[32m"
	colorCyan  = "\033[36m"
)

// PrintError outputs a formatted error message in red
func PrintError(message string) {
	fmt.Printf("%s[ERROR] %s%s\n", colorRed, message, colorReset)
}

// PrintSuccess outputs a confirmation message in green
func PrintSuccess(message string) {
	fmt.Printf("%s[SUCCESS] %s%s\n", colorGreen, message, colorReset)
}

// PrintInfo outputs a general highlight message in cyan
func PrintInfo(message string) {
	fmt.Printf("%s[INFO] %s%s\n", colorCyan, message, colorReset)
}
