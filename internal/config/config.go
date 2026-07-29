package config

import (
	"os"
	"strconv"
)

// Config holds all application configuration from environment variables.
type Config struct {
	DataDir         string
	Port            string
	QualityDefault  int
	MaxWidthDefault int
	MinSavingKB     int64
}

// Load reads configuration from environment variables with defaults.
func Load() *Config {
	return &Config{
		DataDir:         getEnv("MC_DATA_DIR", "./data"),
		Port:            getEnv("MC_PORT", "8080"),
		QualityDefault:  getEnvInt("MC_QUALITY_DEFAULT", 80),
		MaxWidthDefault: getEnvInt("MC_MAX_WIDTH_DEFAULT", 1920),
		MinSavingKB:     int64(getEnvInt("MC_MIN_SAVING_KB", 50)),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
