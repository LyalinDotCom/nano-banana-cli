# Nanobanana CLI

AI-powered image generation and manipulation CLI using Google's Gemini models.

## Features

- **Image Generation** - Generate images from text prompts
- **Image Editing** - Edit existing images with natural language
- **Icon Generation** - Generate app icons in multiple sizes
- **Pattern Generation** - Create seamless, tileable patterns
- **Image Transformation** - Resize, crop, rotate, flip images
- **Transparency** - Remove backgrounds, inspect alpha channels
- **Image Combining** - Create sprite sheets, strips, and grids

## Installation

Download the latest release for your platform from [GitHub Releases](https://github.com/lyalindotcom/nano-banana-cli/releases).

Or build from source:

```bash
go install github.com/lyalindotcom/nano-banana-cli/cmd/nanobanana@latest
```

## Quick Start

```bash
# Set your API key
export GEMINI_API_KEY=your-api-key

# Generate an image
nanobanana generate "a robot playing guitar" -o robot.png

# Generate icons
nanobanana icon "app logo" -o ./icons/ --sizes 64,128,256,512

# Resize an image
nanobanana transform large.png -o thumb.png --resize 200x200

# Remove background
nanobanana transparent make sprite.png -o sprite-clean.png

# Combine images into sprite sheet
nanobanana combine frame1.png frame2.png frame3.png -o spritesheet.png
```

## Commands

| Command | Description |
|---------|-------------|
| `generate` | Generate images from text prompts |
| `icon` | Generate icons in multiple sizes |
| `pattern` | Generate seamless patterns and textures |
| `transform` | Resize, crop, rotate, flip images |
| `transparent` | Remove backgrounds, inspect transparency |
| `combine` | Combine multiple images into one |
| `version` | Print version information |

Run `nanobanana --help` or `nanobanana [command] --help` for detailed usage.

## Models

- **flash** (default): Gemini 2.5 Flash - Fast, optimized for high-volume tasks
- **pro**: Gemini 3 Pro - Professional quality, supports 4K resolution

## Authentication

Set your Gemini API key via:
- Environment variable: `GEMINI_API_KEY` or `NANOBANANA_API_KEY`
- Flag: `--api-key YOUR_KEY`
- `.env` file in current directory

Get your API key at [Google AI Studio](https://aistudio.google.com/apikey).

## JSON Output

All commands support `--json` flag for programmatic parsing:

```bash
nanobanana generate "sunset" -o sunset.png --json
```

## License

MIT
