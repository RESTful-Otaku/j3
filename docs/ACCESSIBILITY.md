# Accessibility

`prefers-reduced-motion`, `prefers-color-scheme`, `prefers-contrast`, `forced-colors` honored. WCAG-AA readable ink (`readableInk`, `readableInk` in event-brand); contrast-safe bright fills; colour **and** glyph (W/L/D); safe areas.

**User modes** `j3-accessibility` + `applyAccessibility` + `<html data-*>`:
- `data-font-scale=normal|large|x-large` (112.5%/125%)
- `data-contrast=normal|high`
- `data-colour-mode=colour|monochrome`
- `data-dyslexic=on|off` (spacing + type; `Atkinson Hyperlegible`/`OpenDyslexic` if provided, else sans)
- `data-motion=system|reduced`
Settings `/settings`. Keyboard + ARIA + focus-visible + `ConfirmDialog` focus trap. No `@html` unsanitized (`markdown` sanitize; `sanitizeText/sanitizeUrl`). `svelte-check`, axe, Playwright, Storybook a11y.
