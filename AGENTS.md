# J3 (J Cubed) — Agents & Project Brief

> **Living document** — read this first, keep it current, and update §§11–14 when work lands.
> Keep it short: detail belongs in `docs/`, `docs/API.md` and git history.

| Field | Value |
|---|---|
| Status | Delivered: branded offline-capable venue platform, full auth/invite chain, officials console, mid-event Okami 2027 seed, app-wide Live/Sim replay, immediate real-time view updates, full per-discipline medals, containerised QA stack, phone-scoped UX, QA mobile builds, PolyForm NC licence, free hosting plan. Hosted CI/Release/Deploy green and `v0.7.5` published (2026-11-24); local gates green. Repo is public and hosted CI runs on free public runners. |
| Repo | `git@github.com:RESTful-Otaku/j3.git` (`main`) |
| Working dir | `/var/home/restful-otaku/Dev/projects/j3` |
| Created / updated | 2026-11-19 / 2026-11-24 |

---

## 0. How to use this file

1. Read §§1–3 before changing code.
2. Check §§11–12 before choosing an approach.
3. Update this file when you settle a decision, finish a roadmap item or complete a session.
4. Prefer small reversible decisions recorded here over silent assumptions.
5. Domain truth beats convenience: model configurable federation rules as config, never constants.

## 1. Mission

J3 is a **white-label platform for creating and running any Japanese Jiu Jitsu event** — one
configurable engine, any organiser/ruleset/discipline/format/display, online or with no venue
internet. Users: organisers, clubs, fighters, officials, spectators.

Core promise: *one configurable engine, every Jiu Jitsu event, fully branded, offline-capable.*

## 2. Non-negotiable principles

1. **Everything is data-driven** — rules, formats, branding, roles and layouts are versioned config.
2. **Offline-first** — local scoring reconciles deterministically; one SQLite file is a backup.
3. **Deterministic, auditable engine** — same config + events = same results; every write has actor,
   timestamp and undo/correction trail.
4. **Multi-mat/multi-display first-class**; every view has its own URL and stays live-synced.
5. **Accessible by default** — W/L/D use colour *and* glyph; phone → LED wall layouts.
6. **Local-first, cloud-optional** — venue LAN box authoritative; cloud publish later.
7. **One source of truth for types** — generated TS from OpenAPI/Go.
8. **Small vertical slices**, demoable early; QA feedback loops outrank breadth.
9. **Usable by non-technical people** — guided wizards, defaults, validation, progressive
   disclosure; raw config is an opt-in escape hatch only.
10. **Runnable full demo is a deliverable** — `bun run dev` boots the whole app with realistic seed.
11. **Automate the heavy lifting** — eligibility, pools, seeding, schedules and promotions from
   organiser constraints, with controlled randomisation and same-club spread.
12. **Final quality bar** — no bad states; polished, animated, dependable, thoroughly tested.

## 3. Config model

| Layer | Controls |
|---|---|
| Tenant / Organisation | branding, isolation, plan gates, locale |
| Theme | logo, banner, colours, fonts, display layouts |
| RuleSet | points, penalties, timers, rounds, golden score, DQ thresholds |
| CategoryScheme | age/weight/height/belt/gender ranges and policies |
| FormatPolicy | format by entrant count, seeding, BYEs, separation, bronze |
| Discipline | Fighting, Ne-Waza, Randori, Kata, drills, sparring… |
| RankingPolicy | medal/team points, circuits, star rating |
| RolePermissions | authority matrix for create/invite/score/override/publish |

Config is **immutable per event** (snapshot at start); new events get new snapshots.

## 4. Domain entities (short form)

`Tenant → Theme, Users, Clubs, Events`. `Event → Disciplines/Categories/Entrants`,
`RuleSet+FormatPolicy+CategoryScheme snapshot`, `Mats`, `Brackets/Stages/Slots/Fights`,
`FightEvents`, `KataPerformances/PanelScores`, `Schedule/Callouts`, `Rankings/Medals`,
`Announcements`. Supporting: invites, registrations, check-in, weigh-in, assignments, audit,
incidents, notifications, media, exports.

## 5. Domain rules (research inputs — configure, don’t hard-code)

- **Disciplines**: JJ Fighting System (strike/throw/ground, perfect = Ippon), Ne-Waza/Katamewaza,
  Tachi-Waza/Randori, continuous fighting, Kata (individual/pair/team/bunkai), training/sparring.
- **People**: age policy (calendar-year or event-day), weight (Gi/No-Gi, male/female/youth,
  metric stored canonical), height, belt/rank (full JJJ spectrum + custom ranks), optional gender,
  match-length by age/belt/Gi, weigh-in tolerance/re-weigh/out-of-class flags.
