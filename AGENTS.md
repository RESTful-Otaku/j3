# J3 (J Cubed) — Agents & Project Brief

> **LIVING DOCUMENT.** This is the source of truth for context, scope, decisions and direction.
> Every agent/session should **read this file first**, act on it, and **update it** when they
> learn something, make a decision, or finish work. Keep it accurate and current.
>
> | Field | Value |
> |---|---|
> | Status | Implementation — officials console, full JJJ catalogue, technical results, mid-event 2027 seed, branded standalone offline-capable Android APK and phone-scoped UX complete; all local gates green (2026-11-22) |
> | Created | 2026-11-19 |
> | Last updated | 2026-11-22 |
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
13. **Final quality bar — thoroughness and polish.** Before wrap-up: rigorous inspection,
    refinement, bug-fixing, testing and verification of absolute correctness. No bad states.
    A solid, simple, informative, beautiful UI/UX that is a joy to use, with stunning animations,
    snappy Street-Fighter-style "wow" factor, awesome transitions, movement and dynamism — while
    remaining a dependable tournament application that minimises organiser effort.
14. **Production-grade verification.** Extensive automated coverage: unit, property/fuzz, API
    integration, deterministic full-tournament simulation from signup to medals, varied seeding,
    edge cases and unusual progression paths. Fix every defect found before shipping.

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
- **User-built custom views:** a builder mode where any authorised user assembles a screen from
  live widgets by **drag and drop**, rearranges and **resizes** them, then **saves, clones, loads
  and deletes** named views. Views open in their own tab and stay live. The goal is minimal manual
  work: the app supplies components and data, the organiser composes the screens they want.
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
- Must remain easy for non-technical users: plain-language choices, clear status, **nested menus**
  with up/down arrow navigation, Enter to select and q/Esc to step back to the parent menu.
- One command must launch a **fully seeded, internally consistent tournament** for QA: pools,
  six-round bracket with promotion history, scored fights, kata scores, medals, standings and audit.

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
- **All data must be queryable** so views, charts, tables and rankings can sort, filter and search
  it: officials need deep query access; spectators and fighters get a safe public subset. A generic
  allow-listed query endpoint plus a data-explorer view supports this without exposing private data.
- **Belt/rank and safety presentation:**
  - belt selection is a dropdown showing a **colour block + name**, with a final "enter my own"
    option exposing a **colour picker + text field** (full JJJ spectrum included by default);
  - **rank/grade affects eligibility** — Kata vs Weapon Kata, Free Fighting vs Randori, age and
    safety/insurance limits; protection requirements per event (**head guard, mouth guard,
    gloves/MMA, instep/shin guards, clean pressed Gi, belt**, etc.) must be listed and public;
  - **competition corner colours are configurable** (red/white, red/blue, whatever the runner
    decides) to distinguish fighters on the mat.
- **Public tournament details view:** rules, requirements, event times, opening/closing ceremonies,
  non-competition schedule, location, venue, amenities, lunch and mandated rest between fights.
- **Main dashboard clock:** large clock, upcoming events with projected start/end times, and the
  ability to adapt projections when the event runs early or late — an authoritative timeline
  everyone can follow.
- **Locales & account settings:** support English, French, German, Spanish, Portuguese, Hindi,
  Punjabi, Arabic, Japanese, Korean and Mandarin. Account settings hold language, UI units
  (metric/imperial), name, email, mobile number, height, weight, rank/grade belt, DOB/age, club
  affiliation, light/dark/system appearance and other useful preferences.
- **Every configurable value needs presets plus custom input:** point values per technique/action,
  timers, rounds, categories, requirements and equipment are dropdown/list/toggle/slider choices
  with "add your own" always available. **Templates must exist at every level** — individual
  events, categories, disciplines, rule sets, scoring schemes and whole tournaments — so a runner
  can reuse and tweak a known-good setup reliably at any time.

- **Uploads are capability-scoped:** fighters upload a photo; officials a photo; clubs a logo and
  banner; runners/organisers a logo and banner. Formats PNG/JPG/WEBP/SVG with a size limit. The
  backend extracts a **main colour and highlight colour from the image** so club/brand colours
  can be auto-generated, with manual override.

- **Operational edge cases must be handled explicitly** (user QA):
  - uneven fight pools and categories with too few or no entrants;
  - **voluntary promotion** of a participant into a heavier/older category when they consent;
  - **participant caps**: reject applications when full, and grey out full/incompatible categories
    on sign-up so users do not waste time applying;
  - dynamic suggestion to switch to **round robin** when elimination is incompatible with the
    actual pool sizes/entrant count;
  - automatic **third-place rematch only if the two fighters have not already met**, otherwise
    award **two bronze medals**;
  - **dropouts, no-shows and injuries on the day**: forfeit/withdraw handling, pool rebalance and
    promotion adjustment so the event continues smoothly and fairly.

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
- [x] **Phase 1c — Setup pipeline & data platform**: five-step persisted wizard with validation,
      tournament/template creation API, stacked toasts with undo, generic allow-listed query API,
      data explorer, audit view, capacity limits + voluntary promotion, forfeit/no-show handling,
      auto-generated interactive bracket, settings page (11 locales, units, belt picker).
- [x] **Phase 1 — Setup completeness**: clubs/categories imported from templates, rule-set
      snapshots per discipline, manual result entry and medal table in one end-to-end flow.
- [x] **Phase 2 — Formats & rules engine**: persistent stages/slots, BYEs/seeding, tie-break
      chains, dual bronze and no-rematch rule; verified by property and simulation tests.
- [x] **Phase 3 — Live scoring**: authoritative server timers, SSE sync, holds, submission
      instant-win, penalties, undo, per-judge assignment and multi-mat control console.
- [x] **Phase 4 — Displays & scheduling**: scoreboard/brackets/tables/pools/public views,
      drift-aware schedule, on-deck, TTS/PA, broadcast overlay and custom view builder.
