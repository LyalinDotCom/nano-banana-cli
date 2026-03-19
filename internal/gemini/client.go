package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "golang.org/x/image/webp"
)

const apiBaseURL = "https://generativelanguage.googleapis.com/v1beta"

type Client struct {
	httpClient *http.Client
	apiKey     string
	model      ValidatedModel
	timeout    time.Duration
}

type GenerateOptions struct {
	AspectRatio     string
	ImageSize       string
	Count           int
	InputPaths      []string
	GroundWeb       bool
	GroundImage     bool
	IncludeThoughts bool
	ThinkingLevel   string
	History         *ConversationHistory
}

type GeneratedImage struct {
	Data             []byte
	MimeType         string
	Width            int
	Height           int
	Thought          bool
	ThoughtSignature string
}

type TextPart struct {
	Text             string
	Thought          bool
	ThoughtSignature string
}

type GroundingSource struct {
	URI      string `json:"uri,omitempty"`
	ImageURI string `json:"image_uri,omitempty"`
	Title    string `json:"title,omitempty"`
}

type SearchEntryPoint struct {
	RenderedContent string `json:"rendered_content,omitempty"`
}

type GroundingMetadata struct {
	GroundingChunks    []GroundingSource `json:"grounding_chunks,omitempty"`
	WebSearchQueries   []string          `json:"web_search_queries,omitempty"`
	ImageSearchQueries []string          `json:"image_search_queries,omitempty"`
	SearchEntryPoint   *SearchEntryPoint `json:"search_entry_point,omitempty"`
}

type GenerateResult struct {
	Model     string
	Images    []*GeneratedImage
	Thoughts  []*GeneratedImage
	Texts     []TextPart
	Grounding *GroundingMetadata
	History   *ConversationHistory
}

type ConversationHistory struct {
	Model    string        `json:"model"`
	Contents []*apiContent `json:"contents"`
}

type apiGenerateContentRequest struct {
	Contents         []*apiContent        `json:"contents"`
	Tools            []apiTool            `json:"tools,omitempty"`
	GenerationConfig *apiGenerationConfig `json:"generationConfig,omitempty"`
}

type apiGenerateContentResponse struct {
	Candidates []apiCandidate `json:"candidates"`
	Error      *apiErrorBody  `json:"error,omitempty"`
}

type apiErrorBody struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

type apiCandidate struct {
	Content           *apiContent           `json:"content"`
	GroundingMetadata *apiGroundingMetadata `json:"groundingMetadata,omitempty"`
}

type apiContent struct {
	Role  string     `json:"role,omitempty"`
	Parts []*apiPart `json:"parts,omitempty"`
}

type apiPart struct {
	Text             string   `json:"text,omitempty"`
	InlineData       *apiBlob `json:"inline_data,omitempty"`
	Thought          bool     `json:"thought,omitempty"`
	ThoughtSignature string   `json:"thought_signature,omitempty"`
}

type apiBlob struct {
	MIMEType string `json:"mime_type,omitempty"`
	Data     []byte `json:"data,omitempty"`
}

type apiTool struct {
	GoogleSearch *apiGoogleSearch `json:"google_search,omitempty"`
}

type apiGoogleSearch struct {
	SearchTypes *apiSearchTypes `json:"searchTypes,omitempty"`
}

type apiSearchTypes struct {
	WebSearch   map[string]any `json:"webSearch,omitempty"`
	ImageSearch map[string]any `json:"imageSearch,omitempty"`
}

type apiGenerationConfig struct {
	ResponseModalities []string           `json:"responseModalities,omitempty"`
	ImageConfig        *apiImageConfig    `json:"imageConfig,omitempty"`
	ThinkingConfig     *apiThinkingConfig `json:"thinkingConfig,omitempty"`
}

type apiImageConfig struct {
	AspectRatio string `json:"aspectRatio,omitempty"`
	ImageSize   string `json:"imageSize,omitempty"`
}

type apiThinkingConfig struct {
	ThinkingLevel   string `json:"thinkingLevel,omitempty"`
	IncludeThoughts bool   `json:"includeThoughts,omitempty"`
}

type apiGroundingMetadata struct {
	GroundingChunks    []apiGroundingChunk  `json:"groundingChunks,omitempty"`
	WebSearchQueries   []string             `json:"webSearchQueries,omitempty"`
	ImageSearchQueries []string             `json:"imageSearchQueries,omitempty"`
	SearchEntryPoint   *apiSearchEntryPoint `json:"searchEntryPoint,omitempty"`
}

type apiGroundingChunk struct {
	Web   *apiGroundingChunkWeb   `json:"web,omitempty"`
	Image *apiGroundingChunkImage `json:"image,omitempty"`
}

type apiGroundingChunkWeb struct {
	URI   string `json:"uri,omitempty"`
	Title string `json:"title,omitempty"`
}

