package image

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
)

// TransformOptions contains options for image transformation
type TransformOptions struct {
	Resize string // WxH or percentage (e.g., "800x600" or "50%")
	Fit    string // cover, contain, fill, inside, outside
	Crop   string // left,top,width,height
	Rotate int    // degrees (-360 to 360)
	Flip   bool   // vertical flip
	Flop   bool   // horizontal flip (mirror)
}

// TransformResult contains information about the transformed image
type TransformResult struct {
	Width  int
	Height int
	Format string
}

// Transform applies transformations to an image
func Transform(inputPath, outputPath string, opts *TransformOptions) (*TransformResult, error) {
	// Open the input image
	src, err := imaging.Open(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %w", err)
	}

	// Apply transformations in order: Crop -> Resize -> Rotate -> Flip -> Flop
	result := src

	// 1. Crop
	if opts.Crop != "" {
		result, err = applyCrop(result, opts.Crop)
		if err != nil {
			return nil, fmt.Errorf("crop failed: %w", err)
		}
	}

	// 2. Resize
	if opts.Resize != "" {
		result, err = applyResize(result, opts.Resize, opts.Fit)
		if err != nil {
			return nil, fmt.Errorf("resize failed: %w", err)
		}
	}

	// 3. Rotate
	if opts.Rotate != 0 {
		result = imaging.Rotate(result, float64(opts.Rotate), image.Transparent)
	}

	// 4. Flip (vertical)
	if opts.Flip {
		result = imaging.FlipV(result)
	}

	// 5. Flop (horizontal)
	if opts.Flop {
		result = imaging.FlipH(result)
	}

	// Ensure output directory exists
	dir := filepath.Dir(outputPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Save the result
	if err := imaging.Save(result, outputPath); err != nil {
		return nil, fmt.Errorf("failed to save image: %w", err)
	}

	bounds := result.Bounds()
	return &TransformResult{
		Width:  bounds.Dx(),
		Height: bounds.Dy(),
		Format: strings.TrimPrefix(filepath.Ext(outputPath), "."),
	}, nil
}

// applyCrop crops the image to the specified region
func applyCrop(img image.Image, cropSpec string) (image.Image, error) {
	parts := strings.Split(cropSpec, ",")
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid crop format, expected: left,top,width,height")
	}

	left, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, fmt.Errorf("invalid left value: %w", err)
	}
	top, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, fmt.Errorf("invalid top value: %w", err)
	}
	width, err := strconv.Atoi(strings.TrimSpace(parts[2]))
	if err != nil {
		return nil, fmt.Errorf("invalid width value: %w", err)
	}
	height, err := strconv.Atoi(strings.TrimSpace(parts[3]))
	if err != nil {
		return nil, fmt.Errorf("invalid height value: %w", err)
	}

	return imaging.Crop(img, image.Rect(left, top, left+width, top+height)), nil
}

// applyResize resizes the image
func applyResize(img image.Image, sizeSpec string, fit string) (image.Image, error) {
	bounds := img.Bounds()
	origWidth := bounds.Dx()
	origHeight := bounds.Dy()

	var width, height int

	// Check if it's a percentage
	if strings.HasSuffix(sizeSpec, "%") {
		pct, err := strconv.ParseFloat(strings.TrimSuffix(sizeSpec, "%"), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid percentage: %w", err)
		}
		width = int(float64(origWidth) * pct / 100)
		height = int(float64(origHeight) * pct / 100)
	} else {
		// Parse WxH format
		parts := strings.Split(strings.ToLower(sizeSpec), "x")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid size format, expected: WxH or percentage")
		}

		var err error
		width, err = strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return nil, fmt.Errorf("invalid width: %w", err)
		}
		height, err = strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, fmt.Errorf("invalid height: %w", err)
		}
	}

	// Apply fit mode
	switch fit {
	case "fill":
		// Resize to exact dimensions (may distort)
		return imaging.Resize(img, width, height, imaging.Lanczos), nil
	case "contain":
		// Fit within bounds, preserve aspect ratio
		return imaging.Fit(img, width, height, imaging.Lanczos), nil
	case "cover":
		// Fill bounds, crop excess
		return imaging.Fill(img, width, height, imaging.Center, imaging.Lanczos), nil
	case "inside", "":
		// Default: fit inside bounds (same as contain)
		return imaging.Fit(img, width, height, imaging.Lanczos), nil
	case "outside":
		// Resize so smallest dimension matches, may exceed bounds
		aspectOrig := float64(origWidth) / float64(origHeight)
		aspectTarget := float64(width) / float64(height)
		if aspectOrig > aspectTarget {
			// Original is wider, match height
			return imaging.Resize(img, 0, height, imaging.Lanczos), nil
		}
		// Original is taller, match width
		return imaging.Resize(img, width, 0, imaging.Lanczos), nil
	default:
		return nil, fmt.Errorf("invalid fit mode: %s (use: cover, contain, fill, inside, outside)", fit)
	}
}

// GetImageDimensions returns the dimensions of an image file
func GetImageDimensions(path string) (int, int, error) {
	img, err := imaging.Open(path)
	if err != nil {
		return 0, 0, err
	}
	bounds := img.Bounds()
	return bounds.Dx(), bounds.Dy(), nil
}