- [x] **Phase 5 — Kata & drills**: catalogue, numbered rounds with bunkai, panel judging,
      advancement, weapon kata and fair sparring rotation.
- [~] **Phase 6 — White-label depth & scale**: check-in/weigh-in, team points, CSV import/export,
      broadcast overlay and 11-locale shared labels delivered; payments/waivers, VAR, full content
      translation and real-device mobile verification remain.
- [x] **Phase 7 — Verification & hardening**: fuzz/property/simulation tests, 60 smoke checks,
      49 Playwright/axe checks, load/race runs, cleanup pass and QA sign-off.

> Order may change; keep this list honest. Each phase should be demoable.

---

## 12. Definition of done & testing
- Rules engine: deterministic golden tests + property tests (no duplicate pairings, promotion
  consistency, medal uniqueness, dual bronze cases, third-place rematch avoidance).
- Sync: authoritative server timers with SSE fan-out, request-idempotent mutations, undo and
  per-contest assignment. Offline replica/outbox remains a future enhancement.
- UI: Vitest + Playwright for critical flows; drag & drop; axe WCAG A/AA scans on every route.
- Platforms: Capacitor iOS/Android projects generated and synced; real-device install pending.
- Perf: 200-request concurrent load suite; 20+ mat simulation covered by deterministic tests.
- Docs: this live brief, `docs/CI.md` and the OpenAPI schema cover configuration and operations.

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
- **D-029 | 2026-11-19 | Allow-listed generic query API + data explorer**
  — tables/columns are explicitly allow-listed; search/sort/filter/pagination are parameterised;
  private tables hidden from public views. Status: accepted.
- **D-030 | 2026-11-19 | Final verification and polish pass is mandatory**
  — user requires thorough inspection, bug fixing, correctness/no-bad-states verification, and
  high-polish animation-rich UX before considering the app complete (§2 principle 13). Status:
  accepted; track as a late roadmap phase.
- **D-031 | 2026-11-19 | Production-grade automated verification suite**
  — fuzz/property tests for conversions and algorithms, deterministic full-tournament simulation,
  varied seed sets, edge cases, API integration and validation tests; defects fixed before ship.
  Status: accepted (§12).
- **D-032 | 2026-11-19 | Capacity, promotion and on-the-day disruption rules**
  — full categories greyed out/rejected; voluntary upward promotion; format fallback to round
  robin; rematch-or-dual-bronze; forfeit/no-show/injury handled dynamically. Status: accepted.
- **D-033 | 2026-11-19 | Rank/safety presentation and authoritative event timeline**
  — colour-block belt picker + custom belt, configurable corner colours, per-event protection
  requirements, public details view and a live dashboard clock with drift-adaptive projections.
  Status: accepted.
- **D-034 | 2026-11-19 | Eleven locales and per-user account settings**
  — EN/FR/DE/ES/PT/HI/PA/AR/JA/KO/ZH plus language, units, profile stats, belt, club affiliation
  and appearance preferences on the user's account. Status: accepted.
- **D-036 | 2026-11-20 | Drag-and-drop custom view builder**
  — widgets are composable live components; users arrange/resize them, save/clone/delete named
  layouts and open them as standalone live screens. Status: accepted.
- **D-037 | 2026-11-21 | Demo auth always resolves a real runner and browser sessions self-heal on 401**
  — demo mode must never 500 because a particular seed is absent; the web client retries once with
  a fresh demo session so a DB reset cannot leave every page permanently unauthorized. Status: accepted.
- **D-039 | 2026-11-21 | Android QA uses Android Studio's SDK and a Pixel 10 Pro AVD with `-gpu auto`**
  — the Linuxbrew emulator and `-gpu swiftshader` segfault on this host (exit 139), so the mobile
  helper now prefers `~/Android/Sdk`, defaults to `j3-pixel-10-pro`, and avoids the crashing
  SwiftShader path. Status: accepted.
- **D-038 | 2026-11-21 | Demo identity switching mints a real bearer session**
  — the old "act as" controls only changed display state, so QA never saw private/public RBAC
  boundaries. `POST /api/v1/auth/demo` now accepts an optional `userId` and issues a real
  server-side session for that account; both nav and dashboard switchers use it and re-apply the
  UI role guard. A demo-only, non-sensitive identity directory keeps the switcher available to
  low-privilege users without exposing private account fields. Status: accepted.
- **D-035 | 2026-11-19 | Capability-scoped uploads with auto-generated colours**
  — photos for fighters/officials, logo+banner for clubs and runners; PNG/JPG/WEBP/SVG, size
  limit; backend extracts main/highlight colours with manual override. Status: accepted.
- **D-040 | 2026-11-22 | Skipping a hands-on guide marks it complete and dismissed**
  — user QA: "when skipping it should set guide to completed and not come up again unless the
  user wants to". `setGuideSkipped` now completes every step and stores the skipped flag;
  auto-start honours it. Help, /guides → Reset and the start button remain explicit ways to
  see it again. Status: accepted.
- **D-041 | 2026-11-22 | Persistent app shell; J3 icon is always the public-board home**
  — nav stays visible on every normal route (including display routes) and mobile has a
  persistent bottom nav; in-app navigation stays client-side, while multi-screen board launches
  remain explicit "↗ New tab" actions. Fixed the latent Svelte issue where grouped nav items
  were never re-rendered after `onMount`, leaving the rail empty. Status: accepted.
- **D-042 | 2026-11-22 | Direct pointer manipulation in the view builder**
  — widgets reorder and resize by dragging handles with a floating ghost, drop indicator and
  eased snap animation; S/M/L, H1/H2/H3 and arrow keys remain accessible fallbacks. Status:
  accepted.
- **D-043 | 2026-11-22 | Full JJJ catalogue with server-resolved points**
  — the officials console searches 400 catalogue entries across 24 groups by English, romaji,
  Japanese script, style and movement tags. Clients send `techniqueId` plus an optional
  `pointLabel`; the server resolves points, kind, card and display label from the event rule
  snapshot, so a browser cannot award arbitrary values. Status: accepted.
