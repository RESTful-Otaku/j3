# Frequently asked questions

## Can I use J3 with no internet at the venue?

Yes. The venue server and LAN displays work offline. The Android APK bundles a seeded database and
read-only snapshot, so phone screens work even with the server stopped; scoring needs the LAN.

## Do I need to know federation rules?

No. Pick a preset or enter plain-language values; J3 stores per-event rule snapshots. Advanced raw
config is optional.

## Can one fighter enter more than one category?

Only if the event's category policy allows it; some disciplines restrict athletes to one age and
one weight division.

## What happens if a category has only two entrants?

J3 suggests best-of-three or round robin instead of a broken elimination bracket. Empty categories
are flagged, not rendered as invalid matches.

## How are same-club meetings avoided?

Seeding and pool generation spread clubs/countries with controlled randomisation where the rules
allow it. Organisers can override with drag-and-drop.

## Who can change a score?

Only an assigned official or an admin/runner override. Every write is attributed, audited and
correctable with a recorded undo.

## How do I put different views on different screens?

Each display route has its own URL. Open it in a new tab on the target TV/projector; use Display
control to assign saved dashboard templates.

## Which languages are supported?

Eleven locales for shared labels: EN, FR, DE, ES, PT, HI, PA, AR, JA, KO, ZH.

## Can I restore a tournament after a crash?

Yes. SQLite WAL plus a single-file backup allow recovery; restore the database and restart the venue
server.

## Where do I find the raw API contract?

See the [API guide](/docs/api-guide) or the interactive API reference at `/docs/api`.
