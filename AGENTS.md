# J3 (J Cubed) — Agents & Project Brief

> **LIVING DOCUMENT.** This is the source of truth for context, scope, decisions and direction.
> Every agent/session should **read this file first**, act on it, and **update it** when they
> learn something, make a decision, or finish work. Keep it accurate and current.
>
> | Field | Value |
> |---|---|
> | Status | Planning — not yet implemented |
> | Created | 2026-11-19 |
> | Last updated | 2026-11-19 |
> | Working dir | `/var/home/restful-otaku/Dev/projects/j3` |
> | Repo | none yet (fresh workspace) |

---

## 0. How to use this document

1. **Read** §§1–3 for mission and principles before writing any code.
2. **Check** §14 Decision log and §15 Open questions before choosing an approach.
3. **Update** this file when you:
   - settle an open question (move it to §14 with rationale),
   - discover a new requirement or domain rule (add to §5, §6 or §7),
   - complete a roadmap item (tick it in §11 and add a §16 change-log line),
   - finish a session (write a short "Next actions" entry in §17).
4. Prefer **small, reversible decisions** recorded in §14 over silent assumptions.
5. Domain truth beats convenience: if a rule is configurable in real federations,
   model it as configuration, not a hard-coded constant (see §3).

---

## 1. Product identity & mission

**J3 (J Cubed)** is a **white-label Japanese Jiu Jitsu tournament creator and operator**.

It is not one tournament system — it is a *platform for generating any tournament*: any
organiser, any club, any rule set, any discipline, any category, any format, any display.
A J3 customer should be able to brand it entirely as their own and run anything from a
single-club sparring night to a multi-mat, multi-day, 1,000-fighter championship with
broadcast graphics — **online or with no internet at the venue at all**.

Core promise: *"One configurable engine, every Jiu Jitsu event, fully branded, fully offline-capable."*

### Primary users
- **Tournament organiser** — configures everything, runs the event.
- **Club / club leader** — registers fighters, sees their people, tracks results.
- **Fighter / participant** — profile, avatar, club logo, schedule, "you're up next".
- **Moderator / official / judge / referee** — scores, assigns, oversees.
- **Spectator / coach / parent** — follows along on displays and public pages.

---

## 2. Non-negotiable principles

1. **Everything is data-driven.** Rules, categories, scoring, formats, branding, roles and
   display layouts are *configuration*, never hard-coded. A new federation = a config, not a fork.
2. **Offline-first.** Venue Wi-Fi is unreliable by default. Scoring must work locally and
   reconcile deterministically. A single file must be a recoverable backup.
3. **Deterministic, auditable rules engine.** Given the same config + events, results are
   reproducible. Every score change has an actor, timestamp and undo trail.
4. **Multi-mat and multi-display are first-class**, not an afterthought.
5. **Accessible by default.** W/L/D must not depend on colour alone; giant screens, projectors,
   tablets portrait/landscape, phones, and LED walls all supported.
6. **Local-first, cloud-optional.** Run on a LAN box at the venue; publish to cloud later.
7. **Single source of truth for types** shared between Go backend and TypeScript frontend.
8. **Small vertical slices.** Ship a working end-to-end path early; deepen breadth after.
9. **Usable by non-technical people — fundamental.** Setting up and running the app must never
   require a tech-savvy person: plain-language guided wizards, sensible defaults, visual
   editors, validation with clear error messages, progressive disclosure (advanced options
   hidden until needed), and *never* raw config/YAML/script editing in normal flows. Power-user
   config remains available as an opt-in escape hatch, not the front door.
10. **A runnable local full demo is a first-class deliverable.** At every phase there must be a
    single command to run the **whole app locally with realistic seeded data** so the user can
    play with it and give meaningful feedback early and often. QA feedback loops outrank
    breadth of features.
11. **Automate the heavy lifting.** Organisers set requirements and constraints; the system
    does the work: category eligibility, seeding, pool generation, scheduling and promotion.
    Algorithms optimise for fairness and enjoyable events, with controlled randomisation so
    groups are well mixed and same-club fighters are spread across pools wherever possible.
12. **Multi-screen by design.** Every view must be openable in its own tab/window so organisers
    can put the control room, public dashboard, several full-screen scoreboards and rotating
    stat pages on different TVs/projectors/LED walls. All views are live-synced with the backend.

---

## 3. The core idea: everything is configuration (white label)

This is the architectural heart of J3. Model these as first-class, versioned, tenant-owned
entities so the same binary serves every customer:

| Config layer | What it controls |
|---|---|
| `Tenant` / `Organisation` | Branding, domain, data isolation, plan/feature gates, locale(s) |
| `Theme` | Logo, banner, colours, fonts, display layouts, medal/podium visuals |
| `RuleSet` | Scoring values, penalties, hold timers, match length, golden score, DQ thresholds |
| `CategoryScheme` | Age/weight/belt/gender ranges and the policy for calculating them |
| `FormatPolicy` | Which bracket format is chosen for N entrants, seeding, BYEs, separation, bronze |
| `Discipline` | Fighting System, Ne-Waza, Randori, Katamewaza, Kata, continuous, sparring, drills… |
| `RankingPolicy` | Medal points, win bonuses, team points, star rating, season/circuit rules |
| `RolePermissions` | Who can create/invite/score/override/publish |

**Config is versioned and immutable per event** (snapshot at event start) so a late rule edit
cannot retroactively change a finished result. New event = new snapshot; old events stay intact.

---

## 4. Domain model (entities)