- **D-044 | 2026-11-22 | Technical outcomes reuse the fight outcome vocabulary**
  — injury/forfeit/no-show/withdrawal call `win_red`/`win_blue`, while cheating/DQ call
  `dq_red`/`dq_blue`; the bout locks to `forfeit`, writes an audited technical score event and
  fans out over SSE. Assignment guards and runner override are unchanged. Status: accepted.
- **D-045 | 2026-11-22 | Okami 2027 is a mid-event QA fixture**
  — the default seed is "Okami Jiu Jitsu Challengers Tournament 2027" with live, paused,
  complete, drawn, forfeit and upcoming contests, partly judged kata, drifted schedule and audit
  events, so RBAC and public displays can be verified without running a fixture first. Status:
  accepted.
- **D-046 | 2026-11-22 | Branded, LAN-baked Android QA build**
  — `scripts/mobile-icons.py` regenerates launcher and splash assets from `web/static/brand/logo.png`;
  the APK bakes the detected venue URL and local demo CORS allows Capacitor's `http://localhost`
  and `capacitor://localhost` origins. Verified on the Pixel 10 Pro AVD with screenshot/logcat
  evidence under `docs/qa/`. Status: accepted.
- **D-047 | 2026-11-22 | Phones are participant/spectator screens; operator tools stay on larger displays**
  — phone-width (`<=700px`) route guarding and nav filtering restrict the app to public information,
  schedules, brackets, tables, pools, club, notifications, settings, guides and docs. Scoring,
  setup, control, audit, query, builder and simulator routes redirect to the appropriate phone home
  and are absent from the phone drawer/bottom nav. Tables reflow into labelled cards instead of
  scrolling sideways, long lists have explicit “show more” controls, and the public dashboard uses
  phone-only section tabs so a single screen never becomes an endless scroll. Tablet and desktop
  retain every role-appropriate operator tool. Status: accepted.
- **D-048 | 2026-11-22 | Native builds bundle a seeded SQLite fixture plus a read-only API snapshot**
  — `scripts/mobile-demo-data.sh` seeds a fresh mid-event Okami 2027 database and captures the
  public/participant GET responses used by phone screens. Both files ship in the APK (`j3-demo.db`
  and `j3-demo-data.json`). While the venue LAN is reachable the app uses live data; otherwise it
  falls back to the bundled snapshot, so the APK can be installed and demoed anywhere. The client
  still sends `techniqueId`/`pointLabel` for scoring and all privileged mutations require the live
  server. Mobile CI builds set `VITE_OFFLINE_DEMO=1`. Status: accepted.

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
- **2026-11-22** — Phone-scoped UX, standalone offline APK and final QA pass. Phones (`<=700px`) now show only public/participant/club/guardian screens: operator routes redirect to `/public` (or `/club` for club leaders) and disappear from the drawer/bottom nav, tables reflow into labelled cards with no horizontal scrolling, long lists use explicit “show more” controls, and the public dashboard has compact Live/Next/Table/Results tabs. Added `web/src/lib/responsive.ts`, phone no-overflow/nav/offline E2E coverage and fixed a parallel-run race in the officials console test by waiting for a selected contest. Every mobile build now regenerates the J3 launcher/splash icons from the supplied logo, seeds a fresh mid-event Okami 2027 database via `scripts/mobile-demo-data.sh`, captures a read-only API snapshot, bundles both as `j3-demo.db`/`j3-demo-data.json`, and sets `VITE_OFFLINE_DEMO=1`; mobile CI does the same. When the venue LAN is reachable the APK uses live data; with the venue server stopped, phone screens fall back to the bundled snapshot. Fixed the fixture script so its compiled binary (not just `go run`) is cleaned up, and fixed the stale docs API-explorer test to read the operation count from the generated contract. The final APK bakes `http://192.168.1.102:8787` and was verified on `j3-pixel-10-pro`: branded launcher icon, live Okami dashboard, offline fallback with server stopped, screenshot/logcat evidence under `docs/qa/`. Local gates green: `go vet`, `go test`, `go test -race`, `bun run check` (0 errors/warnings), `bun run test` (19), `bun run e2e` (110), `bun run build`, smoke 73/73, load 200/200, TUI 44/44, generated-type parity, workflow and shell syntax checks.
- **2026-11-22** — Officials console, catalogue and mid-event QA pass. Added a deterministic 400-entry/24-group JJJ technique catalogue (English, romaji, Japanese script, style and movement tags) and a searchable endpoint that merges configured points from rule sets and discipline rules. The officials console now lists only assigned contests for officials (all for runners), searches the whole catalogue, shows a configured-point dropdown, attributes yellow/red cards and technical wins (injury, forfeit, cheating, DQ, no-show, withdrawal, referee, other), controls the authoritative mat clock and confirms Red/Blue/Draw results with a draw hint. Scoring payloads resolve `techniqueId`/`pointLabel` server-side so browsers cannot spoof points, technique labels, card values or instant wins; technical outcomes reuse `win_*`/`dq_*` and lock to `forfeit` with audit + SSE fan-out. The default seed is now "Okami Jiu Jitsu Challengers Tournament 2027", deliberately mid-event (live, paused, complete, drawn, forfeit and upcoming contests, partly judged kata, schedule drift and audit events) for RBAC/public QA. Added 24 API/seed/catalogue tests or assertions, Vitest client tests, five Playwright console/public checks, smoke checks and updated official/scoring/API/CLI guides. The bundled Android APK now regenerates J3 launcher/splash icons from the supplied logo via `scripts/mobile-icons.py`, bakes the detected LAN venue URL and allows Capacitor origins in local demo CORS. Verified on `j3-pixel-10-pro` with seeded server, screenshot and logcat in `docs/qa/`; no CORS or console errors. Local gates green: `go vet`, `go test`, `go test -race`, `bun run check` (0 errors/warnings), `bun run test` (19), `bun run e2e` (108), `bun run build`, smoke 73/73, load 200/200, generated-type parity and workflows.
- **2026-11-22** — Guide, shell-navigation and view-builder refinement pass. Hands-on guide steps now open the wizard tab (`wizardStep`) or require a contest selection before spotlighting controls; added real `mat-contest-select` and `mat-scores` targets, fixed the old `mat-outcome` misuse, and added an official mat-clock step. Skipping a guide now marks every step complete, persists the skipped state and suppresses auto-start until the user restarts it; /guides shows "Guide skipped · 100%" with a Restart action. The app shell is now persistent: nav renders on every normal route including display boards, mobile has a bottom quick-nav plus the top-bar icon, the J3 icon always goes to `/public`, and display routes size to the content area instead of fighting the nav. Fixed a latent Svelte reactivity bug where `groupItems(group)` hid all nav links because the template never referenced `items` directly. Dashboard, control and display links now switch views in place; explicit "↗ New tab" buttons remain for TVs/projectors. View builder gained pointer drag-to-reorder and drag-to-resize (right, bottom and corner grips) with a ghost, drop indicator and eased FLIP snap, while keeping button/keyboard sizing, save/clone/load/delete and standalone `/view` layouts. Added `view-types` helpers/tests, nav persistence E2E, skip-completion coverage and pointer-gesture builder coverage. Verified: `go vet`, `go test`, `go test -race`, `bun run check`, `bun run test` (17), `bun run e2e` (103), `bun run build`, smoke 70/70, load 200/200 and docs sync all green. Pushed as `dffe0f3`; GitHub Actions CI run 45050614850 and Mobile builds run 45050614927 both completed green.
- **2026-11-22** — Fixed the CI-only production build failure that followed the docs site: SvelteKit's prerender
  crawler could not discover `/docs/[slug]` because the docs cards load `manifest.json` after hydration, so a clean
  GitHub runner failed with `prerender_unseen_routes`. `web/vite.config.ts` now seeds prerender entries from the docs
  manifest (alongside `'*'`), and a clean local `bun run build` prerenders every guide successfully.
