---
name: nano-banana-go-cli
description: Use when you need the repo's Go Nano Banana CLI for Gemini image generation or editing, including model selection across Gemini 2.5, Gemini 3.1, and Pro, multi-image references, Google Search grounding, thought output, and scripted multi-turn history files.
---

# Nano Banana Go CLI

Use this skill when the task is to drive the repo's Go CLI instead of calling the Gemini image API directly.

The CLI lives in the current repo and builds to `./nanobanana`.

An API key is required for any Gemini-backed command. Prefer `GEMINI_API_KEY`. The CLI also accepts `NANOBANANA_API_KEY` and `GOOGLE_API_KEY`.

Default model guidance:
- Use `banana2` unless the task specifically needs `2.5` or `pro`.
- `banana2` / `3.1` -> `gemini-3.1-flash-image-preview`
- `banana` / `2.5` -> `gemini-2.5-flash-image`
- `pro` -> `gemini-3-pro-image-preview`

## What It Enables

- Text-to-image generation
- Image editing with one or more reference images
- Google Search grounding via `--ground-web`
- Google Image Search grounding on Gemini 3.1 via `--ground-image`
- Thought output via `--include-thoughts` and `--thoughts-dir`
- Scripted multi-turn image workflows via `--history-in` and `--history-out`
- A single discovery command: `./nanobanana docs`

## Recommended Workflow

1. Build or rebuild the CLI:

```bash
make build
```

2. Use the manual when you need the whole interface in one place:

```bash
./nanobanana docs
```

3. Use command-specific help when needed:

```bash
./nanobanana generate --help
./nanobanana icon --help
```

## Common Commands

Basic generation:

```bash
./nanobanana generate "a robot playing guitar" -o robot.png
```

Explicit Pro usage:

```bash
./nanobanana generate "designer perfume bottle" -m pro --image-size 4K -o bottle.png
```

Reference-image editing:

```bash
./nanobanana generate "add sunglasses and keep the face unchanged" -i face.png -o face-edit.png
```

Multiple references:

```bash
./nanobanana generate "office group photo of these people" -i p1.png -i p2.png -i p3.png -o group.png
```

Grounding:

```bash
./nanobanana generate "a stylish weather poster for New York today" --ground-web -o weather.png
```

Google Image Search grounding on 3.1:

```bash
./nanobanana generate "a detailed painting of a Timareta butterfly resting on a flower" -m banana2 --ground-image -o butterfly.png
```

Scripted multi-turn flow:

```bash
./nanobanana generate "Create a colorful infographic about photosynthesis" -o photo.png --history-out photo-history.json
./nanobanana generate "Translate the infographic to Spanish and change nothing else" -o photo-es.png --history-in photo-history.json --history-out photo-history.json
```

## Important Constraints

- `banana2` is the default and recommended model.
- `--ground-image` is only for Gemini 3.1.
- Gemini 3.1 supports `512` plus aspect ratios `1:4`, `4:1`, `1:8`, `8:1`.
- Pro supports `1K`, `2K`, and `4K`, but not `512`.
- Gemini 2.5 behaves like fixed `1K`; do not expect `512`, `2K`, or `4K`.
- History files are for scripted automation, not an interactive chat UI.
- When grounded image search is used, preserve containing-page source URLs from JSON output for attribution.
