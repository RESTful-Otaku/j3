# Displays & multi-screen views

Every view has its own URL, so the control room, public dashboard and several full-screen boards can
run on separate screens and stay live-synced.

## Core routes

`/public` dashboard, `/display` scoreboard, `/brackets`, `/tables`, `/pools`, `/control`,
`/schedule`, `/medals`, `/ceremony`. In-app navigation stays client-side; use **↗ New tab** for a
TV/projector board.

## What the scoreboard shows

Event title, mat, banner/logo, fighter names/photos/belts/stats, club logos or independent mark,
officials with roles, per-side moves and cards, authoritative clock and recent technique callouts.
Winners animate with configurable celebrations.

## Responsive layouts

Design range is 6" phone to 100" LED. Ultrawide screens use full width; phones get labelled cards,
section tabs and explicit "show more" controls instead of horizontal or endless scrolling.

## Themes and branding

White-label tokens control colours, fonts, logos and banners. Corner colours and result glyphs are
configurable; the Street Fighter theme is the default high-energy palette, with dark ink on its bright
pink/orange/yellow fills so labels stay legible. Banners use `contain` so the entire uploaded image is
visible against a themed backdrop.

## Navigation and fit

The desktop navigation is drag-resizable (180–380 px, arrow keys supported) and stores the chosen
width; labels and spacing scale with it. The active view gets a bright themed left bar and glow, and
the mobile drawer/bottom bar use the same treatment. The account switcher is clipped to the nav pane.

## Demo replay

When the server has `J3_DEMO=1`, a **Demo replay** launcher follows you across every screen. It runs
a deterministic event and lets you play, pause, restart, step, scrub and change speed. The panel shows
the same playhead as Mat, Table, Bracket and Ceremony state cards; `/tables`, `/public` and
`/ceremony` also reflect the selected bout, standings and medals live. `/simulate` uses the same
controller, so scrubbing in one place stays coherent everywhere.

## PA and announcements

PA toasts flash fighter cards and call competitors to a mat/podium. Browser TTS speaks announcements;
timed toasts auto-dismiss.

## Custom views

The builder lets authorised users drag/resize live widgets, save, clone, load and delete named
layouts, and open them as standalone `/view` screens.

## Broadcast

A transparent OBS/vMix overlay and public live page are available; VAR/camera replay needs external
streaming integration.