- **2026-11-22** — Docs, guides and whole-app refinement pass. Added a real documentation site generated from
  `docs/*.md` plus OpenAPI-derived operations/search indexes: 13 rendered guides, live search, per-page anchors,
  sidebar/table-of-contents navigation, API explorer with “Try it”, `/apidocs` redirect, `bun run docs` and a
  nested **Docs & guides** TUI menu (including **Hands-on guides**). Added hands-on role guides that gate progress
  on real actions, success-only action tracking, `/guides` checklist with jump/reset, mobile bottom-sheet targeting,
  and settings “restart guide” using the active session role. Fixed the core docs navigation defect where SvelteKit
  reused `/docs/[slug]` after a link click and kept the previous guide on screen; sidebar/inline/pagination links now
  swap content and hash links scroll to rendered headings. Fixed the FAQ link to `/docs/api-guide` and the dead
  search-index intro anchors by rendering an `id` on the page `h1`. App UX: canonical labels/icons shared by nav,
  dashboard and display control; grouped navigation with active state and collapsed QA tools; role-filtered dashboard
  shortcuts/journey; one help button instead of duplicates; fullscreen display routes no longer show app chrome;
  `/overlay` guard fixed; public board made read-only for spectators; shared route allow-list drives nav and guard.
  Event-day fixes: mat console now selects the real win/loss/draw instead of hard-coding Red, confirm dialogs stay
  usable after first use, official state loads on contest change, bracket rebuild asks for confirmation and announces
  the real winner, check-in has search/status filters and no silent default weight, setup launch shows a success
  panel with next steps, PA copy no longer promises venue-wide sync, and pools/brackets/schedule follow the active
  event instead of `ev-demo`. Hardening from the same pass: CORS origin echo is allow-listed, CSV export uses the
  session bearer token, category `kind` defaults to `combined`, Okami seed name/email order fixed, Vite proxy regexes
  stop `/api` swallowing `/apidocs`. Added Playwright coverage for docs anchors/links, guide actions and the refined
  controls; 101 E2E checks, 12 Vitest checks, Go tests/race/vet, check, smoke, load, TUI harness, type parity and
  build green locally.
- **2026-11-22** — Full static verification pass; see §19. Checker A (SQL `?` count vs positional
  args over all Go DB calls), Checker B (unused vars/imports), full manual read of all non-test
  Go files and 52 frontend files, OpenAPI↔server route parity, TS import-target resolution.
  Three fixes: `/operations` CSV export (now authenticated download), `/club` invite (category
  `kind` defaults to `combined`), Okami seed name/email swap.
- **2026-11-21** — Android emulator QA is working: diagnosed repeated "terminated"/SIGSEGV crashes as
  the Linuxbrew emulator plus `-gpu swiftshader` graphics backend (backtrace in SwiftShader/GLES).
  `scripts/mobile.sh` now prefers Android Studio's `~/Android/Sdk`, targets the `j3-pixel-10-pro`
  AVD, and launches with `-gpu auto`; the Pixel 10 Pro boots in ~22s, the APK installs, and J3
  launches in the foreground. Screenshot/logcat saved under `docs/qa/`. GitHub Actions and local
  CI/CD green; TUI harness passed all 36 menu actions.
- **2026-11-21** — Demo QA identity switching now enforces real RBAC: `POST /api/v1/auth/demo`
  accepts an optional `userId`, both switchers exchange it for a real bearer session, the nav and
  route guard reapply the user role, and a demo-only identity directory keeps switching available
  without leaking private account fields. Added Go/Vitest/Playwright coverage for the private/public
  boundary and auto-login; smoke 70/70 + 70/70, E2E 83/83, load 200/200, race and build green.
