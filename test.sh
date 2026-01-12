#!/bin/bash
#
# Visual Test Suite for Nanobanana CLI
#
# Usage:
#   ./test.sh              Show available test groups
#   ./test.sh all          Run all tests
#   ./test.sh basic        Run basic tests only
#   ./test.sh game-assets  Run game asset tests only
#   ./test.sh basic icons  Run multiple test groups
#
# Requirements:
# - GEMINI_API_KEY environment variable or .env file
# - Built nanobanana binary in current directory
#

# Don't exit on error - we want to continue and report all failures
set +e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
DIM='\033[2m'
NC='\033[0m' # No Color

# Test directory
TESTS_DIR=".tests"
PROMPTS_DIR="$TESTS_DIR/prompts"

# Binary path
NANOBANANA="./nanobanana"

# Counters
PASSED=0
FAILED=0
SKIPPED=0

# Delay between API calls to avoid rate limiting (seconds)
API_DELAY=2

# Available test groups
declare -A TEST_GROUPS
TEST_GROUPS=(
    ["basic"]="Basic image generation with simple prompts"
    ["complex"]="Complex multi-line prompts from files"
    ["stdin"]="Stdin/pipe input for AI agent integration"
    ["icons"]="Icon generation in multiple sizes"
    ["patterns"]="Seamless patterns and textures"
    ["game-assets"]="Game development assets (characters, backgrounds, buildings)"
    ["transform"]="Image transformations (resize, rotate, crop) - no API"
    ["transparency"]="Transparency operations (remove background)"
    ["combine"]="Image combining (strips, grids)"
    ["json"]="JSON output format"
    ["edge-cases"]="Unicode and special characters"
)

# Test group order for 'all'
TEST_ORDER="basic complex stdin icons patterns game-assets transform transparency combine json edge-cases"

#═══════════════════════════════════════════════════════════════════════════════
# HELPER FUNCTIONS
#═══════════════════════════════════════════════════════════════════════════════

show_help() {
    echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║         Nanobanana CLI Visual Test Suite                   ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${CYAN}Usage:${NC}"
    echo "  ./test.sh              Show this help"
    echo "  ./test.sh all          Run all tests"
    echo "  ./test.sh <group>      Run specific test group"
    echo "  ./test.sh <g1> <g2>    Run multiple test groups"
    echo ""
    echo -e "${CYAN}Available test groups:${NC}"
    echo ""
    for group in $TEST_ORDER; do
        printf "  ${GREEN}%-15s${NC} %s\n" "$group" "${TEST_GROUPS[$group]}"
    done
    echo ""
    echo -e "${CYAN}Examples:${NC}"
    echo "  ./test.sh all                    # Run everything"
    echo "  ./test.sh basic                  # Just basic tests"
    echo "  ./test.sh game-assets            # Game dev assets only"
    echo "  ./test.sh basic icons patterns   # Multiple groups"
    echo ""
    echo -e "${CYAN}Output:${NC}"
    echo "  Images saved to: ${YELLOW}$TESTS_DIR/${NC}"
    echo ""
}

setup_dirs() {
    mkdir -p "$TESTS_DIR"
    mkdir -p "$PROMPTS_DIR"
    mkdir -p "$TESTS_DIR/icons"
    mkdir -p "$TESTS_DIR/game-assets/characters"
    mkdir -p "$TESTS_DIR/game-assets/backgrounds"
    mkdir -p "$TESTS_DIR/game-assets/buildings"
    mkdir -p "$TESTS_DIR/game-assets/items"
}

# Run a test with API call
run_test() {
    local name="$1"
    local cmd="$2"
    local output_file="$3"

    echo -n "  Testing: $name... "

    local output
    output=$(eval "$cmd" 2>&1)
    local exit_code=$?

    if [ $exit_code -eq 0 ] && [ -f "$output_file" ]; then
        size=$(ls -lh "$output_file" | awk '{print $5}')
        echo -e "${GREEN}✓ PASS${NC} ${DIM}($size)${NC}"
        ((PASSED++))
    else
        echo -e "${RED}✗ FAIL${NC}"
        echo "$output" | head -3 | sed 's/^/      /'
        ((FAILED++))
    fi

    sleep $API_DELAY
}

