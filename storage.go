package main

// 
import (
	"encoding/json"
	"os"
	"gymrat/models" // This gives your storage room access to the models package room
)

// Challenge 1: Write a function that accepts a file path, filename, and the VaultData struct.
// It should use json.MarshalIndent and os.WriteFile to save the file.
func SaveVault( dir string, filename string, data models.GymRatVaultData) error {

	gymRatFile,err := json.MarshalIndent(data, "", "  ")

	if err != nil {
		return err
	}

	fullFilePath := dir + "/" + filename
 
	err = os.WriteFile(fullFilePath,gymRatFile,0644)
	
	return err
}

// Challenge 2: Write a function that accepts a file path and filename.
// It should use os.ReadFile and json.Unmarshal to return a populated models.GymRatVaultData struct.
func LoadVault(dir string, filename string) (models.GymRatVaultData, error) {
	var gymRatVaultData models.GymRatVaultData
	fullFilePath := dir + "/" + filename 

	gymRateFileData,err := os.ReadFile(fullFilePath)
	if err != nil {
		return gymRatVaultData,err
	}

	
	err = json.Unmarshal(gymRateFileData, &gymRatVaultData)
	if err != nil {
		return gymRatVaultData, err
	}

	return gymRatVaultData, nil
}