- **2026-11-21** — Fixed the reported `401 /api/v1/demo/accounts` and cascading 500s. Root cause:
  `POST /api/v1/auth/demo` hard-coded `u-tony`, which only exists in the Okami seed, so on an empty
  venue DB or a full-seed DB (`bun run demo:full`, TUI "Run live executable") it returned 500 with a
  foreign-key error; the browser cached no session and every protected endpoint then returned 401.
  Fixed by resolving/bootstrapping a runner in demo mode and returning `expiresAt`/`userId`;
  tolerant session-expiry parsing; one-shot 401 recovery in the web API client (stale sessions after
  a DB reset now self-heal); panic-recovery middleware; DB-derived event IDs for display, medals,
  schedule, kata, event config and attendance; `rows.Err()` checks; duplicate medals now 409;
  `bun run api` enables demo mode. Added Go, Vitest and Playwright regression tests. Verified with
  `go test -race ./...`, smoke 70/70 + full smoke 70/70, E2E 80/80, load 200/200, check/build green.
- **2026-11-21** — TUI verification: added numeric menu shortcuts, action logging and a scripted
  harness that drives and proves all 36 menu actions; added release-artifact build, local CI/CD and
  GitHub Actions verification options; fixed stale labels, seeded-tournament command and deploy
  source path; `scripts/cicd.sh` now starts its own venue server. Local CI/CD and release scripts
  executed successfully.
- **2026-11-21** — GitHub Actions verified green on main: fixed missing Bun setup in two jobs,
  replaced broken `setup-android` action with the preinstalled SDK, corrected the iOS build to use
  `App.xcodeproj`, updated all actions to current majors, fixed a stale generated-types check, and
  changed the runner landing assertion after role-aware login. Added an interactive `/apidocs`
  explorer with live "try it" requests, synced OpenAPI contract + guide into the web app, and a TUI
  menu option to open it. Both CI and Mobile builds now pass on GitHub runners.
- **2026-11-21** — Hardening/stabilisation pass: live scoring authorization now derives identity
  from the verified session (spoofed headers rejected), RBAC defaults moved to migrations so they
  cannot be self-initialised, privileged mutations guarded, signup passwords + reset flow added,
  SQLite pragmas moved into the DSN for pooled connections, MFA rate limiting/constant-time
  comparison, absolute-time clock, atomic score sequences/medals/kata advancement, scoped bracket
  rebuilds, role-aware landing fix, uploads authenticated, public scoreboard made view-only,
  full-route axe scans and stronger CI (pinned Bun, timeouts, generated-type parity, seeded smoke).
- **2026-11-21** — Branding/mobile: integrated supplied J3 logo and banner artwork, applied the
  final palette (Midnight/Steel/Sky Blue, Energetic Orange, Light, Silver, Charcoal) with a bolder
  display font, club logos now render on fighter cards, tables, brackets and pools. Fixed Android
  build (JDK 21 + SDK 36), added LAN server mode for phone QA, Pixel emulator AVD support and a
  full Mobile menu in dev.sh; GitHub Actions now uses JDK 21 + Android SDK setup.
- **2026-11-21** — Final closure: real sessions/RBAC/PII hardening, full 72-path OpenAPI contract
  with examples, route-parity and 401-boundary tests, guardian consent for minors, four seeded
  tournament templates, public landing/signup pages, notification centre, account admin, guided
  walkthroughs with help tooltips and role-aware login destinations. Smoke 70/70 and E2E 58/58.
- **2026-11-21** — Security & production hardening: server-side sessions with bcrypt password
  verification and six-digit MFA, bearer authentication middleware (public routes allow-listed),
  role enforced from the database, per-contest assignment checks, real RBAC CRUD enforcement,
  private-route gating for query/audit/admin/export/attendance, guardian consent for under-16s,
  safe error mapping (404/400/401/403, generic 500), SVG uploads disabled, wildcard CORS removed,
  localhost bind by default, full 72-path OpenAPI spec with security schemes and route-parity test,
  API documentation with worked examples, four seeded tournament templates, landing/signup pages,
  guided walkthroughs with help tooltips and role-based login destinations.
- **2026-11-21** — Responsive/auth/RBAC push: icon-only collapsed nav with proper spacing, full-screen
  sliding mobile drawer, phone/tablet/landscape layouts across all 24 routes, login with password,
  Google/Apple demo providers, passkey/biometric flow and six-digit MFA, runner-configurable
  CRUD permissions with audit logging, and direct-URL role guards. Added mobile/tablet and auth E2E
  suites (56 checks total) and expanded smoke suite to 65 checks.
- **2026-11-21** — Rebuilt Okami day order (kata → juniors → teens → adults with weight bands) and
  competition formats: round robin then seeded single elimination for fighting/ground, full round
  robin for randori, individual kata judged youngest→oldest/lowest→highest with tie retries, dual
  bronze or third-place rematch configured per discipline. Added collapsible role-aware navigation
  and a dedicated `/official` console with bilingual (English/romaji) technique palette, per-corner
  scoring, penalties, kata panel scoring, submission instant-win override and confirm-before-commit.
- **2026-11-21** — Okami Jiu Jitsu Challenger is now the default demo seed: 40 fighters across 4
  clubs, Tony Jamieson as organiser, 2 mats with 3 judges + a scorekeeper each, 8 volunteer
  officials, 8 discipline rule sets with exact timings/points, 7 kit requirements, 3-section kata
  panels, 12-event schedule, 3 pools per discipline group, course attendance rules and full venue
  info. Added schema/config endpoints and `/event` review screen. 62 smoke + 50 E2E checks pass.
- **2026-11-20** — UI/UX refinement: scoreboard clock moved to its own full-width row so it can no
  longer be squashed; fighter cards reflow into a vertical stack on small screens; foul cards use
  compact Y/R badges; tables show a friendly retry banner instead of a raw proxy 500; global
  mobile breakpoints and resilient JSON parsing added; SQLite connection pool widened for file
  databases; custom views support vertical (row-span) resizing as well as horizontal spans.
