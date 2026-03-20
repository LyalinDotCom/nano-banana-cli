package cli

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	appconfig "github.com/lyalindotcom/nano-banana-cli/internal/config"
	"github.com/lyalindotcom/nano-banana-cli/internal/gemini"
	"github.com/lyalindotcom/nano-banana-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	apiKey   string
	model    string
	jsonMode bool
	quiet    bool
	verbose  bool
	noColor  bool

	formatter *output.Formatter

	rootCmd = &cobra.Command{
		Use:   "nanobanana",
		Short: "AI-powered image generation and manipulation CLI",
		Long: `Nanobanana is a command-line tool for generating and manipulating images
using Google's Gemini image generation models.

CAPABILITIES:
  - Text-to-image and image editing with multiple reference images
  - Script-friendly multi-turn image workflows via history files
  - Grounding with Google Search on supported Gemini image models
  - Optional thought output for Gemini 3 image generation
  - Icon generation in multiple sizes
  - Seamless pattern and texture generation
  - Image manipulation: resize, crop, rotate, flip
  - Transparency manipulation: remove backgrounds, inspect alpha channels
  - Image combining: horizontal strips, vertical strips, grids

MODELS:
  - banana2 (default): Gemini 3.1 Flash Image Preview
  - banana / 2.5: Gemini 2.5 Flash Image
  - pro: Gemini 3 Pro Image Preview

AUTHENTICATION:
  Set your Gemini API key via:
  - User config: nanobanana config set-api-key
  - Environment variable: GEMINI_API_KEY or NANOBANANA_API_KEY
  - Flag: --api-key YOUR_KEY
  - .env file in current directory

OUTPUT:
  All commands support --json flag for programmatic parsing.

EXAMPLES:
  # Print the full manual for all commands
  nanobanana docs

  # Generate an image
  nanobanana generate "a robot playing guitar" -o robot.png

  # Edit an existing image
  nanobanana generate "make it look like watercolor" -i photo.jpg -o watercolor.png

  # Generate with Pro at 4K
  nanobanana generate "designer perfume bottle" -m pro --image-size 4K -o bottle.png

  # Ground with Google Search
  nanobanana generate "five day weather poster for NYC" --ground-web -o weather.png

For detailed help on any command:
  nanobanana [command] --help

For a single-command manual covering the whole CLI:
  nanobanana docs`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			godotenv.Load()
			formatter = output.NewFormatter(jsonMode, quiet, noColor)

			if apiKey == "" {
				apiKey = appconfig.GetAPIKey()
			}

		},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "Gemini API key (or set GEMINI_API_KEY env var)")
	rootCmd.PersistentFlags().StringVarP(
		&model,
		"model",
		"m",
		"banana2",
		fmt.Sprintf("Model alias or full model ID. Aliases: %s", strings.Join(gemini.ListModelAliases(), "; ")),
	)
	rootCmd.PersistentFlags().BoolVar(&jsonMode, "json", false, "Output in JSON format")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress non-essential output")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")

	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(versionCmd)
}

func Execute() error {
	return rootCmd.Execute()
}

func GetAPIKey() string {
	return apiKey
}

func GetModel() string {
	return model
}

func GetFormatter() *output.Formatter {
	return formatter
}

func IsVerbose() bool {
	return verbose
}
