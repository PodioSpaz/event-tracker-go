package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds application configuration
type Config struct {
	Database DatabaseConfig
	Logging  LoggingConfig
	App      AppConfig
}

// DatabaseConfig holds database-specific settings
type DatabaseConfig struct {
	Path string
}

// LoggingConfig holds logging settings
type LoggingConfig struct {
	Level  string
	Format string // "json" or "console"
}

// AppConfig holds application settings
type AppConfig struct {
	Name    string
	Version string
	DataDir string
	LogsDir string
}

// Load reads configuration from file and environment variables
func Load() (*Config, error) {
	// Set defaults
	viper.SetDefault("database.path", "data/events.db")
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "console")
	viper.SetDefault("app.name", "Event Tracker")
	viper.SetDefault("app.version", "0.1.0")
	viper.SetDefault("app.data_dir", "data")
	viper.SetDefault("app.logs_dir", "logs")

	// Set config file name and paths
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Look for config in multiple locations
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.event-tracker")
	viper.AddConfigPath("/etc/event-tracker")

	// Read environment variables with prefix
	viper.SetEnvPrefix("EVENT_TRACKER")
	viper.AutomaticEnv()

	// Attempt to read config file (optional - use defaults if not found)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found; using defaults
	}

	// Unmarshal config
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	// Ensure absolute paths
	cfg.Database.Path = resolvePath(cfg.Database.Path)
	cfg.App.DataDir = resolvePath(cfg.App.DataDir)
	cfg.App.LogsDir = resolvePath(cfg.App.LogsDir)

	// Create directories if they don't exist
	if err := os.MkdirAll(cfg.App.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}
	if err := os.MkdirAll(cfg.App.LogsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %w", err)
	}

	return &cfg, nil
}

// resolvePath converts relative paths to absolute paths
func resolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return path
	}

	return filepath.Join(cwd, path)
}

// GetDatabasePath returns the absolute path to the database file
func (c *Config) GetDatabasePath() string {
	return c.Database.Path
}

// GetLogFilePath returns the absolute path to the log file
func (c *Config) GetLogFilePath() string {
	return filepath.Join(c.App.LogsDir, "event-tracker.log")
}

// IsDebug returns true if logging level is set to debug
func (c *Config) IsDebug() bool {
	return c.Logging.Level == "debug" || c.Logging.Level == "trace"
}

// SaveDefaults creates a default config.yaml file if it doesn't exist
func SaveDefaults(path string) error {
	// Check if file already exists
	if _, err := os.Stat(path); err == nil {
		return nil // File exists, don't overwrite
	}

	// Create default config content
	defaultConfig := `# Event Tracker Configuration

# Database settings
database:
  path: data/events.db

# Logging settings
logging:
  level: info  # Options: debug, info, warn, error
  format: console  # Options: console, json

# Application settings
app:
  name: Event Tracker
  version: 0.1.0
  data_dir: data
  logs_dir: logs
`

	// Write to file
	if err := os.WriteFile(path, []byte(defaultConfig), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
