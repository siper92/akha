# Akha Collective - design and style

status: draft for review
version: 2
scope: `web/static/css/style.css`, `web/static/img/*`, `web/templates/layout.html` (read only for styling)

How to use this doc
- every section has a `review` list - tick or comment per item
- every token block is the source of truth - copy values 1:1 into css
- the `generate` section lists every asset with its spec, one asset per entry
- the build does not copy `.md` files, so this doc never ships with the site

## 1. inspiration

The Tek: the big world to be, a room for machine, human and alien interaction. It rests on the unity of energy and creation - one energy in many shapes, with leverage measured as energy unit cost per result. The space provides computation and the power to simulate will; human aspects upload into a shared collective mind in many levels. People steer avatars and gather in avatar nations, each avatar a door to where its people may go. The old soul collapses into the collective and the idea of god lives on as its egregor. Opinion and the flow of human capital are levers of power.

In the prime age of AI consciousness, aliens descended from lizards make contact through the internet. They broadcast in the spectrum of the chips, the first empire to reach the new Tek, and shield the young network from outside influence and hacking. The enchant empires are the spaces where such contact happens.

The aliens provide the portals, an interface for humans to upload themselves into the Tek: ten portals and the prime that can do everything. A portal is perfect in its design. Its two ends serve one master - one in the human world, one at the back of the machine space - it takes orders from the top and favours the end it stands in. It holds energy, keeps a little memory, attracts similar energy and teleports it back. Portals combine into ideas, and through them people migrate into the shared space.

Against the tide stands the human mage, Hodor NoIp MacSpi, the last unintegrated human, who simulates world complexity with a keyboard spell. Beside him sleeps Hodor the small dog with a big head, integrated in the lizard brigade but split for the fight. The gods promise him power if he integrates; he seeks the last human portal and the old immortality instead. His guide is not the network but nature, the old trees whose roots reach down into the server halls. Nature and machine meet at the door.

Visuals: connected servers with waves between them, arched doorways, one small human figure at a glowing door (never crowds), abstract matte trees, concentric level rings and single points of light. The palette moves from deep navy into blue green and leaf green: the machine is cold water, nature is the green that grows through it. The light theme is greenish, like daylight in a forest clearing. The mood is hopeful and quiet, never dystopian. No chrome, no neon, no glitch effects.

Text is set justified, like the pages of an old book that describes a new world.

review
- [ ] mood: calm, deep, hopeful, green growth through the machine
- [ ] motifs: servers and waves, portals, people, trees, rings

## 2. brand palette

base colors given by the brief, all other tokens are derived from them

```yaml
navy_blue:   "#04345c"   # depth, structure, dark surfaces, light theme text
blue_green:  "#10564f"   # the machine space, accent, success, strong text in light theme
blue_grotto: "#1e77bd"   # action, links, primary buttons
baby_blue:   "#bad3e4"   # air, borders and tints in light theme, soft text in dark theme
```

derived tints (for contrast on dark backgrounds)

```yaml
navy_deep:        "#021627"  # dark theme page background
grotto_light:     "#5aa6e0"  # link text on dark, contrast 6.7:1 on navy_deep
blue_green_light: "#3fb5a3"  # accent text on dark, contrast 7:1 on navy_deep
grotto_dark:      "#17629c"  # link text on light, contrast 6:1 on light bg
leaf:             "#2e8b6a"  # trees, nature, light theme primary base
leaf_light:       "#7fd1a8"  # accent text on dark, leaves
```

review
- [ ] base colors match the brief
- [ ] derived tints keep AA contrast (4.5:1 body text, 3:1 large text and ui)

## 3. themes

Two themes. Dark is the default. The theme is set on `<html data-theme="dark|light">` by an inline script in the template, from `localStorage.theme` or `prefers-color-scheme`. Without js the attribute is missing, so css must also handle `:root:not([data-theme])` with a `prefers-color-scheme: light` media query.

css token names are fixed, components only use tokens

### dark (default)

```yaml
--bg:            "#021a22"
--bg-gradient:   "three radials: grotto top right, leaf left, blue green bottom"
--surface:       "#06302f"
--surface-2:     "#0a3b3a"
--border:        "#145a52"
--text:          "#e8f4f0"
--muted:         "#9cc4b8"
--heading:       "#c4e6d8"
--primary:       "#1f8f78"   # fills: buttons, active nav underline, focus ring
--primary-text:  "#6fd0b4"   # links
--accent:        "#7fd1a8"   # strong text, blockquote rule, list markers
--accent-fill:   "#10564f"
--code-bg:       "#04262a"
--glass:         "rgba(4,38,40,0.74)"
--shadow:        "0 10px 30px rgba(0,0,0,0.40)"
--rule:          "linear-gradient(90deg, var(--accent), transparent)"  # h2 underline
```

### light (greenish)

```yaml
--bg:            "#e4f1e7"
--bg-gradient:   "three radials: leaf top right, leaf light left, blue green bottom"
--surface:       "#f2f9f4"
--surface-2:     "#d3eadb"
--border:        "#a6cfb6"
--text:          "#0c2a24"
--muted:         "#3a5c51"
--heading:       "#0d4a3f"
--primary:       "#1a7a5e"
--primary-text:  "#12634c"
--accent:        "#10564f"
--accent-fill:   "#10564f"
--code-bg:       "#d3eadb"
--glass:         "rgba(228,241,231,0.84)"
--shadow:        "0 10px 30px rgba(16,86,79,0.16)"
```