```
Tenant ─┬─ Theme / Brand
        ├─ Users (organiser, club leader, fighter, official, moderator)
        ├─ Clubs (logo, colours, leader) ── Fighters (avatar, club, belt, age, weight, gender?)
        └─ Events
             ├─ Disciplines / Divisions / Categories (age/weight/belt/gender ranges)
             ├─ Entrants (fighter ↔ category; optional club affiliation override)
             ├─ RuleSet + FormatPolicy + CategoryScheme snapshot
             ├─ Mats / Areas (+ own queue, officials, display)
             ├─ Brackets / Pools / Stages / Fights (nodes, edges, seeds, BYEs)
             ├─ Fight events (score, technique, penalty, timer, hold, result, undo)
             ├─ Kata performances (panel judges, scores, deductions, ranking)
             ├─ Schedule (slots, drift, on-deck, callouts)
             ├─ Rankings / Medal tables / Leaderboards (overall + per club)
             └─ Announcements / PA messages / Broadcast overlays
```

Supporting: `Invitation`, `Registration`, `CheckIn`, `WeighIn`, `Assignment`, `AuditLog`,
`Incident`, `Notification`, `MediaClip`, `Export`.

---

## 5. Domain rules research — what real federations actually do

Sources consulted: JJIF Fighting Rules 2026, JJIF Jiu-Jitsu (Ne-Waza) rules 2026, SJJIF
Rulebook 2026, JJWL rules, TJJF Randori rules, WKF Kata rules 2024/2026, Kyokushin kata
rules, AJP/Abu Dhabi World Pro regulations, SJJJF/ASJJF event pages, Tap Cancer Out,
Alliance, plus commercial systems (Jhanu, Fightflow, ScoriUp, Martial Cup, Kihapp, Sidekick).
These are **inputs to the config model**, not rules to hard-code.

### 5.1 Disciplines & event types to support
- **JJ Fighting System** — 3 phases: striking (distance), throws/takedowns after grip, ground.
  Points per phase; "perfect technique across all three" = Ippon and instant end.
- **Ne-Waza / Katamewaza** — groundwork, holds (Osae-komi 10s = Waza-ari, 20s = Ippon),
  submissions = Ippon.
- **Tachi-Waza / Randori** — standing, throws; Ippon ends, Waza-ari can accumulate;
  "open" elite = 1 Ippon wins.
- **Continuous ground fighting** and **continuous fighting** (standing → throw → ground).
- **Free sparring sessions / rotation** (non-tournament, open mat, rotation schedules).
- **Kata judged competitions** — individual, pair, team, bunkai.
- **Training / drills** with guest Sensies and Shihans (seminars, not just contests).
- **Duel / sport variants**, Gi and No-Gi (with different weight sets and match times).

### 5.2 Age / weight / belt / gender
- **Age policy varies**: some federations use *age reached in the calendar year*; others
  require the athlete to be that age **on the day** (U21/Adult often 18 on event day).
- Ranges seen: U8/U10/U12/U14/U16/U18/U21, Adult, Masters 30/35/36/40+/46/51/56+, kids
  by birth year (Kid 1–6).
- **A competitor may compete in >1 age division in some rulesets (JJIF Fighting) but only
  one age and one weight in others (JJIF Ne-Waza).** Must be configurable.
- **Weight categories**: JIUDO/JJIF sets (~40–94 kg bands), SJJIF IBJJF-style names
  (Rooster 57.6, Light Feather 64, Feather 70, Light 76, Middle 82.1, Medium Heavy 88.2,
  Heavy 94.1, Super Heavy 100.2, Ultra Heavy +100.2), separate Gi/No-Gi values, separate
  male/female and youth/adult tables.
- **Belt / grade divisions** (missing from initial brief — critical): White→Blue→Purple→
  Brown→Black; youth Gray/Yellow/Orange/Green groups; masters by birth year. Some federations
  gate white belts (e.g. max 10 wins, then must move up; no jumping guard for white belts
  from 2027).
- **Full JJJ belt-colour spectrum must be representable** as ordered rank data points (each
  with `order`, `colour`, `label`, `min/max age`, `stripes`, `dan/kyu`, custom flag): White,
  Yellow, Orange, Green, Blue, Purple, Brown, Black; youth Grey (Grey/White, solid Grey,
  Grey/Black), Yellow/Orange/Green striped variants; high-grade Red-and-White *mon* and Red;
  plus **arbitrary custom colours/ranks** per tenant. Belt grade feeds category ranges,
  eligibility/rank gating and display styling — never a hard-coded enum.
- **"Ranking" is two distinct concepts**: (a) belt/grade rank (above) and (b) leaderboard
  ranking (§5.7). Both are first-class data points, not display-only decoration.
- **Gender is optional**: support male/female/open/co-ed, and "no gender split" mode. Some
  kids divisions forbid cross-gender transfer — configurable.
- **Height is a first-class stat** alongside weight (ranges and category eligibility can use
  either). Users may enter **metric or imperial**; the **backend converts and stores metric**
  (kg, cm) as canonical, keeping the original input for display fidelity.
- **Match length policy** by age/belt/Gi-NoGi: 1.5 min (U8/U10) up to 5 min (adult), 6 min
  adult black belt in some rules, 5 min all adult No-Gi.
- **Weigh-in** with tolerance, re-weigh, out-of-class flagging (see §7).

### 5.3 Formats by entrant count (real reference policies)
Common patterns to encode as selectable `FormatPolicy` strategies:
- **1** → no contest / single winner (with its own ranking points rule).
- **2** → best of three (stop at 2–0).
- **3** → round robin.
- **4** → round robin **or** single elimination + 3rd-place match.
- **5** → round robin (pool of five).
- **6** → two pools of 3 (top 2 cross to semis) **or** multi-stage.
- **7+** → double elimination with repechage / 3rd-place playoff.
- **9+/17+** → single elimination until round of 16/8, then double elimination; limits vary.
- **True double elimination** (lower-bracket winner can win Gold via reset) vs **consolation
  double** (loser can only win Bronze) — must be explicit, not ambiguous.
