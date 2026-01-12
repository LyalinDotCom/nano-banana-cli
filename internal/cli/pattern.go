package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/lyalindotcom/nano-banana-cli/internal/gemini"
	"github.com/lyalindotcom/nano-banana-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	// Pattern command flags
	patternOutput string
	patternSize   string
	patternStyle  string
	patternType   string
)

var patternCmd = &cobra.Command{
	Use:   "pattern [prompt]",
	Short: "Generate seamless patterns and textures",
	Long: `Generate seamless, tileable patterns and textures from a text description.

This command generates patterns that can be tiled seamlessly, perfect for:
  - Game textures and backgrounds
  - Website backgrounds
  - Fabric and material designs
  - Wallpapers

TYPES:
  seamless (default) - Tileable pattern that repeats seamlessly
  texture            - Surface texture (wood, metal, fabric, etc.)
  wallpaper          - Decorative wallpaper pattern

STYLES:
  geometric - Shapes, lines, mathematical patterns
  organic   - Natural, flowing, biological forms
  abstract  - Non-representational, artistic
  floral    - Flowers, leaves, botanical elements
  tech      - Digital, circuit-like, futuristic

SIZE:
  Default: 512x512 pixels
  Format: WxH (e.g., 256x256, 1024x512)

EXAMPLES:
  # Generate a seamless geometric pattern
  nanobanana pattern "hexagon grid" -o hex-pattern.png

  # Generate a wood texture
  nanobanana pattern "oak wood grain" -o wood.png --type texture

  # Floral wallpaper pattern
  nanobanana pattern "vintage roses" -o roses.png --type wallpaper --style floral

  # Large abstract pattern
  nanobanana pattern "colorful waves" -o waves.png --size 1024x1024 --style abstract`,
	Args: cobra.MinimumNArgs(1),
	RunE: runPattern,
}

func init() {
	patternCmd.Flags().StringVarP(&patternOutput, "output", "o", "", "Output file path (required)")
	patternCmd.Flags().StringVar(&patternSize, "size", "512x512", "Pattern tile size WxH")
	patternCmd.Flags().StringVar(&patternStyle, "style", "", "Style: geometric, organic, abstract, floral, tech")
	patternCmd.Flags().StringVar(&patternType, "type", "seamless", "Type: seamless, texture, wallpaper")

	patternCmd.MarkFlagRequired("output")

	rootCmd.AddCommand(patternCmd)
}

