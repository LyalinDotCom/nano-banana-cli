package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

const docsText = `Nanobanana CLI Manual

Discovery:
  nanobanana docs
  nanobanana --help
  nanobanana <command> --help

Authentication:
  Gemini-backed commands require an API key.
  Recommended: nanobanana config set-api-key
  Supported env vars: GEMINI_API_KEY, NANOBANANA_API_KEY, GOOGLE_API_KEY
  Default model: banana2 (Gemini 3.1 Flash Image Preview)

Installation and first run:
  Recommended from a clone:
    cd nano-banana-cli
    make build
    ./nanobanana config set-api-key
    ./nanobanana docs
    ./nanobanana generate "a robot playing guitar" -o robot.png

  With go install:
    go install github.com/lyalindotcom/nano-banana-cli/cmd/nanobanana@latest
    If "nanobanana" is not found, run:
      "$(go env GOPATH)/bin/nanobanana" docs
    Or add "$(go env GOPATH)/bin" to PATH.

Models:
  banana2, 3.1 -> gemini-3.1-flash-image-preview
  banana, 2.5  -> gemini-2.5-flash-image
  pro          -> gemini-3-pro-image-preview
  Raw model IDs are also accepted.

Commands:

1. generate
   Generate or edit images with Gemini image models.
   Key flags:
     -o, --output
     -i, --input (repeatable)
     -m, --model
     --aspect-ratio
     --image-size
     --ground-web
     --ground-image
     --thinking-level
     --include-thoughts
     --thoughts-dir
     --history-in
     --history-out
   Examples:
     nanobanana generate "a robot playing guitar" -o robot.png
     nanobanana generate "add sunglasses" -i face.png -o face-edit.png
     nanobanana generate "group photo of these people" -i p1.png -i p2.png -o group.png
     nanobanana generate "weather poster for New York today" --ground-web -o weather.png
     nanobanana generate "designer perfume bottle" -m pro --image-size 4K -o bottle.png

2. icon
   Generate icons in multiple sizes.
   Key flags:
     -o, --output
     --sizes
     --style
     --background
   Examples:
     nanobanana icon "coffee cup logo" -o ./icons/
     nanobanana icon "settings gear" -o ./icons/ --sizes 16,32,64,128

3. pattern
   Generate seamless patterns and textures.
   Key flags:
     -o, --output
     --size
     --style
     --type
   Examples:
     nanobanana pattern "hexagon grid" -o hex.png
     nanobanana pattern "oak wood grain" -o wood.png --type texture

4. transform
   Apply local image transforms.
   Key flags:
     -o, --output
     --resize
     --fit
     --crop
     --rotate
     --flip
     --flop
   Examples:
     nanobanana transform photo.jpg -o thumb.jpg --resize 200x200
     nanobanana transform image.png -o cropped.png --crop 100,50,400,300

5. transparent make
   Remove a background color and save a transparent PNG.
   Key flags:
     -o, --output
     --color
     --tolerance
     --overwrite
   Example:
     nanobanana transparent make sprite.png -o sprite-clean.png

6. transparent inspect
   Inspect transparency details for an image.
   Example:
     nanobanana transparent inspect sprite.png

7. combine
   Combine multiple images into one strip or grid.
   Key flags:
     -o, --output
     --direction
     --gap
     --columns
     --align
     --background
   Examples:
     nanobanana combine frame1.png frame2.png frame3.png -o spritesheet.png
     nanobanana combine *.png -o grid.png --direction grid --columns 4

8. version
   Print version and build information.

9. config
   Manage persistent user-level configuration.
   Subcommands:
     path
     show
     set-api-key [key]
     clear-api-key
   Examples:
     nanobanana config set-api-key
     nanobanana config show

10. docs
   Print this manual.
`

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Print a full manual for all commands",
	Long:  "Print a single-page manual covering every command, key flags, model aliases, and examples.",
	Run: func(cmd *cobra.Command, args []string) {
		f := GetFormatter()
		if f.JSONMode {
			f.Success("docs", map[string]any{
				"help_command":     "nanobanana docs",
				"subcommand_help":  "nanobanana <command> --help",
				"default_model":    "banana2",
				"api_key_required": true,
				"commands": []string{
					"generate",
					"icon",
					"pattern",
					"transform",
					"transparent make",
					"transparent inspect",
					"combine",
					"version",
					"config",
					"docs",
				},
			}, nil)
			return
		}
		fmt.Print(docsText)
	},
}

func init() {
	rootCmd.AddCommand(docsCmd)
}
