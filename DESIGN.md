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
    --text-muted: #8696A0;
    --accent: #8A6B7C;
    --accent-contrast: #0B141A;
    --border: #2A3942;
    --error: #F15C6D;
    --success: #4ADE80;
  }
}
```

- `--accent` (muted plum, `#6B4E5E` light / `#8A6B7C` dark) is the **only** accent color. Do not introduce a second competing accent (no default terracotta/orange, no generic SaaS blue) unless explicitly requested.
- If using Tailwind in `admin/`, map these CSS variables into `tailwind.config` (`colors: { bg: 'var(--bg)', accent: 'var(--accent)', ... }`) rather than using Tailwind's default palette directly, so both apps stay token-consistent.

---

## 3. Typography

- **Body / UI face:** a clean, neutral humanist sans (e.g. Inter or General Sans). Should feel invisible, not decorative — this is a utility-dense app (feed, DMs, forms), not a marketing landing page.
- **Monospace face** (JetBrains Mono or IBM Plex Mono): used narrowly — for things like verification codes, technical/admin-panel data tables, timestamps in dense admin views. Not used for general body text.
- **No heavy display serif.** This product does not need an editorial personality; legibility and speed of scanning matter more than a memorable type voice.
- Font sizes fluid via `clamp()` where practical (e.g. `font-size: clamp(0.875rem, 2vw, 1rem)`) to reduce the number of hard breakpoint overrides needed.

---

## 4. Layout Reference (from Peermeld, reinterpreted with our tokens)

**Top bar:** search input, a few status/utility icons on the right (kept minimal — RexiO City doesn't need Peermeld's specific gift/points/streak icons unless a future gamification feature is explicitly scoped), user avatar.

**Feed card:**
- Author row: avatar, display name, `@username`, timestamp (muted color, right-aligned or trailing).
- Body: text content, optional media, optional link-preview card.
- Action row: like, comment, repost, save/bookmark — icon + count, muted color, `--accent` on active/pressed state.
- Hairline `--border` divider between cards, not heavy shadows.

**Bottom tab bar (mobile):** Feed, (other primary sections as scoped), Messaging, Notifications/Alerts. Icons + labels, active tab in `--accent`.

**Profile page:** cover image, avatar overlapping the cover (circular, bordered in `--bg`), display name, `@username`, bio, follow/following counts, action buttons (Follow / Message / Edit Profile depending on viewer), tabs for Posts / Replies / Media.

**Compose:** a floating action button (mobile) opening a bottom sheet with tabs/sections for text, photo, video, voice attachment.

---

## 5. Motion

- Keep motion purposeful and restrained — this is a dense, frequently-used utility app, not a marketing site. Avoid decorative page-load animations.
- Micro-interactions only where they clarify state: e.g. a like button's brief scale/pulse on tap, a subtle fade when a new DM arrives in an open thread.
- Respect `prefers-reduced-motion` — disable non-essential motion when set.

---

## 6. Accessibility Floor (non-negotiable, every screen)

- Visible keyboard focus states on all interactive elements (use `--accent` for the focus ring, not the browser default only).
- Color contrast: body text against `--bg` must meet WCAG AA at minimum in both themes.
- All icon-only buttons (like, repost, bookmark, etc.) need an `aria-label`.
- Forms (login, signup, compose) must have real `<label>`s, not placeholder-only inputs.

---

## 7. Copy / Voice

- Plain, active voice. "Save changes," not "Submit." A control's label stays the same word through the whole flow it triggers (a "Post" button produces a "Posted" confirmation, not "Success!").
- Empty states are instructional, not cute: e.g. "No messages yet — start a conversation from someone's profile," not a joke or filler illustration caption.
- Errors state what happened and what to do next, in the interface's voice — never "Oops!" or an apology tone.

---

## 8. Admin App (`admin/`) Specific Notes

- Same token system as `frontend/` — this is one brand, not a visually distinct "internal tool" look.
- Denser layouts are acceptable here (more data per screen, smaller touch targets are fine on desktop) but must still be fully usable on mobile per the PRD's responsiveness requirement.
- Use the monospace face more liberally here (IDs, timestamps, raw data tables) than on the main user-facing app.

---

*Any component that introduces a color, spacing, or font value not defined here should trigger a pause — add the token here first (and note the addition in `WORKLOGS.md`), then use it. Don't inline a one-off value "just this once."*
