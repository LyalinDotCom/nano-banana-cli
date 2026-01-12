package output

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
)

// Response is the standard output structure
type Response struct {
	Success bool        `json:"success"`
	Command string      `json:"command,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Timing  *Timing     `json:"timing,omitempty"`
}

// ErrorInfo contains error details
type ErrorInfo struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
	Hint    string            `json:"hint,omitempty"`
}

// Timing contains performance metrics
type Timing struct {
	APICallMs int64 `json:"api_call_ms,omitempty"`
	TotalMs   int64 `json:"total_ms,omitempty"`
}

// ImageResult represents a generated or processed image
type ImageResult struct {
	Path   string     `json:"path"`
	Size   *ImageSize `json:"size,omitempty"`
	Format string     `json:"format,omitempty"`
}

// ImageSize contains image dimensions
type ImageSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Formatter handles output formatting
type Formatter struct {
	JSONMode bool
	Quiet    bool
	NoColor  bool
}

// NewFormatter creates a new output formatter
func NewFormatter(jsonMode, quiet, noColor bool) *Formatter {
	if noColor {
		color.NoColor = true
	}
	return &Formatter{
		JSONMode: jsonMode,
		Quiet:    quiet,
		NoColor:  noColor,
	}
}

// Success outputs a success response
func (f *Formatter) Success(command string, data interface{}, timing *Timing) {
	if f.JSONMode {
		resp := Response{
			Success: true,
			Command: command,
			Data:    data,
			Timing:  timing,
		}
		f.outputJSON(resp)
		return
	}

	// Text mode
	if !f.Quiet {
		green := color.New(color.FgGreen).SprintFunc()
		fmt.Fprintf(os.Stdout, "%s %s\n", green("✓"), "Success")
	}
}

// Error outputs an error response
func (f *Formatter) Error(command string, code string, message string, hint string) {
	if f.JSONMode {
		resp := Response{
			Success: false,
			Command: command,
			Error: &ErrorInfo{
				Code:    code,
				Message: message,
				Hint:    hint,
			},
		}
		f.outputJSON(resp)
		return
	}

	// Text mode
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Fprintf(os.Stderr, "%s [%s] %s\n", red("Error:"), code, message)
	if hint != "" {
		fmt.Fprintf(os.Stderr, "%s %s\n", yellow("Hint:"), hint)
	}
}

// Info outputs an informational message (only in non-quiet, non-JSON mode)
func (f *Formatter) Info(format string, args ...interface{}) {
	if f.JSONMode || f.Quiet {
		return
	}
	fmt.Fprintf(os.Stdout, format+"\n", args...)
}

// Progress outputs a progress message
func (f *Formatter) Progress(format string, args ...interface{}) {
	if f.JSONMode || f.Quiet {
		return
	}
	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Fprintf(os.Stdout, "%s %s\n", cyan("→"), fmt.Sprintf(format, args...))
}

// ImageSaved outputs a message about a saved image
func (f *Formatter) ImageSaved(path string, width, height int) {
	if f.JSONMode || f.Quiet {
		return
	}
	green := color.New(color.FgGreen).SprintFunc()
	dim := color.New(color.Faint).SprintFunc()
	fmt.Fprintf(os.Stdout, "%s Saved: %s %s\n", green("✓"), path, dim(fmt.Sprintf("(%dx%d)", width, height)))
}

func (f *Formatter) outputJSON(v interface{}) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.Encode(v)
}

// MeasureDuration creates a timer and returns a function to get elapsed time
func MeasureDuration() func() time.Duration {
	start := time.Now()
	return func() time.Duration {
		return time.Since(start)
	}
}