- **2026-11-20** — Closed the documented outstanding list: check-in/weigh-in with out-of-class
  flagging, team rankings, CSV import/export, broadcast overlay, 11-locale shared label dictionary,
  full-route Playwright/axe coverage and final roadmap/definition-of-done update. Added entrants to
  the query allow-list. 60 smoke + 49 E2E checks passing.
- **2026-11-20** — Cleanup pass: live scoring events now feed the scoreboard; one RBAC/assignment
  guard covers all fight mutations; audit attribution derives from the request; bracket promotion
  returns the correct event; launch auto-runs the created event; kata table uses committed panel
  results; fixed stale mat selection and podium call; removed stale wizard stub, dead modes and
  duplicate helpers; unified request helpers and shared formatting; gated widget pollers per type;
  persisting wizard scoring points and category ranges; consolidated audit writes and simplified
  collection helpers. Verified with 56 smoke + 46 E2E checks.
- **2026-11-20** — General review pass: fixed football-bracket vertical spacing and de-cluttered
  promotion controls; added readiness checklist + one-tap auto-run, a Storybook-style `/ui`
  component gallery, a drag-and-drop `/builder` for custom live views with save/clone/delete/open,
  a tournament journey bar, and E2E coverage for all of it (46 tests).
- **2026-11-20** — Added a complete deterministic tournament seeder (`seed -full`, `bun run
  demo:full`, `bun run smoke:full`) with pools, six-round bracket, scored fights, kata, medals and
  audit; rebuilt the official bracket as a football-style diagram with SVG connectors, promotion
  history and a bronze decider; rebuilt the TUI with nested keyboard-navigable menus. Hardened
  standings with COALESCE to prevent 500s on empty data.
- **2026-11-20** — Added kata round advancement, environment-gated dev/staging/production deploy
  workflow, and an optional signed-IPA build step gated on Apple signing secrets. Smoke runner now
  builds a real binary to avoid stale servers; 52/52 checks passing.
- **2026-11-20** — Verification complete: added Playwright E2E + axe accessibility suites (26/26
  passing), load test (200 concurrent requests), and wired both into CI. Fixed defects found:
  missing document title, mat console initial state, wizard validation copy. Added guided medal
  ceremony sequence and round-robin sparring rotation generator. Smoke suite now 51/51.
- **2026-11-19** — Repository pushed to `git@github.com:RESTful-Otaku/j3.git` (`main`). Added
  GitHub Actions CI (branch naming, Go vet/test/race/build, Bun check/test/build, 50-check smoke),
  mobile artifact builds (Android APK, iOS simulator .app), release workflow, PR template,
  dependabot, CODEOWNERS and local pre-push hooks. Kata rounds/bunkai and template people-import
  added; toast unit tests added.
- **2026-11-19** — Final gaps: configurable tie-break chains, per-contest official assignment with
  scorer enforcement, template import to recreate events, animated medal ceremony view, and a smoke
  suite expanded to 49/49 passing checks.
- **2026-11-19** — Release hardening: hold/pin timers, RBAC enforcement on live scoring, dedicated
  `/mat` scorekeeper console, tie-break-ordered standings, rest warnings, browser TTS announcements,
  seeded kata catalogue and training sessions; smoke suite expanded to 46/46 passing checks.
- **2026-11-19** — Completion push: club-spread seeding, BYE padding and ordered tie-break chains
  in `rules`; direct-elimination bracket option; dual-bronze medal awards; `bun run smoke` automated
  end-to-end verification (41/41 passing across API and every UI route).
- **2026-11-19** — Release push: persisted bracket stages/slots with promotion and result history,
  schedule with drift projections + live clock, kata panel judging with high/low drop, plus their
  API tests. Live scoring, simulator fix and Capacitor projects were already committed.
- **2026-11-19** — Production push: added authoritative fight state/timers, score and penalty
  events, confirm/undo, SSE live hub and display wiring; generated and synced Capacitor Android
  and iOS projects (`bun run mobile:sync|android|ios`). APK build still needs an Android SDK and
  iOS build needs macOS/Xcode.
- **2026-11-19** — Refinement batch: added `internal/simulator` with deterministic 50-fighter
  end-to-end tournament simulation plus `/simulate` animated replay and `/details` public
  requirements/rules view; setup QoL (scoring point values with presets/custom, combined range
  rows, second-level timers, invite pagination + bulk delete, clone, config template library, drag
  handles); scoreboard player cards with last-three moves and foul cards, explicit event/category/
  round, paused-timer flash; J3 logo links to dashboard. All checks green.
- **2026-11-19** — Full document refresh: added §18 current state/outstanding register, updated
  header status and roadmap (Phase 1c done, Phase 7 in progress), and recorded D-035. Completed
  capability-scoped uploads with dominant/highlight colour extraction, belt picker with custom
  grades, 11-locale settings, capacity/disruption rules, format/third-place helpers and their
  automated tests. Fixed validation status codes and unvalidated unit-system defects found by tests.
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
1. Commit this officials-console/mid-event/mobile pass to `main`, push to `origin/main`, verify
   both GitHub Actions workflows complete green and inspect their job logs.
2. Continue only with the outstanding external-dependency items in §18 (payments/waivers,
   VAR, physical-device install, venue hardware tuning, offline outbox); all in-app roadmap slices
   are delivered and covered by local gates.
3. For mobile QA, run `bash scripts/mobile.sh build` for the standalone seeded APK, then
   `bash scripts/mobile.sh install`. For live scoring QA over the LAN, start **Run demo for phone QA**
   in `dev.sh` and rebuild with `J3_SERVER_URL="$(bash scripts/mobile.sh url)"`. Screenshot/logcat
   evidence is in `docs/qa/`.
4. On the day, use `bun run dev` → **Docs & guides** for the live guides. The nav stays visible;
   use **Display control** → "↗ New tab" only when a board needs its own TV/projector screen.

---

## 18. Current state & outstanding work (living summary)

