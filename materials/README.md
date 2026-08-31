# beep brand assets

Logo concept: **Option 13** — minimalist line-art bell with notification arcs and base dash.

Tagline: **Write less, do more.**

## Structure

```
materials/
├── svg/                         # Source vectors (edit these, then regenerate PNGs)
│   ├── icon-light.svg           # Symbol on transparent background (dark ink)
│   ├── icon-dark.svg            # Symbol on transparent background (light ink)
│   ├── logo-light.svg           # Horizontal lockup with wordmark (light background)
│   ├── logo-dark.svg            # Horizontal lockup with wordmark (dark background)
│   ├── icon-light-square.svg    # App icon squircle (white tile, dark bell)
│   └── icon-dark-square.svg     # App icon squircle (dark tile, light bell)
├── png/
│   ├── logo-horizontal-light.png
│   ├── logo-horizontal-dark.png
│   └── icons/
│       ├── icon-light-64.png
│       ├── icon-dark-64.png
│       ├── icon-light-square-64.png
│       └── icon-dark-square-64.png
├── app/
│   └── icon-64.png              # Primary app-store / PWA icon
└── favicon/
    ├── favicon-16.png
    ├── favicon-32.png
    ├── icon-192.png
    └── icon-512.png
```

## Naming convention

| Asset              | Light theme                         | Dark theme                          |
|--------------------|-------------------------------------|-------------------------------------|
| Symbol (SVG)       | `svg/icon-light.svg`                | `svg/icon-dark.svg`                 |
| Symbol (PNG)       | `png/icons/icon-light-64.png`       | `png/icons/icon-dark-64.png`        |
| Horizontal logo    | `png/logo-horizontal-light.png`     | `png/logo-horizontal-dark.png`      |
| App icon squircle  | `svg/icon-light-square.svg`         | `svg/icon-dark-square.svg`          |
| App icon (export)  | `app/icon-64.png`                   | —                                   |

## Regenerating PNGs

From the repo root (requires `rsvg-convert`):

```bash
MAT=materials
rsvg-convert -w 1600 "$MAT/svg/logo-light.svg" -o "$MAT/png/logo-horizontal-light.png"
rsvg-convert -w 1600 "$MAT/svg/logo-dark.svg" -o "$MAT/png/logo-horizontal-dark.png"
rsvg-convert -w 64 -h 64 "$MAT/svg/icon-light.svg" -o "$MAT/png/icons/icon-light-64.png"
rsvg-convert -w 64 -h 64 -b '#111111' "$MAT/svg/icon-dark.svg" -o "$MAT/png/icons/icon-dark-64.png"
rsvg-convert -w 64 -h 64 "$MAT/svg/icon-light-square.svg" -o "$MAT/png/icons/icon-light-square-64.png"
rsvg-convert -w 64 -h 64 "$MAT/svg/icon-dark-square.svg" -o "$MAT/png/icons/icon-dark-square-64.png"
cp "$MAT/png/icons/icon-light-square-64.png" "$MAT/app/icon-64.png"
rsvg-convert -w 16 -h 16 "$MAT/svg/icon-light.svg" -o "$MAT/favicon/favicon-16.png"
rsvg-convert -w 32 -h 32 "$MAT/svg/icon-light.svg" -o "$MAT/favicon/favicon-32.png"
rsvg-convert -w 192 -h 192 "$MAT/svg/icon-light-square.svg" -o "$MAT/favicon/icon-192.png"
rsvg-convert -w 512 -h 512 "$MAT/svg/icon-light-square.svg" -o "$MAT/favicon/icon-512.png"
```
