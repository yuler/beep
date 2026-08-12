# beep brand assets

Organized to match the brand sheet layout (`brand/brand-sheet.png`).

## Structure

```
materials/
├── brand/
│   └── brand-sheet.png          # Source brand specification (AI-generated)
├── svg/
│   ├── icon-light.svg           # Black symbol (for light backgrounds)
│   └── icon-dark.svg            # White symbol (for dark backgrounds)
├── png/
│   ├── logo-horizontal-light.png
│   └── icons/
│       ├── icon-light-64.png    # Black symbol, 64px
│       ├── icon-dark-64.png     # White symbol on black, 64px
│       └── app-icon-64.png      # App icon (symbol on white squircle), 64px
└── app/
    └── icon-64.png              # App store icon, 64px
```

## Naming convention

| Asset | Light theme | Dark theme |
|-------|-------------|------------|
| Symbol (SVG) | `svg/icon-light.svg` | `svg/icon-dark.svg` |
| Symbol (PNG) | `png/icons/icon-light-{size}.png` | `png/icons/icon-dark-{size}.png` |
| Horizontal logo | `png/logo-horizontal-light.png` | `png/logo-horizontal-dark.png` *(not yet extracted)* |
| App icon | `app/icon-{size}.png` | — |

Tagline: **Write less, do more.**