# Run a test without API delay (local operations)
run_test_local() {
    local name="$1"
    local cmd="$2"
    local output_file="$3"

    echo -n "  Testing: $name... "

    local output
    output=$(eval "$cmd" 2>&1)
    local exit_code=$?

    if [ $exit_code -eq 0 ] && [ -f "$output_file" ]; then
        size=$(ls -lh "$output_file" | awk '{print $5}')
        echo -e "${GREEN}✓ PASS${NC} ${DIM}($size)${NC}"
        ((PASSED++))
    else
        echo -e "${RED}✗ FAIL${NC}"
        echo "$output" | head -3 | sed 's/^/      /'
        ((FAILED++))
    fi
}

# Optional test (failures counted as skip)
run_test_optional() {
    local name="$1"
    local cmd="$2"
    local output_file="$3"

    echo -n "  Testing: $name... "

    local output
    output=$(eval "$cmd" 2>&1)
    local exit_code=$?

    if [ $exit_code -eq 0 ] && [ -f "$output_file" ]; then
        size=$(ls -lh "$output_file" | awk '{print $5}')
        echo -e "${GREEN}✓ PASS${NC} ${DIM}($size)${NC}"
        ((PASSED++))
    else
        echo -e "${YELLOW}○ SKIP${NC} ${DIM}(optional)${NC}"
        ((SKIPPED++))
    fi

    sleep $API_DELAY
}

section_header() {
    echo ""
    echo -e "${BLUE}━━━ $1 ━━━${NC}"
}

#═══════════════════════════════════════════════════════════════════════════════
# TEST GROUPS
#═══════════════════════════════════════════════════════════════════════════════

test_basic() {
    section_header "Basic Image Generation"

    run_test "Simple prompt" \
        "$NANOBANANA generate 'a red apple on white background' -o '$TESTS_DIR/basic_01_simple.png'" \
        "$TESTS_DIR/basic_01_simple.png"

    run_test "Prompt with quotes" \
        "$NANOBANANA generate 'A neon sign that says \"OPEN 24/7\"' -o '$TESTS_DIR/basic_02_quotes.png'" \
        "$TESTS_DIR/basic_02_quotes.png"

    run_test "Landscape (16:9)" \
        "$NANOBANANA generate 'mountain landscape at golden hour sunset' -o '$TESTS_DIR/basic_03_landscape.png' --aspect-ratio 16:9" \
        "$TESTS_DIR/basic_03_landscape.png"

    run_test "Portrait (9:16)" \
        "$NANOBANANA generate 'tall lighthouse on rocky coast' -o '$TESTS_DIR/basic_04_portrait.png' --aspect-ratio 9:16" \
        "$TESTS_DIR/basic_04_portrait.png"

    run_test "Square (1:1)" \
        "$NANOBANANA generate 'cute cat sitting in a box' -o '$TESTS_DIR/basic_05_square.png' --aspect-ratio 1:1" \
        "$TESTS_DIR/basic_05_square.png"
}

