# Official & moderator guide

Your console is designed for one tap per action and confirmation before anything is committed. Use **Guides & progress → Officials** for a checked, hands-on run-through.

## Get assigned

A runner assigns officials to mats and contests. Sign in with your official account and open **Official console**. The contest picker shows only contests assigned to you; runners see the whole event so they can override assignments. If you cannot see a contest, you are not assigned to it — ask the runner rather than trying to score from another account.

## Select the contest and corner

1. Choose the contest in **Select contest**. Live contests are offered first.
2. Tap **RED** or **BLUE** to choose the active corner. The button shows the fighter's name.
3. Use the search box and group chips to find a technique by English name, romaji, Japanese script, style or movement type.

Points, penalties and submissions always apply to the active corner. Check the name before every entry.

## Log a technique

The catalogue covers the full JJJ vocabulary: punches, kicks, elbows, knees, palm and nerve strikes; throws, takedowns, sweeps, reaps and hooks including sacrifice and flying variants; ground controls, hold-downs and pins; submissions, chokes, locks and breaks. Search is client-side and instant; the **Points** dropdown on the selected card lists the values configured for this event's rules (for example `Throw-Defended Kick · 3`). The server resolves the final value from the rule set, so a browser cannot award arbitrary points. Tap a technique card, review the confirmation dialogue (technique, romaji and point value), then press **Confirm & log**. The score appears on the live scoreboard immediately. If you tapped the wrong corner or technique, use **Undo last entry** — the correction is audited.

## Penalties and cards

Use **Yellow / shido** or **Red / hansoku** to attribute a card to the active corner. The server applies the configured penalty ladder and records the card, actor and before/after points. Never award a card without the confirmation step.

## Control the clock

The authoritative clock is on the same console as your scoring tools. **Start**, **Pause** and **Resume** write to the venue server; the public board mirrors that state. Pause for a referee stoppage and resume only on the scorekeeper's instruction. Do not run a separate stopwatch as the source of truth.

## Technical wins and losses

For injury, forfeit, no-show, withdrawal, referee decision, cheating or disqualification, choose the winner and reason under **Technical win / loss**, add an optional detail, then call the result. J3 locks the bout, records an audited technical score event and publishes the configured `win_red`, `win_blue`, `dq_red` or `dq_blue` outcome to every display.

## Submissions and instant wins

A submission overrides accumulated points. Search for a choke, lock or tap, select its card and use **Log for RED/BLUE · INSTANT WIN**; J3 records the tap and asks you to confirm the final result. See [scoring rules](/docs/scoring) for holds, golden score and tie-breaks.

## Kata panel judging

Switch to the **Kata panel** tab. Each performance has sections such as etiquette, demonstration and application, scored by each judge from 0–10. J3 drops the highest and lowest marks where the rule set requires it and publishes the authoritative total. Change a score and confirm; totals update on **Tables & rankings**.

## Finish the bout

Use **Red wins**, **Blue wins** or **Draw** in the final decision panel, then confirm. When scores are level a hint reminds you to confirm the draw or call a technical decision. The result is locked, pushed to every display and written to the audit trail. If you need to correct it, ask a runner to use the recorded undo path.

## Handover

At the end of your assignment, check **Audit trail** for your actions and tell the runner about any incidents. Do not share your account with another official — assignments and audit entries must stay attributable.