### Delivered and verified
- Go venue server + pure-Go SQLite with migrations, RBAC role catalogue, seeded demo
  (**2 runners, 8 officials, 5 club leaders, 48 fighters, 1 spectator, 4 clubs, 6 categories,
  24 mid-event Okami 2027 fights** with live, paused, complete, drawn, forfeit and upcoming states).
- Sign-up with imperial→metric conversion, compatibility (strict + consent move-up), capacity
  checks, role request/approval, forfeit/no-show/injury handling and audit logging.
- Deterministic, property-tested algorithms: balanced pool generation with same-club separation,
  eligibility, format fallback and third-place planning; 120-seed full-tournament simulation.
- Five-step setup wizard, tournament/template creation, capability-scoped uploads with image
  colour extraction, query API, data explorer, audit view, settings (11 locales, units,
  colour-block belt picker + custom grade), stacked toasts with undo.
- Multi-screen routes: `/display`, `/public`, `/brackets`, `/tables`, `/pools`, `/control`,
  `/setup`, `/data`, `/audit`, `/settings`, `/details`, `/simulate`; live polling, PA toasts,
  celebration animations, Street Fighter theme, ultrawide CSS, TUI `dev.sh`, Capacitor config.
- Local demo/simulated runs auto-login as a runner; the dashboard and expanded nav both offer a
  real identity switcher that mints a server-side session for any seeded account, so QA sees the
  exact RBAC/public-private boundaries for that user.
- Android QA: Android Studio SDK + Pixel 10 Pro AVD (`-gpu auto`) boots reliably; the branded APK
  with regenerated J3 launcher/splash icons installs and launches in the foreground. It bundles a
  fresh mid-event Okami 2027 SQLite database and read-only API snapshot, so it works standalone;
  when a venue LAN URL is baked in, the same APK uses live data instead. Screenshot/logcat evidence
  is in `docs/qa/`. Local demo CORS allows Capacitor's `http://localhost` and `capacitor://localhost`
  origins.
- Phone scope: `<=700px` viewports show only public/participant/club/guardian screens. Operator
  routes redirect to `/public` (or `/club` for club leaders) and disappear from the drawer and
  bottom nav, tables reflow into labelled cards with no horizontal scroll, long lists expose
  “show more” controls, and the public dashboard uses compact section tabs. Tablet/desktop layouts
  retain all role-appropriate operator tools.
- Officials console: assigned-contest feed for officials and full-event override for runners; a
  400-entry/24-group JJJ catalogue searchable by English, romaji, Japanese script, style and
  movement tags; configured-point dropdowns resolved from rule snapshots; attributed yellow/red
  cards; technical wins for injury/forfeit/cheating/DQ/no-show/withdrawal/referee/other;
  authoritative clock start/pause/resume and final Red/Blue/Draw confirmation with draw hint.
  Every score payload is validated server-side; technical outcomes lock the bout and broadcast.
- Refinement batch delivered: editable point values per technique with presets + custom actions,
  combined age/height/weight/rank category ranges, second-level round durations, invite
  pagination + bulk delete, clone setup, per-config template library, drag handles separated from
  widgets, scoreboard player cards (last three moves, foul cards, belt, height/weight, club logo),
  explicit event title/category/round, distinct paused-timer flash, public details/requirements
  view, deterministic animated full-tournament simulator and spectator charts (top moves, club
  points, medals, standings).
- Documentation & guided QA: 13 rendered markdown guides, live docs search, OpenAPI-derived API
  explorer with 81 operations and “Try it”, `/apidocs` redirect, `bun run docs`, TUI **Docs &
  guides** menu, and hands-on role guides with real-action gating, progress checklist, jump and
  reset. Guide steps open the wizard tab or require a contest selection before highlighting its
  control; skipping marks the guide complete/dismissed until explicitly restarted. Every docs page
  and in-page anchor has a real destination.
- Navigation & event-day clarity: canonical labels/icons across nav/dashboard/display control,
  grouped nav with active state and collapsed QA tools, role-aware journey, persistent desktop and
  mobile shell (the J3 icon always returns to `/public`), in-place view switching with explicit
  “↗ New tab” launches for TVs/projectors, display routes sized to the content area, read-only
  public schedule, explicit mat result selector, reusable confirmation dialogs, searchable
  check-in, launch success panel, and active-event data instead of `ev-demo`.
- View builder: direct pointer drag-to-reorder and drag-to-resize (right/bottom/corner) with a
  floating ghost, drop indicator and eased snap animation; S/M/L, H1/H2/H3 and arrow keys remain
  accessible fallbacks. Saved layouts, standalone `/view` screens and live widgets are unchanged.
- Test suite: `go vet ./...`, `go test ./...`, `go test -race ./...`, `bun run check` (0 errors),
  `bun run test` (19), `bun run e2e` (110 browser journey/axe/docs/guide/nav/builder/official/offline
  checks), `bun run build`, smoke 73/73, load 200/200, generated-type parity and workflow checks
  all green at the time of this update.

### Outstanding — must finish before a real tournament
1. **Authoritative live scoring — delivered**: server-owned timers, score/penalty/submission
   events, hold/pin timers, RBAC enforcement, per-contest official assignment with scorer checks,
   confirm, undo, SSE fan-out and a dedicated `/mat` scorekeeper console.
2. **Production format engine — delivered**: server-side bracket stages/slots with format
   suggestion, club-spread seeding, BYE padding for odd fields, persisted results, winner
   promotion, per-round history, configurable tie-break chains (`?chain=...`) and dual-bronze
   medal awards.
3. **Scheduling authority — delivered**: persisted timeline with drift offsets, projected
   start/end times, mark-complete, mandated-rest warnings and a large live dashboard clock with
   "next up" on `/public`; PA announcements now also speak via browser TTS.