test_complex() {
    section_header "Complex Multi-line Prompts"

    # Create prompt files
    cat > "$PROMPTS_DIR/photorealistic.txt" << 'PROMPT'
A photorealistic close-up portrait of an elderly Japanese ceramicist
with deep, sun-etched wrinkles and a warm, knowing smile. He is
carefully inspecting a freshly glazed tea bowl. The setting is his
rustic, sun-drenched workshop with pottery wheels and shelves of clay
pots in the background. The scene is illuminated by soft, golden hour
light streaming through a window, highlighting the fine texture of the
clay and the fabric of his apron. Captured with an 85mm portrait lens,
resulting in a soft, blurred background (bokeh).
PROMPT

    cat > "$PROMPTS_DIR/sticker.txt" << 'PROMPT'
A kawaii-style sticker of a happy red panda wearing a tiny bamboo hat.
It's munching on a green bamboo leaf. The design features bold, clean
outlines, simple cel-shading, and a vibrant color palette. The
background must be solid white for easy cutting.
PROMPT

    cat > "$PROMPTS_DIR/infographic.txt" << 'PROMPT'
Create a visually stunning infographic about the water cycle.
Include these stages with clear icons and arrows:
1. Evaporation from oceans and lakes (sun heating water)
2. Condensation forming clouds (water vapor rising)
3. Precipitation as rain and snow (clouds releasing water)
4. Collection in rivers, lakes, and groundwater
Use a clean, modern design with a blue color palette.
Add simple labels for each stage. Educational style.
PROMPT

    run_test "Photorealistic portrait" \
        "$NANOBANANA generate --prompt-file '$PROMPTS_DIR/photorealistic.txt' -o '$TESTS_DIR/complex_01_photo.png'" \
        "$TESTS_DIR/complex_01_photo.png"

    run_test "Kawaii sticker" \
        "$NANOBANANA generate --prompt-file '$PROMPTS_DIR/sticker.txt' -o '$TESTS_DIR/complex_02_sticker.png'" \
        "$TESTS_DIR/complex_02_sticker.png"

    run_test "Infographic" \
        "$NANOBANANA generate --prompt-file '$PROMPTS_DIR/infographic.txt' -o '$TESTS_DIR/complex_03_infographic.png'" \
        "$TESTS_DIR/complex_03_infographic.png"
}

test_stdin() {
    section_header "Stdin Input (AI Agent Simulation)"

    run_test "Pipe input" \
        "echo 'A friendly robot waving hello, cartoon style, white background' | $NANOBANANA generate - -o '$TESTS_DIR/stdin_01_pipe.png'" \
        "$TESTS_DIR/stdin_01_pipe.png"

    run_test "Heredoc input" \
        "cat << 'EOF' | $NANOBANANA generate - -o '$TESTS_DIR/stdin_02_heredoc.png'
A cozy coffee shop interior with warm lighting,
wooden furniture, exposed brick walls, and
plants hanging from the ceiling.
Watercolor painting style with soft edges.
EOF" \
        "$TESTS_DIR/stdin_02_heredoc.png"
}

test_icons() {
    section_header "Icon Generation"

    run_test "App icon (multiple sizes)" \
        "$NANOBANANA icon 'modern coffee cup logo, minimalist, steam rising' -o '$TESTS_DIR/icons/' --sizes 64,128,256,512" \
        "$TESTS_DIR/icons/icon_64.png"

    run_test "Flat style icon" \
        "$NANOBANANA icon 'gear settings icon' -o '$TESTS_DIR/icons/flat/' --style flat --sizes 64,128" \
        "$TESTS_DIR/icons/flat/icon_64.png"

    run_test "Minimal style icon" \
        "$NANOBANANA icon 'play button triangle' -o '$TESTS_DIR/icons/minimal/' --style minimal --sizes 64,128" \
        "$TESTS_DIR/icons/minimal/icon_64.png"
}

test_patterns() {
    section_header "Pattern Generation"

    run_test "Geometric pattern" \
        "$NANOBANANA pattern 'hexagon honeycomb grid, blue and gold' -o '$TESTS_DIR/pattern_01_geometric.png' --style geometric" \
        "$TESTS_DIR/pattern_01_geometric.png"

    run_test "Organic pattern" \
        "$NANOBANANA pattern 'flowing water ripples' -o '$TESTS_DIR/pattern_02_organic.png' --style organic" \
        "$TESTS_DIR/pattern_02_organic.png"

    run_test "Wood texture" \
        "$NANOBANANA pattern 'oak wood grain, natural' -o '$TESTS_DIR/pattern_03_wood.png' --type texture" \
        "$TESTS_DIR/pattern_03_wood.png"

    run_test "Tech pattern" \
        "$NANOBANANA pattern 'circuit board traces, dark background' -o '$TESTS_DIR/pattern_04_tech.png' --style tech" \
        "$TESTS_DIR/pattern_04_tech.png"
}