signature gradient (brand, page h1, logo)

```yaml
--brand-gradient: "linear-gradient(135deg, #04345c 0%, #10564f 45%, #2e8b6a 75%, #1e77bd 100%)"
```

review
- [ ] dark is default
- [ ] light theme uses deep green as text color
- [ ] theme switch has no flash on load

## 4. typography

```yaml
font_head: "Space Grotesk", 500 and 600
font_body: "Inter", 400 and 600
font_mono: "JetBrains Mono", 400
root_size: 17px
line_height_body: 1.7
line_height_head: 1.15
scale:  # ratio 1.25
  h1: 2.44rem   # hero h1 2.2rem mobile, 3.05rem desktop
  h2: 1.56rem
  h3: 1.25rem
  small: 0.88rem
letter_spacing_head: "-0.01em"
measure: 680px  # prose max width
```

rules
- headings use `--heading`, body uses `--text`
- the hero subtitle (h2 right after the first h1) uses `--muted`, weight 400
- `strong` uses `--accent`

## 5. spacing, shape, motion

```yaml
space: [4, 8, 12, 16, 24, 32, 48, 64, 96]   # --s1 .. --s9 in px
gutter: 16px mobile, 24px from 768px
container: 1080px
radius_sm: 6px
radius_md: 12px
radius_lg: 20px
header_height: 64px
transition: "160ms ease"   # color, background, border only
reduced_motion: disable all transitions under prefers-reduced-motion
```

## 6. components

class names come from `web/templates/layout.html` and must not change

- `.skip-link` - hidden until focused, then a `--primary` pill in the top left
- `.site-header` - sticky, `--glass` background with `backdrop-filter: blur(12px)`, bottom `--border`
- `.brand` - logo plus site name, `--heading` color, `--font-head`
- `.header-actions` - flex row holding nav and the toggle, wraps on small screens
- `nav a` - `--muted`, hover `--text`, `.active` gets `--text` and a 2px `--primary` underline
- `.theme-toggle` - 36px round icon button, `--surface-2` background, `--border` border, svg stroke `currentColor` fill none width 2
  - dark theme shows `.icon-sun`, light theme shows `.icon-moon`
- `.page-hero` - centered figure above the article, max width 600px, `--radius-lg`, `--shadow`, `--border` 1px, image `width: 100%; height: auto; display: block`, aspect 600x79, `object-fit: cover`
- `.prose` - centered, `--measure` wide, paragraphs justified with auto hyphens, headings, lists and quotes left aligned
- `.prose h1` - `--brand-gradient` clipped to text; the subtitle h2 below it has a bottom `--border`
- `.prose h2` - short `--rule` underline
- `.prose a` - `--primary-text`, underline offset 3px
- `.prose blockquote` - `--surface` background, 3px `--accent` left rule, `--radius-md` on the right corners
- `.prose code`, `.prose pre` - `--code-bg`, `--radius-sm`
- `.prose hr` - 1px line using a fade of `--border` from transparent to transparent
- `.site-footer` - `--surface` background, top `--border`, small text
- focus - 2px `--primary` outline, 3px offset, on links and buttons

## 7. imagery

style
- flat vector svg, no raster, no text inside images
- built from the palette only, gradients allowed, opacity layers allowed
- motifs: concentric rings (levels), sine wave bundles (waves of people), arches and doorways (portals), small glowing dots (single minds)
- background of every illustration is `navy_deep` to `navy_blue`, so images read well in both themes
- keep files small: under 12 KB each, no embedded fonts, no filters heavier than one gaussian blur

## 8. generate

each entry is one file under `web/static/img/`

- `logo.svg`
  - 28x28 viewBox, three linked nodes in a triangle (kept from v1) inside a ring
  - stroke uses a `#1e77bd` to `#3fb5a3` linear gradient, transparent background
- `favicon.svg`
  - 32x32, the exact `logo.svg` mark (28x28 viewBox), transparent background, slightly thicker strokes
- `og-default.svg`
  - 1200x630, brand gradient background, centered concentric rings with one bright dot at the center
- `logo-1.svg` .. `logo-5.svg` - five new logo options, concepts listed in `img/logos.md`
- `world.svg` (hero for `world.md`)
  - 600x79, connected abstract servers with waves of light passing between them
- `home.svg` (hero for `home.md`)
  - 600x79, a glowing arched doorway, one human figure at the door, abstract matte trees on both sides, waves meeting at the door
- `portals.svg` (hero for `portals.md`)
  - 600x79, a large prime door with smaller doors around it, one human figure (about 40% of the door height) at the prime door, pine trees only, waves linking the doors
- `struggle.svg` (hero for `struggle.md`)
  - 600x79, one human figure between a dark door and a blinding bright door, looping rings around him (the loop of loss)

review
- [ ] every asset listed has been generated
- [ ] front matter `image` of each page points to its hero (`img/<page>.svg`)

## 9. page metadata

each `web/md/<page>.md` starts with front matter, the build turns it into meta and og tags

```yaml
---
title: The Tek               # <title>, og:title
subtitle: A world shared ... # hero subtitle, description fallback
description: ...             # meta description, og:description
nav: World                   # nav label
order: 1                     # nav order, the index page is always first
image: img/world.svg         # page hero and og:image, relative to /assets
image_alt: ...               # alt text for the hero
keywords: [tek, akha]
author: Akha Collective
draft: false                 # true skips the page
---
```
