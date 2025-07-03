package config

import (
	"fmt"
	"os"
)

// AppConfig holds all application-wide configuration settings.
type AppConfig struct {
	// Database
	DatabaseURL string

	// SMTP Email Sender
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string

	// Application Base URL (for email links, etc.)
	AppBaseURL string

	// Add other configurations as your app grows
	// ServerPort string
	// DebugMode  bool
}

// LoadConfig loads application configuration from environment variables.
// It returns an AppConfig struct and an error if any required variable is missing.
func LoadConfig() (*AppConfig, error) {
	cfg := &AppConfig{
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     os.Getenv("SMTP_PORT"),
		SMTPUsername: os.Getenv("SMTP_USERNAME"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:     os.Getenv("SMTP_FROM_EMAIL"),
		AppBaseURL:   os.Getenv("APP_BASE_URL"),
	}

	// Basic validation for critical configuration
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is not set")
	}
	if cfg.SMTPHost == "" || cfg.SMTPPort == "" || cfg.SMTPUsername == "" || cfg.SMTPPassword == "" || cfg.SMTPFrom == "" {
		// Note: For actual production, you might want more granular checks or allow some to be optional
		return nil, fmt.Errorf("one or more SMTP environment variables (SMTP_HOST, SMTP_PORT, SMTP_USERNAME, SMTP_PASSWORD, SMTP_FROM_EMAIL) are not set")
	}
	if cfg.AppBaseURL == "" {
		return nil, fmt.Errorf("APP_BASE_URL environment variable is not set")
	}

	return cfg, nil
}

// Example of how to get an integer or boolean from env (if you had them)
/*
func getEnvInt(key string, defaultValue int) int {
	s := os.Getenv(key)
	if s == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		fmt.Printf("Warning: Invalid integer for %s, using default %d. Error: %v\n", key, defaultValue, err)
		return defaultValue
	}
	return val
}

func getEnvBool(key string, defaultValue bool) bool {
	s := os.Getenv(key)
	if s == "" {
		return defaultValue
	}
	val, err := strconv.ParseBool(s)
	if err != nil {
		fmt.Printf("Warning: Invalid boolean for %s, using default %t. Error: %v\n", key, defaultValue, err)
		return defaultValue
	}
	return val
}
*/
