# Formats, pools & promotion

Format choice is a policy, not a hard-coded rule. J3 selects a format from the actual number of entrants and the event's constraints, then lets a moderator override it.

## Selecting a format

| Entrants | Common J3 policy |
|---|---|
| 1 | No contest; single winner with its own ranking points |
| 2 | Best of three, stop at 2–0 |
| 3 | Round robin |
| 4 | Round robin or single elimination with third-place match |
| 5 | Round robin pool of five |
| 6 | Two pools of three, top two cross to semis |
| 7+ | Double elimination, repechage or multi-stage |
| 9+/17+ | Single elimination until last 8/16, then double elimination |

If elimination is incompatible with the entrant count, J3 suggests round robin instead of creating an invalid bracket.

## Round robin

The circle method generates every pairing once. Odd fields receive a BYE without duplicating fixtures. Standings use the configured tie-break chain, not just wins.

## Pools

Automated pool generation balances pool sizes, seeds by rank and stats, and uses controlled randomisation to spread clubs. Organisers set the target pool size, maximum pools and separation rules. **Regenerate balanced pools** gives a different controlled seed; moderators can drag fighters between slots.

## Elimination brackets

J3 supports single elimination, true double elimination (lower winner can win gold via a reset) and consolation double (lower winner can only win bronze). The distinction is stored in the format policy so it is never ambiguous.

BYEs are allocated by seed or title. First-round club or country separation avoids early same-club meetings where the policy allows it.

## Third place and bronze

If the two losing semi-finalists have already met, J3 awards two bronze medals instead of scheduling a repeat. If they have not met, a third-place match is generated. This rule is configurable per discipline.

## Promotion and history

Winners, losers and draws advance according to the rule set. Each round keeps a copy of the fighter card, so the bracket shows the full path. Rematches are avoided where configured. Forfeits and no-shows rebalance the affected pool and promotion path.

## See also

- [Live scoring rules](/docs/scoring)
- [Runner guide](/docs/runner)
- [Tournament formats in the organiser guide](/docs/guide)
