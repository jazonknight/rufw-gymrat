package main

import (
	"gymrat/models"
	"gymrat/storage"
)

func SaveVault(dir string, filename string, data models.GymRatVaultData) error {
	return storage.SaveVault(dir, filename, data)
}

func LoadVault(dir string, filename string) (models.GymRatVaultData, error) {
	return storage.LoadVault(dir, filename)
}

func SaveExercises(dir string, filename string, catalog models.ExerciseCatalogData) error {
	return storage.SaveExercises(dir, filename, catalog)
}

func LoadExercises(dir string, filename string) (models.ExerciseCatalogData, error) {
	return storage.LoadExercises(dir, filename)
}