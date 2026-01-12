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
	// Transform command flags
	transformOutput string
	transformResize string
	transformFit    string
	transformCrop   string
	transformRotate int
	transformFlip   bool
	transformFlop   bool
)

var transformCmd = &cobra.Command{
	Use:   "transform [input-image]",
	Short: "Apply transformations to images (resize, crop, rotate, flip)",
	Long: `Apply transformations to images including resize, crop, rotate, and flip.

OPERATIONS (applied in order):
  1. Crop    - Extract a region from the image
  2. Resize  - Scale the image to new dimensions
  3. Rotate  - Rotate by specified degrees
  4. Flip    - Flip vertically
  5. Flop    - Flip horizontally (mirror)

RESIZE MODES (--fit):
  inside (default) - Fit within bounds, preserve aspect ratio
  contain          - Same as inside
  cover            - Fill bounds, crop excess to fit
  fill             - Stretch to exact dimensions (may distort)
  outside          - Resize so smallest side matches target

EXAMPLES:
  # Resize to specific dimensions
  nanobanana transform photo.jpg -o thumb.jpg --resize 200x200

  # Resize by percentage
  nanobanana transform large.png -o small.png --resize 50%

  # Resize with cover mode (fill and crop)
  nanobanana transform photo.jpg -o square.jpg --resize 500x500 --fit cover

  # Crop a region
  nanobanana transform image.png -o cropped.png --crop 100,50,400,300

  # Rotate 90 degrees
  nanobanana transform photo.jpg -o rotated.jpg --rotate 90

  # Flip and mirror
  nanobanana transform sprite.png -o flipped.png --flip --flop

  # Combined operations
  nanobanana transform input.png -o output.png --crop 0,0,800,600 --resize 400x300 --rotate 45`,
	Args: cobra.ExactArgs(1),
	RunE: runTransform,
}

func init() {
	transformCmd.Flags().StringVarP(&transformOutput, "output", "o", "", "Output file path (required)")
	transformCmd.Flags().StringVar(&transformResize, "resize", "", "Resize to WxH or percentage (e.g., 800x600, 50%)")
	transformCmd.Flags().StringVar(&transformFit, "fit", "inside", "Fit mode: cover, contain, fill, inside, outside")
	transformCmd.Flags().StringVar(&transformCrop, "crop", "", "Crop region: left,top,width,height")
	transformCmd.Flags().IntVar(&transformRotate, "rotate", 0, "Rotation angle in degrees (-360 to 360)")
	transformCmd.Flags().BoolVar(&transformFlip, "flip", false, "Flip vertically")
	transformCmd.Flags().BoolVar(&transformFlop, "flop", false, "Flip horizontally (mirror)")

	transformCmd.MarkFlagRequired("output")

	rootCmd.AddCommand(transformCmd)
}

func runTransform(cmd *cobra.Command, args []string) error {
	inputPath := args[0]
	f := GetFormatter()
	startTime := time.Now()

	// Validate input file exists
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		f.Error("transform", "FILE_NOT_FOUND",
			fmt.Sprintf("Input file not found: %s", inputPath), "")
		return err
	}

	// Check that at least one operation is specified
	if transformResize == "" && transformCrop == "" && transformRotate == 0 && !transformFlip && !transformFlop {
		f.Error("transform", "NO_OPERATION",
			"No transformation specified",
			"Use --resize, --crop, --rotate, --flip, or --flop")
		return fmt.Errorf("no operation specified")
	}

	// Validate rotation range
	if transformRotate < -360 || transformRotate > 360 {
		f.Error("transform", "INVALID_ROTATION",
			"Rotation must be between -360 and 360 degrees", "")
		return fmt.Errorf("invalid rotation")
	}

	f.Progress("Transforming image...")

	// Build options
	opts := &image.TransformOptions{
		Resize: transformResize,
		Fit:    transformFit,
		Crop:   transformCrop,
		Rotate: transformRotate,
		Flip:   transformFlip,
		Flop:   transformFlop,
	}

	// Apply transformations
	result, err := image.Transform(inputPath, transformOutput, opts)
	if err != nil {
		f.Error("transform", "TRANSFORM_FAILED", err.Error(), "")
		return err
	}

	f.ImageSaved(transformOutput, result.Width, result.Height)

	// Output success
	elapsed := time.Since(startTime)
	timing := &output.Timing{
		TotalMs: elapsed.Milliseconds(),
	}

	data := map[string]interface{}{
		"input":  inputPath,
		"output": transformOutput,
		"image": output.ImageResult{
			Path:   transformOutput,
			Format: result.Format,
			Size:   &output.ImageSize{Width: result.Width, Height: result.Height},
		},
		"operations": map[string]interface{}{
			"resize": transformResize,
			"fit":    transformFit,
			"crop":   transformCrop,
			"rotate": transformRotate,
			"flip":   transformFlip,
			"flop":   transformFlop,
		},
	}

	f.Success("transform", data, timing)
	return nil
}