- **BYE allocation** by seed/title; **first-round club/country separation**; avoid same-club
  early meetings; configurable 3rd-place match; "two-match guarantee" policies.

### 5.4 Scoring systems
- **Ippon / Waza-ari / Shido** (JJIF/Judo-style); Ippon ends the contest (except formats
  where 2 Ippon required). Control must be held ~3 seconds; points count even if escaped
  immediately after the 3rd second.
- **BJJ-style points**: takedown 2, sweep 2, guard pass 3, mount/back 4, knee-on-belly 2;
  advantage as 1; submission = instant win overriding points.
- **Fighting System**: unblocked strike 2 / blocked 1; perfect throw 2 / partial 1;
  submission 2–3; Osae-komi hold 10s = 1, 15s = 2, 20s = 3.
- **Hold timers** with live countdown; **submission/tap (or verbal tap) = instant win** that
  negates prior points and triggers promotion.
- **Best of N** (2 of 3 / 3 of 5): distinct from multi-round kata.
- **Golden score / overtime**: e.g. 2 × 2 min, first point wins.
- **Kata panel scoring**: scale 5.0–10.0 (0.1 or 0.2 increments); drop highest & lowest;
  average the middle 3 or 5; 5 or 7 judges; tie-breaks reinstate low then high then majority
  vote; deduction guidelines (−0.2…−1.6); compulsory (shitei, drawn) vs choice (tokuyi),
  must differ each round; 3 rounds, top 16 → 8; bunkai for team medal bouts.

### 5.5 Penalties / disqualification
- **Shido ×3 = Hansoku-make** (DQ) in JJIF/Judo style; Waza-ari awarded to opponent for
  light fouls in Fighting System.
- **BJJ style**: −1, −2, −3 on the board → DQ on 3rd; verbal warning for brown/black.
- **Yellow card / red card** (red = 2 fouls or a bad foul) — user requested; model as
  configurable penalty ladder with outcome (point deduction, opponent point, DQ).
- **Forfeit** from injury / technical reason — user requested; must record reason and apply
  consistent result rules (who advances, ranking points, "must fight at least once" rule).

### 5.6 Tie-breakers (configurable ordered chain)
Wins → submissions → points scored in winning bouts → head-to-head direct combat →
fewest points conceded → fastest submission → positive play / corner-judge majority.
Different federations pick different chains — make it an ordered, editable list.

### 5.7 Rankings & team points
- Medal points (e.g. Gold 30 / Silver 14 / Bronze 6; star-rated events scale this up to
  225/105/49). Single-entry champion often = silver points.
- Win bonuses: submission +3, other win +1 (SJJIF/ASJJF).
- Team points accumulate Gold/Silver/Bronze of all members; separate Kids and Adult/Master
  tallies; brackets composed solely of one team may score 0.
- Federations own season rankings, regions/circuits, leaderboards, public pages + API.

### 5.8 Scheduling reality
Rest 5 min between bouts, 10 min before medal bouts; multi-day, multi-mat, judge/competitor
conflicts ("judges not competing during assigned judging"); estimated finish times from
round lengths; drift recalculation; on-deck / double-deck callouts.

---

## 6. Feature map (what to build)

### 6.1 Setup & white-label
- **Guided, plain-language setup wizard for non-technical organisers** (principle 9): step-by-step
  questions, live preview, sensible defaults, inline validation, "advanced" collapsed by
  default, and a one-click "use a template" path. No raw config files in the normal flow.
- Wizard to create organiser, event, branding (logo, banner, colours, fonts), domain.
- Customisable categories, disciplines, events; versioned rule-set snapshots; templates
  ("copy last year's event"). Invite clubs and participants; moderation roles.

### 6.2 People & roles — RBAC + invite-only hierarchy
The v1 authorisation model is **RBAC with an invite-only chain of authority**:

1. **Tournament runner** — full admin; configures everything, can create/invite/delete/override.
2. **Officials / moderators** — less control than the runner; score, officiate, assigned per
   mat/bout; cannot change event configuration or other people's authority.
3. **Club leaders** — moderate **their own club's** participants only; can register/add fighters
   under their affiliation, invite them, manage club branding, view their club's data.
4. **Fighters** — regular users with basic access: self-register independently, or accept a
   club leader's invitation to register under that club affiliation.

- RBAC governs **both visibility and capability**: private vs public data, and create / read /
  update / delete rights, locked to the authority each role legitimately has in the tournament.
- Purpose: integrity, fairness, trust and **anti-abuse / anti-cheating** — nobody can alter
  results outside their authority.
- **Full auditability:** every material action is an event (actor, role, target, timestamp,
  before/after), so accountability is assured and disputes can be resolved.
- Official assignment per mat/bout; **re-authorise every join** (unassigned corners cannot score).
- Fighter profile with avatar + club logo.
- Public/private boundary: spectators only see explicitly published data; internal scoring,
  drafts, notes and personal data stay scoped to their role.

### 6.3 Registration, check-in, weigh-in
- **Self-service sign-up** collects email, phone number, name, age/DOB, **height and weight**
  (metric or imperial; backend converts to metric), **rank/grade** (belt colour / Dan / kyu /
  youth / custom), plus account type.
- Account type drives the next step: choose a **position/role** (official/moderator/runner/
  organiser), **or** choose event participation from the tournament's available selections —
  only the categories the user is **compatible with based on their stats** are offered.
- Invitations (link/QR), club bulk entry, club-affiliation choice, payments, waitlists, refunds.
- Check-in, QR badges, weigh-in with tolerance/re-weigh/out-of-class flags (see §7).
- Light and dark appearance with system-following default and per-user override (§9).

### 6.4 Divisions & categories
- Age ranges (minors/adults) with calendar-year vs event-day policy; weight ranges (Gi/No-Gi);
  belt/grade divisions; optional gender; ability to merge/split categories by entrant count.

