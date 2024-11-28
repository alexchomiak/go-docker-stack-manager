package utility

import (
	"errors"
	"log"
	"os"
	"path/filepath"
)

func CreateTempFile(directory string, pattern string, contents []byte) (*os.File, error) {

	dir, err := filepath.Abs(directory)
	if err != nil {
		return nil, errors.New("Failed to get absolute path")
	}

	tempFile, err := os.CreateTemp(dir, pattern)
	if err != nil {
		log.Printf("Error creating temporary file: %v", err)
		return nil, errors.New("Failed to create temporary file")
	}
	if _, err := tempFile.Write(contents); err != nil {
		log.Printf("Error writing to temporary file: %v", err)
		return nil, errors.New("Failed to write to temporary file")
	}
	tempFile.Close()

	return tempFile, nil
}
