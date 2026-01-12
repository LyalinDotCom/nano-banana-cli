package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/lyalindotcom/nano-banana-cli/internal/image"
	"github.com/lyalindotcom/nano-banana-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	// Transparent make flags
	transparentMakeOutput    string
	transparentMakeColor     string
	transparentMakeTolerance int
	transparentMakeOverwrite bool
)

var transparentCmd = &cobra.Command{
	Use:   "transparent",
	Short: "Transparency manipulation operations",
	Long: `Transparency manipulation operations for images.

SUBCOMMANDS:
  make    - Remove a background color and make it transparent
  inspect - Analyze image transparency and get recommendations

EXAMPLES:
  # Remove white background
  nanobanana transparent make sprite.png -o sprite-clean.png

  # Remove black background with tolerance
  nanobanana transparent make icon.png -o icon-clean.png --color black --tolerance 15

  # Inspect image transparency
  nanobanana transparent inspect image.png`,
}

var transparentMakeCmd = &cobra.Command{
	Use:   "make [input-image]",
	Short: "Remove background color and make it transparent",
	Long: `Remove a specified background color from an image and make those pixels transparent.

COLORS:
  white (default) - Remove white background (#FFFFFF)
  black           - Remove black background (#000000)
  #RRGGBB         - Remove custom hex color

TOLERANCE:
  0   - Exact color match only
  10  - Default, allows slight variations
  100 - Very loose matching

The output is always saved as PNG since transparency requires alpha channel.

EXAMPLES:
  # Remove white background (default)
  nanobanana transparent make sprite.png -o sprite-clean.png

  # Remove black background
  nanobanana transparent make icon.png -o icon-clean.png --color black

  # Remove custom color with high tolerance
  nanobanana transparent make image.png -o output.png --color #E0E0E0 --tolerance 20`,
	Args: cobra.ExactArgs(1),
	RunE: runTransparentMake,
}

var transparentInspectCmd = &cobra.Command{
	Use:   "inspect [input-image]",
	Short: "Analyze image transparency",
	Long: `Analyze an image's transparency and get recommendations.

OUTPUT INCLUDES:
  - Whether the image has an alpha channel
  - Percentage of transparent pixels
  - Image format and dimensions
  - Dominant background color (from edge pixels)
  - Recommendation for transparency processing

EXAMPLES:
  # Inspect an image
  nanobanana transparent inspect sprite.png

  # Get JSON output for programmatic use
  nanobanana transparent inspect sprite.png --json`,
	Args: cobra.ExactArgs(1),
	RunE: runTransparentInspect,
}

func init() {
	// Make subcommand flags
	transparentMakeCmd.Flags().StringVarP(&transparentMakeOutput, "output", "o", "", "Output file path")
	transparentMakeCmd.Flags().StringVar(&transparentMakeColor, "color", "white", "Background color to remove: white, black, or #RRGGBB")
	transparentMakeCmd.Flags().IntVar(&transparentMakeTolerance, "tolerance", 10, "Color matching tolerance (0-100%)")
	transparentMakeCmd.Flags().BoolVar(&transparentMakeOverwrite, "overwrite", false, "Overwrite the original file")

	// Add subcommands
	transparentCmd.AddCommand(transparentMakeCmd)
	transparentCmd.AddCommand(transparentInspectCmd)

	rootCmd.AddCommand(transparentCmd)
}

func runTransparentMake(cmd *cobra.Command, args []string) error {
	inputPath := args[0]
	f := GetFormatter()
	startTime := time.Now()

	// Validate input file exists
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		f.Error("transparent make", "FILE_NOT_FOUND",
			fmt.Sprintf("Input file not found: %s", inputPath), "")
		return err
	}

	// Determine output path
	outputPath := transparentMakeOutput
	if outputPath == "" {
		if transparentMakeOverwrite {
			outputPath = inputPath
		} else {
			// Default: input_transparent.png
			outputPath = inputPath[:len(inputPath)-len(".png")] + "_transparent.png"
			if inputPath[len(inputPath)-4:] != ".png" {
				outputPath = inputPath + "_transparent.png"
			}
		}
	}

	// Validate tolerance
	if transparentMakeTolerance < 0 || transparentMakeTolerance > 100 {
		f.Error("transparent make", "INVALID_TOLERANCE",
			"Tolerance must be between 0 and 100", "")
		return fmt.Errorf("invalid tolerance")
	}

	f.Progress("Removing %s background...", transparentMakeColor)

	// Build options
	opts := &image.TransparencyOptions{
		Color:     transparentMakeColor,
		Tolerance: transparentMakeTolerance,
	}

	// Make transparent
	result, err := image.MakeTransparent(inputPath, outputPath, opts)
	if err != nil {
		f.Error("transparent make", "TRANSPARENCY_FAILED", err.Error(), "")
		return err
	}

	f.ImageSaved(outputPath, result.Width, result.Height)

	// Output success
	elapsed := time.Since(startTime)
	timing := &output.Timing{
		TotalMs: elapsed.Milliseconds(),
	}

	data := map[string]interface{}{
		"input":  inputPath,
		"output": outputPath,
		"image": output.ImageResult{
			Path:   outputPath,
			Format: "png",
			Size:   &output.ImageSize{Width: result.Width, Height: result.Height},
		},
		"options": map[string]interface{}{
			"color":     transparentMakeColor,
			"tolerance": transparentMakeTolerance,
		},
	}

	f.Success("transparent make", data, timing)
	return nil
}

func runTransparentInspect(cmd *cobra.Command, args []string) error {
	inputPath := args[0]
	f := GetFormatter()
	startTime := time.Now()

	// Validate input file exists
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		f.Error("transparent inspect", "FILE_NOT_FOUND",
			fmt.Sprintf("Input file not found: %s", inputPath), "")
		return err
	}

	// Inspect transparency
	result, err := image.InspectTransparency(inputPath)
	if err != nil {
		f.Error("transparent inspect", "INSPECTION_FAILED", err.Error(), "")
		return err
	}

	// Output
	if f.JSONMode {
		elapsed := time.Since(startTime)
		timing := &output.Timing{
			TotalMs: elapsed.Milliseconds(),
		}
		data := map[string]interface{}{
			"input":   inputPath,
			"results": result,
		}
		f.Success("transparent inspect", data, timing)
	} else {
		// Text output
		fmt.Printf("Transparency Analysis: %s\n", inputPath)
		fmt.Printf("  Format:              %s\n", result.Format)
		fmt.Printf("  Dimensions:          %dx%d\n", result.Width, result.Height)
		fmt.Printf("  Has Alpha Channel:   %v\n", result.HasAlphaChannel)
		fmt.Printf("  Transparent Pixels:  %.1f%%\n", result.TransparentPixelPercent)
		fmt.Printf("  Dominant Background: %s\n", result.DominantBackgroundColor)
		fmt.Printf("\n  Recommendation: %s\n", result.Recommendation)
	}

	return nil
}