- **Formats**: 1 no contest, 2 best-of-3, 3–5 round robin, 4 single elim + bronze, 6 pools of 3,
  7+ double elim/repechage, true-double vs consolation-double explicit; BYEs, separation,
  two-match guarantee, configurable 3rd place.
- **Scoring**: Ippon/Waza-ari/Shido, BJJ points, Fighting System points, holds (10/15/20s),
  submission/tap instant win, best-of-N, golden score, kata panel 5.0–10.0 with high/low drop.
- **Penalties**: shido×3 = DQ, BJJ −1/−2/−3, yellow/red ladder, forfeit with recorded reason.
- **Tie-breakers**: ordered chain (wins → subs → points → head-to-head → conceded → fastest …).
- **Rankings**: medal points, win bonuses, club totals, star-rated events, seasons/circuits.
- **Scheduling**: rest intervals, multi-day/multi-mat, conflicts, drift, on-deck, callouts.

## 6. Feature map (delivered unless noted)

- **Setup**: five-step persisted wizard, templates at every level, clubs/fighters imported from
  templates, rule snapshots, validation with jump-to-section, capacity + voluntary move-up.
- **People/auth**: see §7. Fighter profile with photo, club logo, stats and belt.
- **Registration**: fighter/spectator self-signup, compatible-category offer, club affiliation or
  independent, guardian consent for minors, check-in/weigh-in, QR badges.
- **Formats**: pools, BYEs, seeding, club/country spread, round robin, single/double elim,
  repechage, dual bronze or third-place rematch only if not already met, drag/drop override with
  per-round history.
- **Live scoring**: authoritative server clock, hold/submission/penalty events, undo, assignment
  checks, result locking, SSE fan-out, officials console with 400-entry JJJ catalogue and
  server-resolved points. Material writes also emit cheap `state.changed` invalidations, so brackets,
  officials, scoreboard/mat/display, public, tables, pools, overlay, widgets and ceremonies refetch
  immediately in live mode while replay mode stays isolated.
- **Medals**: deterministic podiums per category × discipline × event type (fighting, Ne-Waza,
  Randori, continuous/sparring, kata/weapons), dual-bronze policy, manual overrides and one-shot
  recompute from completed results via `POST /api/v1/demo/medals/recompute`.
- **Kata/drills**: catalogue, numbered rounds, bunkai, panel judging, advancement, weapon kata,
  fair sparring rotation.
- **Displays**: `/display`, `/public`, `/brackets`, `/tables`, `/pools`, `/control`, `/setup`,
  `/data`, `/audit`, `/settings`, `/details`, `/simulate`, `/official`, `/mat`, `/ceremony`,
  `/view`, `/ui`; PA/TTS toasts, celebrations, Street Fighter default with contrast-safe bright fills,
  ultrawide CSS; resizable nav with themed active state; full-bleed host banners; explicit
  **Live event / Simulated tournament** nav toggle. Sim mode shows a global replay widget on every
  view (play/pause, restart, step, scrub, ¼×–8× speed, live/table/bracket/ceremony state); `/simulate`
  is the full event-by-event console. Fighter corners default to **Red vs White** and are fully
  configurable per event; internal scoring slots stay `red`/`blue` for deterministic results.
- **Builder**: drag/resize/save/clone/load/delete live widget layouts; standalone `/view` screens.
- **Data**: generic allow-listed query API, data explorer, audit view, CSV import/export, uploads
  with dominant/highlight colour extraction, 11-locale label dictionary.
- **Scheduling**: persisted timeline, drift, projections, mark-complete, rest warnings, big clock.
- **CLI/mobile**: TUI `dev.sh` with Docker QA submenu, Capacitor Android/iOS, branded APK with bundled
  seeded SQLite and read-only API snapshot; live LAN data when reachable.
- **Local QA stack**: Docker/Podman images, one-shot seed/sim containers, shared SQLite volume, nginx
  same-origin `/api` proxy (SSE-safe), `full`/`okami`/`clean`/`light`/`dev` modes and `j3d simulate`.
  Simulated fights use catalogue point values (typical totals 10–40) and instant-win submissions.
- **Not yet wired (external dependency)**: payments/waivers, VAR, full content translation,
  physical-device install, venue hardware tuning, offline write outbox.

## 7. Auth & authority chain

- Roles: **admin → runner → official/moderator → club leader → participant**; `admin` is stored as
  `tournament_runner` + `is_admin=1` but exposed as `admin` everywhere.
