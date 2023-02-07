package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
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

// ReadFilePerRows reads a file and returns the rows in an array of strings
func ReadFilePerRows(filePath string, delimiter string) ([]string, error) {
	rows := make([]string, 0)
	f, err := os.Open(filePath)
	if err != nil {
		return rows, err
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.Trim(line, delimiter)
		rows = append(rows, line)
	}
	return rows, nil
}
