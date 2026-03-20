package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	APIKey    string        `mapstructure:"api_key"`
	Model     string        `mapstructure:"model"`
	Timeout   time.Duration `mapstructure:"timeout"`
	OutputDir string        `mapstructure:"output_dir"`
}

const DefaultModel = "gemini-3.1-flash-image-preview"
const ProModel = "gemini-3-pro-image-preview"
const DefaultTimeout = 2 * time.Minute

func ConfigFilePath() (string, error) {
	if customDir := strings.TrimSpace(os.Getenv("NANOBANANA_CONFIG_DIR")); customDir != "" {
		return filepath.Join(customDir, "config.yaml"), nil
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "nanobanana", "config.yaml"), nil
}

func Load() (*Config, error) {
	v := newViper()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func Save(cfg *Config) error {
	if cfg == nil {
		return errors.New("config is required")
	}

	path, err := ConfigFilePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	v := viper.New()
	v.Set("api_key", cfg.APIKey)
	v.Set("model", cfg.Model)
	v.Set("timeout", cfg.Timeout.String())
	v.Set("output_dir", cfg.OutputDir)
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	return v.WriteConfigAs(path)
}

func SetAPIKey(apiKey string) error {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return errors.New("api key cannot be empty")
	}

	cfg, err := Load()
	if err != nil {
		return err
	}
	cfg.APIKey = apiKey
	if cfg.Model == "" {
		cfg.Model = DefaultModel
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = DefaultTimeout
	}
	if cfg.OutputDir == "" {
		cfg.OutputDir = "."
	}
	return Save(cfg)
}

func ClearAPIKey() error {
	cfg, err := Load()
	if err != nil {
		return err
	}
	cfg.APIKey = ""
	return Save(cfg)
}

func GetAPIKey() string {
	if key := os.Getenv("GEMINI_API_KEY"); key != "" {
		return key
	}
	if key := os.Getenv("NANOBANANA_API_KEY"); key != "" {
		return key
	}
	if key := os.Getenv("GOOGLE_API_KEY"); key != "" {
		return key
	}
	cfg, err := Load()
	if err == nil && cfg.APIKey != "" {
		return cfg.APIKey
	}
	return ""
}

func ResolveModel(model string) string {
	switch model {
	case "", "banana2", "3.1":
		return DefaultModel
	case "banana", "2.5":
		return "gemini-2.5-flash-image"
	case "pro":
		return ProModel
	default:
		return model
	}
}

func newViper() *viper.Viper {
	v := viper.New()
	v.SetDefault("model", DefaultModel)
	v.SetDefault("timeout", DefaultTimeout)
	v.SetDefault("output_dir", ".")

	v.SetEnvPrefix("NANOBANANA")
	v.AutomaticEnv()
	_ = v.BindEnv("api_key", "GEMINI_API_KEY", "NANOBANANA_API_KEY", "GOOGLE_API_KEY")
	_ = v.BindEnv("model", "NANOBANANA_MODEL")

	if path, err := ConfigFilePath(); err == nil {
		v.SetConfigFile(path)
		v.SetConfigType("yaml")
		_ = v.ReadInConfig()
	}

	return v
}