type apiGroundingChunkImage struct {
	URI      string `json:"uri,omitempty"`
	ImageURI string `json:"image_uri,omitempty"`
	Title    string `json:"title,omitempty"`
}

type apiSearchEntryPoint struct {
	RenderedContent string `json:"renderedContent,omitempty"`
}

func NewClient(apiKey, model string, timeout time.Duration) (*Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("API key is required")
	}
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}

	return &Client{
		httpClient: &http.Client{Timeout: timeout},
		apiKey:     apiKey,
		model:      ResolveModel(model),
		timeout:    timeout,
	}, nil
}

func (c *Client) Model() ValidatedModel {
	return c.model
}

func (c *Client) Generate(ctx context.Context, prompt string, opts *GenerateOptions) (*GenerateResult, error) {
	if opts == nil {
		opts = &GenerateOptions{Count: 1}
	}
	if opts.Count <= 0 {
		opts.Count = 1
	}

	if err := c.validateOptions(opts); err != nil {
		return nil, err
	}

	userContent, err := c.buildUserContent(prompt, opts.InputPaths)
	if err != nil {
		return nil, err
	}

	var (
		allImages   []*GeneratedImage
		allThoughts []*GeneratedImage
		allTexts    []TextPart
		lastGround  *GroundingMetadata
		lastHistory *ConversationHistory
	)

	for i := 0; i < opts.Count; i++ {
		history := opts.History
		if history != nil {
			history = history.Clone()
		}

		contents := []*apiContent{}
		if history != nil {
			contents = append(contents, history.Contents...)
		}
		contents = append(contents, cloneContent(userContent))

		reqBody := &apiGenerateContentRequest{
			Contents:         contents,
			GenerationConfig: c.buildGenerationConfig(opts),
			Tools:            c.buildTools(opts),
		}

		resp, err := c.callGenerateContent(ctx, reqBody)
		if err != nil {
			return nil, err
		}

		result, err := c.extractResult(resp)
		if err != nil {
			return nil, err
		}

		history = &ConversationHistory{
			Model:    c.model.Spec.ID,
			Contents: append(contents, resp.firstContent()...),
		}

		allImages = append(allImages, result.Images...)
		allThoughts = append(allThoughts, result.Thoughts...)
		allTexts = append(allTexts, result.Texts...)
		lastGround = result.Grounding
		lastHistory = history
	}

	return &GenerateResult{
		Model:     c.model.Spec.ID,
		Images:    allImages,
		Thoughts:  allThoughts,
		Texts:     allTexts,
		Grounding: lastGround,
		History:   lastHistory,
	}, nil
}

func (c *Client) SaveImage(img *GeneratedImage, outputPath string) error {
	dir := filepath.Dir(outputPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	if err := os.WriteFile(outputPath, img.Data, 0644); err != nil {
		return fmt.Errorf("failed to write image: %w", err)
	}

	return nil
}

func (c *Client) SaveHistory(history *ConversationHistory, path string) error {
	if history == nil {
		return nil
	}
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create history directory: %w", err)
		}
	}
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write history file: %w", err)
	}
	return nil
}

func LoadHistory(path string) (*ConversationHistory, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read history file: %w", err)
	}

	var history ConversationHistory
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, fmt.Errorf("failed to parse history file: %w", err)
	}
	if len(history.Contents) == 0 {
		return nil, fmt.Errorf("history file %s does not contain contents", path)
	}
	return &history, nil
}

func (h *ConversationHistory) Clone() *ConversationHistory {
	if h == nil {
		return nil
	}
	cloned := &ConversationHistory{
		Model:    h.Model,
		Contents: make([]*apiContent, 0, len(h.Contents)),
	}
	for _, content := range h.Contents {
		cloned.Contents = append(cloned.Contents, cloneContent(content))
	}
	return cloned
}

