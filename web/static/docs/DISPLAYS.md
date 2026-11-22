# Displays & multi-screen views

Every view is a normal URL, so you can open as many as you need on separate TVs, projectors, tablets and phones. They all poll or stream committed server data and stay live.

## Core routes

| Route | Use |
|---|---|
| `/public` | Spectator dashboard, clock, next-up and live results |
| `/display` | Main full-screen scoreboard |
| `/mat` | Scorekeeper and timekeeper console |
| `/brackets` | Progression, results and history |
| `/tables` | Standings, kata, medals and club points |
| `/pools` | All matchups and current status |
| `/control` | Assign dashboard templates to screens and trigger PA |
| `/ceremony` | Guided medal podium sequence |

Use fullscreen in the browser (`F11` or kiosk mode) for venue displays.

## What the scoreboard shows

The board includes the event title, mat number, banner and logo, both fighters with large avatars, belt designation, height/weight, club logo, corner colours, active referee and judges, recent moves with points, penalties on the correct side, and the authoritative clock. W/L/D always show a glyph as well as colour.

## Responsive layouts

Layouts adapt from a 6-inch phone to an ultrawide LED wall. On narrow screens, fighter cards stack vertically and the card slides into a bottom sheet. Ultrawide screens use the full available width rather than a centred column. Avatar areas stay large enough to read across a hall.

## Themes and branding

Themes are CSS custom properties plus uploaded assets: logo, banner, colours and fonts. The Street Fighter theme adds lineups, matchups and announcer energy. Club colours extracted from uploaded logos appear on cards, tables and brackets.

## PA and announcements

Use **Display control** or the announce button on Event config to trigger timed toasts, flash the next competitor's card and speak text with browser TTS. Multilingual labels are available for eleven locales.

## Custom views

**View builder** lets an authorised user drag live widgets onto a canvas, resize them, save/clone/delete named layouts and open one as its own URL. Pull a widget's ⠿ handle to reorder, drag its right or bottom edge (or the corner grip) to resize horizontally, vertically or both, and watch a floating ghost with a snap indicator while you drag; the layout eases into place on release. Arrow keys on the same grips and the S/M/L and H1/H2/H3 buttons remain as accessible fallbacks. Combine clock, next-up, scoreboard, standings and medal widgets for a dedicated stat screen.

## Broadcast

`/overlay` provides a transparent HTML layer for OBS or vMix with a scorebug and lower thirds. Async network video replay remains a future enhancement.
