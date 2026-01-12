package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/lyalindotcom/nano-banana-cli/internal/gemini"
	"github.com/lyalindotcom/nano-banana-cli/internal/image"
	"github.com/lyalindotcom/nano-banana-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	// Icon command flags
	iconOutput     string
	iconSizes      []int
	iconStyle      string
	iconBackground string
)

var iconCmd = &cobra.Command{
	Use:   "icon [prompt]",
	Short: "Generate icons in multiple sizes",
	Long: `Generate icons in multiple sizes from a text description.

This command generates a base icon and then resizes it to multiple sizes,
making it perfect for:
  - App icons (iOS, Android, desktop)
  - Favicons
  - UI elements and buttons

STYLES:
  modern (default) - Clean, contemporary design
  flat             - Flat design without gradients
  minimal          - Simple, minimalist approach
  detailed         - More detailed and complex

BACKGROUNDS:
  transparent (default) - No background (requires PNG output)
  white                 - White background
  black                 - Black background
  #RRGGBB              - Custom hex color

SIZES:
  Default: 64, 128, 256, 512 pixels
  Common icon sizes: 16, 32, 48, 64, 128, 256, 512, 1024

OUTPUT:
  If output is a directory, files are named: icon_<size>.png
  If output is a file pattern with {size}, it's replaced with the size

EXAMPLES:
  # Generate icons in default sizes
  nanobanana icon "coffee cup logo" -o ./icons/

  # Generate specific sizes
  nanobanana icon "settings gear" -o ./icons/ --sizes 16,32,64,128

  # Flat style with white background
  nanobanana icon "play button" -o ./icons/ --style flat --background white

  # Custom naming pattern
  nanobanana icon "app logo" -o ./icons/myapp_{size}.png --sizes 128,256,512`,
	Args: cobra.MinimumNArgs(1),
	RunE: runIcon,
}

func init() {
	iconCmd.Flags().StringVarP(&iconOutput, "output", "o", "", "Output directory or file pattern (required)")
	iconCmd.Flags().IntSliceVar(&iconSizes, "sizes", []int{64, 128, 256, 512}, "Icon sizes in pixels")
	iconCmd.Flags().StringVar(&iconStyle, "style", "modern", "Style: modern, flat, minimal, detailed")
	iconCmd.Flags().StringVar(&iconBackground, "background", "transparent", "Background: transparent, white, black, or #RRGGBB")

	iconCmd.MarkFlagRequired("output")

	rootCmd.AddCommand(iconCmd)
}

