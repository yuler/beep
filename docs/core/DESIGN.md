# Design System

A minimal, modern styling system for the Rails core app. Vanilla CSS, no build
step — inspired by Basecamp's [Fizzy](https://github.com/basecamp/fizzy) and the
[Modern CSS patterns in Campfire](https://dev.37signals.com/modern-css-patterns-and-techniques-in-campfire/)
write-up (and its [notes](https://yuler.dev/posts/modern-css-design/)). Linear/Geist
flavor: near-neutral grays plus a single indigo/violet accent, in light and dark
modes.

## Principles

- **oklch over hex/rgb.** Colors are defined as `Lightness Chroma Hue` so you can
  adjust shade without shifting hue, and add alpha trivially.
- **Two layers of tokens.** Raw `--lch-*` values feed semantic `--color-*` tokens.
  Components and utilities only ever consume the semantic layer.
- **Theme via one attribute.** `[data-theme]` on `<html>` switches everything.
  Default follows the OS; the user can override and the choice is remembered.
- **No Tailwind, no Sass.** Styles are hand-written CSS organized into cascade
  layers, shipped through Propshaft's no-build asset pipeline.

## Pipeline (no build step)

`core/app/assets/stylesheets/application.css` declares the layer order and
`@import`s every partial:

```css
@layer reset, base, components, utilities;

@import url("reset.css");
@import url("theme.css");
@import url("base.css");
/* …buttons, forms, tables, components, prose, header, footer, flash… */
```

Propshaft rewrites each `@import url("…")` to a fingerprinted, digested path at
request time, so there is no compile step and nothing to watch. The layout links
the single entry file:

```erb
<%= stylesheet_link_tag "application", "data-turbo-track": "reload" %>
```

## Cascade layers

Order is fixed in `application.css`; import order does not affect precedence:

1. `reset` — modern CSS reset (`.reset` file, sits in `@layer reset`).
2. `base` — element defaults (`body`, `a`, `h1`–`h6`, `code`, focus rings…).
3. `components` — reusable component classes (`.btn`, `.card`, `.alert`,
   `.table`, `.field`, `.menu`, `.header`, `.footer`…).
4. `utilities` — single-purpose helpers (`.flex`, `.gap`, `.txt-*`, `.pad`,
   `.margin-*`, `.stack`, spacing and color utilities).

Utilities win over components when both apply.

## Token architecture

### 1. Raw scale — `--lch-*`

Just the three oklch channels, no color function. Readable and easy to tweak.

```css
--lch-gray-100: 98% 0.003 285;
--lch-violet:   285 80% 60%;
--lch-green:    145 60% 45%;
```

Grays sit at hue `285`, the accent at hue `285` too — same family, so the UI
reads as one coherent palette. Accent is a Linear-style indigo/violet.

### 2. Semantic tokens — `--color-*`

Raw values wrapped in `oklch()`. This is the API you use.

| Token | Purpose |
| --- | --- |
| `--color-canvas` | Page background |
| `--color-shade` | Slightly recessed background |
| `--color-ink` | Primary text |
| `--color-ink-dark` | Secondary text |
| `--color-ink-medium` | Tertiary text, placeholders |
| `--color-ink-inverted` | Text on a dark/accent fill |
| `--color-border` | Default hairline borders |
| `--color-border-strong` | Emphasized borders |
| `--color-accent` | Primary accent (links, active, focus) |
| `--color-accent-hover` | Accent hover/active |
| `--color-white` | Solid white (text on accent) |
| `--color-link` | Inline link color |
| `--color-selected` | Selection / highlight |
| `--color-positive` | Success status |
| `--color-negative` | Destructive / error status |
| `--color-info` | Informational status |

Alpha is trivial with oklch:

```css
--color-accent-subtle: oklch(var(--lch-violet) / 0.12);
```

## Light & dark

- **Light** is the default, defined on `:root`.
- **Dark** is defined on `:root[data-theme="dark"]`.
- **No-JS fallback:** when `<html>` has no `data-theme`, `@media
  (prefers-color-scheme: dark)` applies the dark values, so the OS preference is
  honored even before any script runs.

Only the semantic `--color-*` values change between themes; component CSS never
needs color overrides for dark mode — it reads the tokens.

Files:

- `core/app/assets/stylesheets/theme.css` — all `--lch-*` and `--color-*` tokens
  and the light/dark blocks.

## Theme switching

1. **No-flash script** (`_head.html.erb`): runs before first paint. If
   `localStorage.theme` holds `light`/`dark` it sets `data-theme` on `<html>`;
   otherwise it leaves the attribute unset so the OS preference drives the theme
   live (via the `:not([data-theme])` media block in `theme.css`).
2. **Toggle** (`theme_controller.js`, button `.theme-toggle` in `_header.html.erb`):
   flips light ⇄ dark, persists to `localStorage`, and swaps the sun/moon icon.
3. **First visit** (nothing stored): follows `prefers-color-scheme`, and keeps
   following it if the OS changes.

## Components & utilities

Component classes live in their own partials and are documented by their class
names:

- `.btn` with variants `.btn--primary`, `.btn--reversed`, `.btn--negative`,
  `.btn--ghost`, `.btn--outline`, `.btn--plain`, sizes `.btn--small` /
  `.btn--large`, `.btn--block`; circular icon buttons are automatic when the
  label is visually hidden.
- `.card` / `.card--raised` with `.card__body` and `.card__title`.
- `.alert` with `.alert--success`, `.alert--error`, `.alert--info`.
- `.table` / `.table--zebra`, `.table-scroll`.
- `.field` (`.field__label`, `.field__hint`, `.field--with-icon`) + `.input`,
  `.textarea`, `.select` (`.input--code`).
- `.menu` dropdown (`.menu__trigger`, `.menu__content`, `.menu__title`,
  `.menu__item`) — opens on `:focus-within`, no JS required.
- `.header` / `.header__bar` / `.header__logo` / `.header__actions`,
  `.theme-toggle`, `.footer` / `.footer__meta`.
- `.page` width variants (`.page--sm/md/lg/xl`) and `.stage` for centered auth
  screens.
- Utilities: `.flex`, `.grid`, `.stack`, `.gap*`, `.items-*`, `.justify-*`,
  `.txt-*` (size/color/weight), `.pad*`, `.margin-*`, `.fill*`, `.border*`,
  `.visually-hidden`, etc.

## Modern CSS conventions

- **Custom properties as a mini-API.** Expose component knobs with fallback
  values, e.g. `color: var(--btn-color, var(--color-ink));`, and declare variants
  by overriding the custom property, not the property.
- **`:has()` instead of state classes.** Query content rather than adding
  server- or JS-managed modifier classes.
- **Capability queries.** Use `@media (any-hover: hover)` for hover affordances.
- **Logical properties.** Prefer `inline-size`/`block-size`, `padding-inline`,
  `margin-block-start` over physical `width`/`height`, `padding-left`, etc.

## Usage examples

```html
<!-- Card -->
<article class="card">
  <div class="card__body">
    <h3 class="card__title">Title</h3>
    <p class="txt-muted">Supporting copy.</p>
  </div>
</article>

<!-- Primary button -->
<button class="btn btn--primary">Continue</button>

<!-- Field -->
<div class="field">
  <label class="field__label" for="email">Email</label>
  <input id="email" class="input" type="email">
</div>

<!-- Status -->
<span class="txt-negative">Something went wrong</span>
```
