# Nanobanana CLI

> **⚠️ DEPRECATED:** This repository is no longer maintained. Please use the actively maintained version at **[https://github.com/LyalinDotCom/nanobanana-cli](https://github.com/LyalinDotCom/nanobanana-cli)**.

AI-powered image generation and manipulation CLI for Gemini image models.

> **Note:** This is a personal project and is not actively supported. Pull requests are not accepted. Feel free to [open an issue](https://github.com/lyalindotcom/nano-banana-cli/issues) if you run into something.

## Features

- Generate and edit images with Gemini 2.5, Gemini 3.1, and Gemini 3 Pro image models
- Use one or many reference images in a single prompt
- Ground image generation with Google Search, including Google Image Search on Gemini 3.1
- Save and resume scripted multi-turn image workflows with history files
- Optionally include thought parts and save thought images
- Generate icons, patterns, and other derived image assets
- Resize, crop, rotate, flip, combine, and clean transparency locally

## Installation

Use one of these install paths.

### Recommended: clone and build

```bash
git clone https://github.com/lyalindotcom/nano-banana-cli.git
cd nano-banana-cli/nano-banana-cli
make build
./nanobanana docs
```

This is the clearest path for contributors, local users, and agents working from the repo.

### Go users: install globally

```bash
go install github.com/lyalindotcom/nano-banana-cli/cmd/nanobanana@latest
```

After `go install`, the binary is placed in your Go bin directory, usually:

```bash
$(go env GOPATH)/bin
```

If `nanobanana` is not found, either run it directly:

```bash
"$(go env GOPATH)/bin/nanobanana" docs
```

Or add that directory to your `PATH`.

### Release binary

Download a release from [GitHub Releases](https://github.com/lyalindotcom/nano-banana-cli/releases), make it executable, and run `nanobanana docs`.

## Quick Start

```bash
# Configure your API key once at the user level
nanobanana config set-api-key

# See the full manual for every command, key flag, and example
nanobanana docs

# Default model: Nano Banana 2 / Gemini 3.1 Flash Image Preview
nanobanana generate "a robot playing guitar" -o robot.png

# Use Gemini 2.5 Flash Image explicitly
nanobanana generate "a product icon" -m banana -o icon.png

# Use Gemini 3 Pro at 4K
nanobanana generate "designer perfume bottle" -m pro --image-size 4K -o bottle.png

# Edit an existing image
nanobanana generate "make it look like watercolor" -i photo.jpg -o watercolor.png

# Use multiple reference images
nanobanana generate "an office group photo of these people" \
  -i person1.png -i person2.png -i person3.png \
  -o group.png

# Ground with Google Search
nanobanana generate "a stylish poster showing today's weather in New York" \
  --ground-web \
  -o weather.png

# Ground with Google Search + Image Search (Gemini 3.1 only)
nanobanana generate "a detailed painting of a Timareta butterfly resting on a flower" \
  --ground-image \
  -o butterfly.png
```

If you prefer environment variables instead of persistent config:

```bash
export GEMINI_API_KEY=your-api-key
```

## Models

The CLI accepts aliases or raw Google model IDs.

| Alias | Model ID | Notes |
| --- | --- | --- |
| `banana2` (default), `3.1` | `gemini-3.1-flash-image-preview` | Recommended default, supports `512`, extra aspect ratios, grounding, thought controls |
| `banana`, `2.5` | `gemini-2.5-flash-image` | Fast Nano Banana model for lower-latency runs |
| `pro` | `gemini-3-pro-image-preview` | Best for professional asset production and 4K output |

## Documentation and Discovery

Use one of these commands to understand the CLI:

```bash
nanobanana docs
nanobanana --help
nanobanana <command> --help
```

`nanobanana docs` is the single-command manual intended for agents and humans who need the full interface in one place.

## Authentication

Gemini-backed commands require an API key.

Recommended persistent setup:

```bash
nanobanana config set-api-key
nanobanana config show
nanobanana config path
```

Alternative sources, in practice:

- `--api-key`
- `GEMINI_API_KEY`
- `NANOBANANA_API_KEY`
- `GOOGLE_API_KEY`
- local `.env`

The config command is intended to keep the CLI usable even when local project `.env` files are absent or agents run in sandboxes that do not preserve them.

## Commands

| Command | Description |
| --- | --- |
| `generate` | Generate or edit images with Gemini image models |
| `icon` | Generate icons in multiple sizes |
| `pattern` | Generate seamless patterns and textures |
| `transform` | Resize, crop, rotate, flip images |
| `transparent make` | Remove a background color and save a transparent PNG |
| `transparent inspect` | Inspect transparency details for an image |
| `combine` | Combine multiple images into one |
| `version` | Print version information |
| `config` | Manage persistent user-level configuration |
| `docs` | Print the full CLI manual |

## Command Reference

### `generate`

Usage:

```bash
nanobanana generate [prompt] -o OUTPUT
```

Key flags:

- Repeatable `-i/--input` reference images
- `-o/--output` output file path
- `-m/--model` model alias or raw model ID
- `--image-size 512|1K|2K|4K`
- `--aspect-ratio` with model-aware validation
- `--ground-web`
- `--ground-image` for Gemini 3.1
- `--thinking-level minimal|high` for Gemini 3.1
- `--include-thoughts`
- `--thoughts-dir`
- `--history-in` / `--history-out` for scriptable multi-turn workflows

### Scripted Multi-Turn Editing

```bash
nanobanana generate \
  "Create a colorful infographic about photosynthesis" \
  -o photosynthesis.png \
  --history-out photosynthesis-history.json

nanobanana generate \
  "Translate the infographic to Spanish and change nothing else" \
  -o photosynthesis-es.png \
  --history-in photosynthesis-history.json \
  --history-out photosynthesis-history.json
```

History files preserve the conversation contents needed for follow-up turns, including thought signatures returned by the API.

### `icon`

Usage:

```bash
nanobanana icon [prompt] -o OUTPUT
```

Key flags:

- `-o/--output` output directory or naming pattern
- `--sizes` comma-separated icon sizes
- `--style` `modern|flat|minimal|detailed`
- `--background` `transparent|white|black|#RRGGBB`

Examples:

```bash
nanobanana icon "coffee cup logo" -o ./icons/
nanobanana icon "settings gear" -o ./icons/ --sizes 16,32,64,128
```

### `pattern`

Usage:

```bash
nanobanana pattern [prompt] -o OUTPUT
```

Key flags:

- `-o/--output` output file path
- `--size` tile size as `WxH`
- `--style` `geometric|organic|abstract|floral|tech`
- `--type` `seamless|texture|wallpaper`

Examples:

```bash
nanobanana pattern "hexagon grid" -o hex.png
nanobanana pattern "oak wood grain" -o wood.png --type texture
```

### `transform`

Usage:

```bash
nanobanana transform INPUT -o OUTPUT
```

Key flags:

- `-o/--output` output file path
- `--resize` resize to `WxH` or `%`
- `--fit` resize behavior
- `--crop` crop rectangle `left,top,width,height`
- `--rotate` angle
- `--flip` vertical flip
- `--flop` horizontal mirror

Examples:

```bash
nanobanana transform photo.jpg -o thumb.jpg --resize 200x200
nanobanana transform image.png -o cropped.png --crop 100,50,400,300
```

### `transparent make`

Usage:

```bash
nanobanana transparent make INPUT -o OUTPUT
```

Key flags:

- `-o/--output` output PNG path
- `--color` `white|black|#RRGGBB`
- `--tolerance` color matching tolerance
- `--overwrite` overwrite original file

Example:

```bash
nanobanana transparent make sprite.png -o sprite-clean.png
```

### `transparent inspect`

Usage:

```bash
nanobanana transparent inspect INPUT
```

Example:

```bash
nanobanana transparent inspect sprite.png
```

### `combine`

Usage:

```bash
nanobanana combine IMAGE... -o OUTPUT
```

Key flags:

- `-o/--output` output file path
- `--direction` `horizontal|vertical|grid`
- `--gap` gap in pixels
- `--columns` number of columns for grids
- `--align` `start|center|end`
- `--background` `transparent|white|black|#RRGGBB`

Examples:

```bash
nanobanana combine frame1.png frame2.png frame3.png -o spritesheet.png
nanobanana combine *.png -o grid.png --direction grid --columns 4
```

### `version`

Usage:

```bash
nanobanana version
```

### `config`

Usage:

```bash
nanobanana config <subcommand>
```

Subcommands:

- `nanobanana config path`
- `nanobanana config show`
- `nanobanana config set-api-key [key]`
- `nanobanana config clear-api-key`

Examples:

```bash
nanobanana config set-api-key
nanobanana config set-api-key YOUR_API_KEY
nanobanana config show
nanobanana config clear-api-key
```

### `docs`

Usage:

```bash
nanobanana docs
```

This prints the full CLI manual for all commands, key flags, model aliases, and examples.

## JSON Output

All commands support `--json`:

```bash
nanobanana generate "sunset over mountains" -o sunset.png --json
```

For grounded runs, JSON output includes grounding metadata and source URLs. When Google Image Search grounding is used, the response includes containing-page URLs for attribution.

## License

Apache 2.0. See [LICENSE](LICENSE).
