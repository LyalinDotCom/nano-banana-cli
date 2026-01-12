package image

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
)

// CombineOptions contains options for combining images
type CombineOptions struct {
	Direction  string // horizontal, vertical, grid
	Gap        int    // Gap between images in pixels
	Columns    int    // Number of columns for grid layout
	Align      string // start, center, end
	Background string // transparent, white, black, or hex
}

// CombineResult contains information about the combined image
type CombineResult struct {
	Width  int
	Height int
	Format string
}

// CombineImages combines multiple images into one
func CombineImages(inputPaths []string, outputPath string, opts *CombineOptions) (*CombineResult, error) {
	if len(inputPaths) < 2 {
		return nil, fmt.Errorf("at least 2 images are required")
	}

	if opts == nil {
		opts = &CombineOptions{
			Direction:  "horizontal",
			Gap:        0,
			Align:      "center",
			Background: "transparent",
		}
	}

	// Load all images
	images := make([]image.Image, len(inputPaths))
	for i, path := range inputPaths {
		img, err := imaging.Open(path)
		if err != nil {
			return nil, fmt.Errorf("failed to open %s: %w", path, err)
		}
		images[i] = img
	}

	// Calculate dimensions and create canvas
	var result *image.NRGBA
	switch opts.Direction {
	case "horizontal":
		result = combineHorizontal(images, opts)
	case "vertical":
		result = combineVertical(images, opts)
	case "grid":
		result = combineGrid(images, opts)
	default:
		return nil, fmt.Errorf("invalid direction: %s (use: horizontal, vertical, grid)", opts.Direction)
	}

	// Ensure output directory exists
	dir := filepath.Dir(outputPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Save the result
	outFile, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	if err := png.Encode(outFile, result); err != nil {
		return nil, fmt.Errorf("failed to encode PNG: %w", err)
	}

	bounds := result.Bounds()
	return &CombineResult{
		Width:  bounds.Dx(),
		Height: bounds.Dy(),
		Format: "png",
	}, nil
}

func combineHorizontal(images []image.Image, opts *CombineOptions) *image.NRGBA {
	// Calculate total width and max height
	totalWidth := 0
	maxHeight := 0

	for i, img := range images {
		bounds := img.Bounds()
		totalWidth += bounds.Dx()
		if bounds.Dy() > maxHeight {
			maxHeight = bounds.Dy()
		}
		if i > 0 {
			totalWidth += opts.Gap
		}
	}

	// Create canvas
	canvas := image.NewNRGBA(image.Rect(0, 0, totalWidth, maxHeight))
	fillBackground(canvas, opts.Background)

	// Draw images
	x := 0
	for i, img := range images {
		bounds := img.Bounds()
		y := calculateAlignment(maxHeight, bounds.Dy(), opts.Align)
		draw.Draw(canvas, image.Rect(x, y, x+bounds.Dx(), y+bounds.Dy()), img, bounds.Min, draw.Over)
		x += bounds.Dx()
		if i < len(images)-1 {
			x += opts.Gap
		}
	}

	return canvas
}

func combineVertical(images []image.Image, opts *CombineOptions) *image.NRGBA {
	// Calculate max width and total height
	maxWidth := 0
	totalHeight := 0

	for i, img := range images {
		bounds := img.Bounds()
		if bounds.Dx() > maxWidth {
			maxWidth = bounds.Dx()
		}
		totalHeight += bounds.Dy()
		if i > 0 {
			totalHeight += opts.Gap
		}
	}

	// Create canvas
	canvas := image.NewNRGBA(image.Rect(0, 0, maxWidth, totalHeight))
	fillBackground(canvas, opts.Background)

	// Draw images
	y := 0
	for i, img := range images {
		bounds := img.Bounds()
		x := calculateAlignment(maxWidth, bounds.Dx(), opts.Align)
		draw.Draw(canvas, image.Rect(x, y, x+bounds.Dx(), y+bounds.Dy()), img, bounds.Min, draw.Over)
		y += bounds.Dy()
		if i < len(images)-1 {
			y += opts.Gap
		}
	}

	return canvas
}

func combineGrid(images []image.Image, opts *CombineOptions) *image.NRGBA {
	// Calculate columns if not specified
	cols := opts.Columns
	if cols <= 0 {
		// Auto-calculate: aim for roughly square grid
		cols = int(float64(len(images))*0.5 + 0.5)
		if cols < 1 {
			cols = 1
		}
	}

	// Calculate rows
	rows := (len(images) + cols - 1) / cols

	// Find max cell dimensions
	maxCellWidth := 0
	maxCellHeight := 0
	for _, img := range images {
		bounds := img.Bounds()
		if bounds.Dx() > maxCellWidth {
			maxCellWidth = bounds.Dx()
		}
		if bounds.Dy() > maxCellHeight {
			maxCellHeight = bounds.Dy()
		}
	}

	// Calculate canvas size
	totalWidth := cols*maxCellWidth + (cols-1)*opts.Gap
	totalHeight := rows*maxCellHeight + (rows-1)*opts.Gap

	// Create canvas
	canvas := image.NewNRGBA(image.Rect(0, 0, totalWidth, totalHeight))
	fillBackground(canvas, opts.Background)

	// Draw images in grid
	for i, img := range images {
		row := i / cols
		col := i % cols

		bounds := img.Bounds()
		cellX := col * (maxCellWidth + opts.Gap)
		cellY := row * (maxCellHeight + opts.Gap)

		// Center within cell
		x := cellX + (maxCellWidth-bounds.Dx())/2
		y := cellY + (maxCellHeight-bounds.Dy())/2

		draw.Draw(canvas, image.Rect(x, y, x+bounds.Dx(), y+bounds.Dy()), img, bounds.Min, draw.Over)
	}

	return canvas
}

func calculateAlignment(containerSize, itemSize int, align string) int {
	switch align {
	case "start":
		return 0
	case "end":
		return containerSize - itemSize
	case "center", "":
		return (containerSize - itemSize) / 2
	default:
		return (containerSize - itemSize) / 2
	}
}

func fillBackground(canvas *image.NRGBA, bg string) {
	var bgColor color.NRGBA

	switch strings.ToLower(bg) {
	case "transparent", "":
		bgColor = color.NRGBA{R: 0, G: 0, B: 0, A: 0}
	case "white":
		bgColor = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	case "black":
		bgColor = color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	default:
		// Try to parse as hex
		if c, err := parseColor(bg); err == nil {
			bgColor = color.NRGBA{R: c.R, G: c.G, B: c.B, A: c.A}
		} else {
			bgColor = color.NRGBA{R: 0, G: 0, B: 0, A: 0}
		}
	}

	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)
}
