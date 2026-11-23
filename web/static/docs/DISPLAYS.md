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

## Live event / Simulated tournament

The nav has an explicit **Live event / Simulated tournament** toggle (including compact and mobile
nav). Live is the real venue state. Simulated mode is app-wide and persists in the URL and local
storage; it shows a persistent replay widget on every view with play/pause, restart, previous/next
bout, speed and full scrubbing. `/simulate` is the full event-by-event console and switches the app
into simulated mode; **Stop & go live** returns to real data. Replay frames never mutate live views.

Mat, Table, Bracket and Ceremony state cards share the same playhead, and `/public`, `/tables` and
`/ceremony` also reflect the selected bout, standings and medals. Scrubbing in one place stays
coherent everywhere.

## Live invalidation

The venue server broadcasts cheap `state.changed` SSE frames after material writes: results,
brackets and promotions, officials, kata scores/advancement, medals and schedule changes. Open
brackets, officials console, scoreboard/mat/display, public, tables, pools, overlay, custom widgets
and the medal ceremony refetch immediately on those frames instead of waiting for a poll interval.
Fight scores and the authoritative clock keep their existing detailed fight payloads. Polling remains
as a fallback if a connection drops.

## PA and announcements

PA toasts flash fighter cards and call competitors to a mat/podium. Browser TTS speaks announcements;
timed toasts auto-dismiss.

## Custom views

The builder lets authorised users drag/resize live widgets, save, clone, load and delete named
layouts, and open them as standalone `/view` screens.

## Broadcast

A transparent OBS/vMix overlay and public live page are available; VAR/camera replay needs external
streaming integration.
