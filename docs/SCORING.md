# Live scoring rules

Scoring is authoritative on the venue server. Clients send `techniqueId` and optional `pointLabel`;
points, kind, card and label are resolved from the event rule snapshot.

## Score events

Every point, technique, advantage and move is a timestamped event attributed to an actor. The
scoreboard sums committed events; it never keeps a separate score state.

## Holds and submissions

Hold/pin timers count down live and follow configured thresholds (e.g. 10/15/20 seconds). A tap or
verbal submission instantly ends the bout and overrides prior points.

## Penalties

Shido, yellow/red cards, point deductions and DQ ladders are configurable. Cards render on the
correct fighter's side with a glyph; thresholds trigger the configured outcome.

## Clock and stoppages

The server owns the clock. Referee stoppages pause it; the scorekeeper/moderator resumes. Overtime
and golden score follow the rule snapshot.

## Best-of-N and multi-round bouts

Best-of-three/five tracks each round separately, then applies the series rule. Kata rounds and
bunkai scores are stored separately from fight scores.

## Tie-breakers

An ordered chain resolves pool standings: wins, submissions, points scored, head-to-head, fewest
conceded, fastest submission, positive play/majority as configured.

## Result codes

`WP/LP/DP` (points), `WS/LS` (submission), `WT/LT` (technical/forfeit/injury/DQ/red card). W/L/D
always show a glyph as well as a colour.

## Confirmation and undo

Major actions ask for confirmation. Results lock after confirmation; corrections are explicit
audited undo events with before/after values.
