package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/lyalindotcom/nano-banana-cli/internal/gemini"
	"github.com/lyalindotcom/nano-banana-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	outputPath      string
	inputPaths      []string
	promptFile      string
	count           int
	aspectRatio     string
	imageSize       string
	noOverwrite     bool
	thinkingLevel   string
	includeThoughts bool
	thoughtsDir     string
	groundWeb       bool
	groundImage     bool
	historyIn       string
	historyOut      string
)

var generateCmd = &cobra.Command{
	Use:   "generate [prompt]",
	Short: "Generate or edit images with Gemini image models",
	Long: `Generate images from text prompts using Gemini's native image generation.

This command supports:
  - Text-to-image generation
  - Image editing with one or more reference images
  - Scriptable multi-turn workflows via --history-in/--history-out
  - Grounding with Google Search on supported models
  - Optional thought output for Gemini 3 image models

PROMPT INPUT:
  - As arguments: nanobanana generate "your prompt here"
  - From file: nanobanana generate --prompt-file prompt.txt -o out.png
  - From stdin: echo "prompt" | nanobanana generate - -o out.png
  - From stdin: nanobanana generate -o out.png < prompt.txt

MODELS:
  - banana2 (default): Gemini 3.1 Flash Image Preview
  - banana / 2.5: Gemini 2.5 Flash Image
  - pro: Gemini 3 Pro Image Preview

ASPECT RATIOS:
  Standard: 1:1, 3:2, 2:3, 3:4, 4:3, 4:5, 5:4, 9:16, 16:9, 21:9
  Gemini 3.1 only: 1:4, 4:1, 1:8, 8:1

IMAGE SIZES:
  Gemini 3.1: 512, 1K, 2K, 4K
  Gemini 3 Pro: 1K, 2K, 4K
  Gemini 2.5: fixed 1K behavior

EXAMPLES:
  # Generate a simple image
  nanobanana generate "a sunset over mountains" -o sunset.png

  # Generate with a specific aspect ratio
  nanobanana generate "landscape photo" --aspect-ratio 16:9 -o landscape.png

  # Generate with Pro at 4K
  nanobanana generate "detailed portrait" -m pro --image-size 4K -o portrait.png

  # Edit an existing image
  nanobanana generate "add sunglasses" -i face.png -o face-sunglasses.png

  # Use multiple reference images
  nanobanana generate "office group photo of these people" -i person1.png -i person2.png -o group.png

  # Ground with Google Search
  nanobanana generate "stylish graphic of today's weather in NYC" --ground-web -o weather.png

  # Ground with web + image search (Gemini 3.1 only)
  nanobanana generate "a detailed painting of a Timareta butterfly" --ground-image -o butterfly.png

  # Save and resume scripted history
  nanobanana generate "Create a colorful infographic about photosynthesis" -o photo.png --history-out photo-history.json
  nanobanana generate "Translate the infographic to Spanish and keep everything else the same" -o photo-es.png --history-in photo-history.json --history-out photo-history.json`,
	Args: cobra.MinimumNArgs(0),
	RunE: runGenerate,
}

