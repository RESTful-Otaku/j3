# Organiser guide

J3 lets a non-technical organiser create, run and finish a tournament from one browser. Use the
seeded local demo to try every role before your event.

## Before you start

Have your event name, dates, venue, logo/banner, disciplines, categories and rules ready. You do
not need federation rule files: J3 asks plain-language questions and supplies presets.

## Open the demo

Run `bun run dev`, sign in as the seeded admin (`j3@local` or any account on `/demo-accounts`,
password `demo`). The J3 icon always returns to `/public`; nav stays visible on every normal route.
Use **↗ New tab** only when a board needs its own TV/projector screen.

## Create your event in five steps

1. **Event & identity** — name, dates, venue, branding, preset or template.
2. **Disciplines & rules** — timings, points, rounds and style per discipline.
3. **Categories & constraints** — age, weight, height, belt, gender, caps and split/merge rules.
4. **People & invitations** — invite officials and club owners; fighters self-register.
5. **Review & launch** — validation jumps to the offending step; launch creates pools, brackets,
   assignments and dashboard views, then saves a reusable template.

Progress is persisted, so you can move forward and back without losing work.

## Invite clubs and register fighters

Admins invite runners/admins; runners invite officials and club owners; club owners invite only
their own fighters/spectators. Invites return a `/invite/<token>` link that works before email/SMS
is connected. Fighters choose a compatible division, club or independent status, and a belt from
the colour picker. Full/incompatible categories are greyed out.

| Role | Invites | Main job |
|---|---|---|
| Admin | every role | owns tenant and authority matrix |
| Runner | officials, club owners, participants | configures and overrides the event |
| Official | club owners, participants | scores assigned contests |
| Club leader | own-club participants | manages club brand and people |
| Fighter / spectator | — | self-register and follow results |

## Let J3 build the draw

The readiness checklist runs eligibility, balanced pools, same-club spread, BYEs, seeding,
official assignments and format fallback to round robin for small pools. Drag fighters or match
order to override; every round keeps a copy of each fighter for full history.

## Score a contest

Officials open `/official` or the mat console, select their assigned bout, log techniques from the
catalogue, award cards, run the authoritative clock and confirm Red/Blue/Draw. Points come from the
event rule snapshot, so a browser cannot invent scores. Results lock with a recorded undo path.

## Kata and training

Kata uses numbered shitei/tokuyi rounds, panel judging with high/low dropped, bunkai scores and
automatic advancement. Training sessions and fair sparring rotation are available without changing
the competition draw.

## Screens and displays

Open `/public`, `/display`, `/brackets`, `/tables`, `/pools`, `/control` or `/view` on separate
screens. The control room switches views in place; displays have their own URLs for TVs and LED
walls. Use **Display control** to assign saved dashboard templates to screens.

## Keep the event on track

The public dashboard has a large clock, next-up callouts and drift-aware projected times. Mark
schedule items complete as the day moves; J3 recalculates. PA toasts and browser TTS announce
fighters, medals and results.

## Dealing with changes on the day

For no-shows, injuries, withdrawals or voluntary promotion, record the technical outcome or move
the fighter in the draw. J3 rebalances the affected pool/promotion and keeps an audit entry. Small
or empty categories get format suggestions instead of broken brackets.

## Audit and exports

`/audit` shows actor, role, time, target and before/after for every material action. `/data`
queries allow-listed tables; CSV export/import and printable brackets/scoresheets are available to
authorised users.

## Need help?

Use **Docs & guides** in `dev.sh`, `/guides` for role-specific hands-on steps, or the API explorer
at `/docs/api`. Skipping a guide marks it complete until you explicitly restart it.
