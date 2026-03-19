package gemini

import (
	"fmt"
	"slices"
	"strings"
)

const (
	ModelFlash25   = "gemini-2.5-flash-image"
	ModelFlash31   = "gemini-3.1-flash-image-preview"
	ModelPro       = "gemini-3-pro-image-preview"
	DefaultModelID = ModelFlash31
)

type ModelSpec struct {
	ID                    string
	Aliases               []string
	DefaultImageSize      string
	SupportedAspectRatios []string
	SupportedImageSizes   []string
	SupportsGrounding     bool
	SupportsImageSearch   bool
	SupportsThinking      bool
	SupportsThinkingLevel bool
	MaxInputImages        int
}

var (
	standardAspectRatios = []string{
		"1:1",
		"2:3", "3:2",
		"3:4", "4:3",
		"4:5", "5:4",
		"9:16", "16:9",
		"21:9",
	}

	flash31AspectRatios = []string{
		"1:1",
		"1:4", "4:1",
		"1:8", "8:1",
		"2:3", "3:2",
		"3:4", "4:3",
		"4:5", "5:4",
		"9:16", "16:9",
		"21:9",
	}

	allAspectRatios = []string{
		"1:1",
		"1:4", "4:1",
		"1:8", "8:1",
		"2:3", "3:2",
		"3:4", "4:3",
		"4:5", "5:4",
		"9:16", "16:9",
		"21:9",
	}

	allImageSizes = []string{"512", "1K", "2K", "4K"}

	modelsByID = map[string]ModelSpec{
		ModelFlash25: {
			ID:                    ModelFlash25,
			Aliases:               []string{"banana", "2.5"},
			DefaultImageSize:      "1K",
			SupportedAspectRatios: standardAspectRatios,
			SupportedImageSizes:   []string{"1K"},
			MaxInputImages:        3,
		},
		ModelFlash31: {
			ID:                    ModelFlash31,
			Aliases:               []string{"banana2", "3.1"},
			DefaultImageSize:      "1K",
			SupportedAspectRatios: flash31AspectRatios,
			SupportedImageSizes:   allImageSizes,
			SupportsGrounding:     true,
			SupportsImageSearch:   true,
			SupportsThinking:      true,
			SupportsThinkingLevel: true,
			MaxInputImages:        14,
		},
		ModelPro: {
			ID:                    ModelPro,
			Aliases:               []string{"pro"},
			DefaultImageSize:      "1K",
			SupportedAspectRatios: standardAspectRatios,
			SupportedImageSizes:   []string{"1K", "2K", "4K"},
			SupportsGrounding:     true,
			SupportsThinking:      true,
			MaxInputImages:        14,
		},
	}

	aliasToModelID = func() map[string]string {
		m := map[string]string{}
		for id, spec := range modelsByID {
			m[id] = id
			for _, alias := range spec.Aliases {
				m[alias] = id
			}
		}
		return m
	}()
)

type ValidatedModel struct {
	Spec  ModelSpec
	Alias string
}

func ResolveModelName(name string) string {
	return ResolveModel(name).Spec.ID
}

func ResolveModel(name string) ValidatedModel {
	normalized := strings.TrimSpace(strings.ToLower(name))
	if normalized == "" {
		return ValidatedModel{Spec: modelsByID[DefaultModelID], Alias: "banana2"}
	}

	if id, ok := aliasToModelID[normalized]; ok {
		return ValidatedModel{Spec: modelsByID[id], Alias: normalized}
	}

	// Unknown values are treated as raw model IDs.
	return ValidatedModel{
		Spec: ModelSpec{
			ID:                    name,
			DefaultImageSize:      "1K",
			SupportedAspectRatios: allAspectRatios,
			SupportedImageSizes:   allImageSizes,
			SupportsGrounding:     true,
			SupportsImageSearch:   true,
			SupportsThinking:      true,
			SupportsThinkingLevel: true,
			MaxInputImages:        14,
		},
		Alias: name,
	}
}

func ListModelAliases() []string {
	return []string{
		"banana2 (default), 3.1",
		"banana, 2.5",
		"pro",
	}
}

func ListAllAspectRatios() []string {
	return append([]string(nil), allAspectRatios...)
}

func IsKnownModel(name string) bool {
	_, ok := aliasToModelID[strings.ToLower(strings.TrimSpace(name))]
	return ok
}

func IsValidAspectRatio(ratio string) bool {
	return slices.Contains(allAspectRatios, strings.TrimSpace(ratio))
}

func IsValidImageSize(size string) bool {
	size = strings.TrimSpace(strings.ToUpper(size))
	return size == "" || slices.Contains(allImageSizes, size)
}

func ValidateAspectRatio(spec ModelSpec, ratio string) error {
	ratio = strings.TrimSpace(ratio)
	if ratio == "" {
		return nil
	}
	if !slices.Contains(spec.SupportedAspectRatios, ratio) {
		return fmt.Errorf("aspect ratio %q is not supported by model %s", ratio, spec.ID)
	}
	return nil
}

func ValidateImageSize(spec ModelSpec, size string) error {
	size = strings.TrimSpace(strings.ToUpper(size))
	if size == "" {
		return nil
	}
	if !slices.Contains(spec.SupportedImageSizes, size) {
		return fmt.Errorf("image size %q is not supported by model %s", size, spec.ID)
	}
	return nil
}

func DefaultImageSize(spec ModelSpec, requested string) string {
	requested = strings.TrimSpace(strings.ToUpper(requested))
	if requested != "" {
		return requested
	}
	return spec.DefaultImageSize
}