### 6.5 Formats, brackets & promotion
- Round robin, pools, single/double elimination, multi-stage, repechage, best-of-N, BYEs,
  seeding, club/country separation, 3rd-place playoff, **dual bronze winners** option.
- Automatic, consistent promotion of winners/losers/draws per configured rules; **rematches
  for 3rd place if the two fighters have not already met**; avoid repeat pairings where configured.
- **Automated pool generation:** given organiser constraints and eligible entrants, create as
  many balanced pools as each event needs; seed by rank/stats; use controlled randomisation to
  mix groups well and **avoid same-club clustering**; expose a human override with re-balance.
- Drag & drop fighter cards and match-order lists between brackets / into next rounds / to
  promote, with full animation.

### 6.6 Live bout engine (scoreboard)
- Score, round timer, countdown, hold/pin timers, technique & move callouts with points,
  last-move ticker, instant-win submission/tap, penalties (yellow/red/shido), forfeit/injury.
- W/L/D colour coding **plus glyphs**, undo per judge/corner, adjudication, result locking.
- Multi-round bouts (best of 3/5) and golden score.

### 6.7 Kata / judged performances
- Panel judges on phones/tablets, scores/deductions, drop high/low, tie-break chain,
  compulsory vs choice kata catalogue (multi-style, multi-language), team kata & bunkai,
  order draw, live leaderboard.

### 6.8 Displays & visualisations
- Bracket charts, pool tables, leaderboards, rankings, medal tables, club comparisons —
  all using club/fighter logos.
- **Dynamic scoreboard** with fight pools & categories for discipline/overall.
- **Dynamic live fight display**: avatars, names, score, timer, countdown, technique callouts,
  animations, last moves + points, penalties, forfeit. Same for Kata demonstrations.
- Themable layouts for phone, tablet (portrait/landscape), projector, huge LED wall; zoom,
  dark/light, kiosk/fullscreen.

### 6.9 Scheduling, clock & callouts
- Schedule with planned times, drift (ahead/behind) auto-update, on-deck/upcoming lists,
  medal-award callouts with meeting point (podium/mat/area), **PA announcement** trigger
  showing name/message on main screen (TTS + multilingual).

### 6.10 Rankings, medals & leaderboards
- Overall + individual, per-club + per-category, live updating, medal table with
  Bronze/Silver/Gold and dual bronze.

### 6.11 Multi-mat / multi-display
- Several mats and several live scoreboards simultaneously + overall event dashboard +
  "fighters next up" so everyone stays oriented.

### 6.12 Broadcast & media
- Transparent HTML overlays for OBS/vMix (scorebug, lower thirds), public live page,
  multi-camera phone replay (VAR), clip library, stream links.

### 6.13 Notifications & portals
- Push/SMS/email, "you're up next", QR check-in, organiser portal, competitor portal,
  coach/club view.

### 6.14 Reporting & data
- Printable/PDF brackets, scoresheets, programmes, medal certificates; CSV/XLSX import/export;
  historical archive; API access; scoring audit + incident log.

### 6.15 Real-time data flow & integrity
- Fight events (score changes, moves/techniques, penalties, results) are captured live,
  tallied, and **committed to the database as they are confirmed**. All visualisations —
  brackets, charts, tables, cards, leaderboards, displays — **query this committed data**
  rather than keeping separate state.
- Data model: a `users` table plus dedicated tables for everything that needs tracking
  (stats, scoring events, audit, rankings, etc.) so logic and analytics read one truth.
- Every write is attributable and auditable (§6.2); results are locked after confirmation,
  with explicit, recorded undo/correction paths.

### 6.16 Multi-screen views, displays and spectator experience
From direct user QA feedback (2026-11-19) — all of this is required, not optional polish:

- **Responsive to the extremes.** Ultrawide monitors must use the full available width;
  phones, tablets portrait/landscape, projectors and LED walls all get appropriate layouts.
  Fighter photo/avatar areas must be **large enough to read from across a hall**.
- **Every view opens in a new tab/window** (own URL) so the control room, public dashboard,
  several full-screen scoreboards and rotating stats pages can be placed on separate screens.
- **Public dashboard:** key info at a glance — upcoming events, next group of participants who
  must report to a mat/podium, current results and general event status.
- **Scoreboard must show:** accurate event title; mat number; tournament banner + logo; each
  fighter's name, avatar, **belt designation**, height/weight, **affiliated club logo** (or a
  default independent mark); **active referee and judges/officials with names, faces and roles**;
  tracked moves **and penalties on each fighter's respective side**; timer that **pauses when the
  referee stops the fight** and resumes on the scorekeeper's/moderator's command (tablet/laptop).
- **Dedicated table/ranking views:** overall tables, category tables, medal/score rankings,
  queryable/filterable by group, age bracket, weight bracket, height bracket and discipline.
  **Kata needs its own view** with panel judges and gymnastics-style score display.
- **Dedicated bracket view:** classic elimination and round-robin bracket charts showing
  progression through rounds (quarter → semi → final → third-place/bronze rematch), who won/lost,
  who is paired next. Moderators can move fighters into generated slots as they win, **keeping a
  copy of each fighter at every round so the full history is visible**.
- **Result icons/configurable outcomes:** WP/LP/DP (win/loss/draw on points), WS/LS (submission),
  WT/LT (technical/forfeit/injury/DQ/red card), plus yellow/red card badges on fighter cards.
  All enabled/disabled by organiser rules and constraints.
- **Dedicated fight-pools view:** all matchups, ongoing fights highlighted, finished fights
  marked with result, upcoming fights listed.
- **Configurable dashboard templates:** savable templates defining which widgets each screen
  shows; runners assign templates to screens. All views live-synced with the backend.
