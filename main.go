// Main entry point for the GymRat application.
// Supports running in interactive terminal CLI mode or launching the REST API server & Web UI dashboard.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gymrat/cmd"
	"gymrat/models"
	"gymrat/server"
	"gymrat/storage"
)

func getEnv(key string, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func main() {
	defaultPort := getEnv("PORT", "8080")
	defaultSessionDir := getEnv("SESSIONS_DIR", filepath.Join(".", "data", "sessions"))
	defaultWebDir := getEnv("WEB_DIR", filepath.Join(".", "web"))
	defaultStorageMode := getEnv("STORAGE_MODE", "memory")
	defaultSeedFile := getEnv("SEED_FILE", storage.DefaultSeedFilename)
	defaultMaxPayloadMB := 5
	if val := os.Getenv("MAX_PAYLOAD_MB"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			defaultMaxPayloadMB = parsed
		}
	}
	defaultServerMode := os.Getenv("SERVER_MODE") == "true" || os.Getenv("MODE") == "server"

	serverMode := flag.Bool("server", defaultServerMode, "Start GymRat REST API Server & Web UI")
	port := flag.String("port", defaultPort, "Port for REST API Server")
	sessionDir := flag.String("sessions-dir", defaultSessionDir, "Base directory for sessions")
	webDir := flag.String("web-dir", defaultWebDir, "Directory containing static web assets")
	storageMode := flag.String("storage-mode", defaultStorageMode, "Session storage mode ('memory' or 'disk')")
	seedFile := flag.String("seed-file", defaultSeedFile, "Path to immutable default exercises JSON seed file")
	maxPayloadMB := flag.Int("max-payload-mb", defaultMaxPayloadMB, "Max upload request payload size in MB")
	flag.Parse()

	targetDir := "."
	exercisesFile := storage.DefaultExercisesFilename
	plansFile := storage.DefaultPlansFilename

	fmt.Println("---- RupertFrameworks: GymRat Boot Sequence ----")

	sessionMgr, err := storage.NewSessionManager(*sessionDir, *seedFile, *storageMode)
	if err != nil {
		fmt.Printf("Fatal Error initializing session manager: %v\n", err)
		os.Exit(1)
	}

	if *serverMode {
		srv := server.NewServer(sessionMgr, *webDir, *port, *maxPayloadMB)
		if err := srv.Start(); err != nil {
			fmt.Printf("Server Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// CLI Mode: Load catalog and plans from targetDir
	catalog, err := storage.LoadExercises(targetDir, exercisesFile)
	if err != nil {
		fmt.Printf("Notice reading exercises file: %v. Initializing empty catalog.\n", err)
		catalog = models.ExerciseCatalogData{Exercises: []models.Exercise{}}
		_ = storage.SaveExercises(targetDir, exercisesFile, catalog)
	}

	vault, err := storage.LoadVault(targetDir, plansFile)
	if err != nil {
		fmt.Printf("Notice reading plans file: %v. Initializing empty vault.\n", err)
		vault = models.GymRatVaultData{WorkoutPlans: []models.Plan{}}
		_ = storage.SaveVault(targetDir, plansFile, vault)
	}

	fmt.Println("GymRat Catalog & Vault Loaded Successfully")
	if err := cmd.MenuCLI(&catalog, &vault, targetDir, exercisesFile, plansFile); err != nil {
		fmt.Printf("CLI exited with error: %v\n", err)
	}
}