func runIcon(cmd *cobra.Command, args []string) error {
	prompt := strings.Join(args, " ")
	f := GetFormatter()
	startTime := time.Now()

	// Validate API key
	apiKey := GetAPIKey()
	if apiKey == "" {
		f.Error("icon", "MISSING_API_KEY", "No API key provided",
			"Set GEMINI_API_KEY environment variable or use --api-key flag")
		return fmt.Errorf("missing API key")
	}

	// Validate sizes
	for _, size := range iconSizes {
		if size < 16 || size > 2048 {
			f.Error("icon", "INVALID_SIZE",
				fmt.Sprintf("Invalid icon size: %d (must be 16-2048)", size),
				"Common sizes: 16, 32, 64, 128, 256, 512, 1024")
			return fmt.Errorf("invalid size")
		}
	}

	// Validate style
	validStyles := []string{"modern", "flat", "minimal", "detailed"}
	styleValid := false
	for _, s := range validStyles {
		if iconStyle == s {
			styleValid = true
			break
		}
	}
	if !styleValid {
		f.Error("icon", "INVALID_STYLE",
			fmt.Sprintf("Invalid style: %s", iconStyle),
			fmt.Sprintf("Valid styles: %s", strings.Join(validStyles, ", ")))
		return fmt.Errorf("invalid style")
	}

	// Build enhanced prompt for icon generation
	enhancedPrompt := buildIconPrompt(prompt, iconStyle, iconBackground)

	// Create client
	modelName := gemini.ResolveModelName(GetModel())
	client, err := gemini.NewClient(apiKey, modelName, 2*time.Minute)
	if err != nil {
		f.Error("icon", "CLIENT_ERROR", err.Error(), "Check your API key")
		return err
	}

	// Generate base icon at largest size
	largestSize := iconSizes[0]
	for _, size := range iconSizes {
		if size > largestSize {
			largestSize = size
		}
	}

	f.Progress("Generating base icon with %s...", modelName)

	ctx := context.Background()
	config := &gemini.ImageConfig{
		AspectRatio: "1:1",
		Count:       1,
	}

	images, err := client.GenerateImage(ctx, enhancedPrompt, config)
	if err != nil {
		if geminiErr, ok := err.(*gemini.GeminiError); ok {
			f.Error("icon", geminiErr.Code, geminiErr.Message, "")
		} else {
			f.Error("icon", "GENERATION_FAILED", err.Error(), "")
		}
		return err
	}

	if len(images) == 0 {
		f.Error("icon", "NO_IMAGE", "No image was generated", "Try rephrasing your prompt")
		return fmt.Errorf("no image generated")
	}

	// Determine output paths
	outputDir := iconOutput
	filePattern := "icon_{size}.png"
	isPattern := strings.Contains(iconOutput, "{size}")

	if isPattern {
		outputDir = filepath.Dir(iconOutput)
		filePattern = filepath.Base(iconOutput)
	} else {
		// Check if output is a directory path
		stat, err := os.Stat(iconOutput)
		if err == nil && stat.IsDir() {
			outputDir = iconOutput
		} else if strings.HasSuffix(iconOutput, "/") || strings.HasSuffix(iconOutput, "\\") {
			outputDir = iconOutput
		}
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		f.Error("icon", "DIR_ERROR", fmt.Sprintf("Failed to create directory: %s", outputDir), "")
		return err
	}

	// Save base image to temp location
	tempFile := filepath.Join(outputDir, "_temp_base.png")
	if err := client.SaveImage(images[0], tempFile); err != nil {
		f.Error("icon", "SAVE_FAILED", err.Error(), "")
		return err
	}
	defer os.Remove(tempFile)

	// Generate all sizes
	var results []output.ImageResult
	for _, size := range iconSizes {
		filename := strings.Replace(filePattern, "{size}", strconv.Itoa(size), -1)
		if !strings.Contains(filePattern, "{size}") {
			filename = fmt.Sprintf("icon_%d.png", size)
		}
		outputPath := filepath.Join(outputDir, filename)

		f.Progress("Creating %dx%d icon...", size, size)

		// Resize the base image
		opts := &image.TransformOptions{
			Resize: fmt.Sprintf("%dx%d", size, size),
			Fit:    "cover",
		}
		_, err := image.Transform(tempFile, outputPath, opts)
		if err != nil {
			f.Error("icon", "RESIZE_FAILED", err.Error(), "")
			return err
		}

		f.ImageSaved(outputPath, size, size)
		results = append(results, output.ImageResult{
			Path:   outputPath,
			Format: "png",
			Size:   &output.ImageSize{Width: size, Height: size},
		})
	}

	// Output success
	elapsed := time.Since(startTime)
	timing := &output.Timing{
		TotalMs: elapsed.Milliseconds(),
	}

	data := map[string]interface{}{
		"prompt": prompt,
		"model":  modelName,
		"style":  iconStyle,
		"sizes":  iconSizes,
		"images": results,
	}

	f.Success("icon", data, timing)
	return nil
}

func buildIconPrompt(basePrompt, style, background string) string {
	styleDesc := ""
	switch style {
	case "flat":
		styleDesc = "flat design, no gradients, solid colors, simple shapes"
	case "minimal":
		styleDesc = "minimalist, simple, clean lines, limited colors"
	case "detailed":
		styleDesc = "detailed, polished, professional, refined"
	case "modern":
		styleDesc = "modern, clean, contemporary design"
	default:
		styleDesc = "modern, clean design"
	}

	bgDesc := ""
	switch background {
	case "transparent", "":
		bgDesc = "on a transparent background"
	case "white":
		bgDesc = "on a clean white background"
	case "black":
		bgDesc = "on a black background"
	default:
		bgDesc = fmt.Sprintf("on a %s colored background", background)
	}

	return fmt.Sprintf("Create an icon of %s. Style: %s. The icon should be %s. Square format, centered, suitable for app icon or UI element.",
		basePrompt, styleDesc, bgDesc)
}