func init() {
	generateCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output file path (required)")
	generateCmd.Flags().StringArrayVarP(&inputPaths, "input", "i", nil, "Input/reference image (repeat up to model limit)")
	generateCmd.Flags().StringVarP(&promptFile, "prompt-file", "p", "", "Read prompt from file (supports multi-line)")
	generateCmd.Flags().IntVarP(&count, "count", "c", 1, "Number of images to generate (1-10)")
	generateCmd.Flags().StringVar(&aspectRatio, "aspect-ratio", "1:1", "Aspect ratio")
	generateCmd.Flags().StringVar(&imageSize, "image-size", "", "Image size: 512, 1K, 2K, 4K")
	generateCmd.Flags().BoolVar(&noOverwrite, "no-overwrite", false, "Fail if output file exists")
	generateCmd.Flags().StringVar(&thinkingLevel, "thinking-level", "", "Thinking level: minimal, high (Gemini 3.1 only)")
	generateCmd.Flags().BoolVar(&includeThoughts, "include-thoughts", false, "Include thought parts in the response")
	generateCmd.Flags().StringVar(&thoughtsDir, "thoughts-dir", "", "Directory for saving thought images when --include-thoughts is used")
	generateCmd.Flags().BoolVar(&groundWeb, "ground-web", false, "Enable grounding with Google Search")
	generateCmd.Flags().BoolVar(&groundImage, "ground-image", false, "Enable Google Image Search grounding (Gemini 3.1 only)")
	generateCmd.Flags().StringVar(&historyIn, "history-in", "", "Resume a scripted image conversation from a JSON history file")
	generateCmd.Flags().StringVar(&historyOut, "history-out", "", "Write updated conversation history to a JSON file")

	generateCmd.MarkFlagRequired("output")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	f := GetFormatter()
	startTime := time.Now()

	prompt, err := getPrompt(args)
	if err != nil {
		f.Error("generate", "PROMPT_ERROR", err.Error(), "Provide prompt as argument, --prompt-file, or stdin")
		return err
	}
	if strings.TrimSpace(prompt) == "" {
		f.Error("generate", "EMPTY_PROMPT", "Prompt cannot be empty", "Provide prompt as argument, --prompt-file, or stdin")
		return fmt.Errorf("empty prompt")
	}

	apiKey := GetAPIKey()
	if apiKey == "" {
		f.Error("generate", "MISSING_API_KEY", "No API key provided", "Set GEMINI_API_KEY environment variable or use --api-key flag")
		return fmt.Errorf("missing API key")
	}

	if count < 1 || count > 10 {
		f.Error("generate", "INVALID_COUNT", "Count must be between 1 and 10", "")
		return fmt.Errorf("invalid count")
	}

	if aspectRatio != "" && !gemini.IsValidAspectRatio(aspectRatio) {
		f.Error("generate", "INVALID_ASPECT_RATIO", fmt.Sprintf("Invalid aspect ratio: %s", aspectRatio), fmt.Sprintf("Valid ratios: %s", strings.Join(gemini.ListAllAspectRatios(), ", ")))
		return fmt.Errorf("invalid aspect ratio")
	}

	selectedImageSize := strings.TrimSpace(imageSize)
	if !gemini.IsValidImageSize(selectedImageSize) {
		f.Error("generate", "INVALID_IMAGE_SIZE", fmt.Sprintf("Invalid image size: %s", selectedImageSize), "Valid sizes: 512, 1K, 2K, 4K")
		return fmt.Errorf("invalid image size")
	}

	if strings.TrimSpace(thinkingLevel) != "" && !slices.Contains([]string{"minimal", "high"}, strings.ToLower(thinkingLevel)) {
		f.Error("generate", "INVALID_THINKING_LEVEL", fmt.Sprintf("Invalid thinking level: %s", thinkingLevel), "Valid levels: minimal, high")
		return fmt.Errorf("invalid thinking level")
	}

	if (historyIn != "" || historyOut != "") && count != 1 {
		f.Error("generate", "INVALID_HISTORY_USAGE", "History files require --count 1", "Use a single scripted turn per history file update")
		return fmt.Errorf("history requires count 1")
	}

	if noOverwrite {
		if _, err := os.Stat(outputPath); err == nil {
			f.Error("generate", "FILE_EXISTS", fmt.Sprintf("Output file already exists: %s", outputPath), "Use a different output path or remove --no-overwrite flag")
			return fmt.Errorf("file exists")
		}
	}

	for _, inputPath := range inputPaths {
		if _, err := os.Stat(inputPath); os.IsNotExist(err) {
			f.Error("generate", "FILE_NOT_FOUND", fmt.Sprintf("Input file not found: %s", inputPath), "")
			return err
		}
	}

	var history *gemini.ConversationHistory
	if historyIn != "" {
		history, err = gemini.LoadHistory(historyIn)
		if err != nil {
			f.Error("generate", "INVALID_HISTORY", err.Error(), "")
			return err
		}
	}

	client, err := gemini.NewClient(apiKey, GetModel(), 3*time.Minute)
	if err != nil {
		f.Error("generate", "CLIENT_ERROR", err.Error(), "Check your API key")
		return err
	}
	modelInfo := client.Model()

	options := &gemini.GenerateOptions{
		AspectRatio:     aspectRatio,
		ImageSize:       selectedImageSize,
		Count:           count,
		InputPaths:      inputPaths,
		GroundWeb:       groundWeb,
		GroundImage:     groundImage,
		IncludeThoughts: includeThoughts,
		ThinkingLevel:   thinkingLevel,
		History:         history,
	}

	f.Progress("Generating image with %s...", modelInfo.Spec.ID)

	result, err := client.Generate(context.Background(), prompt, options)
	if err != nil {
		if geminiErr, ok := err.(*gemini.GeminiError); ok {
			hint := ""
			switch geminiErr.Code {
			case gemini.ErrInvalidAPIKey:
				hint = "Check your API key at https://aistudio.google.com/apikey"
			case gemini.ErrQuotaExceeded:
				hint = "Wait before retrying or check your quota"
			case gemini.ErrSafetyBlocked:
				hint = "Try rephrasing your prompt"
			}
			f.Error("generate", geminiErr.Code, geminiErr.Message, hint)
		} else {
			f.Error("generate", "GENERATION_FAILED", err.Error(), "")
		}
		return err
	}

	var imageResults []output.ImageResult
	for i, img := range result.Images {
		savePath := outputPath
		if len(result.Images) > 1 {
			ext := filepath.Ext(outputPath)
			base := strings.TrimSuffix(outputPath, ext)
			savePath = fmt.Sprintf("%s_%d%s", base, i+1, ext)
		}

		if err := client.SaveImage(img, savePath); err != nil {
			f.Error("generate", "SAVE_FAILED", err.Error(), "")
			return err
		}

		f.ImageSaved(savePath, img.Width, img.Height)
		imageResults = append(imageResults, output.ImageResult{
			Path:   savePath,
			Format: strings.TrimPrefix(img.MimeType, "image/"),
			Size:   &output.ImageSize{Width: img.Width, Height: img.Height},
		})
	}

	var thoughtResults []output.ImageResult
	if includeThoughts && thoughtsDir != "" {
		for i, thought := range result.Thoughts {
			thoughtPath := filepath.Join(thoughtsDir, fmt.Sprintf("thought_%02d%s", i+1, extensionForMime(thought.MimeType)))
			if err := client.SaveImage(thought, thoughtPath); err != nil {
				f.Error("generate", "THOUGHT_SAVE_FAILED", err.Error(), "")
				return err
			}
			thoughtResults = append(thoughtResults, output.ImageResult{
				Path:   thoughtPath,
				Format: strings.TrimPrefix(thought.MimeType, "image/"),
				Size:   &output.ImageSize{Width: thought.Width, Height: thought.Height},
			})
		}
	}

	if historyOut != "" {
		if err := client.SaveHistory(result.History, historyOut); err != nil {
			f.Error("generate", "HISTORY_SAVE_FAILED", err.Error(), "")
			return err
		}
		f.Info("Saved history: %s", historyOut)
	}

	for _, text := range result.Texts {
		if !text.Thought && strings.TrimSpace(text.Text) != "" {
			f.Info(strings.TrimSpace(text.Text))
		}
	}

	if result.Grounding != nil && len(result.Grounding.GroundingChunks) > 0 {
		f.Info("Grounding sources:")
		for _, src := range dedupeGroundingSources(result.Grounding.GroundingChunks) {
			if src.URI != "" {
				f.Info("  %s", src.URI)
			}
		}
	}

	timing := &output.Timing{TotalMs: time.Since(startTime).Milliseconds()}
	data := map[string]interface{}{
		"prompt":             prompt,
		"model":              result.Model,
		"input_images":       inputPaths,
		"aspect_ratio":       aspectRatio,
		"image_size":         gemini.DefaultImageSize(modelInfo.Spec, selectedImageSize),
		"images":             imageResults,
		"grounding_metadata": result.Grounding,
	}
	if len(result.Texts) > 0 {
		data["parts"] = result.Texts
	}
	if len(thoughtResults) > 0 {
		data["thought_images"] = thoughtResults
	}
	if historyOut != "" {
		data["history_file"] = historyOut
	}

	f.Success("generate", data, timing)
	return nil
}

func getPrompt(args []string) (string, error) {
	if promptFile != "" {
		data, err := os.ReadFile(promptFile)
		if err != nil {
			return "", fmt.Errorf("failed to read prompt file: %w", err)
		}
		return string(data), nil
	}

	if len(args) == 1 && args[0] == "-" {
		return readStdin()
	}

	if len(args) == 0 {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			return readStdin()
		}
		return "", fmt.Errorf("no prompt provided")
	}

	return strings.Join(args, " "), nil
}

func readStdin() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	var builder strings.Builder

	for {
		line, err := reader.ReadString('\n')
		builder.WriteString(line)
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read stdin: %w", err)
		}
	}

	return strings.TrimSpace(builder.String()), nil
}

func extensionForMime(mime string) string {
	switch mime {
	case "image/jpeg":
		return ".jpg"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ".png"
	}
}

func dedupeGroundingSources(in []gemini.GroundingSource) []gemini.GroundingSource {
	seen := map[string]bool{}
	var out []gemini.GroundingSource
	for _, src := range in {
		key := src.URI + "|" + src.ImageURI
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, src)
	}
	return out
}
