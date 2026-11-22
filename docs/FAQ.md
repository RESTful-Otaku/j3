# Frequently asked questions

## Can I use J3 with no internet at the venue?

Yes. The venue server, database and web app all run on a local machine or LAN box. Phones and
displays connect to it over Wi-Fi. Nothing in the normal scoring flow needs the cloud. A single
SQLite file is the backup.

## Do I need to know federation rules?

No. Pick a template or accept the defaults. Every value — points, match length, penalties,
categories, tie-break order, display layout — is editable in plain language before the event.
Nothing is hard-coded per federation.

## Can one fighter enter more than one category?

That depends on your rule set. Some federations allow more than one age division, others allow one
age and one weight. The compatibility policy is configured in the event, not fixed in the code.

## What happens if a category has only two entrants?

J3 selects an appropriate format from your policy. Two entrants can use best of three, three use a
round robin, and small or uneven pools fall back to round robin rather than producing a broken
bracket.

## How are same-club meetings avoided?

Seeding uses controlled randomisation with a same-club separation constraint. You can override a
placement by dragging a fighter; the move is recorded in the audit trail.

## Who can change a score?

Only an official assigned to that contest, with a role that has the `fights` capability. The
scorekeeper controls the board, the referee controls the clock, and a runner can override a locked
result. Every change records the actor, timestamp and before/after values.

## How do I put different views on different screens?

Each view has its own URL. Open it in a new tab, then drag that tab onto the second monitor or
full-screen it. Use **View builder** to compose a saved layout from live widgets, then assign that
template to a screen in **Display control**.

## Which languages are supported?

The shared label dictionary includes English, French, German, Spanish, Portuguese, Hindi,
Punjabi, Arabic, Japanese, Korean and Mandarin. Appearance follows the system by default and can
be overridden per user.

## Can I restore a tournament after a crash?

Yes. SQLite runs in WAL mode and the venue file can be backed up or restored mid-event. Results
are committed as they are confirmed, so a restart continues from the last committed state.

## Where do I find the raw API contract?

Use the [API reference](/docs/api), the raw [OpenAPI file](/openapi.yaml) or the
[API guide](/docs/api-guide).
