package cli

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/lyalindotcom/nano-banana-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	// Global flags
	apiKey   string
	model    string
	jsonMode bool
	quiet    bool
	verbose  bool
	noColor  bool

	// Formatter for output
	formatter *output.Formatter

	// Root command
	rootCmd = &cobra.Command{
		Use:   "nanobanana",
		Short: "AI-powered image generation and manipulation CLI",
		Long: `Nanobanana is a command-line tool for generating and manipulating images
using Google's Gemini image generation models.

CAPABILITIES:
  - Text-to-image generation with various aspect ratios
  - Image editing using natural language instructions
  - Icon generation in multiple sizes
  - Seamless pattern and texture generation
  - Image manipulation: resize, crop, rotate, flip
  - Transparency manipulation: remove backgrounds, inspect alpha channels
  - Image combining: horizontal strips, vertical strips, grids

MODELS:
  - flash (default): Gemini 2.5 Flash - Fast, optimized for high-volume tasks
  - pro: Gemini 3 Pro - Professional quality, supports 4K resolution

AUTHENTICATION:
  Set your Gemini API key via:
  - Environment variable: GEMINI_API_KEY or NANOBANANA_API_KEY
  - Flag: --api-key YOUR_KEY
  - .env file in current directory

OUTPUT:
  All commands support --json flag for programmatic parsing.

EXAMPLES:
  # Generate an image
  nanobanana generate "a robot playing guitar" -o robot.png

  # Edit an existing image
  nanobanana generate "make it look like watercolor" -i photo.jpg -o watercolor.png

  # Generate app icons in multiple sizes
  nanobanana icon "coffee cup logo" -o ./icons/ --sizes 64,128,256,512

  # Remove background from image
  nanobanana transparent make sprite.png -o sprite-clean.png

  # Resize an image
  nanobanana transform image.png -o thumb.png --resize 200x200

For detailed help on any command:
  nanobanana [command] --help`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Load .env file if present
			godotenv.Load()

			// Initialize formatter
			formatter = output.NewFormatter(jsonMode, quiet, noColor)

			// Get API key from flag or environment
			if apiKey == "" {
				apiKey = os.Getenv("GEMINI_API_KEY")
			}
			if apiKey == "" {
				apiKey = os.Getenv("NANOBANANA_API_KEY")
			}
			if apiKey == "" {
				apiKey = os.Getenv("GOOGLE_API_KEY")
			}
		},
	}
)

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "Gemini API key (or set GEMINI_API_KEY env var)")
	rootCmd.PersistentFlags().StringVarP(&model, "model", "m", "flash", "Model: flash (default), pro")
	rootCmd.PersistentFlags().BoolVar(&jsonMode, "json", false, "Output in JSON format")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress non-essential output")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")

	// Add subcommands
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(versionCmd)
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// GetAPIKey returns the configured API key
func GetAPIKey() string {
	return apiKey
}

// GetModel returns the configured model
func GetModel() string {
	return model
}

// GetFormatter returns the output formatter
func GetFormatter() *output.Formatter {
	return formatter
}

// IsVerbose returns whether verbose mode is enabled
func IsVerbose() bool {
	return verbose
}
