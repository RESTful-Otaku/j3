# Organiser guide

J3 runs a whole tournament from one place. This guide uses plain language and assumes no
technical experience. Start with the demo: it is seeded with two clubs, officials, fighters,
fights, kata scores and medals so you can click around before creating your own event.

## Before you start

Have these ready:

- the event name, dates and venue;
- your logo and banner images (PNG, JPG, WEBP or SVG);
- the clubs you expect and a rough number of fighters;
- the age, weight, height and belt ranges you want to offer;
- which disciplines you are running (fighting, ground, randori, kata and so on).

You do not need to know rules or point values from memory. J3 offers sensible defaults for
everything, and you can change a value later before the event starts.

## Open the demo

From the developer console choose **Run full local demo**, then open
[http://localhost:5173](http://localhost:5173). The demo signs you in as a tournament runner
automatically. Use the account switcher in the navigation to see exactly what an official,
club leader, fighter or spectator sees.

## Create your event in five steps

The setup wizard is at **Setup wizard** in the navigation. Your answers are saved as you go, so
you can leave and come back without losing work.

1. **Event & identity** — name, dates, venue, logo, banner and colours. Choose a template if one
   looks similar to your event; everything in it can be changed.
2. **Disciplines & rules** — pick the contests you will run. Each one has match length, point
   values, timers, penalties and protection requirements. Use a preset where possible, then add
   your own values with **Add your own**.
3. **Categories & constraints** — age, weight, height and belt ranges, gender policy, participant
   caps and how many fight pools to create. Full or incompatible categories are shown greyed out
   on sign-up.
4. **People & invitations** — invite runners, officials, club leaders and fighters by email (and
   SMS where configured). Fighters can also sign up themselves; privileged roles always need a
   runner's approval.
5. **Review & launch** — J3 checks your setup. If something needs attention it takes you straight
   to the section that needs it. When it is happy, launch: pools, brackets, officials, displays
   and a reusable template are created for you.

> Each step has its own URL. Open a step in a new tab if you want to check something else and come
> back — nothing is lost.

## Invite clubs and register fighters

A club leader can register fighters under their club, upload a club logo and see only their own
people. A fighter can register independently or accept a club invitation. Registration collects
name, contact details, date of birth, height, weight, belt/grade and emergency information.

Enter metric or imperial measurements — whichever is easier. J3 stores the metric values and
keeps what you typed for display. Only categories compatible with a fighter's details are offered;
a fighter may ask to move up to a heavier or older division, and you can approve or reject that
request.

On the day, use **Check-in & import** to check people in, record weigh-ins and mark out-of-class
fighters. QR badges let officials find the right entrant quickly.

## Let J3 build the draw

**Auto-run** in the readiness checklist creates balanced pools, seeds fighters by rank, spreads
same-club fighters across different pools, adds BYEs where needed and assigns officials to mats.
You can always drag a fighter to a different slot: J3 keeps a copy at every round so the full
history remains visible.

If a pool is too small for elimination, J3 suggests round robin instead. For third place it
checks whether the two fighters have already met: if they have, both receive bronze; if not, it
creates the rematch.

## Score a contest

Open the **Official console** on a tablet or laptop:

1. Select the contest and confirm you are assigned to it. Unassigned officials cannot score.
2. Start the clock. The referee can stop it; the scorekeeper resumes it on command.
3. Add techniques and penalties to the correct side. Moves appear on the scoreboard immediately
   and every change is attributed to you.
4. A submission or tap is an instant win and overrides earlier points.
5. Confirm the result. J3 asks you to confirm before any score, outcome or promotion is committed,
   then publishes it to every display and records an audit entry.

Mistakes are fixed with **Undo last event**; the correction is recorded too. Results lock after
confirmation and only a runner can override them.

## Kata and training

Kata competitions use a judge panel on phones or tablets. J3 drops the highest and lowest scores,
averages the middle scores and applies the tie-break chain you configured. Rounds can be
compulsory or chosen, include bunkai, and advance the top performers automatically. Weapon kata
and training sessions are supported as well.

## Screens and displays

Every view has its own URL, so open each one in its own browser tab and drag it onto the projector,
TV or LED wall:

| View | Use |
|---|---|
| `/public` | Public dashboard: next up, live results and event status |
| `/display` | Full-screen scoreboard with fighter photos, clubs, officials and timers |
| `/brackets` | Elimination and round-robin progression |
| `/tables` | Standings, rankings, medal tables and kata results |
| `/pools` | Every matchup with finished, live and upcoming fights |
| `/control` | Display control and PA announcements |
| `/ceremony` | Guided medal ceremony |
| `/view` or `/builder` | Your own saved, live dashboard layouts |

Use **Display control** to assign a saved template to a screen, trigger announcements and flash
a fighter card when they are due on the mat. Themes are white-label: colours, fonts, banners and
the Street Fighter display style are all configuration, not code.

## Keep the event on track

The public clock and dashboard timeline show projected start and end times. If the day runs early
or late, adjust the drift once and every projection updates. Rest warnings remind you when a
fighter has another bout too soon, and the on-deck list shows who is next.

## Dealing with changes on the day

- **Withdrawal or no-show** — mark a forfeit with a reason; the pool rebalances and promotions
  adjust without deleting the original result.
- **Injury** — record it as a technical outcome so ranking points stay consistent.
- **Voluntary move up** — with the fighter's consent, place them in a heavier or older category.
- **Too few entrants** — switch the affected category to round robin from the bracket view.
- **Duplicate bronze** — handled automatically when the two third-place fighters already met.

## Audit and exports

Every material action records who did it, when, what it affected and the before/after values.
Runners and officials can filter the **Audit trail** to resolve disputes. **Data explorer** and the
CSV buttons in **Check-in & import** export fights, entrants and standings. The full machine API is
documented in the [API reference](/docs/api).

## Need help?

- [Developer guide](/docs/development) explains commands, architecture and troubleshooting.
- [CI & releases](/docs/ci) explains the automated checks.
- The TUI menu **Docs & guides** opens this site and refreshes the generated API reference.