- Hosted owner bootstraps from `J3_OWNER_EMAIL`/`J3_OWNER_NAME` (defaults `j3@local` / `J3 Owner`).
  `J3_OWNER_PASSWORD` wins; otherwise a strong one-time password is logged and first login forces a
  change via `POST /api/v1/auth/password`.
- Public signup creates only **fighter/spectator**. All privileged accounts arrive through hashed
  token invites with `/invite/<token>` URLs; create/accept/revoke are audited. Club leaders are
  scoped to their own club; password reset never activates a pending invite.
- Demo: `GET /api/v1/demo/credentials` (password `demo`) backs login sheet, `/demo-accounts` and
  one-click identity switching that mints a real server session. Offline APK bundles the sheet so
  phones can switch identities without LAN; privileged writes still need the venue server.
- RBAC controls visibility and capability; per-contest assignment required to score (admin/runner
  override). Every material action is auditable. Details: `docs/AUTH_GUIDE.md`.

## 8. Architecture & platforms

- Go single binary + pure-Go SQLite (`modernc.org/sqlite`), REST + SSE, OpenAPI contract.
- SvelteKit + TypeScript + Bun (never npm); generated `web/src/lib/api-types.d.ts`.
- Venue server authoritative; devices use local replica/outbox for reads; client sends
  `techniqueId`/`pointLabel`, server resolves points. A single SQLite file is portable backup.
- Pure Go algorithm packages: eligibility, pools, seeding, brackets, promotion, scoring resolver,
  tie-breaks, scheduler — golden/property/simulation tested.
- Responsive 6" phone → LED wall; phones (`<=700px`) show only public/participant screens;
  operator tools stay on tablet/desktop/projector. Capacitor iOS/Android, PWA.

## 9. Roadmap

| Phase | Status | Scope |
|---|---|---|
| 0 Foundation | ✅ | Go API/SQLite, SvelteKit shell, generated types, seed, QA dashboard |
| 1a Sign-up/pools | ✅ | units, stats, compatibility, constrained pools |
| 1b Displays/CLI | ✅ | URL boards, ultrawide, PA, toasts, TUI, Capacitor |
| 1c Setup/data | ✅ | wizard, query API, audit, settings, capacity rules |
| 1 Setup completeness | ✅ | templates, rule snapshots, medals end-to-end |
| 2 Formats/rules | ✅ | stages/slots, BYEs, tie-breaks, dual bronze |
| 3 Live scoring | ✅ | server timers, SSE, holds, penalties, undo, multi-mat |
| 4 Displays/scheduling | ✅ | boards, drift, TTS, builder, overlays |
| 5 Kata/drills | ✅ | catalogue, panels, advancement, rotation |
| 6 White-label depth | ~ | check-in/weigh-in, team points, CSV, overlays, 11 locales; external gaps remain |
| 7 Verification | ✅ | fuzz/property/sim, 118 E2E, smoke, load, race, QA |

## 10. Definition of done

Deterministic golden/property tests; 401 boundary and route-parity tests; Playwright + axe on all
routes; smoke + load + race + build + generated-type parity green; docs synced; no known bad
states. Offline write outbox remains a future enhancement.

## 11. Decision log (compressed)

