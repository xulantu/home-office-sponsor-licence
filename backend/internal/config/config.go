package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Name string `yaml:"name"`
	User string `yaml:"user"`
}

// ServerConfig holds HTTP server settings
type ServerConfig struct {
	Port int `yaml:"port"`
}

// Config holds all application configuration
type Config struct {
	Server       ServerConfig   `yaml:"server"`
	Database     DatabaseConfig `yaml:"database"`
	TestDatabase DatabaseConfig `yaml:"test_database"`
}

// Load reads configuration from config.yaml and .env files.
// configPath is the path to config.yaml.
// envPath is the path to .env (can be empty to skip).
func Load(configPath, envPath string) (*Config, error) {
	// Load .env file first (if it exists) to set environment variables
	if envPath != "" {
		if err := loadEnvFile(envPath); err != nil {
			// Ignore if file doesn't exist, but return other errors
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to load .env: %w", err)
			}
		}
	}

	// Read config.yaml
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// ConnectionString builds a PostgreSQL connection string.
// Password is read from DATABASE_PASSWORD environment variable.
func (d *DatabaseConfig) ConnectionString() string {
	password := os.Getenv("DATABASE_PASSWORD")
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		d.User, password, d.Host, d.Port, d.Name)
}

// loadEnvFile reads a .env file and sets environment variables
func loadEnvFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on first =
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Only set if not already set (environment takes precedence)
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}

	return scanner.Err()
}
