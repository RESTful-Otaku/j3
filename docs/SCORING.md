# Live scoring rules

J3 does not hard-code one federation's rules. The values below are the defaults shipped in the generic rule set; a runner can change every one in the setup wizard.

## Score events

A score event has an actor, timestamp, corner, technique, point value and audit entry. Confirmed events are committed to SQLite immediately; every visualisation reads that committed data.

Common scoring families:

| Family | Typical values |
|---|---|
| Ippon / Waza-ari | Ippon 3, Waza-ari 1, Shido penalty; Ippon ends the contest |
| BJJ points | takedown 2, sweep 2, guard pass 3, mount/back 4, advantage 1 |
| Fighting System | strike 2/1, throw 2/1, submission 2–3, hold 10s/15s/20s |
| Kata panel | 0–10 in 0.1 or 0.2 steps, high/low drop, middle average |

See the [API guide](/docs/api-guide) for the request shapes behind these actions.

## Holds and submissions

A hold timer counts down live on the scoreboard. Reaching the configured threshold awards the configured score. A tap or verbal submission is an instant win that negates prior points and triggers promotion. Both paths require confirmation.

## Penalties

Penalties use a configurable ladder:

- **yellow card** — light foul, may award an opponent point;
- **red card** — serious foul or two light fouls;
- **shido** — three accumulate to hansoku-make (disqualification);
- **forfeit / injury / no-show** — recorded with a reason and consistent advancement rules.

Penalty badges appear on the offending fighter's side of the scoreboard, never only as a colour.

## Clock and stoppages

The venue server owns the clock. The scorekeeper starts, pauses, resumes and submits results. Paused time is visually distinct so spectators can see a referee stoppage. Golden score or overtime is a separate configured phase with its own first-point rule.

## Best-of-N and multi-round bouts

A bout can be a single contest, best of three or best of five. J3 tracks each round separately, then applies the configured series rule. The bracket keeps a copy of every fighter at each round so the full history remains visible.

## Tie-breakers

Runners choose an ordered chain, for example: wins → submissions → points in winning bouts → head-to-head → fewest points conceded → fastest submission → positive play majority. The first rule that separates fighters decides the rank.

## Result codes

Configurable outcome icons include WP/LP/DP (points), WS/LS (submission), WT/LT (technical/forfeit/DQ) plus yellow/red card badges. A result icon is always paired with text.

## Confirmation and undo

No score, card or result becomes authoritative until confirmed. Undo is available per entry, attributed to the actor and visible in the audit trail. Results are locked after confirmation; corrections need an explicit, recorded path.
