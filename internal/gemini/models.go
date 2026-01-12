package gemini

// Model constants for Gemini image generation
const (
	// ModelFlash is the fast Gemini 2.5 Flash model for image generation
	ModelFlash = "gemini-2.5-flash-image-preview"

	// ModelPro is the higher quality Gemini 3 Pro model
	ModelPro = "gemini-3-pro-image-preview"
)

// Supported aspect ratios
var AspectRatios = []string{
	"1:1",
	"3:2", "2:3",
	"3:4", "4:3",
	"4:5", "5:4",
	"9:16", "16:9",
	"21:9",
}

// Supported resolutions (4K only available for Pro model)
var Resolutions = []string{
	"1K",
	"2K",
	"4K", // Pro model only
}

// IsValidAspectRatio checks if the aspect ratio is supported
func IsValidAspectRatio(ratio string) bool {
	for _, r := range AspectRatios {
		if r == ratio {
			return true
		}
	}
	return false
}

// IsValidResolution checks if the resolution is supported
func IsValidResolution(resolution string) bool {
	for _, r := range Resolutions {
		if r == resolution {
			return true
		}
	}
	return false
}

// ResolveModelName converts short names to full model IDs
func ResolveModelName(name string) string {
	switch name {
	case "flash", "":
		return ModelFlash
	case "pro":
		return ModelPro
	default:
		return name
	}
}
