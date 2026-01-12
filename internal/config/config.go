package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration values
type Config struct {
	APIKey    string        `mapstructure:"api_key"`
	Model     string        `mapstructure:"model"`
	Timeout   time.Duration `mapstructure:"timeout"`
	OutputDir string        `mapstructure:"output_dir"`
}

// DefaultModel is the default Gemini model for image generation
const DefaultModel = "gemini-2.5-flash-image-preview"

// ProModel is the higher quality Gemini 3 Pro model
const ProModel = "gemini-3-pro-image-preview"

// DefaultTimeout for API requests
const DefaultTimeout = 2 * time.Minute

// Load reads configuration from environment variables and config file
func Load() (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("model", DefaultModel)
	v.SetDefault("timeout", DefaultTimeout)
	v.SetDefault("output_dir", ".")

	// Environment variables
	v.SetEnvPrefix("NANOBANANA")
	v.AutomaticEnv()

	// Bind specific env vars
	v.BindEnv("api_key", "GEMINI_API_KEY", "NANOBANANA_API_KEY", "GOOGLE_API_KEY")
	v.BindEnv("model", "NANOBANANA_MODEL")

	// Config file (optional)
	home, err := os.UserHomeDir()
	if err == nil {
		configDir := filepath.Join(home, ".config", "nanobanana")
		v.AddConfigPath(configDir)
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.ReadInConfig() // Ignore error if file doesn't exist
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// GetAPIKey returns the API key from config or environment
func GetAPIKey() string {
	// Check environment variables in order of priority
	if key := os.Getenv("GEMINI_API_KEY"); key != "" {
		return key
	}
	if key := os.Getenv("NANOBANANA_API_KEY"); key != "" {
		return key
	}
	if key := os.Getenv("GOOGLE_API_KEY"); key != "" {
		return key
	}
	return ""
}

// ResolveModel converts short model names to full model IDs
func ResolveModel(model string) string {
	switch model {
	case "flash", "":
		return DefaultModel
	case "pro":
		return ProModel
	default:
		return model
	}
}
