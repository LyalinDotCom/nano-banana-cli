package gemini

import (
	"path/filepath"
	"testing"
	"time"
)

func TestResolveModelAliases(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		wantID string
	}{
		{name: "default", input: "", wantID: ModelFlash31},
		{name: "banana2", input: "banana2", wantID: ModelFlash31},
		{name: "banana", input: "banana", wantID: ModelFlash25},
		{name: "pro", input: "pro", wantID: ModelPro},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ResolveModel(tc.input)
			if got.Spec.ID != tc.wantID {
				t.Fatalf("ResolveModel(%q) = %q, want %q", tc.input, got.Spec.ID, tc.wantID)
			}
		})
	}
}

func TestValidateOptionsByModel(t *testing.T) {
	tests := []struct {
		name    string
		model   string
		opts    GenerateOptions
		wantErr bool
	}{
		{
			name:    "flash25 rejects 2k",
			model:   "banana",
			opts:    GenerateOptions{ImageSize: "2K", Count: 1},
			wantErr: true,
		},
		{
			name:    "flash25 rejects image search",
			model:   "banana",
			opts:    GenerateOptions{GroundImage: true, Count: 1},
			wantErr: true,
		},
		{
			name:  "flash31 accepts 512 and wide ratio",
			model: "banana2",
			opts:  GenerateOptions{ImageSize: "512", AspectRatio: "1:8", Count: 1},
		},
		{
			name:    "pro rejects 512",
			model:   "pro",
			opts:    GenerateOptions{ImageSize: "512", Count: 1},
			wantErr: true,
		},
		{
			name:  "pro accepts 4k grounding",
			model: "pro",
			opts:  GenerateOptions{ImageSize: "4K", GroundWeb: true, Count: 1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client, err := NewClient("test-key", tc.model, time.Second)
			if err != nil {
				t.Fatalf("NewClient: %v", err)
			}
			err = client.validateOptions(&tc.opts)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateOptions() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestHistoryRoundTrip(t *testing.T) {
	client, err := NewClient("test-key", "banana2", time.Second)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	history := &ConversationHistory{
		Model: "gemini-3.1-flash-image-preview",
		Contents: []*apiContent{
			{
				Role: "user",
				Parts: []*apiPart{
					{Text: "make a poster"},
				},
			},
			{
				Role: "model",
				Parts: []*apiPart{
					{Text: "done", ThoughtSignature: "sig-a"},
					{InlineData: &apiBlob{MIMEType: "image/png", Data: []byte("abc")}, ThoughtSignature: "sig-b"},
				},
			},
		},
	}

	path := filepath.Join(t.TempDir(), "history.json")
	if err := client.SaveHistory(history, path); err != nil {
		t.Fatalf("SaveHistory: %v", err)
	}

	loaded, err := LoadHistory(path)
	if err != nil {
		t.Fatalf("LoadHistory: %v", err)
	}

	if loaded.Model != history.Model {
		t.Fatalf("loaded model = %q, want %q", loaded.Model, history.Model)
	}
	if got := loaded.Contents[1].Parts[1].ThoughtSignature; got != "sig-b" {
		t.Fatalf("loaded thought signature = %q, want %q", got, "sig-b")
	}
}

func TestTranslateGrounding(t *testing.T) {
	meta := translateGrounding(&apiGroundingMetadata{
		WebSearchQueries:   []string{"weather nyc"},
		ImageSearchQueries: []string{"timareta butterfly"},
		SearchEntryPoint:   &apiSearchEntryPoint{RenderedContent: "<div>search</div>"},
		GroundingChunks: []apiGroundingChunk{
			{Web: &apiGroundingChunkWeb{URI: "https://example.com/a", Title: "A"}},
			{Image: &apiGroundingChunkImage{URI: "https://example.com/b", ImageURI: "https://img.example.com/b.png", Title: "B"}},
		},
	})

	if meta == nil {
		t.Fatal("translateGrounding returned nil")
	}
	if len(meta.GroundingChunks) != 2 {
		t.Fatalf("grounding chunk count = %d, want 2", len(meta.GroundingChunks))
	}
	if meta.SearchEntryPoint == nil || meta.SearchEntryPoint.RenderedContent == "" {
		t.Fatal("search entry point was not preserved")
	}
}