4. **Kata & training — delivered**: panel judging with 3–9 judges, high/low drop, authoritative
   totals, numbered rounds (shitei/tokuyi), bunkai scores and top-performer advancement; a seeded
   kata catalogue (including weapon kata), training sessions and a fair round-robin sparring
   rotation generator are available.
5. **Setup & ceremony — delivered**: saved tournament templates can recreate an event **or import
   its clubs/fighters**; a readiness checklist and one-tap auto-run generate pools, brackets and
   official assignments; `/ceremony` provides a guided animated podium sequence with PA callouts.
   A drag-and-drop `/builder` lets users compose, resize, save, clone, load and delete live widget
   dashboards, with `/ui` as a Storybook-style component gallery.
6. **Verification — delivered**: `bun run smoke:full`, `bun run e2e` (101 browser journey, docs,
   guide + axe accessibility tests), `bun run load` (200 concurrent requests), Go race tests and
   web Vitest suites all pass. CI runs every gate on PRs before merge.

### Latest refinement batch (user QA, 2026-11-19)
- **Setup QoL:** pagination, bulk create/delete, clone events, template presets, more optional
  fields, combined age/height/weight/rank ranges per category, round duration in seconds, drag
  handles separated from sliders/widgets inside movable cards.
- **Display clarity:** more colour/pop/spacing for distance viewing; scoreboard shows event title,
  category and round; player cards always show name, avatar, club logo/colours, belt, height,
  weight, last three moves and foul cards; paused timer flashes a distinct colour; fight ledger
  and pools show winners/losers, medals as coloured circles and penalty cards.
- **Officials:** multiple judges per contest (kata panels and mat assistants); one scorekeeper/
  timekeeper per mat who controls the board and advances events.
- **Brackets/graphs:** a proper football-style bracket diagram with connectors and drag/shuffle;
  real score tables, rankings and medal counts; charts for top moves, highest-scoring clubs,
  most medals by club/fighter.
- **Automation:** a deterministic full-tournament simulation run in real time end-to-end, plus a
  compact realistic 50-participant seed for rapid QA.
- **Navigation:** clicking the J3 logo returns to the signed-in account's main dashboard.
- **Animation:** significantly more wow-factor, transitions and dynamic visuals without losing
  clarity or role-appropriate access.

### Outstanding — polish & scale (external dependencies)
- Payments and signed waivers require gateway credentials; the flows are not wired yet.
- Full content translation beyond the shared label dictionary (11 locales supported for labels).
- Video assistant replay (VAR) needs a camera/streaming integration.
- Capacitor APK/IPA installation on physical devices needs a handset and Apple signing identity;
  the branded Android APK is already verified on the Pixel 10 Pro emulator.
- Projector/LED aspect-ratio tuning needs the actual venue hardware.
- Offline replica/outbox sync for **writes** is a future enhancement; the venue server remains
  authoritative with SSE. The bundled APK currently ships a read-only seeded snapshot for offline
  demos.

## 19. Second full static verification pass (2026-11-22)

This environment exposes no shell tools (no `go`/`bun` binaries on disk), so the live gates
(`go test`, `go test -race`, `go vet`, `bun run check`, `bun run test`, `bun run smoke`,
`bun run e2e`, `bun run build`) could not be re-executed here; GitHub Actions is green on
`main` per the latest change-log entry. Instead the whole tree was verified statically and by
deep manual review:

- **Go backend (all non-test files in `internal/{api,auth,config,db,rules,seed,simulator,units}`
  and `cmd/j3d/main.go`):**
  - Checker A — for every `ExecContext`/`QueryContext`/`QueryRowContext`/`Exec`/`Query` call
    (~200 sites), the SQL `?` count vs the number of positional args (the only flags were the
    four concatenated queries in `permissions.go`/`security.go` and two in `signup.go`,
    hand-verified: 3 `?`s + 3 args, 2 `?`s + 2 args, 18 `?`s + 18 args).
  - Checker B — unused variables/imports (none).
  - Full manual read of every file (e.g., SQL-string column counts, FK-constraint-driven
    updates like `categories.kind`, Scan shapes vs SELECT column counts).
- **OpenAPI spec ↔ server routes:** method-level parity complete — all 79 registered routes
  (including `GET /uploads/`) appear in `api/openapi.yaml` with the right methods
  (the parity test in `openapi_test.go` only checks `HandleFunc` routes).
- **Frontend (all 52 files under `web/src`):** full read of every route/lib/component page;
  API path parity vs server routes; TS import-target resolution (all resolve); no type/syntax
  issues found; `api-types.d.ts` contains every type referenced by `api.ts`.
- **SQLite/Go timestamp plumbing:** `strftime('%Y-%m-%dT%H:%M:%fZ', …)` produces "SS.SSS"
  (SQLite 3.54.1 per `lib/sqlite.go`; output format per `testdata/tcl/date.test`) which parses
  under the Go layouts in `security.go`/`reset.go`/`live.go` (`.000`/`.999` + literal `Z`).
- **Fixed (3):**
  1. `/operations` CSV export links (Fights/Entrants/Standings) returned 401 — they were plain
     `<a>` anchors to the authenticated `/api/v1/demo/export`. Now `api.exportCsv` fetches with
     the session bearer token and triggers a blob download
     (`web/src/lib/api.ts`, `web/src/routes/operations/+page.svelte`).
  2. `/club` "invite member" returned 500 — tournament creation inserts `categories` with an
     empty `kind`, violating `CHECK (kind IN ('age','weight','belt','gender','combined'))` which
     has no default. `createFromRequest` now defaults the kind to `combined`
     (`internal/api/tournament.go`).
  3. Cosmetic: Okami seed keepers and volunteers had `display_name` and `email` swapped in the
     `users` inserts (`internal/seed/okami.go`).

Findings summary: **no compile errors, no unused vars/imports, no SQL count mismatches, no
route/method or path-parity gaps, no logic bugs found** — the app is internally consistent, and
the three bugs listed above are the only defects found by this pass.

---

_End of living document. Agents: please update §§14–18 as you work._
