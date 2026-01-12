package gemini

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/genai"
)

// Client wraps the Google GenAI client for image generation
type Client struct {
	client  *genai.Client
	model   string
	timeout time.Duration
}

// ImageConfig contains configuration for image generation
type ImageConfig struct {
	AspectRatio string
	Resolution  string // "1K", "2K", "4K" (4K only for Pro model)
	Count       int    // Number of images to generate (1-10)
}

// GeneratedImage represents a generated image
type GeneratedImage struct {
	Data     []byte
	MimeType string
	Width    int
	Height   int
}

// NewClient creates a new Gemini client
func NewClient(apiKey, model string, timeout time.Duration) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &Client{
		client:  client,
		model:   model,
		timeout: timeout,
	}, nil
}

// GenerateImage generates an image from a text prompt
func (c *Client) GenerateImage(ctx context.Context, prompt string, config *ImageConfig) ([]*GeneratedImage, error) {
	if config == nil {
		config = &ImageConfig{
			AspectRatio: "1:1",
			Count:       1,
		}
	}

	// Build generation config with image output
	genConfig := &genai.GenerateContentConfig{
		ResponseModalities: []string{"IMAGE", "TEXT"},
	}

	// Note: ImageConfig is not directly available in the SDK
	// We need to include aspect ratio instructions in the prompt
	enhancedPrompt := prompt
	if config.AspectRatio != "" && config.AspectRatio != "1:1" {
		enhancedPrompt = fmt.Sprintf("%s. Generate this image with %s aspect ratio.", prompt, config.AspectRatio)
	}

	// Generate content
	result, err := c.client.Models.GenerateContent(ctx, c.model, genai.Text(enhancedPrompt), genConfig)
	if err != nil {
		return nil, c.classifyError(err)
	}

	// Extract images from response
	images, err := c.extractImages(result)
	if err != nil {
		return nil, err
	}

	if len(images) == 0 {
		return nil, fmt.Errorf("no images generated")
	}

	return images, nil
}

// EditImage edits an existing image based on a text prompt
func (c *Client) EditImage(ctx context.Context, imagePath string, prompt string, config *ImageConfig) ([]*GeneratedImage, error) {
	// Read the input image
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read input image: %w", err)
	}

	mimeType := detectMimeType(imagePath)

	// Build parts with image and text
	parts := []*genai.Part{
		{InlineData: &genai.Blob{
			MIMEType: mimeType,
			Data:     imageData,
		}},
		{Text: prompt},
	}

	contents := []*genai.Content{
		{Parts: parts},
	}

	// Build generation config
	genConfig := &genai.GenerateContentConfig{
		ResponseModalities: []string{"IMAGE", "TEXT"},
	}

	// Generate content
	result, err := c.client.Models.GenerateContent(ctx, c.model, contents, genConfig)
	if err != nil {
		return nil, c.classifyError(err)
	}

	// Extract images from response
	images, err := c.extractImages(result)
	if err != nil {
		return nil, err
	}

	if len(images) == 0 {
		return nil, fmt.Errorf("no images generated")
	}

	return images, nil
}

// extractImages extracts image data from the generation response
func (c *Client) extractImages(result *genai.GenerateContentResponse) ([]*GeneratedImage, error) {
	var images []*GeneratedImage

	if result == nil || len(result.Candidates) == 0 {
		return nil, fmt.Errorf("empty response from API")
	}

	for _, candidate := range result.Candidates {
		if candidate.Content == nil {
			continue
		}
		for _, part := range candidate.Content.Parts {
			if part.InlineData != nil && strings.HasPrefix(part.InlineData.MIMEType, "image/") {
				images = append(images, &GeneratedImage{
					Data:     part.InlineData.Data,
					MimeType: part.InlineData.MIMEType,
				})
			}
		}
	}

	return images, nil
}

// SaveImage saves a generated image to a file
func (c *Client) SaveImage(img *GeneratedImage, outputPath string) error {
	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// If data is base64 encoded string, decode it
	var data []byte
	if len(img.Data) > 0 {
		// Check if it's base64 encoded
		if decoded, err := base64.StdEncoding.DecodeString(string(img.Data)); err == nil {
			data = decoded
		} else {
			data = img.Data
		}
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write image: %w", err)
	}

	return nil
}

// Error codes
const (
	ErrInvalidAPIKey    = "INVALID_API_KEY"
	ErrQuotaExceeded    = "QUOTA_EXCEEDED"
	ErrRateLimited      = "RATE_LIMITED"
	ErrInvalidInput     = "INVALID_INPUT"
	ErrSafetyBlocked    = "SAFETY_BLOCKED"
	ErrAPIError         = "API_ERROR"
	ErrNoImageGenerated = "NO_IMAGE_GENERATED"
)

// GeminiError represents a classified error
type GeminiError struct {
	Code    string
	Message string
}

func (e *GeminiError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// classifyError converts API errors to classified errors
func (c *Client) classifyError(err error) error {
	errStr := err.Error()

	if strings.Contains(errStr, "API key") || strings.Contains(errStr, "unauthorized") || strings.Contains(errStr, "401") {
		return &GeminiError{Code: ErrInvalidAPIKey, Message: "Invalid or unauthorized API key"}
	}
	if strings.Contains(errStr, "quota") || strings.Contains(errStr, "429") {
		return &GeminiError{Code: ErrQuotaExceeded, Message: "API quota exceeded"}
	}
	if strings.Contains(errStr, "rate limit") {
		return &GeminiError{Code: ErrRateLimited, Message: "Rate limit exceeded, please wait"}
	}
	if strings.Contains(errStr, "safety") || strings.Contains(errStr, "blocked") {
		return &GeminiError{Code: ErrSafetyBlocked, Message: "Content blocked by safety filters"}
	}

	return &GeminiError{Code: ErrAPIError, Message: err.Error()}
}

// detectMimeType detects the MIME type from file extension
func detectMimeType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "image/png"
	}
}
