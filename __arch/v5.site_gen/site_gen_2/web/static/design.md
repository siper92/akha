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

A world where technology is the center of gravity. Access is not a password but a level: how far into the machine you are allowed to go, and how much of yourself you bring along. The site should feel like standing at the edge of that system, calm and ordered, with depth below the surface.

Humans enter the machine space in waves. Single minds arrive, merge into small groups, then into collectives large enough to read as weather. Identity is layered, and there is a quiet higher order of beings who speak slowly from cold server halls. Visuals use soft wave lines, concentric rings and overlapping translucent shapes to show many becoming one, without losing the single point of light that was a person.

Every page is a portal. Doorways, rings and gradients from deep navy into blue green suggest a step into a VR space. The mood is hopeful and quiet, never dystopian: ocean depth, harbor light, glass, gentle glow. No chrome, no neon, no glitch effects.

review
- [ ] mood: calm, deep, hopeful
- [ ] motifs: waves, rings, portals, layered translucency

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
```

review
- [ ] base colors match the brief
- [ ] derived tints keep AA contrast (4.5:1 body text, 3:1 large text and ui)

## 3. themes

Two themes. Dark is the default. The theme is set on `<html data-theme="dark|light">` by an inline script in the template, from `localStorage.theme` or `prefers-color-scheme`. Without js the attribute is missing, so css must also handle `:root:not([data-theme])` with a `prefers-color-scheme: light` media query.

css token names are fixed, components only use tokens

### dark (default)

```yaml
--bg:            "#021627"
--bg-gradient:   "radial-gradient(1200px 600px at 80% -10%, rgba(30,119,189,0.22), transparent 60%), radial-gradient(900px 500px at -10% 20%, rgba(16,86,79,0.35), transparent 60%)"
--surface:       "#052a49"
--surface-2:     "#04345c"
--border:        "#0f4a75"
--text:          "#eaf2f8"
--muted:         "#9fbdd3"
--heading:       "#bad3e4"
--primary:       "#1e77bd"   # fills: buttons, active nav underline, focus ring
--primary-text:  "#5aa6e0"   # links
--accent:        "#3fb5a3"   # strong text, blockquote rule
--accent-fill:   "#10564f"
--code-bg:       "#031f36"
--glass:         "rgba(5,42,73,0.72)"
--shadow:        "0 10px 30px rgba(0,0,0,0.35)"
```

### light

```yaml
--bg:            "#f5f9fc"
--bg-gradient:   "radial-gradient(1200px 600px at 80% -10%, rgba(30,119,189,0.12), transparent 60%), radial-gradient(900px 500px at -10% 20%, rgba(16,86,79,0.10), transparent 60%)"
--surface:       "#ffffff"
--surface-2:     "#e8f1f8"
--border:        "#bad3e4"
--text:          "#08243d"
--muted:         "#3d5a73"
--heading:       "#04345c"
--primary:       "#1e77bd"
--primary-text:  "#17629c"
--accent:        "#10564f"
--accent-fill:   "#10564f"
--code-bg:       "#e8f1f8"
--glass:         "rgba(255,255,255,0.78)"
--shadow:        "0 10px 30px rgba(4,52,92,0.12)"
```

signature gradient (brand, hero frames, logo)

```yaml
--brand-gradient: "linear-gradient(135deg, #04345c 0%, #10564f 50%, #1e77bd 100%)"
```

review
- [ ] dark is default
- [ ] light theme keeps navy as text color
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
- `.page-hero` - figure above the article, max width 1080px, `--radius-lg`, `--shadow`, `--border` 1px, image `width: 100%; height: auto; display: block`, aspect 1200x630
- `.prose` - centered, `--measure` wide
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
  - 32x32 viewBox, same mark as the logo on a `#04345c` rounded square, radius 7
- `og-default.svg`
  - 1200x630, brand gradient background, centered concentric rings with one bright dot at the center
- `world.svg` (hero for `world.md`)
  - 1200x630, a low harbor skyline in `navy_blue` at the bottom
  - above it a large portal made of 5 to 7 concentric rings in blue green and blue grotto, fading outward
  - small baby blue dots drifting toward the portal
- `home.svg` (hero for `home.md`)
  - 1200x630, one tall arched doorway glowing in blue grotto in the center
  - a single small figure silhouette in front of it
  - 11 soft wave lines in blue green and baby blue flowing in from both sides and meeting at the door

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
