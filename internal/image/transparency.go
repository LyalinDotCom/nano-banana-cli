package image

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
)

// TransparencyOptions contains options for transparency manipulation
type TransparencyOptions struct {
	Color     string // Color to remove: "white", "black", or hex "#RRGGBB"
	Tolerance int    // Color matching tolerance (0-100%)
}

// TransparencyResult contains information about the transparency operation
type TransparencyResult struct {
	Width  int
	Height int
}

// InspectionResult contains information about image transparency
type InspectionResult struct {
	HasAlphaChannel           bool    `json:"has_alpha_channel"`
	TransparentPixelPercent   float64 `json:"transparent_pixel_percent"`
	Format                    string  `json:"format"`
	Width                     int     `json:"width"`
	Height                    int     `json:"height"`
	DominantBackgroundColor   string  `json:"dominant_background_color"`
	Recommendation            string  `json:"recommendation"`
}

// MakeTransparent removes a background color and makes it transparent
func MakeTransparent(inputPath, outputPath string, opts *TransparencyOptions) (*TransparencyResult, error) {
	if opts == nil {
		opts = &TransparencyOptions{
			Color:     "white",
			Tolerance: 10,
		}
	}

	// Parse the target color
	targetColor, err := parseColor(opts.Color)
	if err != nil {
		return nil, fmt.Errorf("invalid color: %w", err)
	}

	// Open the image
	src, err := imaging.Open(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %w", err)
	}

	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Create a new NRGBA image (with alpha channel)
	result := image.NewNRGBA(bounds)

	// Calculate tolerance threshold
	tolerance := float64(opts.Tolerance) / 100.0 * 255.0

	// Process each pixel
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixel := src.At(x, y)
			r, g, b, a := pixel.RGBA()

			// Convert from 16-bit to 8-bit
			r8 := uint8(r >> 8)
			g8 := uint8(g >> 8)
			b8 := uint8(b >> 8)
			a8 := uint8(a >> 8)

			// Check if this pixel matches the target color within tolerance
			if colorDistance(r8, g8, b8, targetColor) <= tolerance {
				// Make transparent
				result.SetNRGBA(x, y, color.NRGBA{R: r8, G: g8, B: b8, A: 0})
			} else {
				// Keep original
				result.SetNRGBA(x, y, color.NRGBA{R: r8, G: g8, B: b8, A: a8})
			}
		}
	}

	// Ensure output directory exists
	dir := filepath.Dir(outputPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Save as PNG (required for transparency)
	outFile, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	if err := png.Encode(outFile, result); err != nil {
		return nil, fmt.Errorf("failed to encode PNG: %w", err)
	}

	return &TransparencyResult{
		Width:  width,
		Height: height,
	}, nil
}

// InspectTransparency analyzes an image's transparency
func InspectTransparency(inputPath string) (*InspectionResult, error) {
	// Open the image
	file, err := os.Open(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %w", err)
	}
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	totalPixels := width * height

	// Count transparent pixels and analyze background
	transparentCount := 0
	colorCounts := make(map[color.RGBA]int)
	hasAlpha := false

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixel := img.At(x, y)
			r, g, b, a := pixel.RGBA()

			// Convert to 8-bit
			r8 := uint8(r >> 8)
			g8 := uint8(g >> 8)
			b8 := uint8(b >> 8)
			a8 := uint8(a >> 8)

			if a8 < 255 {
				hasAlpha = true
				if a8 < 128 {
					transparentCount++
				}
			}

			// Track edge pixels for dominant background color
			if x == bounds.Min.X || x == bounds.Max.X-1 || y == bounds.Min.Y || y == bounds.Max.Y-1 {
				c := color.RGBA{R: r8, G: g8, B: b8, A: 255}
				colorCounts[c]++
			}
		}
	}

	// Find dominant background color
	var dominantColor color.RGBA
	maxCount := 0
	for c, count := range colorCounts {
		if count > maxCount {
			maxCount = count
			dominantColor = c
		}
	}

	dominantColorStr := fmt.Sprintf("#%02X%02X%02X", dominantColor.R, dominantColor.G, dominantColor.B)
	if dominantColor.R == 255 && dominantColor.G == 255 && dominantColor.B == 255 {
		dominantColorStr = "white"
	} else if dominantColor.R == 0 && dominantColor.G == 0 && dominantColor.B == 0 {
		dominantColorStr = "black"
	}

	// Calculate percentage
	transparentPercent := float64(transparentCount) / float64(totalPixels) * 100

	// Generate recommendation
	recommendation := ""
	if !hasAlpha {
		recommendation = fmt.Sprintf("Image has no alpha channel. Use 'nanobanana transparent make' with --color %s to add transparency.", dominantColorStr)
	} else if transparentPercent < 1 {
		recommendation = "Image has alpha channel but very few transparent pixels. Background removal may not have been applied."
	} else {
		recommendation = "Image already has transparency."
	}

	return &InspectionResult{
		HasAlphaChannel:         hasAlpha,
		TransparentPixelPercent: transparentPercent,
		Format:                  format,
		Width:                   width,
		Height:                  height,
		DominantBackgroundColor: dominantColorStr,
		Recommendation:          recommendation,
	}, nil
}

// parseColor converts a color string to RGB values
func parseColor(colorStr string) (color.RGBA, error) {
	switch strings.ToLower(colorStr) {
	case "white":
		return color.RGBA{R: 255, G: 255, B: 255, A: 255}, nil
	case "black":
		return color.RGBA{R: 0, G: 0, B: 0, A: 255}, nil
	default:
		// Try to parse as hex
		if strings.HasPrefix(colorStr, "#") {
			colorStr = colorStr[1:]
		}
		if len(colorStr) != 6 {
			return color.RGBA{}, fmt.Errorf("invalid hex color format, expected #RRGGBB")
		}
		r, err := strconv.ParseUint(colorStr[0:2], 16, 8)
		if err != nil {
			return color.RGBA{}, err
		}
		g, err := strconv.ParseUint(colorStr[2:4], 16, 8)
		if err != nil {
			return color.RGBA{}, err
		}
		b, err := strconv.ParseUint(colorStr[4:6], 16, 8)
		if err != nil {
			return color.RGBA{}, err
		}
		return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}, nil
	}
}

// colorDistance calculates the Euclidean distance between two colors
func colorDistance(r, g, b uint8, target color.RGBA) float64 {
	dr := float64(r) - float64(target.R)
	dg := float64(g) - float64(target.G)
	db := float64(b) - float64(target.B)
	return (dr*dr + dg*dg + db*db) / 3.0 // Simplified, not true Euclidean
}
