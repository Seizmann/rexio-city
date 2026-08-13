# DESIGN.md — RexiO City Design System

This is the binding visual specification for `frontend/` and `admin/`. Do not introduce hardcoded colors, spacing, or font sizes anywhere — everything routes through the tokens below. If a design decision isn't covered here, ask/flag it in `WORKLOGS.md` rather than inventing a one-off value.

Reference inspiration: Peermeld (layout structure — top bar, feed card actions, bottom tab bar, profile layout) and Twitter/X (card density, interaction icon set). **Not a visual clone** — colors and theme are our own, defined below. Peermeld is light-mode only; we must support both light and dark, following device preference.

---

## 1. Theme Strategy

- **Light and dark mode, both required.** No app-level toggle needed in V1 — follow the OS/browser preference via `prefers-color-scheme`. (If a manual override toggle is added later, it must still map onto these same CSS variables.)
- **No hardcoded hex/px values anywhere in components.** All color, spacing, radius, and font-size values come from CSS custom properties defined once in a global stylesheet.
- **Mobile-first, required.** Write base styles for mobile, then enhance upward with `min-width` media queries. Desktop is not an afterthought — verify both, but mobile is the default reading order of the CSS.

---

## 2. Color Tokens

```css
:root {
  /* Light mode (default) */
  --bg: #FFFFFF;
  --bg-secondary: #F0F2F5;
  --bg-elevated: #FFFFFF;
  --text: #111B21;
  --text-muted: #667781;
  --accent: #6B4E5E;
  --accent-contrast: #FFFFFF;
  --border: #E9EDEF;
  --error: #B00020;
  --success: #1F9254;

  --radius-sm: 8px;
  --radius-md: 12px;
  --radius-lg: 20px;

  --space-1: 4px;
  --space-2: 8px;
  --space-3: 16px;
  --space-4: 24px;
  --space-5: 32px;
  --space-6: 48px;
}

@media (prefers-color-scheme: dark) {
  :root {
    --bg: #0B141A;
    --bg-secondary: #202C33;
    --bg-elevated: #17212B;
    --text: #E9EDEF;
    --text-muted: #8B98A5;
    --accent: #7B5EA7;
    --accent-contrast: #FFFFFF;
    --border: #38444D;
    --error: #F4212E;
    --success: #00BA7C;
  }
}
```

---

## 3. Typography

- **Font family:** System font stack (`-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif`)
- **Base size:** 14px on mobile, 16px on desktop
- **Line height:** 1.5 for body, 1.2 for headings

---

## 4. Layout

- **Max width:** 600px for feed, full-width for profiles
- **Padding:** 16px mobile, 24px desktop
- **Gap:** 12px between cards

---

## 5. Components

### Buttons
- Primary: accent background, white text
- Secondary: border, accent text
- Danger: error color

### Cards
- Background: `--bg-elevated`
- Border: `1px solid --border`
- Radius: `--radius-md`
- Shadow: subtle on elevation

### Inputs
- Background: `--bg-secondary`
- Border: `1px solid --border`
- Focus: `--accent` border
- Radius: `--radius-sm`

---

## 6. Spacing Scale

Use the spacing tokens above. Never hardcode pixel values in components.

---

## 7. Responsive Breakpoints

```css
/* Mobile first */
@media (min-width: 640px) { /* sm */ }
@media (min-width: 768px) { /* md */ }
@media (min-width: 1024px) { /* lg */ }
```

---

## 8. Dark Mode

Automatically follows system preference. No user toggle in V1.

---

## 9. Accessibility

- Minimum contrast ratio: 4.5:1 for text
- Focus visible on all interactive elements
- Semantic HTML (button, nav, main, article)

---

*This design system is binding. Deviations require explicit approval and must be documented.*