- **Confirmation dialogues** for major actions: submitting a score/point, result, winner/loser/
  draw and promotions.
- **Celebrations & animation:** winner round/overall zoom-in of the player card, medal wins,
  slick transitions; optional **Street Fighter style** theme with lineups, matchups and announcer
  energy.
- **PA/announcer interaction system:** timed toasts/popups and shoutouts; flashing fighter cards
  to call competitors to their mat/area. Helps organisers, competitors and spectators stay
  informed and keeps the event running smoothly.

### 6.17 Developer & operator CLI
- A **nice TUI-style, navigation-friendly `dev.sh`** with emoji, spinners and menus for:
  dry running the app, QA testing, resetting demo data, building a production/live executable,
  deploying it, and building **Capacitor mobile variants for Android and iOS/iPhone**.
- Must remain easy for non-technical users: numbered menu, plain-language choices, clear status.

### 6.18 Tournament creation pipeline (full setup from scratch)
The core "create a tournament" experience must be a **simple, coherent ~5-step pipeline** that a
non-technical runner completes in **under 10 minutes**, with most work automated:

1. **Event & identity** — name, title, dates, venue, branding/theme, banner/logo, presets.
2. **Disciplines & rules** — per-discipline rules, requirements, point systems and attribution,
   timers, round counts, event titles, tournament style per event/discipline/category.
3. **Categories & constraints** — weight/height/rank ranges, gender policy, participant cap,
   max fight pools, category fulfilment rules and automatic placement settings.
4. **People & invitations** — bulk email/phone invites (email + optional SMS) to runners,
   moderators/officials, club leaders and participants; delegate roles; individual self-signup.
5. **Review & launch** — sanity check + validation; on error, inform the user with a popup and
   jump to the offending section; on success create everything, auto-generate kata and fight
   pools, assign officials, create the dashboard/views and **save the whole setup as a template**.

- Steps are **URL/persisted views** that can be stepped forward/back without losing data.
- Widgets: text/number fields with **sensible defaults + "choose one of these" presets** and
  "enter my own"; checkboxes, radio/boolean selectors, sliders, and **draggable reordering** of
  event/matchup cards and bracket slots.
