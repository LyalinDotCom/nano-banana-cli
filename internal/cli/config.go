package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	appconfig "github.com/lyalindotcom/nano-banana-cli/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage persistent nanobanana configuration",
	Long: `Manage persistent user-level configuration for nanobanana.

This is the recommended way to persist an API key outside local project files.
It is useful when:
  - the binary is installed globally
  - local .env files are unavailable
  - agents run in sandboxes that do not preserve project-local secrets

Examples:
  nanobanana config set-api-key
  nanobanana config set-api-key YOUR_API_KEY
  nanobanana config show
  nanobanana config path
  nanobanana config clear-api-key`,
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print the user config file path",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := GetFormatter()
		path, err := appconfig.ConfigFilePath()
		if err != nil {
			f.Error("config path", "CONFIG_PATH_ERROR", err.Error(), "")
			return err
		}

		if f.JSONMode {
			f.Success("config path", map[string]any{"path": path}, nil)
			return nil
		}

		fmt.Println(path)
		return nil
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show config status without revealing the API key",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := GetFormatter()
		path, err := appconfig.ConfigFilePath()
		if err != nil {
			f.Error("config show", "CONFIG_PATH_ERROR", err.Error(), "")
			return err
		}

		cfg, err := appconfig.Load()
		if err != nil {
			f.Error("config show", "CONFIG_LOAD_ERROR", err.Error(), "")
			return err
		}

		data := map[string]any{
			"path":               path,
			"api_key_configured": cfg.APIKey != "",
			"default_model":      cfg.Model,
			"output_dir":         cfg.OutputDir,
			"timeout":            cfg.Timeout.String(),
		}

		f.Success("config show", data, nil)
		if f.JSONMode {
			return nil
		}

		fmt.Printf("Config file: %s\n", path)
		fmt.Printf("API key configured: %v\n", cfg.APIKey != "")
		fmt.Printf("Default model: %s\n", cfg.Model)
		fmt.Printf("Output dir: %s\n", cfg.OutputDir)
		fmt.Printf("Timeout: %s\n", cfg.Timeout.String())
		return nil
	},
}

var configSetAPIKeyCmd = &cobra.Command{
	Use:   "set-api-key [key]",
	Short: "Persist an API key in the user config file",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := GetFormatter()

		var apiKey string
		if len(args) == 1 {
			apiKey = strings.TrimSpace(args[0])
		} else {
			fmt.Fprint(os.Stderr, "Enter Gemini API key: ")
			value, err := bufio.NewReader(os.Stdin).ReadString('\n')
			if err != nil {
				f.Error("config set-api-key", "READ_API_KEY_ERROR", err.Error(), "")
				return err
			}
			apiKey = strings.TrimSpace(value)
		}

		if err := appconfig.SetAPIKey(apiKey); err != nil {
			f.Error("config set-api-key", "SET_API_KEY_ERROR", err.Error(), "")
			return err
		}

		path, _ := appconfig.ConfigFilePath()
		f.Success("config set-api-key", map[string]any{
			"path":               path,
			"api_key_configured": true,
		}, nil)
		if !f.JSONMode {
			fmt.Printf("Saved API key to %s\n", path)
		}
		return nil
	},
}

var configClearAPIKeyCmd = &cobra.Command{
	Use:   "clear-api-key",
	Short: "Remove the API key from the user config file",
	RunE: func(cmd *cobra.Command, args []string) error {
		f := GetFormatter()
		if err := appconfig.ClearAPIKey(); err != nil {
			f.Error("config clear-api-key", "CLEAR_API_KEY_ERROR", err.Error(), "")
			return err
		}
		path, _ := appconfig.ConfigFilePath()
		f.Success("config clear-api-key", map[string]any{
			"path":               path,
			"api_key_configured": false,
		}, nil)
		if !f.JSONMode {
			fmt.Printf("Cleared API key from %s\n", path)
		}
		return nil
	},
}

func init() {
	configCmd.AddCommand(configPathCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetAPIKeyCmd)
	configCmd.AddCommand(configClearAPIKeyCmd)

	rootCmd.AddCommand(configCmd)
}
