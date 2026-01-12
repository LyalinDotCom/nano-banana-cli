package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/lyalindotcom/nano-banana-cli/internal/image"
	"github.com/lyalindotcom/nano-banana-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	// Combine command flags
	combineOutput     string
	combineDirection  string
	combineGap        int
	combineColumns    int
	combineAlign      string
	combineBackground string
)

var combineCmd = &cobra.Command{
	Use:   "combine [images...]",
	Short: "Combine multiple images into one",
	Long: `Combine multiple images into a single image using horizontal, vertical, or grid layout.

DIRECTIONS:
  horizontal (default) - Place images side by side
  vertical            - Stack images top to bottom
  grid                - Arrange in a grid layout

ALIGNMENT (for images of different sizes):
  start  - Align to top (vertical) or left (horizontal)
  center - Center align (default)
  end    - Align to bottom (vertical) or right (horizontal)

BACKGROUND:
  transparent (default) - No background
  white                 - White background
  black                 - Black background
  #RRGGBB              - Custom hex color

This command is perfect for creating:
  - Sprite sheets from individual frames
  - Image strips for panoramas
  - Grid layouts for thumbnails or previews

EXAMPLES:
  # Combine horizontally (sprite strip)
  nanobanana combine frame1.png frame2.png frame3.png -o spritesheet.png

  # Stack vertically
  nanobanana combine top.png middle.png bottom.png -o stacked.png --direction vertical

  # Create a 4-column grid
  nanobanana combine *.png -o grid.png --direction grid --columns 4

  # Add gap between images
  nanobanana combine img1.png img2.png -o combined.png --gap 10

  # White background with centered alignment
  nanobanana combine a.png b.png -o result.png --background white --align center`,
	Args: cobra.MinimumNArgs(2),
	RunE: runCombine,
}

func init() {
	combineCmd.Flags().StringVarP(&combineOutput, "output", "o", "", "Output file path (required)")
	combineCmd.Flags().StringVar(&combineDirection, "direction", "horizontal", "Direction: horizontal, vertical, grid")
	combineCmd.Flags().IntVar(&combineGap, "gap", 0, "Gap between images in pixels")
	combineCmd.Flags().IntVar(&combineColumns, "columns", 0, "Number of columns for grid layout (auto if not set)")
	combineCmd.Flags().StringVar(&combineAlign, "align", "center", "Alignment: start, center, end")
	combineCmd.Flags().StringVar(&combineBackground, "background", "transparent", "Background: transparent, white, black, or #RRGGBB")

	combineCmd.MarkFlagRequired("output")

	rootCmd.AddCommand(combineCmd)
}

func runCombine(cmd *cobra.Command, args []string) error {
	f := GetFormatter()
	startTime := time.Now()

	// Expand glob patterns and validate files
	var inputPaths []string
	for _, arg := range args {
		// Check if it's a glob pattern
		matches, err := filepath.Glob(arg)
		if err != nil {
			f.Error("combine", "INVALID_PATTERN",
				fmt.Sprintf("Invalid glob pattern: %s", arg), "")
			return err
		}

		if len(matches) == 0 {
			// Not a glob, treat as regular file
			if _, err := os.Stat(arg); os.IsNotExist(err) {
				f.Error("combine", "FILE_NOT_FOUND",
					fmt.Sprintf("Input file not found: %s", arg), "")
				return err
			}
			inputPaths = append(inputPaths, arg)
		} else {
			inputPaths = append(inputPaths, matches...)
		}
	}

	if len(inputPaths) < 2 {
		f.Error("combine", "NOT_ENOUGH_IMAGES",
			"At least 2 images are required",
			"Provide multiple image paths or use glob patterns like *.png")
		return fmt.Errorf("not enough images")
	}

	// Validate direction
	switch combineDirection {
	case "horizontal", "vertical", "grid":
		// Valid
	default:
		f.Error("combine", "INVALID_DIRECTION",
			fmt.Sprintf("Invalid direction: %s", combineDirection),
			"Use: horizontal, vertical, or grid")
		return fmt.Errorf("invalid direction")
	}

	// Validate gap
	if combineGap < 0 {
		f.Error("combine", "INVALID_GAP",
			"Gap cannot be negative", "")
		return fmt.Errorf("invalid gap")
	}

	f.Progress("Combining %d images (%s)...", len(inputPaths), combineDirection)

	// Build options
	opts := &image.CombineOptions{
		Direction:  combineDirection,
		Gap:        combineGap,
		Columns:    combineColumns,
		Align:      combineAlign,
		Background: combineBackground,
	}

	// Combine images
	result, err := image.CombineImages(inputPaths, combineOutput, opts)
	if err != nil {
		f.Error("combine", "COMBINE_FAILED", err.Error(), "")
		return err
	}

	f.ImageSaved(combineOutput, result.Width, result.Height)

	// Output success
	elapsed := time.Since(startTime)
	timing := &output.Timing{
		TotalMs: elapsed.Milliseconds(),
	}

	data := map[string]interface{}{
		"inputs": inputPaths,
		"output": combineOutput,
		"image": output.ImageResult{
			Path:   combineOutput,
			Format: result.Format,
			Size:   &output.ImageSize{Width: result.Width, Height: result.Height},
		},
		"options": map[string]interface{}{
			"direction":  combineDirection,
			"gap":        combineGap,
			"columns":    combineColumns,
			"align":      combineAlign,
			"background": combineBackground,
		},
	}

	f.Success("combine", data, timing)
	return nil
}