func (c *Client) validateOptions(opts *GenerateOptions) error {
	if err := ValidateAspectRatio(c.model.Spec, opts.AspectRatio); err != nil {
		return &GeminiError{Code: ErrInvalidInput, Message: err.Error()}
	}

	size := DefaultImageSize(c.model.Spec, opts.ImageSize)
	if err := ValidateImageSize(c.model.Spec, size); err != nil {
		return &GeminiError{Code: ErrInvalidInput, Message: err.Error()}
	}

	if len(opts.InputPaths) > c.model.Spec.MaxInputImages {
		return &GeminiError{
			Code:    ErrInvalidInput,
			Message: fmt.Sprintf("model %s supports up to %d input images", c.model.Spec.ID, c.model.Spec.MaxInputImages),
		}
	}

	if opts.GroundWeb && !c.model.Spec.SupportsGrounding {
		return &GeminiError{Code: ErrInvalidInput, Message: fmt.Sprintf("model %s does not support grounding", c.model.Spec.ID)}
	}
	if opts.GroundImage && !c.model.Spec.SupportsImageSearch {
		return &GeminiError{Code: ErrInvalidInput, Message: fmt.Sprintf("model %s does not support Google Image Search grounding", c.model.Spec.ID)}
	}
	if opts.IncludeThoughts && !c.model.Spec.SupportsThinking {
		return &GeminiError{Code: ErrInvalidInput, Message: fmt.Sprintf("model %s does not support thought output", c.model.Spec.ID)}
	}
	if strings.TrimSpace(opts.ThinkingLevel) != "" && !c.model.Spec.SupportsThinkingLevel {
		return &GeminiError{Code: ErrInvalidInput, Message: fmt.Sprintf("model %s does not support configurable thinking levels", c.model.Spec.ID)}
	}
	if opts.GroundImage && !opts.GroundWeb {
		opts.GroundWeb = true
	}
	if opts.History != nil && opts.History.Model != "" && opts.History.Model != c.model.Spec.ID {
		return &GeminiError{
			Code:    ErrInvalidInput,
			Message: fmt.Sprintf("history model %s does not match requested model %s", opts.History.Model, c.model.Spec.ID),
		}
	}
	return nil
}

func (c *Client) buildUserContent(prompt string, inputPaths []string) (*apiContent, error) {
	parts := []*apiPart{{Text: prompt}}
	for _, inputPath := range inputPaths {
		data, err := os.ReadFile(inputPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read input image %s: %w", inputPath, err)
		}
		parts = append(parts, &apiPart{
			InlineData: &apiBlob{
				MIMEType: detectMimeType(inputPath),
				Data:     data,
			},
		})
	}
	return &apiContent{
		Role:  "user",
		Parts: parts,
	}, nil
}

func (c *Client) buildGenerationConfig(opts *GenerateOptions) *apiGenerationConfig {
	cfg := &apiGenerationConfig{
		ResponseModalities: []string{"TEXT", "IMAGE"},
	}

	if opts.AspectRatio != "" || opts.ImageSize != "" {
		cfg.ImageConfig = &apiImageConfig{
			AspectRatio: opts.AspectRatio,
			ImageSize:   DefaultImageSize(c.model.Spec, opts.ImageSize),
		}
	}

	if opts.IncludeThoughts || strings.TrimSpace(opts.ThinkingLevel) != "" {
		cfg.ThinkingConfig = &apiThinkingConfig{
			IncludeThoughts: opts.IncludeThoughts,
			ThinkingLevel:   normalizeThinkingLevel(opts.ThinkingLevel),
		}
	}

	return cfg
}

func (c *Client) buildTools(opts *GenerateOptions) []apiTool {
	if !opts.GroundWeb && !opts.GroundImage {
		return nil
	}

	search := &apiGoogleSearch{}
	if opts.GroundImage {
		search.SearchTypes = &apiSearchTypes{
			WebSearch:   map[string]any{},
			ImageSearch: map[string]any{},
		}
	}

	return []apiTool{{GoogleSearch: search}}
}

func (c *Client) callGenerateContent(ctx context.Context, payload *apiGenerateContentRequest) (*apiGenerateContentResponse, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", apiBaseURL, c.model.Spec.ID, c.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, c.classifyError(err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read API response: %w", err)
	}

	var parsed apiGenerateContentResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		if resp.StatusCode >= 400 {
			return nil, c.classifyError(fmt.Errorf("api error: %s", string(respBody)))
		}
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	if resp.StatusCode >= 400 {
		if parsed.Error != nil {
			return nil, c.classifyError(fmt.Errorf("%s (%s)", parsed.Error.Message, parsed.Error.Status))
		}
		return nil, c.classifyError(fmt.Errorf("api error: %s", resp.Status))
	}

	return &parsed, nil
}

func (c *Client) extractResult(resp *apiGenerateContentResponse) (*GenerateResult, error) {
	if resp == nil || len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return nil, &GeminiError{Code: ErrNoImageGenerated, Message: "empty response from API"}
	}

	candidate := resp.Candidates[0]
	result := &GenerateResult{Model: c.model.Spec.ID}

	for _, part := range candidate.Content.Parts {
		if part == nil {
			continue
		}
		if part.InlineData != nil && strings.HasPrefix(part.InlineData.MIMEType, "image/") {
			img := &GeneratedImage{
				Data:             part.InlineData.Data,
				MimeType:         part.InlineData.MIMEType,
				Thought:          part.Thought,
				ThoughtSignature: part.ThoughtSignature,
			}
			img.Width, img.Height = imageDimensions(img.Data)
			if part.Thought {
				result.Thoughts = append(result.Thoughts, img)
			} else {
				result.Images = append(result.Images, img)
			}
			continue
		}
		if part.Text != "" {
			result.Texts = append(result.Texts, TextPart{
				Text:             part.Text,
				Thought:          part.Thought,
				ThoughtSignature: part.ThoughtSignature,
			})
		}
	}

	if len(result.Images) == 0 {
		return nil, &GeminiError{Code: ErrNoImageGenerated, Message: "no final images were returned by the API"}
	}

	if candidate.GroundingMetadata != nil {
		result.Grounding = translateGrounding(candidate.GroundingMetadata)
	}

	return result, nil
}

