package utils

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func LoadEnvFile(paths ...string) {
	for _, rawPath := range paths {
		if p := filepath.Clean(rawPath); fileExists(p) {
			_ = godotenv.Load(p)
			return
		}
	}

	if fileExists(".env") {
		_ = godotenv.Load(".env")
	}
}

func fileExists(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}
