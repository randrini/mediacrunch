# Design

<!-- impeccable:design-schema 1 -->

## World

The Darkroom — a photographic darkroom aesthetic where every compression job is an exposure decision. Quality is exposure time, progress is the enlarger countdown, the media grid is a contact sheet. Warm near-black ground with safelight-red as the single accent that illuminates action and data; warm paper-white text on a grayscale tonal ramp; ruled monospace data lines; contact-sheet grids with frame-number density; enlarger-bar progress fills with red glow; safelight-glow on active states.

## Palette

### Ground

| Token | Hex | Role |
|-------|-----|------|
| base | #0d0b0a | Page background, the darkroom floor |
| surface | #16120e | Card background, first elevation |
| elevated | #1f1a14 | Input background, table header, second elevation |
| highlight | #2a2218 | Hover state, active surface, third elevation |

### Borders

| Token | Hex | Role |
|-------|-----|------|
| border | #2a2218 | Default borders, table dividers, card outlines |
| border-strong | #3a3024 | Strong borders, focus rings, emphasized dividers |

### Accent — Safelight Red

| Token | Hex | Role |
|-------|-----|------|
| accent | #c8412a | Primary actions, active states, savings data, compressed badges |
| accent-hover | #e85d3a | Hover state for primary actions |
| accent-muted | rgba(200, 65, 42, 0.12) | Accent background for badges, selection bars, muted accents |

### Ink — Dark Text on Accent

| Token | Hex | Role |
|-------|-----|------|
| ink | #0d0b0a | Text on accent/danger/warning backgrounds (near-black, high contrast) |

### Paper — Warm Photographic White

| Token | Hex | Role |
|-------|-----|------|
| text-primary | #f5f0e8 | Primary body text, headings |
| text-secondary | #9a9082 | Secondary text, labels, descriptions |
| text-tertiary | #6b6358 | Tertiary text, placeholders, muted labels |

### Zone System Status

| Token | Hex | Role |
|-------|-----|------|
| success | #8a9a6b | Developed/compressed state, positive indicators |
| warning | #c4943a | Amber filter, caution states, Plex actions |
| danger | #a02828 | Overexposed/failed state, destructive actions |

## Typography

- **UI text:** Saira (weights 300–700), system-ui fallback. Clean grotesk with technical precision — the instrument label of the darkroom.
- **Data text:** JetBrains Mono (weights 400–600). Monospace for all numbers, byte sizes, quality values, timestamps, IDs, and tabular data. Uses tabular-nums and font-variant-numeric: tabular-nums everywhere.

### Scale

- Page titles: text-2xl font-bold tracking-tight
- Section titles: text-lg font-semibold
- Card titles: text-sm font-semibold
- Body text: text-sm
- Data values: text-xs font-mono tabular-nums
- Labels: text-[10px] or text-[11px] uppercase tracking-wider text-text-tertiary
- Stat values: text-[16px] or text-sm font-mono tabular-nums

## Component Language

### Cards

`card-glass` — bg-surface with border-border rounded-lg. No glass, no blur, no gradient. Solid surface with a hairline warm-dark border. Hover adds border-border-strong.

### Size Badges

Three-zone system using the zone system (photographic tonal ramp):
- `.size-badge-large` — red-tinted (bg-danger/15 text-danger-variant) for files > 10 MB
- `.size-badge-medium` — amber-tinted (bg-warning/15 text-warning-variant) for files > 1 MB
- `.size-badge-small` — green-tinted (bg-success/15 text-success-variant) for smaller files

### Buttons

- **Primary:** bg-accent text-ink hover:bg-accent-hover — safelight red with near-black text
- **Secondary:** bg-elevated text-text-secondary hover:bg-highlight — neutral surface
- **Destructive:** bg-danger text-ink hover:bg-danger/80 — overexposed red with near-black text
- **Muted accent:** bg-accent-muted text-accent hover:bg-accent/20 — for scan, media actions
- **Warning:** bg-warning/10 text-warning hover:bg-warning/20 — for lock, Plex-specific actions

### Inputs

bg-elevated border border-border rounded-md px-3 py-2 text-text-primary focus:ring-2 focus:ring-accent/50 focus:border-accent/50. Labels in text-sm font-medium text-text-secondary.

### Progress

`.enlarger-bar` — Linear gradient from accent to accent-hover with box-shadow glow (0 0 8px rgba(200, 65, 42, 0.4)). The safelight bar that measures exposure.

### Active States

`.safelight-glow` — box-shadow: 0 0 12px rgba(200, 65, 42, 0.15), inset 0 0 1px rgba(200, 65, 42, 0.1). Applied to running job indicators and active selection bars.

### Data Lines

`.ruled-line` — border-b border-border. Monospace data on hairline rules, like a contact sheet's frame numbers.

### Type Icons

Inline SVG icons in 2px stroke weight for media types: movie (film strip), series (tv), season (calendar), episode (film frame), collection (folder). No emoji.

## Navigation

Sticky top nav with backdrop blur. Brand mark: aperture/diaphragm SVG in accent color with font-mono "MediaCrunch". Active link: bg-accent-muted text-accent. Instance selector dropdown: bg-elevated with focus:ring-accent. Mobile: hamburger menu with slide-down panel.

## Responsive

Desktop: 3-column instance grid, full media table with all columns. Tablet: 2-column grid, card list for media. Mobile: 1-column grid, stacked cards with compact stat grids. All breakpoints maintain the contact-sheet density — compact, scannable, data-dense.

## Motion

One authored moment: the enlarger-bar progress fill with glow on active jobs. All other transitions are 150ms ease-out (transition-base). No scattered hover effects — the safelight glow is the only sustained animation.

## Browser Surfaces

- Text selection: rgba(200, 65, 42, 0.3) background with #f5f0e8 text
- Scrollbars: thin (5px), scrollbar-color #3a3024 #16120e
- Focus rings: ring-accent/50 on inputs, ring-offset-base
- Custom checkbox accent: accent-color var(--accent)