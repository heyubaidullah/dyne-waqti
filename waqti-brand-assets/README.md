# Waqti brand assets

Developed by Ubaidullah Khan at [Dyne Labs](https://www.dynelabs.org) © 2026

All source files are SVG (vector) so they can be freely recolored, rescaled, or edited later without quality loss. PNG exports are included for places that require a raster file (favicons, app store icons, social previews).

## Folder guide

```
svg/
├── icon/                  Square icon / app-icon tile, and the standalone mark
│   ├── waqti-icon-dark.svg         Full-color icon on a dark navy badge (default app icon)
│   ├── waqti-icon-light.svg        Full-color icon on a white badge
│   ├── waqti-mark-color.svg        Just the crescent + hands, full color, transparent bg
│   ├── waqti-mark-mono-black.svg   Single-color black, transparent bg (print, engraving, stamps)
│   └── waqti-mark-mono-white.svg   Single-color white, transparent bg (dark/reverse contexts)
├── lockup-horizontal/     Icon beside the "Waqti" wordmark
│   ├── waqti-logo-horizontal-dark.svg
│   └── waqti-logo-horizontal-light.svg
├── lockup-stacked/        Icon above the "WAQTI" wordmark — the /display and web-facing version
│   ├── waqti-logo-stacked-dark.svg
│   └── waqti-logo-stacked-light.svg
└── wordmark/              Text only, no icon, for tight spaces
    ├── waqti-wordmark-dark.svg
    └── waqti-wordmark-light.svg

png/
├── favicon-32.png / favicon-64.png     Browser tab icon
├── app-icon-512.png / app-icon-1024.png   App store / desktop shortcut icon
└── social-share-1200x630.png           Open Graph / social link preview image
```

## When to use which

- **App icon, favicon, desktop shortcut, kiosk taskbar** → `svg/icon/waqti-icon-dark.svg` or the matching PNG in `png/`.
- **`/display` idle/splash screen, big on-screen branding** → `svg/lockup-stacked/waqti-logo-stacked-dark.svg`.
- **Website header, `/admin` panel top bar, email signature** → `svg/lockup-horizontal/` (pick dark or light to match the surface).
- **Printed materials, flyers, letterhead** → `svg/lockup-horizontal/waqti-logo-horizontal-light.svg` or the wordmark-only file if space is tight.
- **Single-color use** (engraving, embroidery, stamps, watermarks, laser cutting) → `svg/icon/waqti-mark-mono-black.svg` or `-mono-white.svg`.
- **Placing the mark on a photo or colored background** → `svg/icon/waqti-mark-color.svg` (no square badge, transparent).
- **Social media profile picture / link previews** → `png/app-icon-512.png` and `png/social-share-1200x630.png`.

## Color tokens

| Token | Hex | Use |
|---|---|---|
| Background dark | `#0B0F17` | Badge / dark surfaces |
| Accent gold | `#F59E0B` | Primary mark color |
| Accent gold hover | `#D97706` | Interactive hover state |
| Status emerald | `#10B981` | Small accent dot on the mark, "live/on-time" |
| Text primary | `#F8FAFC` | Wordmark on dark |
| Text secondary | `#94A3B8` | Tagline on dark |
| Alert red | `#EF4444` | Reserved for app alerts, not used in the logo |

## Typography

- **Wordmark / English UI:** Plus Jakarta Sans (fallback: Poppins, Helvetica Neue, Arial, sans-serif)
- **Arabic text in-app:** Amiri (fallback: Noto Naskh Arabic, serif)

## Editing notes

- All icon/mark files use the same underlying geometry (crescent: circle at `205,199` r`104` minus circle at `247,167` r`92`; hub at `217,199`), so if you resize or recolor one, you can copy the same coordinates into any other variant to keep them in sync.
- The mono and transparent mark files use an SVG `<mask>` to cut the crescent shape, so they render correctly on any background — no need to match a fill color to the backdrop.
- To recolor: swap the `fill`/`stroke` hex values only — do not change the numeric coordinates, or the mark will drift out of alignment with the badge (it was deliberately centered and hand-spaced through measurement, not by eye).