func (r *apiGenerateContentResponse) firstContent() []*apiContent {
	if r == nil || len(r.Candidates) == 0 || r.Candidates[0].Content == nil {
		return nil
	}
	return []*apiContent{cloneContent(r.Candidates[0].Content)}
}

func translateGrounding(meta *apiGroundingMetadata) *GroundingMetadata {
	if meta == nil {
		return nil
	}

	out := &GroundingMetadata{
		WebSearchQueries:   meta.WebSearchQueries,
		ImageSearchQueries: meta.ImageSearchQueries,
	}

	if meta.SearchEntryPoint != nil {
		out.SearchEntryPoint = &SearchEntryPoint{RenderedContent: meta.SearchEntryPoint.RenderedContent}
	}

	for _, chunk := range meta.GroundingChunks {
		if chunk.Web != nil {
			out.GroundingChunks = append(out.GroundingChunks, GroundingSource{
				URI:   chunk.Web.URI,
				Title: chunk.Web.Title,
			})
		}
		if chunk.Image != nil {
			out.GroundingChunks = append(out.GroundingChunks, GroundingSource{
				URI:      chunk.Image.URI,
				ImageURI: chunk.Image.ImageURI,
				Title:    chunk.Image.Title,
			})
		}
	}

	return out
}

func cloneContent(content *apiContent) *apiContent {
	if content == nil {
		return nil
	}
	cloned := &apiContent{
		Role:  content.Role,
		Parts: make([]*apiPart, 0, len(content.Parts)),
	}
	for _, part := range content.Parts {
		if part == nil {
			continue
		}
		cp := &apiPart{
			Text:             part.Text,
			Thought:          part.Thought,
			ThoughtSignature: part.ThoughtSignature,
		}
		if part.InlineData != nil {
			data := make([]byte, len(part.InlineData.Data))
			copy(data, part.InlineData.Data)
			cp.InlineData = &apiBlob{
				MIMEType: part.InlineData.MIMEType,
				Data:     data,
			}
		}
		cloned.Parts = append(cloned.Parts, cp)
	}
	return cloned
}

func imageDimensions(data []byte) (int, int) {
	if len(data) == 0 {
		return 0, 0
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return 0, 0
	}
	return cfg.Width, cfg.Height
}

func normalizeThinkingLevel(level string) string {
	level = strings.TrimSpace(strings.ToLower(level))
	switch level {
	case "":
		return ""
	case "minimal":
		return "minimal"
	case "high":
		return "high"
	default:
		return level
	}
}

const (
	ErrInvalidAPIKey    = "INVALID_API_KEY"
	ErrQuotaExceeded    = "QUOTA_EXCEEDED"
	ErrRateLimited      = "RATE_LIMITED"
	ErrInvalidInput     = "INVALID_INPUT"
	ErrSafetyBlocked    = "SAFETY_BLOCKED"
	ErrAPIError         = "API_ERROR"
	ErrNoImageGenerated = "NO_IMAGE_GENERATED"
)

type GeminiError struct {
	Code    string
	Message string
}

func (e *GeminiError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (c *Client) classifyError(err error) error {
	errStr := strings.ToLower(err.Error())
	switch {
	case strings.Contains(errStr, "api key"), strings.Contains(errStr, "unauthorized"), strings.Contains(errStr, "permission_denied"), strings.Contains(errStr, "401"):
		return &GeminiError{Code: ErrInvalidAPIKey, Message: "invalid or unauthorized API key"}
	case strings.Contains(errStr, "quota"), strings.Contains(errStr, "resource_exhausted"), strings.Contains(errStr, "429"):
		return &GeminiError{Code: ErrQuotaExceeded, Message: "API quota exceeded"}
	case strings.Contains(errStr, "rate limit"):
		return &GeminiError{Code: ErrRateLimited, Message: "rate limit exceeded, please wait"}
	case strings.Contains(errStr, "safety"), strings.Contains(errStr, "blocked"):
		return &GeminiError{Code: ErrSafetyBlocked, Message: "content blocked by safety filters"}
	case strings.Contains(errStr, "invalid"), strings.Contains(errStr, "unsupported"), strings.Contains(errStr, "400"):
		return &GeminiError{Code: ErrInvalidInput, Message: strings.TrimSpace(err.Error())}
	default:
		return &GeminiError{Code: ErrAPIError, Message: strings.TrimSpace(err.Error())}
	}
}

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
