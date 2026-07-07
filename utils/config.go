package utils

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

var envLoaded bool

func loadEnv() {
	if envLoaded {
		return
	}

	candidates := []string{".env", "phd-api/.env"}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, ".env"))
	}

	for _, path := range candidates {
		if err := godotenv.Load(path); err == nil {
			envLoaded = true
			return
		}
	}

	envLoaded = true
}

// LoadEnv loads environment variables from .env (safe to call multiple times).
func LoadEnv() {
	loadEnv()
}

// Env returns an environment variable, loading .env once if needed.
func Env(key string) string {
	loadEnv()
	return os.Getenv(key)
}
