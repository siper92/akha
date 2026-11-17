# step 1 - design and style (sub agent 0)

## idea and inspiration

- a world where humans and machines share one distributed network
- feeling: calm, clean, technical, quiet confidence - a terminal at night
- inspiration: network maps, server rack status lights, minimal sci-fi ui
- focus: reading - the story is the product, ui stays out of the way

## style rules

- flat: no gradients, no shadows, no blur, no rounded cards
- radius 4px max (buttons, code), everything else square
- 1px borders for separation, never boxes inside boxes
- one accent per view, accent is for links and the active nav item

## color schema (dark, single theme)

- `--bg` #0E141B - page background
- `--surface` #151E28 - header, footer, blockquote background
- `--border` #233140 - lines and hr
- `--text` #E4EAF0 - body text
- `--muted` #8A9AAB - subtitle, footer, meta
- `--primary` #2EC4B6 - links, logo nodes, active nav underline
- `--accent` #FF9F1C - blockquote left border, strong highlights only
- `--code-bg` #1B2632
- define them as css custom properties on `:root`

## fonts

- headings: `Space Grotesk` 600
- body: `Inter` 400 / 600
- code: `JetBrains Mono` 400
- fallback: `system-ui, -apple-system, Segoe UI, Roboto, sans-serif` and `ui-monospace, monospace`

## size and rhythm

- base 17px, line-height 1.7 for body, 1.2 for headings
- scale 1.25: h3 1.25rem, h2 1.56rem, h1 2.44rem, hero h1 3.05rem
- spacing scale (px): 4, 8, 12, 16, 24, 32, 48, 64, 96
- paragraph gap 24px, section gap 64px

## layout

- header full width, inner container max 1080px, height 64px
- article column max 680px, centered
- side gutter 16px on phones, 24px from 768px
- one breakpoint: 768px
- hero h1 drops to 2.2rem under 768px

## design elements - only these may be used

- header: logo (svg) + site name on the left, nav on the right
- nav item: text link, muted, active item text color + 2px primary underline
- hero: first h1 and the h2 right after it (`.prose > h1:first-child`, `.prose > h1:first-child + h2`), h2 is muted and not bold
- prose: h1, h2, h3, p, a, strong, em, ul, ol, li, blockquote, hr, inline code, pre
- footer: site name, year, a one-line muted tagline
- icons: only `logo.svg` (three connected nodes, primary color stroke) - no icon fonts

## files to create

- `web/templates/layout.html` - go `html/template`, uses only the fields from the shared contract
- `web/static/css/style.css`
- `web/static/img/logo.svg`
- `md/world.md`, `md/home.md` - lorem ipsum in the shared md format

## pages and order

- 1 `world` at `/` - index page
- 2 `home` at `/home`
