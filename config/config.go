package config

import (
	"errors"
	"os"
	"strconv"
)

// Config holds application configuration
type Config struct {
	DatabasePath     string
	FacesDir         string
	ModelsDir        string
	DefaultThreshold float64
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		DatabasePath:     "db.json",
		FacesDir:         "faces",
		ModelsDir:        "models",
		DefaultThreshold: 0.75,
	}
}

// LoadConfig loads configuration from environment variables or uses defaults
func LoadConfig() *Config {
	cfg := DefaultConfig()

	if dbPath := os.Getenv("FACE_CLI_DB_PATH"); dbPath != "" {
		cfg.DatabasePath = dbPath
	}

	if facesDir := os.Getenv("FACE_CLI_FACES_DIR"); facesDir != "" {
		cfg.FacesDir = facesDir
	}

	if modelsDir := os.Getenv("FACE_CLI_MODEL_DIR"); modelsDir != "" {
		cfg.ModelsDir = modelsDir
	}

	if threshold := os.Getenv("FACE_CLI_THRESHOLD"); threshold != "" {
		if t, err := strconv.ParseFloat(threshold, 64); err == nil && t >= 0 && t <= 1 {
			cfg.DefaultThreshold = t
		}
	}

	return cfg
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.DatabasePath == "" {
		return errors.New("database path cannot be empty")
	}
	if c.FacesDir == "" {
		return errors.New("faces directory cannot be empty")
	}
	if c.ModelsDir == "" {
		return errors.New("models directory cannot be empty")
	}
	if c.DefaultThreshold < 0 || c.DefaultThreshold > 1 {
		return errors.New("threshold must be between 0 and 1")
	}
	return nil
}
