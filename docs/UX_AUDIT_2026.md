# J3 UX & UI Audit — 2026

**Scope.** Independent review of three device targets against modern 2026 app expectations.
No single responsive soup: each class gets its own navigation model, density, touch model,
typography scale, and layout.

Device classes in code today:

- **Phone** `max-width: 700px` + `isPhoneWidth()` → phone scope (`PHONE_ROUTES`).
- **Tablet** `min-width: 701px and max-width: 1024px` (+ landscape ≤1366; + ≤560px-height landscape).
- **Web SPA** `min-width: 1025px`, wide `min-width: 1800px`, reduced-motion honoured.
- **Display/LED** `(display)` route group uses `display.css`, separate from app chrome.

---

## 1. Cross-cutting findings (all device types)

### Strengths (keep)
- **White-label branding** applied end-to-end: event logo/banner, brand colours, corner Red/White,
  fighter photos, club logos, club colours on cards, tables, brackets, ceremony, pools.
- **Accessible by default**: colour **and** glyph (W/L/D), `aria-*`, keyboard, `prefers-reduced-motion`,
  readable ink computed, contrast-safe bright fills, safe-area insets on fixed chrome.
- **Safe areas** (`--safe-top/right/bottom/left`) on top bar, drawer, replay, dialogs, toasts.
- **Offline-first / QA**: standalone seeded builds, sim mode, `state.changed` live updates.
- **Minimalism + expandable `Card`**: concise default, `+`/`–` toggle, preview when shrunk,
  never hides glanceable operator displays.

### Gaps identified (this pass)
1. **Wasted bottom space on mobile.** `.app-content.mobile` reserves `calc(72px + safe-bottom)`
   for a bottom tab bar that was removed (drawer-only nav). Content scrolls further than needed.
2. **No enforced touch-target baseline** on touch devices (some `min-height:44`, some not).
3. **Tablet under-serves width.** Tablet 701–1024 inherits phone-ish tightness in places;
   `/settings`, `/club` are narrow single-column; forms could use two-up; no explicit tablet nav.
4. **Breakpoint inconsistency.** Phone CSS `700`, phone scope `700`, mobile-nav `820`, tablet `1024`.
   Works but undocumented; the 701–820 band is "mobile chrome + tablet content".
5. **`app-content` bottom padding on mobile** + `DemoReplay` `bottom:52px` assumes a removed bar.
6. **Desktop density** is high on operator screens (fine) but landing/settings pages are wide-open.

---

## 2. Web SPA (desktop, ≥1025px)

**Aspect/interaction.** Large screens; hover + precise pointer; keyboard; persistent left nav;
operator + display surfaces (official console, mat, brackets on TVs, LED walls, multi-display).

**UX / interaction.** Strong. Drawer→fixed rail; `☰` toggle; resize handle; active-state;
mode switch (live/sim); command-ish density appropriate. `Esc` closes drawer.

**Visibility.** Banner fills width; brand mark = event logo; club colours everywhere; photos.
**Minimalism vs data.** Operator screens intentionally dense; **not** applied to glanceable displays.
Collapsible `Card` available; nothing critical hidden. Good.

**2026 expectations check.**
- Motion reduced-motion ✓; `view-enter`/`reveal` only ≤700 + Card ✓.
- Focus-visible rings partly rely on UA; add consistent `:focus-visible`.
- `prefers-color-scheme` + 4 themes ✓; dark default.
- Keyboard: hamburger + drawer items + Card `Enter/Space` ✓.

**Recommended (implemented):**
- Consistent `:focus-visible` outline token; touch-target baseline on `(hover:none)`.
- Reclaim no-op mobile-only changes don't affect desktop.

---

## 3. Phone app (Capacitor + phone web, ≤700px)

**Aspect/interaction.** Small, portrait-first; thumb reach; touch; safe areas (notch, home
indicator, OS bars); intermittent connectivity; one-hand use.

**UX / interaction.**
- Top bar: J3 home + event title + hamburger. **Drawer-only nav** (old bottom home bar removed ✓).
- Participant-only `PHONE_ROUTES` for production; QA full-phone for testing.
- Concise single-column `/settings`, `/signup`, `/public`, `/club`.
- `FighterPortrait`, tables reflow to labelled cards, no sideways scroll, names wrap, 320px safe.
- 44px+ targets on primary actions; `action-row` stacked.

**Gaps (implemented).**
- Reclaim `padding-bottom` reserved for removed bottom bar.
- Global touch-target baseline on `(hover:none)`.
- `DemoReplay` bottom offset no longer assumes removed bar (kept conservative).

**Minimalism vs data.** Strong. `Card` shrinks; `.phone-section` tabs; `phone-limit`. 
**2026.** Installable PWA/Capacitor, safe-area, reduced-motion, offline seeded QA, drawer-only nav,
no fake home bar. Consider (not blocked): pull-to-refresh, haptics, `viewport-fit=cover` already.

---

## 4. Tablet app (701–1024px, portrait + landscape)

**Aspect/interaction.** Medium; touch + optional pointer; can hold keyboard; two-thumb; landscape/portrait.

**Current.** Tablet 2-col `.grid`; mobile nav drawer at ≤820; **not** phone-scoped (tablets can use
operator screens); landscape ≤1366 adjustments; `orientation` + `max-height:560px` handling.
**Under-served:** `/settings`, `/club`, `/public` sections are tuned for phone; forms single-column.

**Recommended (implemented).**
- Explicit tablet section: restore width for `/settings`, two-up field rows, gentler card padding.
- Keep drawer nav; no bottom bar; no phone scope.

---

## 5. Minimalism vs. important data/state

Keep visible (never hide): **state** (live/pending/complete/forfeit, paused, drift, on-schedule),
**score/clock**, **corner** (Red/White + colour + glyph), **club**, **belt**, **mat/category**,
**win/loss/draw glyphs**, **safe link/offline status**, **accessibility contrast**.
Collapsible detail is for **explanatory/secondary** content, not state. Operator/display cards
remain expanded. `Card` `expandable={false}` for forms and glanceable scoreboards.

---

## 6. Concrete changes applied in this pass

1. **Reclaim bottom space:** `.app-content.mobile` bottom padding → `calc(12px + var(--safe-bottom))`.
2. **Touch baseline** (`@media (hover:none)`): interactive elements `min-height:44px`; icon-only
   `min-width:44px`; no `:hover` transform reliance.
3. **Focus-visible:** consistent `:focus-visible` ring on `a, button, input, select, textarea, [tabindex]`.
4. **Tablet `/settings`:** widen shell at 701–1024; two-up `field-row`; tablet-safe padding.
5. Reduced-motion + safe-area preserved; operator displays untouched; **no behavioural/route changes**.

Verification: `svelte-check`, Go tests, web unit tests, mobile/nav/official E2E, `build:flavors`,
`build:android:qa`, `check:workflows`.
