// Main entry point for the GymRat application.
// Supports running in interactive terminal CLI mode or launching the REST API server & Web UI dashboard.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"gymrat/cmd"
	"gymrat/models"
	"gymrat/server"
	"gymrat/storage"
)

func main() {
	serverMode := flag.Bool("server", false, "Start GymRat REST API Server & Web UI")
	port := flag.String("port", "8080", "Port for REST API Server")
	sessionDir := flag.String("sessions-dir", filepath.Join(".", "data", "sessions"), "Base directory for sessions")
	flag.Parse()

	targetDir := "."
	exercisesFile := storage.DefaultExercisesFilename
	plansFile := storage.DefaultPlansFilename

	fmt.Println("---- RupertFrameworks: GymRat Boot Sequence ----")

	sessionMgr, err := storage.NewSessionManager(*sessionDir)
	if err != nil {
		fmt.Printf("Fatal Error initializing session manager: %v\n", err)
		os.Exit(1)
	}

	if *serverMode {
		webDir := filepath.Join(".", "web")
		srv := server.NewServer(sessionMgr, webDir, *port)
		if err := srv.Start(); err != nil {
			fmt.Printf("Server Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// CLI Mode: Load catalog and plans from targetDir
	catalog, err := LoadExercises(targetDir, exercisesFile)
	if err != nil {
		fmt.Printf("Notice reading exercises file: %v. Initializing empty catalog.\n", err)
		catalog = models.ExerciseCatalogData{Exercises: []models.Exercise{}}
		_ = SaveExercises(targetDir, exercisesFile, catalog)
	}

	vault, err := LoadVault(targetDir, plansFile)
	if err != nil {
		fmt.Printf("Notice reading plans file: %v. Initializing empty vault.\n", err)
		vault = models.GymRatVaultData{WorkoutPlans: []models.Plan{}}
		_ = SaveVault(targetDir, plansFile, vault)
	}

	fmt.Println("GymRat Catalog & Vault Loaded Successfully")
	if err := cmd.MenuCLI(&catalog, &vault, targetDir, exercisesFile, plansFile); err != nil {
		fmt.Printf("CLI exited with error: %v\n", err)
	}
}