func runPattern(cmd *cobra.Command, args []string) error {
	prompt := strings.Join(args, " ")
	f := GetFormatter()
	startTime := time.Now()

	// Validate API key
	apiKey := GetAPIKey()
	if apiKey == "" {
		f.Error("pattern", "MISSING_API_KEY", "No API key provided",
			"Set GEMINI_API_KEY environment variable or use --api-key flag")
		return fmt.Errorf("missing API key")
	}

	// Parse size
	var width, height int
	parts := strings.Split(strings.ToLower(patternSize), "x")
	if len(parts) != 2 {
		f.Error("pattern", "INVALID_SIZE",
			fmt.Sprintf("Invalid size format: %s", patternSize),
			"Use format WxH (e.g., 512x512)")
		return fmt.Errorf("invalid size")
	}
	var err error
	width, err = strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		f.Error("pattern", "INVALID_SIZE", "Invalid width value", "")
		return err
	}
	height, err = strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		f.Error("pattern", "INVALID_SIZE", "Invalid height value", "")
		return err
	}

	if width < 64 || width > 2048 || height < 64 || height > 2048 {
		f.Error("pattern", "INVALID_SIZE",
			"Size must be between 64 and 2048 pixels", "")
		return fmt.Errorf("invalid size")
	}

	// Validate type
	validTypes := []string{"seamless", "texture", "wallpaper"}
	typeValid := false
	for _, t := range validTypes {
		if patternType == t {
			typeValid = true
			break
		}
	}
	if !typeValid {
		f.Error("pattern", "INVALID_TYPE",
			fmt.Sprintf("Invalid type: %s", patternType),
			fmt.Sprintf("Valid types: %s", strings.Join(validTypes, ", ")))
		return fmt.Errorf("invalid type")
	}

	// Validate style if provided
	if patternStyle != "" {
		validStyles := []string{"geometric", "organic", "abstract", "floral", "tech"}
		styleValid := false
		for _, s := range validStyles {
			if patternStyle == s {
				styleValid = true
				break
			}
		}
		if !styleValid {
			f.Error("pattern", "INVALID_STYLE",
				fmt.Sprintf("Invalid style: %s", patternStyle),
				fmt.Sprintf("Valid styles: %s", strings.Join(validStyles, ", ")))
			return fmt.Errorf("invalid style")
		}
	}

	// Build enhanced prompt for pattern generation
	enhancedPrompt := buildPatternPrompt(prompt, patternType, patternStyle)

	// Create client
	modelName := gemini.ResolveModelName(GetModel())
	client, err := gemini.NewClient(apiKey, modelName, 2*time.Minute)
	if err != nil {
		f.Error("pattern", "CLIENT_ERROR", err.Error(), "Check your API key")
		return err
	}

	f.Progress("Generating %s pattern with %s...", patternType, modelName)

	ctx := context.Background()

	// Determine aspect ratio from size
	aspectRatio := "1:1"
	if width != height {
		// Find closest supported ratio
		ratio := float64(width) / float64(height)
		if ratio > 1.5 {
			aspectRatio = "16:9"
		} else if ratio < 0.67 {
			aspectRatio = "9:16"
		} else if ratio > 1.2 {
			aspectRatio = "4:3"
		} else if ratio < 0.83 {
			aspectRatio = "3:4"
		}
	}

	config := &gemini.ImageConfig{
		AspectRatio: aspectRatio,
		Count:       1,
	}

	images, err := client.GenerateImage(ctx, enhancedPrompt, config)
	if err != nil {
		if geminiErr, ok := err.(*gemini.GeminiError); ok {
			f.Error("pattern", geminiErr.Code, geminiErr.Message, "")
		} else {
			f.Error("pattern", "GENERATION_FAILED", err.Error(), "")
		}
		return err
	}

	if len(images) == 0 {
		f.Error("pattern", "NO_IMAGE", "No pattern was generated", "Try rephrasing your prompt")
		return fmt.Errorf("no image generated")
	}

	// Save the pattern
	if err := client.SaveImage(images[0], patternOutput); err != nil {
		f.Error("pattern", "SAVE_FAILED", err.Error(), "")
		return err
	}

	f.ImageSaved(patternOutput, width, height)

	// Output success
	elapsed := time.Since(startTime)
	timing := &output.Timing{
		TotalMs: elapsed.Milliseconds(),
	}

	data := map[string]interface{}{
		"prompt": prompt,
		"model":  modelName,
		"type":   patternType,
		"style":  patternStyle,
		"size":   patternSize,
		"image": output.ImageResult{
			Path:   patternOutput,
			Format: "png",
			Size:   &output.ImageSize{Width: width, Height: height},
		},
	}

	f.Success("pattern", data, timing)
	return nil
}

func buildPatternPrompt(basePrompt, patternType, style string) string {
	typeDesc := ""
	switch patternType {
	case "seamless":
		typeDesc = "a seamless, tileable pattern that can repeat infinitely without visible seams or edges"
	case "texture":
		typeDesc = "a realistic surface texture, detailed material appearance"
	case "wallpaper":
		typeDesc = "a decorative wallpaper pattern, suitable for backgrounds"
	default:
		typeDesc = "a seamless, tileable pattern"
	}

	styleDesc := ""
	if style != "" {
		switch style {
		case "geometric":
			styleDesc = " using geometric shapes, clean lines, and mathematical precision"
		case "organic":
			styleDesc = " with organic, natural, flowing forms"
		case "abstract":
			styleDesc = " in an abstract, artistic, non-representational style"
		case "floral":
			styleDesc = " featuring flowers, leaves, and botanical elements"
		case "tech":
			styleDesc = " with a digital, technological, circuit-like aesthetic"
		}
	}

	return fmt.Sprintf("Create %s of %s%s. The pattern should tile seamlessly. High quality, detailed.",
		typeDesc, basePrompt, styleDesc)
}