| ID | Decision |
|---|---|
| D-001–D-009 | Living AGENTS doc; versioned config snapshots; venue server authoritative; pure-Go SQLite; generic rulesets + full rank data; RBAC invite hierarchy; venue-only v1; git atomic commits; Phase 0 start. |
| D-010–D-019 | Zero-tech setup fundamental; Bun; runnable seeded demo; ~60 demo accounts; self-signup with stats; automated constrained pools; light/dark; approval for privileged role requests (superseded by D-052); strict + consent category compatibility. |
| D-020–D-035 | Dedicated URL boards; dashboard templates; toasts/undo; auto bracket; local uploads; audit view; allow-listed query API; polish + production verification mandatory; capacity/disruption rules; rank/safety presentation; 11 locales; pointer-drag builder; server-resolved points; capability-scoped uploads. |
| D-036–D-048 | Direct pointer manipulation in builder; full catalogue; technical outcomes reuse fight vocabulary; Okami 2027 mid-event seed; branded LAN-baked APK; phones are participant screens; APK bundles SQLite + read-only snapshot. |
| D-049–D-053 | Owner `j3@local` / `J3 Owner` overridable; env password or one-time generated + forced change; `admin` virtual role without role CHECK migration; privileged accounts only via token invites; demo credentials/identity switching work offline. |
| D-054–D-058 | CI mobile builds produce installable Android QA APK, iOS simulator app and signed real-device IPA when Apple secrets exist; emulator/simulator runtime smoke is required; QA builds self-seed and auto-login as admin; white-label title comes from event config; PolyForm Noncommercial 1.0.0 with attribution; Oracle Always Free A1 recommended for free prod/staging. |
| D-059–D-063 | One shared theme store; pure-white light, high-contrast dark gray and loud Street Fighter themes; organiser title/logo/banner/primary/secondary overrides applied from event tokens with computed readable ink; safe `.env.example` plus least-privilege workflows; `v*` tags publish versioned hosted APK/`.app` built with public `J3_SERVER_URL`, with signing secrets optional rather than release-blocking. |
| D-064–D-068 | GitHub releases publish only source, Linux server binary, versioned APK and iOS simulator `.app.zip`; Street Fighter is the default theme and dark ink keeps its neon gradient readable; nav is drag-resizable with a strong active border and banners use `contain`; demo replay is a shared global playhead with pause/restart/scrub and derived Mat/Table/Bracket/Ceremony state; standings/team rankings/kata reads are event-scoped and public-safe so tables never 500. |
| D-069–D-074 | Global Live/Simulated nav toggle with URL/localStorage persistence and a replay widget on every view; full event-by-event `/simulate` console; containerised Docker/Podman QA stack with static web proxy, one-shot seed/sim jobs and `j3d simulate`; `golang:1.27.1-alpine` pin to avoid the modernc sqlite compiler ICE; SSE `state.changed` invalidations for every operational view; deterministic full medal set per category/discipline/event type including kata, with manual overrides and recompute. |
| D-075 | Hosted CI is free-only: every job in every workflow is gated to public-repo visibility and `check:workflows` enforces it, so the private repo never requests a paid or 10× runner; local QA auto-detects a *working* Docker/Podman daemon and falls back to free host ports. |
| D-076 | Seeded events default to Red vs White corner labels/colours applied across mat, official, details, brackets, overlay and replay; stable `red`/`blue` scoring slots keep audit deterministic. |
| D-077 | Hosted Android smoke waits for adb `sys.boot_completed` and the package manager, retries APK installs, launches via `am start` with a monkey fallback, tolerates `pidof` misses under `pipefail`, and treats `uiautomator` dumps as diagnostics while gating on screenshot/logcat. |
| D-078 | Hosted release APK builds generate the standard Android debug keystore when no signing secrets are configured, so free public runners can still publish an installable signed APK. |
| D-079 | Android smoke uninstalls any prior signed build before install and scopes crash detection to the app PID/package, so QA/release key swaps and unrelated system warnings do not fail hosted deploys. |

Rationale and status for older decisions remain in git history; all listed decisions are accepted.

## 12. Open questions

1. True double vs consolation double as default? (configurable either way)
2. Target locales beyond EN/JA? (11 label locales shipped)
3. Preferred Go→TS generation? (OpenAPI codegen shipped)
4. Minimum LED/projector aspect ratios?
5. Existing brand assets/domain? (J3 logo/banner integrated)
6. Final first-migration table set? (core entities agreed)
7. Role requests vs invite-only? (invite-only chosen, D-052)
8. Exact compatibility chains per federation? (configurable ranges + consent)

## 13. Change log (recent)

- **2026-11-24 — `v0.7.5` published**: free-only public-repo gates across all workflows, resilient
  Android smoke (boot wait, explicit launch, prior-build uninstall, app-scoped crash gate), debug-key
  fallback for hosted release APKs and Red vs White corners. Curated assets published from
  `569934c`; no billing or branch protection enabled.
- **2026-11-24 — Resilient Android smoke**: emulator readiness now polls adb `sys.boot_completed`,
  the smoke script waits for the package manager, disables QA side-load verification, retries
  installs, launches via `am start` with a monkey fallback, tolerates `pidof` misses under `pipefail`,
  and keeps `uiautomator` dumps as diagnostics while gating on screenshot/logcat, so hosted failures
  are actionable. Hosted release builds generate the standard debug keystore when signing secrets are
  absent, keeping free public runners able to publish an installable APK. Smoke now uninstalls prior
  signed builds and scopes crash checks to the app PID/package, avoiding key-swap and system-log
  false positives.
- **2026-11-24 — Red/White fighter corners**: seeded demo and Okami events now default to Red vs
  White corner labels and colours. Event config drives mat, official, details, brackets, overlay and
  replay surfaces while stable `red`/`blue` audit slots remain deterministic.
