package utils

import (
	"fmt"
	"os"
)

func CheckFileExists(inputPath string) bool {
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return false
	}
	return true
}

func CopyFileToNewPath(oldPath string, newPath string) error {
	// Check if the file exists
	ok := CheckFileExists(oldPath)
	if !ok {
		return fmt.Errorf("unable to copy file %s, it doesn't exist", oldPath)
	}
	// copy the file with the given name
	err := os.Rename(oldPath, newPath)
	if err != nil {
		return err
	}
	return nil
}

func RemoveFolderOrFile(target string) error {
	// Remove all the directories and files
	// Using RemoveAll() function
	err := os.RemoveAll(target)
	if err != nil {
		return err
	}
	return nil
}
