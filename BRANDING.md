# BRANDING.md — RexiO City Brand Guidelines

Visual identity and branding standards for RexiO City.

---

## Logo Usage

### Primary Logo
- Icon: `rexio_core_icon.svg`
- Wordmark: `rexio_logo_wordmark.svg`

### Clear Space
Maintain minimum clear space equal to the height of the "R" icon around all sides.

### Backgrounds
- Light mode: Logo on `#FFFFFF` or `#F0F2F5`
- Dark mode: Logo on `#0B141A` or `#202C33`
- Never place logo on busy images without overlay

---

## Colors

See `DESIGN.md` for full token system. Key colors:

| Token | Light | Dark |
|---|---|---|
| Background | `#FFFFFF` | `#0B141A` |
| Secondary | `#F0F2F5` | `#202C33` |
| Elevated | `#FFFFFF` | `#17212B` |
| Text | `#111B21` | `#E9EDEF` |
| Muted | `#667781` | `#8696A0` |
| Accent | `#6B4E5E` | `#8A6B7C` |
| Border | `#E9EDEF` | `#2A3942` |
| Error | `#B00020` | `#F15C6D` |
| Success | `#1F9254` | `#4ADE80` |

---

## Typography

### Primary Font
- **Inter** or **General Sans**
- Weights: 400 (regular), 500 (medium), 600 (semibold), 700 (bold)
- Use for: Body text, UI elements, headings

### Monospace Font
- **JetBrains Mono** or **IBM Plex Mono**
- Use for: Verification codes, admin data tables, timestamps

### Size Scale
- 12px — Caption, metadata
- 14px — Body text
- 16px — Headings (H3)
- 20px — Headings (H2)
- 24px — Headings (H1)
- 32px — Display (rare)

---

## Voice & Tone

### Principles
- Plain, active voice
- Direct, not cute
- Helpful, not apologetic
- Technical when needed, simple when possible

### Examples

| Instead of | Use |
|---|---|
| "Oops! Something went wrong." | "Unable to save changes. Please try again." |
| "Great job! You've completed the task!" | "Changes saved." |
| "Can we bother you with a quick question?" | "We need your permission to send notifications." |

### Error Messages
- State what happened
- State what to do next
- No apologies

---

## Empty States

Instructional, not decorative:

- "No messages yet — start a conversation from someone's profile."
- "No posts yet — be the first to share something."
- "No followers — share your profile to connect with others."

---

## Iconography

- Use consistent icon set (e.g., Phosphor Icons or similar)
- Stroke width: 1.5px or 2px
- Rounded corners
- Accessible: All icon-only buttons need `aria-label`

---

## Motion

- Purposeful only
- Micro-interactions for state changes
- Respect `prefers-reduced-motion`
- No decorative page-load animations

---

*Last updated: 2026-08-12*