- **2026-11-23 — Free-only CI + resilient local QA**: every hosted job now checks repository
  visibility and `check:workflows` enforces it, so a private repo skips all runners (including the
  10× macOS job) while the now-public repo runs free; `qa-docker.sh` picks a working Docker/Podman
  daemon, falls back from busy default ports and reports the real published ports from `status`/`url`.
  Docs and release policy unchanged.
- **2026-11-23 — Live/sim modes, real-time views and full medals**: explicit Live event / Simulated
  tournament nav toggle with a persistent replay widget on every view and a full event-by-event
  console; containerised Docker/Podman QA stack with seeded/simulated modes and `j3d simulate`;
  server `state.changed` SSE invalidations make brackets, officials, scoreboards, public, tables,
  pools, overlay, custom widgets and ceremonies update immediately; medals now cover every category,
  discipline and event type including kata, with deterministic recompute and manual overrides.
- **2026-11-23 — Demo replay + display polish**: Street Fighter is the default theme with dark ink on bright fills; nav is drag-resizable with a clear themed active indicator and clipped account switcher; banners show the whole image; global demo replay supports play/pause/restart/step/scrub and follows every screen; `/public`, `/tables` and `/ceremony` render replay state; tables/kata/team rankings reads hardened against 500s.
- **2026-11-22 — Themes + versioned mobile releases**: pure-white light, high-contrast dark gray and Street Fighter themes; shared theme persistence; event token overrides for title/logo/banner/main/secondary colours with contrast-safe text; safe `.env.example` and environment/secrets guide; workflows fixed for Android SDK PATH/boot and iOS simulator parsing; tagged releases publish hosted APK/`.app` with backend URL baked in.
- **2026-11-22 — CI/CD + licensing**: mobile workflows build installable QA apps, run emulator/simulator smoke, release production server/web/API/docs/mobile artifacts; QA builds self-seed and auto-login as admin; event-config title replaces `J Cubed`; PolyForm NC licence and free Oracle hosting plan documented; TUI has one-step build+deploy actions and hosted-APK build.
- **2026-11-22 — Auth rollout**: real admin role, hosted owner bootstrap, authority matrix, token
  invites, demo credential sheet with real/offline identity switching, invite acceptance page, docs
  and full gate/CI/APK verification.
- **2026-11-22 — Phone/APK/QA**: phone-scoped routes, branded launcher assets, seeded offline APK,
  live/offline emulator evidence, E2E/smoke hardening.
- **2026-11-22 — Auth hardening**: live status/role checks, forced password change enforcement,
  admin-only matrix edits, audited invites, usable wizard invites, immediate-write SQLite lock mode.
- **2026-11-22 — Docs compaction**: AGENTS and `docs/*.md` reduced to concise living references.
- Older entries: see git history (`git log --oneline`).

## 14. Current state

**Delivered**: complete app described above; local gates green; APK verified on `j3-pixel-10-pro`
with live and offline evidence in `docs/qa/`. `v0.7.5` (`569934c`) publishes source archives,
`j3-server-linux-amd64`, `j3-android-0.7.5.apk` and `j3-ios-0.7.5.app.zip`; QA apps are built and
smoke-tested when hosted CI runs (public repo) but not published. See `docs/HOSTING.md` for free
Oracle Always Free hosting.

**Hosted CI status**: `RESTful-Otaku/j3` is public, so GitHub-hosted runners are free and the
visibility-gated workflows run normally. Final `v0.7.5` verification on `569934c` is green: CI
`45403247450`, Mobile `45403247408`, Release `45403287849` and Deploy `45403287727`. Earlier
`v0.7.5` attempts were superseded before publication (missing runner debug keystore) or fixed after
a system-log false positive; the old `v0.7.4` runs (`45099245577`, `45099245595`, `45099247549`,
`45099247622`) failed before jobs started when the repo was private. Every hosted job still checks
public visibility and `check:workflows` enforces it, so no paid or 10× runner can be requested if
visibility changes again. No billing action is needed.

**External-dependency remainder**: payment/waiver gateways, VAR/video integration, full content
translation, physical-device install, venue projector/LED tuning, offline write outbox.

**Next actions**: none for the release; `v0.7.5` is published and all hosted gates are green. Do not
enable billing or branch protection. Local fast QA is `bash scripts/qa-docker.sh full` (working
Docker/Podman daemon, free auto-selected ports; see `docs/DOCKER.md`); native QA is `bun run dev`. For
phones rebuild a standalone self-seeding offline APK with `bash scripts/mobile.sh build`, or a
LAN/HTTPS client with `J3_SERVER_URL="$(bash scripts/mobile.sh url)" bash scripts/mobile.sh build`.

---

_End of compact living document. Agents: keep it this way._