test_game_assets() {
    section_header "Game Assets - Characters"

    # Character prompts with transparent background emphasis
    cat > "$PROMPTS_DIR/char_hero.txt" << 'PROMPT'
16-bit pixel art sprite of a brave knight hero character.
- Silver armor with blue cape
- Holding a glowing sword
- Heroic standing pose, facing right
- Style: retro SNES-era pixel art
- Limited color palette (16 colors max)
- NO background - character only on solid white
- Clean pixel edges, no anti-aliasing
- Size suitable for 64x64 sprite
PROMPT

    cat > "$PROMPTS_DIR/char_enemy.txt" << 'PROMPT'
16-bit pixel art sprite of a slime enemy monster.
- Green translucent body with darker green core
- Two simple dot eyes, angry expression
- Bouncy blob shape
- Style: classic JRPG enemy
- Limited color palette
- NO background - solid white background only
- Clean pixels, suitable for 32x32 sprite
PROMPT

    cat > "$PROMPTS_DIR/char_npc.txt" << 'PROMPT'
16-bit pixel art sprite of a friendly merchant NPC.
- Wearing brown robes and hood
- Carrying a large backpack with items
- Warm smile, welcoming pose
- Style: retro RPG shopkeeper
- NO background - solid white only
- Clean pixel art, 48x64 sprite size
PROMPT

    cat > "$PROMPTS_DIR/char_mage.txt" << 'PROMPT'
16-bit pixel art sprite of a wizard mage character.
- Purple robes with gold trim and stars
- Pointed wizard hat
- Holding a glowing staff with crystal
- Long white beard
- Casting pose with magic sparkles
- Style: classic fantasy RPG
- NO background - solid white
- Clean retro pixel art style
PROMPT

    run_test "Hero character" \
        "$NANOBANANA generate --prompt-file '$PROMPTS_DIR/char_hero.txt' -o '$TESTS_DIR/game-assets/characters/hero.png'" \
        "$TESTS_DIR/game-assets/characters/hero.png"

    run_test "Slime enemy" \
        "$NANOBANANA generate --prompt-file '$PROMPTS_DIR/char_enemy.txt' -o '$TESTS_DIR/game-assets/characters/slime.png'" \
        "$TESTS_DIR/game-assets/characters/slime.png"

    run_test "Merchant NPC" \
        "$NANOBANANA generate --prompt-file '$PROMPTS_DIR/char_npc.txt' -o '$TESTS_DIR/game-assets/characters/merchant.png'" \
        "$TESTS_DIR/game-assets/characters/merchant.png"

    run_test "Wizard mage" \
        "$NANOBANANA generate --prompt-file '$PROMPTS_DIR/char_mage.txt' -o '$TESTS_DIR/game-assets/characters/mage.png'" \
        "$TESTS_DIR/game-assets/characters/mage.png"

    section_header "Game Assets - Backgrounds (Parallax Layers)"

    cat > "$PROMPTS_DIR/bg_sky.txt" << 'PROMPT'
Pixel art sky background layer for parallax scrolling.
- Gradient from light blue at bottom to deeper blue at top
- Fluffy white clouds scattered across
- Style: 16-bit retro game background
- Seamless horizontal tiling
- Soft, peaceful daytime sky
- Size: wide landscape format 16:9
PROMPT

    cat > "$PROMPTS_DIR/bg_mountains.txt" << 'PROMPT'
Pixel art distant mountains background layer.
- Purple/blue misty mountains silhouette
- Multiple mountain peaks at different heights
- Subtle color gradient for depth
- Style: 16-bit platformer background
- Seamless horizontal tiling
- Designed for parallax middle layer
- Size: wide 16:9 format
PROMPT

    cat > "$PROMPTS_DIR/bg_forest.txt" << 'PROMPT'
Pixel art forest treeline background layer.
- Dense green forest canopy
- Various tree shapes and sizes
- Darker at bottom, lighter leaves at top
- Style: 16-bit adventure game
- Seamless horizontal tiling
- Foreground parallax layer
- Size: wide 16:9 format
PROMPT

    cat > "$PROMPTS_DIR/bg_city.txt" << 'PROMPT'
Pixel art cyberpunk city skyline at night.
- Tall skyscrapers with neon lights
- Purple and blue color scheme
- Glowing windows and signs
- Flying vehicles in distance
- Style: retro sci-fi game
- Seamless horizontal tiling
- Size: wide 16:9 format
PROMPT

    run_test "Sky layer" \
        "$NANOBANANA generate --prompt-file '$PROMPTS_DIR/bg_sky.txt' -o '$TESTS_DIR/game-assets/backgrounds/sky.png' --aspect-ratio 16:9" \
        "$TESTS_DIR/game-assets/backgrounds/sky.png"

    run_test "Mountains layer" \
        "$NANOBANANA generate --prompt-file '$PROMPTS_DIR/bg_mountains.txt' -o '$TESTS_DIR/game-assets/backgrounds/mountains.png' --aspect-ratio 16:9" \
        "$TESTS_DIR/game-assets/backgrounds/mountains.png"

    run_test "Forest layer" \
        "$NANOBANANA generate --prompt-file '$PROMPTS_DIR/bg_forest.txt' -o '$TESTS_DIR/game-assets/backgrounds/forest.png' --aspect-ratio 16:9" \
        "$TESTS_DIR/game-assets/backgrounds/forest.png"

    run_test "Cyberpunk city" \
        "$NANOBANANA generate --prompt-file '$PROMPTS_DIR/bg_city.txt' -o '$TESTS_DIR/game-assets/backgrounds/city.png' --aspect-ratio 16:9" \
        "$TESTS_DIR/game-assets/backgrounds/city.png"

    section_header "Game Assets - Buildings"

    cat > "$PROMPTS_DIR/building_castle.txt" << 'PROMPT'
Pixel art medieval castle for 2D game.
- Stone walls with battlements
- Central tower with flag
- Wooden gate entrance
- Style: 16-bit fantasy RPG
- Front-facing view
- Solid white background (for transparency)
- Clean pixel art, limited palette
PROMPT

    cat > "$PROMPTS_DIR/building_house.txt" << 'PROMPT'
Pixel art cozy village house.
- Thatched roof cottage style
- Wooden walls, stone chimney with smoke
- Small windows with warm light
- Flower boxes and wooden door
- Style: 16-bit RPG village
- Solid white background
- Clean retro pixel art
PROMPT

    cat > "$PROMPTS_DIR/building_shop.txt" << 'PROMPT'
Pixel art item shop building for RPG.
- Wooden storefront with awning
- Sign hanging with potion bottle icon
- Display window showing items
- Welcoming open door
- Style: classic JRPG shop
- Solid white background
- Retro 16-bit pixel art
PROMPT

    cat > "$PROMPTS_DIR/building_tavern.txt" << 'PROMPT'
Pixel art medieval tavern/inn building.
- Two-story wooden building
- Hanging sign with mug icon
- Warm glowing windows
- Barrel and crates outside
- Style: fantasy RPG tavern
- Solid white background
- 16-bit retro pixel art style
PROMPT

    run_test "Castle" \
        "$NANOBANANA generate --prompt-file '$PROMPTS_DIR/building_castle.txt' -o '$TESTS_DIR/game-assets/buildings/castle.png'" \
        "$TESTS_DIR/game-assets/buildings/castle.png"

    run_test "Village house" \
        "$NANOBANANA generate --prompt-file '$PROMPTS_DIR/building_house.txt' -o '$TESTS_DIR/game-assets/buildings/house.png'" \
        "$TESTS_DIR/game-assets/buildings/house.png"

    run_test "Item shop" \
        "$NANOBANANA generate --prompt-file '$PROMPTS_DIR/building_shop.txt' -o '$TESTS_DIR/game-assets/buildings/shop.png'" \
        "$TESTS_DIR/game-assets/buildings/shop.png"

    run_test "Tavern" \
        "$NANOBANANA generate --prompt-file '$PROMPTS_DIR/building_tavern.txt' -o '$TESTS_DIR/game-assets/buildings/tavern.png'" \
        "$TESTS_DIR/game-assets/buildings/tavern.png"

    section_header "Game Assets - Items & Collectibles"

    cat > "$PROMPTS_DIR/item_sword.txt" << 'PROMPT'
Pixel art legendary sword weapon.
- Glowing blue blade with magical aura
- Golden hilt with gem
- Style: 16-bit RPG inventory item
- Solid white background
- Clean pixel art, 32x32 icon size
- Slight glow effect around blade
PROMPT

    cat > "$PROMPTS_DIR/item_potion.txt" << 'PROMPT'
Pixel art health potion bottle.
- Glass bottle with red liquid
- Cork stopper
- Bubbles inside, magical glow
- Style: classic RPG consumable
- Solid white background
- Clean pixel art, 32x32 size
PROMPT

    cat > "$PROMPTS_DIR/item_coin.txt" << 'PROMPT'
Pixel art gold coin collectible.
- Shiny golden coin
- Star or gem emblem in center
- Sparkle effects around it
- Style: platformer game collectible
- Solid white background
- Clean pixel art, 16x16 size
PROMPT

    cat > "$PROMPTS_DIR/item_chest.txt" << 'PROMPT'
Pixel art treasure chest.
- Wooden chest with golden trim
- Slightly open with golden glow inside
- Gems and coins visible
- Style: RPG loot container
- Solid white background
- 16-bit retro pixel art
PROMPT

    cat > "$PROMPTS_DIR/item_key.txt" << 'PROMPT'
Pixel art golden key item.
- Ornate golden key
- Decorative handle with swirl design
- Magical sparkle effect
- Style: adventure game key item
- Solid white background
- Clean pixel art, 32x32 size
PROMPT

    run_test "Magic sword" \
        "$NANOBANANA generate --prompt-file '$PROMPTS_DIR/item_sword.txt' -o '$TESTS_DIR/game-assets/items/sword.png'" \
        "$TESTS_DIR/game-assets/items/sword.png"

    run_test "Health potion" \
        "$NANOBANANA generate --prompt-file '$PROMPTS_DIR/item_potion.txt' -o '$TESTS_DIR/game-assets/items/potion.png'" \
        "$TESTS_DIR/game-assets/items/potion.png"

    run_test "Gold coin" \
        "$NANOBANANA generate --prompt-file '$PROMPTS_DIR/item_coin.txt' -o '$TESTS_DIR/game-assets/items/coin.png'" \
        "$TESTS_DIR/game-assets/items/coin.png"

    run_test "Treasure chest" \
        "$NANOBANANA generate --prompt-file '$PROMPTS_DIR/item_chest.txt' -o '$TESTS_DIR/game-assets/items/chest.png'" \
        "$TESTS_DIR/game-assets/items/chest.png"

    run_test "Golden key" \
        "$NANOBANANA generate --prompt-file '$PROMPTS_DIR/item_key.txt' -o '$TESTS_DIR/game-assets/items/key.png'" \
        "$TESTS_DIR/game-assets/items/key.png"

    # Post-process: Remove white backgrounds from characters and items
    section_header "Game Assets - Background Removal"

    for img in "$TESTS_DIR/game-assets/characters"/*.png; do
        if [ -f "$img" ]; then
            base=$(basename "$img" .png)
            run_test_local "Transparent: $base" \
                "$NANOBANANA transparent make '$img' -o '$TESTS_DIR/game-assets/characters/${base}_transparent.png' --color white --tolerance 20" \
                "$TESTS_DIR/game-assets/characters/${base}_transparent.png"
        fi
    done

    for img in "$TESTS_DIR/game-assets/items"/*.png; do
        if [ -f "$img" ]; then
            base=$(basename "$img" .png)
            run_test_local "Transparent: $base" \
                "$NANOBANANA transparent make '$img' -o '$TESTS_DIR/game-assets/items/${base}_transparent.png' --color white --tolerance 20" \
                "$TESTS_DIR/game-assets/items/${base}_transparent.png"
        fi
    done

    for img in "$TESTS_DIR/game-assets/buildings"/*.png; do
        if [ -f "$img" ]; then
            base=$(basename "$img" .png)
            run_test_local "Transparent: $base" \
                "$NANOBANANA transparent make '$img' -o '$TESTS_DIR/game-assets/buildings/${base}_transparent.png' --color white --tolerance 20" \
                "$TESTS_DIR/game-assets/buildings/${base}_transparent.png"
        fi
    done
}

test_transform() {
    section_header "Image Transformations (Local)"

    # Find a source image
    local source=""
    for img in "$TESTS_DIR/basic_01_simple.png" "$TESTS_DIR"/game-assets/items/*.png; do
        if [ -f "$img" ]; then
            source="$img"
            break
        fi
    done

    if [ -z "$source" ]; then
        echo -e "  ${YELLOW}No source image found. Run 'basic' or 'game-assets' tests first.${NC}"
        ((SKIPPED+=5))
        return
    fi

    run_test_local "Resize 128x128" \
        "$NANOBANANA transform '$source' -o '$TESTS_DIR/transform_01_resize.png' --resize 128x128" \
        "$TESTS_DIR/transform_01_resize.png"

    run_test_local "Resize 50%" \
        "$NANOBANANA transform '$source' -o '$TESTS_DIR/transform_02_half.png' --resize 50%" \
        "$TESTS_DIR/transform_02_half.png"

    run_test_local "Rotate 90°" \
        "$NANOBANANA transform '$source' -o '$TESTS_DIR/transform_03_rotate.png' --rotate 90" \
        "$TESTS_DIR/transform_03_rotate.png"

    run_test_local "Flip vertical" \
        "$NANOBANANA transform '$source' -o '$TESTS_DIR/transform_04_flip.png' --flip" \
        "$TESTS_DIR/transform_04_flip.png"

    run_test_local "Crop region" \
        "$NANOBANANA transform '$source' -o '$TESTS_DIR/transform_05_crop.png' --crop 100,100,400,400" \
        "$TESTS_DIR/transform_05_crop.png"
}

test_transparency() {
    section_header "Transparency Operations"

    local source="$TESTS_DIR/complex_02_sticker.png"
    if [ ! -f "$source" ]; then
        source=$(find "$TESTS_DIR" -name "*.png" -type f | head -1)
    fi

    if [ -z "$source" ] || [ ! -f "$source" ]; then
        echo -e "  ${YELLOW}No source image found. Run other tests first.${NC}"
        ((SKIPPED+=2))
        return
    fi

    run_test_local "Inspect transparency" \
        "$NANOBANANA transparent inspect '$source'" \
        "$source"

    run_test_local "Remove white bg" \
        "$NANOBANANA transparent make '$source' -o '$TESTS_DIR/transparency_01_removed.png' --color white --tolerance 15" \
        "$TESTS_DIR/transparency_01_removed.png"
}

test_combine() {
    section_header "Image Combining"

    # Find source images
    local sources=()
    for img in "$TESTS_DIR/transform_01_resize.png" "$TESTS_DIR"/game-assets/items/*_transparent.png; do
        if [ -f "$img" ]; then
            sources+=("$img")
            [ ${#sources[@]} -ge 4 ] && break
        fi
    done

    if [ ${#sources[@]} -lt 2 ]; then
        echo -e "  ${YELLOW}Not enough source images. Run 'transform' or 'game-assets' tests first.${NC}"
        ((SKIPPED+=3))
        return
    fi

    run_test_local "Horizontal strip" \
        "$NANOBANANA combine '${sources[0]}' '${sources[1]}' -o '$TESTS_DIR/combine_01_horizontal.png' --direction horizontal --gap 10" \
        "$TESTS_DIR/combine_01_horizontal.png"

    run_test_local "Vertical stack" \
        "$NANOBANANA combine '${sources[0]}' '${sources[1]}' -o '$TESTS_DIR/combine_02_vertical.png' --direction vertical --gap 10" \
        "$TESTS_DIR/combine_02_vertical.png"

    if [ ${#sources[@]} -ge 4 ]; then
        run_test_local "2x2 grid" \
            "$NANOBANANA combine '${sources[0]}' '${sources[1]}' '${sources[2]}' '${sources[3]}' -o '$TESTS_DIR/combine_03_grid.png' --direction grid --columns 2 --gap 5" \
            "$TESTS_DIR/combine_03_grid.png"
    fi
}

test_json() {
    section_header "JSON Output"

    echo -n "  Testing: JSON format... "
    local output
    output=$($NANOBANANA generate "simple test" -o "$TESTS_DIR/json_test.png" --json 2>&1)

    if echo "$output" | grep -q '"success"'; then
        echo -e "${GREEN}✓ PASS${NC}"
        echo "$output" > "$TESTS_DIR/json_output.json"
        ((PASSED++))
    else
        echo -e "${RED}✗ FAIL${NC}"
        ((FAILED++))
    fi

    sleep $API_DELAY
}

test_edge_cases() {
    section_header "Edge Cases"

    run_test "Unicode prompt" \
        "$NANOBANANA generate '日本の桜 cherry blossom 🌸 beautiful' -o '$TESTS_DIR/edge_01_unicode.png'" \
        "$TESTS_DIR/edge_01_unicode.png"

    run_test_optional "Special characters" \
        "$NANOBANANA generate 'A mathematical formula: E=mc² and symbols @#&*' -o '$TESTS_DIR/edge_02_special.png'" \
        "$TESTS_DIR/edge_02_special.png"

    run_test_optional "Very long prompt" \
        "$NANOBANANA generate 'A highly detailed fantasy scene featuring a majestic dragon with iridescent scales perched atop an ancient stone tower overlooking a vast medieval kingdom with rolling hills, dense forests, winding rivers, and distant snow-capped mountains under a dramatic sunset sky with swirling clouds painted in shades of orange, purple, and gold' -o '$TESTS_DIR/edge_03_long.png'" \
        "$TESTS_DIR/edge_03_long.png"
}

#═══════════════════════════════════════════════════════════════════════════════
# SUMMARY
#═══════════════════════════════════════════════════════════════════════════════

show_summary() {
    echo ""
    echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║                      TEST SUMMARY                          ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "  ${GREEN}Passed:${NC}  $PASSED"
    echo -e "  ${RED}Failed:${NC}  $FAILED"
    echo -e "  ${YELLOW}Skipped:${NC} $SKIPPED"
    echo ""

    # Count generated files
    local file_count=$(find "$TESTS_DIR" -name "*.png" -type f 2>/dev/null | wc -l | tr -d ' ')
    echo -e "  Generated files: ${CYAN}$file_count${NC}"
    echo ""

    echo -e "${CYAN}Output locations:${NC}"
    echo "  $TESTS_DIR/"
    echo "  $TESTS_DIR/game-assets/characters/"
    echo "  $TESTS_DIR/game-assets/backgrounds/"
    echo "  $TESTS_DIR/game-assets/buildings/"
    echo "  $TESTS_DIR/game-assets/items/"
    echo ""
    echo -e "${CYAN}To view results:${NC}"
    echo "  open $TESTS_DIR/   # macOS"
    echo ""
}

#═══════════════════════════════════════════════════════════════════════════════
# MAIN
#═══════════════════════════════════════════════════════════════════════════════

main() {
    # Check if binary exists
    if [ ! -f "$NANOBANANA" ]; then
        echo -e "${RED}Error: nanobanana binary not found. Run 'make build' first.${NC}"
        exit 1
    fi

    # No arguments - show help
    if [ $# -eq 0 ]; then
        show_help
        exit 0
    fi

    # Setup directories
    setup_dirs

    echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║         Nanobanana CLI Visual Test Suite                   ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "Output directory: ${YELLOW}$TESTS_DIR/${NC}"

    # Determine which tests to run
    local tests_to_run=()

    if [ "$1" = "all" ]; then
        tests_to_run=($TEST_ORDER)
    else
        for arg in "$@"; do
            if [ -n "${TEST_GROUPS[$arg]}" ]; then
                tests_to_run+=("$arg")
            else
                echo -e "${RED}Unknown test group: $arg${NC}"
                echo "Run './test.sh' to see available groups."
                exit 1
            fi
        done
    fi

    echo -e "Running tests: ${CYAN}${tests_to_run[*]}${NC}"

    # Run selected tests
    for test in "${tests_to_run[@]}"; do
        case "$test" in
            basic)       test_basic ;;
            complex)     test_complex ;;
            stdin)       test_stdin ;;
            icons)       test_icons ;;
            patterns)    test_patterns ;;
            game-assets) test_game_assets ;;
            transform)   test_transform ;;
            transparency) test_transparency ;;
            combine)     test_combine ;;
            json)        test_json ;;
            edge-cases)  test_edge_cases ;;
        esac
    done

    show_summary

    # Exit with error if any tests failed
    [ $FAILED -gt 0 ] && exit 1
    exit 0
}

main "$@"
