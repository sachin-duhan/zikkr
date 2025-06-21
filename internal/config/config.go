package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	// GitHub configuration
	GitHub struct {
		Token string `mapstructure:"token"`
	} `mapstructure:"github"`

	// Clone configuration
	Clone struct {
		MaxConcurrent    int    `mapstructure:"max_concurrent"`
		ConnectTimeout   int    `mapstructure:"connect_timeout"`
		OperationTimeout int    `mapstructure:"operation_timeout"`
		OutputDir        string `mapstructure:"output_dir"`
		ExistingRepos    string `mapstructure:"existing_repos"` // skip, overwrite, fetch-only
	} `mapstructure:"clone"`

	// Logging configuration
	Log struct {
		Level  string `mapstructure:"level"`
		Format string `mapstructure:"format"`
		File   string `mapstructure:"file"`
	} `mapstructure:"log"`

	// Output configuration
	Output struct {
		Format string `mapstructure:"format"` // json, yaml
		File   string `mapstructure:"file"`
	} `mapstructure:"output"`
}

// LoadConfig loads the configuration from various sources
func LoadConfig() (*Config, error) {
	config := &Config{}

	viper.SetDefault("clone.max_concurrent", 5)
	viper.SetDefault("clone.connect_timeout", 60)
	viper.SetDefault("clone.operation_timeout", 600)
	viper.SetDefault("clone.existing_repos", "skip")
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "text")

	// Environment variables
	viper.SetEnvPrefix("ZIKRR")
	viper.AutomaticEnv()

	// Config file
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		configHome = filepath.Join(home, ".config")
	}

	viper.AddConfigPath(configHome)
	viper.AddConfigPath(".")
	viper.SetConfigName(".zikrr")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}

// SaveConfig saves the current configuration to file
func SaveConfig(config *Config) error {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		configHome = filepath.Join(home, ".config")
	}

	configFile := filepath.Join(configHome, ".zikrr.yaml")
	if err := os.MkdirAll(filepath.Dir(configFile), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := viper.WriteConfigAs(configFile); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