- **Notifications/toasts:** errors, validation, warnings and success ("created/updated/deleted/
  saved/changed") stack in arrival order, slide/fade in and out, are dismissable, show a progress
  bar counting down to auto-dismiss, and offer **undo within 10 seconds** for material actions.
- **Auto-generated bracket diagram:** built from fight pools, with fighter cards in each slot,
  progression/knockout/promotion highlighting, result icons, cards/fouls/penalties, and drag/
  shuffle by moderators with the right access. Keep a copy of each fighter at every round.
- Everything saved as reusable **templates** (tournament setup + custom dashboard views) for
  future events. Users fill in their details/stats and chosen events; the app does the rest.
- **Media uploads are mandatory for a professional, official look:** fighter/official photos,
  club symbols/logos, runner/organiser logo and event banner, club colours and belt/rank colours.
  Uploads are stored locally with the venue data and referenced by theme/club/fighter records so
  they appear on scoreboards, brackets, tables and cards.
- **Event auditing is essential and visible to runners and officials:** every action records who
  did it, when, what it affected, and how (before → after). A dedicated audit view lets runners
  and officials review the full trail to prevent cheating, bias or abuse.

---

## 7. Gaps & things not in the original brief (research-backed)

The brief was broad but these are notable omissions that a credible tournament platform needs.
Confirm which to include in v1 vs later:

1. **Belt / grade divisions and rank gating** — arguably the biggest domain gap (§5.2).
2. **Weigh-in workflow** (tolerance, re-weigh, out-of-class, QR badges).
3. **Registration, payments, refunds, waitlists, waivers/consent, insurance** and GDPR data
   ownership (white-label tenant = controller).
4. **Gi vs No-Gi division tables and uniform/belt-display checks.**
5. **Match-length policy by age/belt/Gi-NoGi** rather than one global timer.
6. **"True double" vs "consolation double"** semantics and repechage (avoid ambiguous "double elim").
7. **Explicit tie-break chains** and "must fight at least once to qualify for ranking/podium".
8. **Team/club points, star-rated event ranking points, season circuits** (central to multi-club).
9. **Kata specifics**: compulsory-draw vs choice, 3 rounds, deduction guidelines, bunkai,
   catalogues, judge-panel randomisation.
10. **Rest intervals / judge–competitor conflict-aware scheduling** and drift (§5.8).
11. **Offline LAN architecture** (venue server, QR judge pairing, WebSockets, single-file backup)
    — validates SQLite choice but needs a sync design (§8.3).
12. **Audit trail, undo, score reconciliation, per-join re-authorisation** (tamper resistance).
13. **Broadcast overlays, VAR/multi-camera replay, AI/TTS announcer, multilingual (EN/JP+)**.
14. **Import/export & archive** (Excel/CSV start lists, decades of results, public API).
15. **Accessibility/colour-blind safety** for W/L/D and card colours.
16. **Operational resilience**: UPS, crash recovery mid-event, backup/restore, migrations,
    observability, load headroom (target ≫ 20 mats × 5 judges).

---

## 8. Technical architecture

### 8.1 Stack (as requested)
- **Backend:** Go (single binary; pure-Go SQLite `modernc.org/sqlite` for portability, optional
  PostgreSQL for cloud/scale), REST + WebSocket, OpenAPI spec.
- **Frontend:** SvelteKit + TypeScript (one app, multiple "modes" via routes), Tailwind/CSS vars.
  **Package manager / runner / test runner: Bun (not npm).** Use `bun install`, `bun run`, `bunx`.
- **Mobile/Desktop:** Capacitor wrapping the Svelte app for iOS/Android/tablets; PWA for web.
- **Device DB:** local SQLite (`@capacitor-community/sqlite` / `op-sqlite`) with outbox.
- **Shared types:** generate TS from OpenAPI/Go types — no hand-maintained duplicates.

### 8.2 Topology
```
Cloud (optional) ──publish── Venue Server (Go + SQLite, LAN hub)
                                  │ WebSocket/REST
        ┌─────────────────────────┼─────────────────────────┐
   Judge phones            Operator/control room       Displays/mats
   (QR pair, score)        (brackets/schedule)         (scoreboards, LED, projector)
        └──────────── Capacitor apps / PWA ────────────┘
```
Venue server is authoritative for scoring order; devices keep a local replica + outbox so
they keep working offline and reconcile when the LAN returns.

### 8.3 Sync & consistency (design direction — decide in §14)
- **Event-sourced outbox**: local ops (`ScoreDelta`, `Penalty`, `Result`, `MoveEvent`) queued,
  sent with idempotency keys and Lamport sequence; server assigns global order.
- **Conflict policy**: scoring = authoritative server sequence (last-writer-wins with
  per-judge undo); config = LWW with version; counters = commutative deltas (CRDT-like).
- Explicit **reconcile** check: display board state vs DB; visible "synced/offline" indicator.
- A single SQLite file = portable backup; WAL mode; point-in-time restore mid-event.

### 8.4 Multi-tenancy & theming
- Row-level tenant isolation + per-tenant config snapshots; feature gates per plan.
- Themes as CSS custom properties + asset bundle (logo/banner/fonts) served by backend;
  white-label domain + email templates + localisation per tenant.

### 8.5 Proposed core tables (sketch — refine during implementation)
`tenants, users, clubs, fighters, events, disciplines, categories, entrants, rule_sets,
format_policies, mats, brackets, stages, nodes, fights, fight_events, kata_performances,
panel_scores, assignments, schedules, callouts, announcements, rankings, medals, audit_log,
incidents, invitations, registrations, weigh_ins, outbox, media_clips, themes`.

### 8.6 Algorithms (pure Go packages, golden/property tested)
- **Category assignment** from DOB/weight/belt/gender + policy.
- **Format selection** by entrant count + policy; **seeding** (rank, club/country separation, BYE rules).
- **Bracket generation** (RR round-robin scheduling — circle method; single/double elim; pools;
  multi-stage; repechage; 3rd-place with "have they met?" rematch rule).
- **Result propagation** — winner/loser/draw → next node, statistics, medals, rankings, team points.
- **Scoring resolver** — apply move/point/hold/submission/penalty events to compute live score,
  decide instant win, DQ, golden score, best-of-N, then lock result.
- **Tie-break chain** evaluation (ordered rules).
- **Scheduler** — mat/slot assignment, rest conflicts, drift recalculation, on-deck computation.

### 8.7 Non-functional requirements
Offline correctness · sub-second sync · no data loss on reconnect · accessible/contrast-safe ·
responsive 6" phone → 100" LED · i18n (EN + JA first) · auditability · fast redraw on huge screens ·
crash recovery · migrations · load headroom well beyond a real event.

---

## 9. UX & design language
- Calm, modern, high-contrast; generous spacing; **one obvious primary action per screen**.
- White-label theming via tokens; fight corner colors (red/blue) consistent everywhere.
- W = green + "W", L = red + "L", D = amber + "D" (never colour alone).
- Drag & drop with spring animations; every structural change undoable.
- Dense "operator" views vs calm "display" views for the same data.

---

## 10. Platform targets (Capacitor & responsive)
- **Phones:** judge scoring, fighter/club portal, quick check-in.
- **Tablets portrait & landscape:** mat-side scoreboard, judge panel, operator console.
- **Desktop/LED/projector:** displays, brackets, control room, broadcast output.
- Capacitor plugins: camera (avatars/logos), filesystem, local notifications, share,
  keep-awake/kiosk, status bar; safe-area handling; offline cache.

---

## 11. Roadmap (vertical slices; tick as completed)

- [x] **Phase 0 — Foundation**: repo, Go API + SQLite, SvelteKit shell, shared types/OpenAPI,
      AGENTS doc, basic tenant/theme, theming tokens, testing harness, seeded demo + QA dashboard.
- [x] **Phase 1a — Sign-up & automated pools**: imperial→metric conversion, stats capture,
      compatibility (strict + consent move-up), role request/approval, constrained balanced pool
      generation with same-club separation, QA views. Golden/unit tests.
- [x] **Phase 1b — Multi-screen displays & operator CLI**: responsive ultrawide layouts, URL-
      addressable public dashboard / full-screen scoreboard / brackets / tables+kata / pools /
      display control, big fighter cards, officials, per-side moves & penalties, PA toasts,
      confirmation dialogues, celebration animations, Street Fighter theme, dashboard templates,
      TUI dev.sh with live build and Capacitor Android/iOS targets.
- [ ] **Phase 1 — Setup slice**: create event → clubs → fighters → categories → rule-set
      snapshot → simple single-elim bracket → manual result → medal table. End-to-end offline.
- [ ] **Phase 2 — Formats & rules engine**: RR, pools, double elim, repechage, BYEs/seeding,
      tie-breaks, dual bronze, third-place rematch rule. Golden tests.
- [ ] **Phase 3 — Live scoring**: WebSocket + outbox sync, judge phone UI, bout engine
      (timer, techniques, holds, submission, penalties, forfeit, undo), multi-mat.
- [ ] **Phase 4 — Displays & scheduling**: scoreboard display, bracket/leaderboard/chart
      visualisations, schedule + drift + on-deck, callouts + PA/TTS.
- [ ] **Phase 5 — Kata & drills**: panel judging, kata catalogue, deduction rules, drills/
      sparring rotation, guest sensie/shihan sessions.
- [ ] **Phase 6 — White-label depth & scale**: portals, registration/payments/check-in/
      weigh-in, rankings/team points, broadcast overlays/VAR, i18n JA, mobile polish,
      import/export, hardening.

> Order may change; keep this list honest. Each phase should be demoable.

---

## 12. Definition of done & testing
- Rules engine: deterministic golden tests + property tests (no duplicate pairings, promotion
  consistency, medal uniqueness, dual bronze cases, third-place rematch avoidance).
- Sync: offline disconnect/reconnect, concurrent judges, idempotency, undo, replica reconciliation.
- UI: Vitest + Playwright for critical flows; drag & drop; keyboard/a11y; colour-blind check.
- Platforms: phones, tablets portrait/landscape, projector/LED aspect ratios, kiosk.
- Perf: simulated ≫ 20 mats × 5 judges; sub-second update budget.
- Docs: config reference + rule-set authoring guide + operator runbook.

---

## 13. Glossary
**Ippon** — full point / instant win. **Waza-ari** — half/partial score. **Shido** — penalty.
**Hansoku-make** — disqualification. **Osae-komi** — hold/pin. **Repechage** — second-chance
bracket for 3rd place. **Ne-Waza** — ground grappling. **Tachi-Waza** — standing/throwing.
**Randori** — free practice/sparring. **Kata** — formal pre-arranged sequences. **Bunkai** —
application explanation. **Shitei / Tokuyi** — compulsory / chosen kata. **Tatami** — mat area.

---

## 14. Decision log
_Format: `D-### | date | decision | rationale | status`._

- **D-001 | 2026-11-19 | Use AGENTS.md at repo root as the living context document**
  — requested explicitly by the user. Status: accepted.
- **D-002 | 2026-11-19 | Treat rules/formats/branding as versioned config snapshots per event**
  — required for true white-label + reproducible results. Status: accepted (user confirmed
  generic + custom presets for v1; config editor is mandatory).
- **D-003 | 2026-11-19 | Venue server authoritative; devices keep local replica + outbox**
  — offline venues + consistent scoring; scoring committed to DB on confirmation, all
  visualisations query committed data. Status: accepted (conflict policy to be detailed).
- **D-004 | 2026-11-19 | Pure-Go SQLite (`modernc.org/sqlite`) for the portable venue binary**
  — no CGO, single-file backup. Status: proposed (validate perf during Phase 0/1).
- **D-005 | 2026-11-19 | v1 ships one generic configurable ruleset, no federation presets**
  — user chose "generic + custom only" to avoid baking in federation specifics. Ranking and
  belt grades are nonetheless first-class rules/data points: full JJJ colour spectrum
  (White→Black, youth Grey/striped variants, Red-White *mon*, Red) plus tenant-defined custom
  ranks, feeding category ranges, gating and display. Status: accepted.
- **D-006 | 2026-11-19 | v1 auth = RBAC invite-only hierarchy with full auditability**
  — tournament runner (admin) > officials/moderators (limited) > club leaders (own club only)
  > fighters (independent self-register or club-invited). RBAC controls visibility
  (private/public) and CRUD by authority; designed for integrity, fairness and anti-abuse;
  every material action is an auditable event. Status: accepted.
- **D-007 | 2026-11-19 | v1 is venue-only; cloud publishing comes later**
  — LAN venue server, local displays; no public cloud component in the first release.
  Status: accepted.
- **D-008 | 2026-11-19 | Initialise git repo now; make regular atomic commits**
  — requested by user; enables rollback and easy review. Status: accepted.
- **D-009 | 2026-11-19 | Begin with Phase 0 scaffolding**
  — user selected "Scaffold Phase 0 repo". Status: accepted.
- **D-010 | 2026-11-19 | Zero-tech setup is a fundamental product requirement**
  — user stated the app must be usable by non-technical people; drives guided wizards,
  plain language, defaults, validation and progressive disclosure (§2 principle 9, §6.1).
  Status: accepted.
- **D-011 | 2026-11-19 | Use Bun for all frontend package management, scripts and tests**
  — explicit user instruction; no npm/pnpm. Status: accepted.
- **D-012 | 2026-11-19 | Local full-stack demo + seed data is a priority deliverable**
  — user wants to play with the running app and QA early/often (§2 principle 10). Provide a
  single command (`bun run dev`) that boots the venue server + web UI with seeded demo data,
  plus a reset/seed command. Status: accepted.
- **D-013 | 2026-11-19 | Seed dummy accounts at every RBAC level with instant switching**
  — QA must be able to see that each user type works and looks right. Seed a runner, officials,
  club leaders (per club), independent and club-invited fighters, and a spectator; expose a
  demo account list + "act as" mechanism and a one-click account switcher in the UI.
  Dev/demo only. Status: accepted (scale increased by D-014).
- **D-014 | 2026-11-19 | Demo seed scale: ~60 accounts**
  — user asked for **2 tournament runners, 8 officials, 4 club leaders, 48 fighters across a
  few clubs** (plus a spectator). Seed must be generated programmatically and feel realistic
  enough to QA tables, brackets, lists and role switching at realistic volume. Status: accepted.
- **D-015 | 2026-11-19 | Self-service sign-up with stats and role/participation choice**
  — collect email, phone, name, DOB, height & weight (imperial converted to metric on the
  backend), rank/grade. Account type then selects an official/runner/organiser position **or**
  participation in compatible event categories. Privileged roles must still be approved/invited
  (see §15) to preserve the RBAC integrity principle. Status: accepted, approval flow pending.
- **D-016 | 2026-11-19 | Automated, constrained pool generation**
  — organisers set requirements/constraints; algorithms compute eligibility, balanced pools,
  seeding and promotion, with controlled randomisation and same-club separation, plus manual
  override. Status: accepted (design in §8.6).
- **D-017 | 2026-11-19 | Light and dark appearance**
  — system-following default + per-user override + white-label tokens. Status: accepted.
- **D-018 | 2026-11-19 | Privileged roles require request + approval**
  — sign-up may request official/runner/organiser, but an existing runner must approve;
  fighters and spectators register instantly. Prevents self-elevation. Status: accepted.
- **D-019 | 2026-11-19 | Category compatibility = strict ranges + consent to move up**
  — offer exactly eligible divisions; heavier/older divisions only with explicit user consent.
  Status: accepted.
- **D-020 | 2026-11-19 | Dedicated, URL-addressable display views**
  — public dashboard, full-screen scoreboard, brackets, tables/rankings, kata, pools and control
  room each get their own route so they can be opened on separate screens/tabs. Live-synced.
  Status: accepted (§6.16).
- **D-021 | 2026-11-19 | Savable dashboard templates for screen assignment**
  — runners configure which widgets each screen shows; templates saved and reused. Status:
  accepted (design).
- **D-022 | 2026-11-19 | Confirmation dialogues + celebration animations + PA toasts**
  — major actions require confirmation; results trigger animated celebrations; announcer can
  trigger timed toasts/cards. Street Fighter theme is an optional white-label theme. Status:
  accepted.
- **D-023 | 2026-11-19 | TUI menu for dev.sh incl. Capacitor Android/iOS builds**
  — user requested an emoji/spinner, navigation-friendly CLI for dry runs, QA, live executable
  build/deploy and mobile variants. Status: accepted (§6.17).
- **D-024 | 2026-11-19 | Five-step persisted setup pipeline with validation**
  — create/theme/configure a whole tournament in <10 minutes; presets + custom values; final
  sanity check jumps to the offending step. Status: accepted (§6.18).
- **D-025 | 2026-11-19 | Stacked toasts with progress bar and 10s undo**
  — one notification system for validation, warnings, success and errors, with undo for material
  actions. Status: accepted.
- **D-026 | 2026-11-19 | Bracket diagram auto-generated from pools**
  — interactive slots, drag/shuffle by moderators, per-round history and result highlighting.
  Status: accepted.
- **D-027 | 2026-11-19 | Local media uploads for avatars, logos, banners, club/rank colours**
  — multipart upload endpoint + local file serving; references stored on theme/club/fighter/
  official records and used by every display. Status: accepted.
- **D-028 | 2026-11-19 | Runner/official audit view**
  — queryable audit log showing actor, role, timestamp, action, target and before/after values,
  to deter and detect cheating, bias or abuse. Status: accepted.

## 15. Open questions
1. True double vs consolation double as the default? (Plan: make configurable either way.)
2. Target locales for v1 — English + Japanese confirmed?
3. Preferred Go→TS type-generation approach (OpenAPI codegen vs hand-written schema)?
4. Minimum LED/projector resolutions and aspect ratios to design for?
5. Any existing brand assets/domain before theming is built?
6. Which tables/stats should be in the very first schema migration (users + core entities agreed)?
7. Can a self-signup **request** an official/runner role (pending approval by an existing
   runner), or must privileged roles always arrive by invite? (Recommendation: request +
   approval; never self-elevate.)
8. What is the exact compatibility rule set for offered categories (strict ranges vs nearest
   eligible division / moving up a weight with consent)?

## 16. Change log
- **2026-11-19** — Completed Phase 1b from user QA feedback: dedicated live routes
  (`/display`, `/public`, `/brackets`, `/tables`, `/pools`, `/control`), live polling, ultrawide
  CSS, large fighter portraits, officials/standings/event API, method icons, confirmation
  dialogues, celebration animation, PA toast broadcasts across tabs, Street Fighter theme,
  savable dashboard templates and an emoji TUI `dev.sh` with Capacitor build options. Added
  §6.16/§6.17 and D-020–D-023.
- **2026-11-19** — Created document: mission, principles, config model, domain model,
  research findings (§5), gaps (§7), architecture (§8), roadmap (§11), decision log,
  open questions.
- **2026-11-19** — Recorded user decisions: generic+custom rulesets with full belt/ranking
  data (D-005), RBAC invite-only hierarchy with auditability (D-006), venue-only v1 (D-007),
  git + atomic commits (D-008), start Phase 0 (D-009). Added §5.2 belt-colour detail,
  §6.2 RBAC hierarchy, §6.15 real-time data flow. Trimmed answered questions.
- **2026-11-19** — Completed Phase 1a: `units` (imperial→metric) and `rules` (eligibility +
  deterministic balanced pools) Go packages with tests; `users` stats columns, `role_requests`,
  `pools`/`pool_members` migration; sign-up, preview, approval and pool API endpoints; OpenAPI
  types regenerated; Sign-up and Pools QA tabs. Added D-018/D-019.
- **2026-11-19** — Added zero-tech setup (D-010), Bun tooling (D-011), local full demo (D-012),
  ~60 seeded accounts with role switcher (D-013/D-014), self-service sign-up + automated
  pool generation + light/dark mode (D-015/D-016/D-017). Completed Phase 0: Go venue server,
  SQLite migrations, RBAC roles, OpenAPI-generated types, SvelteKit QA dashboard, `bun run dev`,
  tests and CI.

## 17. Immediate next actions
1. Guided non-technical setup wizard (organiser sets constraints) → clubs → categories →
   rule-set snapshot → bracket/pool generation → manual result → medal table, end-to-end offline.
2. Live scoring engine: authoritative timers, techniques, holds, submission instant-win,
   penalties, W/L/D and multi-mat sync (Phase 3). Confirmation + promotion must be audited.
3. QA each slice with `bun run dev` (TUI menu) or `bun run demo`, plus `bun run reset`. Verify
   that every display route still opens standalone on a second screen.

---

_End of living document. Agents: please update §§14–17 as you work._
