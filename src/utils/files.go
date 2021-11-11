package utils

import "os"

func CheckFileExists(inputPath string) bool {
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return false
	}
	return true
}
