<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { initials } from '../lib/format';
  import {
    api,
    allActions,
    type Action,
    type DemoAccount,
    type DemoIdentity,
    type DemoCounts,
    type DemoFight,
    type Role,
    type Eligibility,
    type SignupResult,
    type RoleRequest,
    type PoolView,
    type SignupRequest,
  } from '../lib/api';
  import { landingForRole, session, switchDemoUser, uiRole } from '../lib/auth';
  import { currentRole, HOME_ROUTE, itemsForRole, type Role as NavRole } from '../lib/nav';
  import { eventTitle } from '../lib/event-title';
  import { cornerConfig } from '../lib/event-brand';
  import { cycleTheme, themeChoice } from '../lib/theme';

  let ready = false;
  let error = '';
  let counts: DemoCounts | null = null;
  let accounts: DemoIdentity[] = [];
  let fights: DemoFight[] = [];
  let actorId = '';
  let actor: DemoAccount | null = null;
  let capabilities: Action[] = [];
  let activeTab = 'overview';
  let roleFilter: Role | 'all' = 'all';
  let search = '';
  let selectedFightId = '';
  let navRole: NavRole = 'runner';

  function syncNavRole() {
    const active = session();
    navRole = active ? uiRole(active.role) : currentRole();
  }
  $: shortcuts = itemsForRole(navRole).filter((item) => ['/setup', '/data', '/audit'].includes(item.href));
  $: journey = itemsForRole(navRole).filter((item) => ['/public', '/display', '/brackets', '/control'].includes(item.href));

  let signupForm: SignupRequest = {
    name: '',
    email: '',
    password: '',
    phone: '',
    dateOfBirth: '1998-01-01',
    unitSystem: 'metric',
    heightCm: 175,
    weightKg: 74,
    heightFt: 5,
    heightIn: 9,
    weightLb: 165,
    beltRankId: 'br-blue',
    gender: 'male',
    accountType: 'fighter',
    clubId: 'club-riverside',
    categoryIds: [],
    consentMoveUp: false,
  };
  let eligible: Eligibility[] = [];
  let signupResult: SignupResult | null = null;
  let roleRequests: RoleRequest[] = [];
  let pools: PoolView[] = [];
  let poolParams = {
    categoryId: '',
    targetPoolSize: 4,
    maxPools: 8,
    sameClubSeparation: true,
    seed: 42,
  };
  let generating = false;
  let eventDetails: Record<string, any> = {};

  let beltRanks: { id: string; label: string; colour?: string }[] = [];
  let clubs: { id: string; label: string }[] = [{ id: '', label: 'Independent / no club' }];

  const tabs = [
    { id: 'overview', label: 'Overview' },
    { id: 'accounts', label: 'Accounts & roles' },
    { id: 'scoreboard', label: 'Live scoreboard' },
    { id: 'fights', label: 'Fights' },
    { id: 'signup', label: 'Sign up' },
    { id: 'pools', label: 'Pools' },
  ];
  const roles: (Role | 'all')[] = [
    'all',
    'tournament_runner',
    'official',
    'club_leader',
    'fighter',
    'spectator',
  ];
  const roleLabel: Record<string, string> = {
    all: 'Everyone',
    tournament_runner: 'Runners',
    official: 'Officials',
    club_leader: 'Club leaders',
    fighter: 'Fighters',
    spectator: 'Spectators',
  };

  function asAccount(identity: DemoIdentity | null | undefined): DemoAccount | null {
    if (!identity) return null;
    return {
      id: identity.id, name: identity.name, email: '',
      role: identity.role as DemoAccount['role'], roleLabel: identity.roleLabel,
      clubName: identity.clubName, source: 'independent', status: 'active',
      actions: identity.actions,
    };
  }

  async function optional<T>(promise: Promise<T>): Promise<T | null> {
    try { return await promise; } catch { return null; }
  }

  onMount(async () => {
    syncNavRole();
    window.addEventListener('j3-session-changed', syncNavRole);
    try {
      const [demo, identitiesResponse, fs, eventInfo, beltInfo, clubInfo] = await Promise.all([
        api.demo(),
        api.identities(),
        api.fights(),
        api.event(),
        api.belts(),
        optional(api.query({ table: 'clubs', limit: 100 })),
      ]);
      try { eventDetails = JSON.parse(eventInfo.detailsJson || '{}'); } catch { eventDetails = {}; }
      beltRanks = beltInfo.belts;
      clubs = [...clubs, ...(clubInfo?.rows ?? []).map((row) => ({ id: String(row.id), label: String(row.name) }))];
      counts = demo.counts;
      accounts = identitiesResponse.identities;
      actorId = identitiesResponse.current?.id ?? session()?.userId ?? '';
      fights = fs.fights;
      const [view, poolResponse, requestResponse] = await Promise.all([
        optional(api.view()), optional(api.pools()), optional(api.roleRequests()),
      ]);
      actor = view?.actor ?? asAccount(identitiesResponse.current);
      capabilities = view?.actions ?? identitiesResponse.current?.actions ?? [];
      const live = fights.find((f) => f.state === 'live') ?? fights[0];
      selectedFightId = live?.id ?? '';
      pools = poolResponse?.pools ?? [];
      roleRequests = requestResponse?.requests ?? [];
      ready = true;
    } catch (e) {
      error = e instanceof Error ? e.message : String(e);
    }
  });

  async function switchActor(id: string) {
    if (id === actorId) return;
    try {
      const next = await switchDemoUser(id);
      actorId = next.userId ?? id;
      await goto(landingForRole(next.role));
    } catch (e) {
      error = e instanceof Error ? e.message : String(e);
    }
  }

  async function previewSignup() {
    try {
      const preview = await api.signupPreview(signupForm);
      eligible = preview.eligible;
    } catch (e) {
      error = e instanceof Error ? e.message : String(e);
    }
  }

  async function submitSignup() {
    try {
      signupResult = await api.signup(signupForm);
      const identitiesResponse = await api.identities();
      accounts = identitiesResponse.identities;
      roleRequests = (await optional(api.roleRequests()))?.requests ?? [];
    } catch (e) {
      error = e instanceof Error ? e.message : String(e);
    }
  }

  async function generatePools() {
    generating = true;
    try {
      const result = await api.generatePools({ eventId: 'ev-demo', ...poolParams });
      pools = result.pools;
    } catch (e) {
      error = e instanceof Error ? e.message : String(e);
    } finally {
      generating = false;
    }
  }

  async function approve(id: string) {
    try {
      await api.approveRoleRequest(id);
      const [reqs, identitiesResponse] = await Promise.all([api.roleRequests(), api.identities()]);
      roleRequests = reqs.requests;
      accounts = identitiesResponse.identities;
    } catch (e) {
      error = e instanceof Error ? e.message : String(e);
    }
  }

  function maxClubLoad(pool: PoolView) {
    const counts: Record<string, number> = {};
    for (const member of pool.members) {
      const key = member.club || 'Independent';
      counts[key] = (counts[key] ?? 0) + 1;
    }
    return Math.max(0, ...Object.values(counts));
  }

  function roleCount(role: Role | 'all') {
    return role === 'all'
      ? accounts.length
      : accounts.filter((a) => a.role === role).length;
  }

  function payload(event: { payload: unknown }): Record<string, unknown> {
    return (event.payload ?? {}) as Record<string, unknown>;
  }

  function score(fight: DemoFight | undefined, corner: 'red' | 'blue') {
    if (!fight) return 0;
    return fight.events
      .filter((e) => e.kind === 'score' && payload(e).corner === corner)
      .reduce((sum, e) => sum + Number(payload(e).points ?? 0), 0);
  }

  $: filteredAccounts = accounts.filter((a) => {
    const matchesRole = roleFilter === 'all' || a.role === roleFilter;
    const matchesSearch = `${a.name} ${a.roleLabel} ${a.clubName}`
      .toLowerCase()
      .includes(search.toLowerCase());
    return matchesRole && matchesSearch;
  });
  $: selectedFight = fights.find((f) => f.id === selectedFightId);
  $: liveCount = fights.filter((f) => f.state === 'live').length;
</script>

<svelte:head>
  <title>{$eventTitle} · J3 QA dashboard</title>
  <meta name="description" content="White-label Japanese Jiu Jitsu tournament platform by J3" />
</svelte:head>

<div class="shell">
  <header class="topbar">
    <div class="brand">
      <a class="brand-mark" href={HOME_ROUTE} aria-label="J3 public dashboard">J3</a>
      <div>
        <div class="brand-name">{$eventTitle}</div>
        <div class="brand-sub">Tournament creator · local QA demo</div>
      </div>
    </div>
    <div style="display:flex; align-items:center; gap:10px; flex-wrap:wrap;">
      {#each shortcuts as item (item.href)}
        <a class="btn" class:primary={item.href === '/setup'} href={item.href}>{item.icon} {item.label}</a>
      {/each}
      <span class="pill"><span class="dot" class:off={!!error}></span>{error ? 'Server offline' : 'Venue server running'}</span>
      <button class="tab" on:click={cycleTheme} aria-label="Cycle theme">
        {$themeChoice === 'dark' ? '☀ Light' : $themeChoice === 'light' ? '⚡ Street' : '🌙 Dark gray'}
      </button>
    </div>
  </header>

  <div class="journey">
    <span class="muted small">Tournament journey</span>
    <a href={HOME_ROUTE}>🏠 Public board</a>
    {#each journey as item (item.href)}
      <span>›</span><a href={item.href}>{item.icon} {item.label}</a>
    {/each}
  </div>

  <div class="notice">
    <strong>Demo mode.</strong> The venue server is seeded with realistic data. Use
    <em>Accounts &amp; roles</em> to instantly act as any user and check that permissions and
    views are right. Run <code>bun run reset</code> to start fresh.
  </div>

  {#if error}
    <div class="card" style="border-color: var(--loss);">
      <h2>Could not reach the venue server</h2>
      <p class="muted small">Start it with <code>bun run dev</code> from the project root.</p>
      <p class="small" style="margin-top:8px; color:var(--loss);">{error}</p>
    </div>
  {:else if !ready}
    <div class="placeholder">Loading demo data…</div>
  {:else}
    <section class="hero">
      <div class="card accent">
        <div class="eyebrow">Ready to play</div>
        <h1>Your tournament,<br />ready to run.</h1>
        <p class="muted" style="margin-top:12px; max-width:52ch;">
          J3 is a white-label Japanese Jiu Jitsu tournament creator. Everything here is live
          data from the local venue server — switch accounts, inspect fights, and explore the
          scoreboard exactly as officials and spectators will see it.
        </p>
        <div class="action-row">
          <button class="btn primary" on:click={() => (activeTab = 'accounts')}>Switch account</button>
          <button class="btn" on:click={() => (activeTab = 'scoreboard')}>Open live scoreboard</button>
          <a class="btn" href="/control">🖥️ Display control</a>
          <a class="btn" href="/display">🥊 Live scoreboard</a>
          <a class="btn" href="/api/v1/demo" target="_blank" rel="noreferrer">Inspect API ↗</a>
        </div>
      </div>
      <div class="card">
        <div class="eyebrow">Acting as</div>
        {#if actor}
          <div style="display:flex; align-items:center; gap:14px;">
            <div class="avatar">{initials(actor.name)}</div>
            <div>
              <h2>{actor.name}</h2>
              <div class="small muted">{actor.email}{actor.clubName ? ` · ${actor.clubName}` : ''}</div>
            </div>
          </div>
          <div style="margin-top:14px;"><span class="badge {actor.role}">{actor.roleLabel}</span></div>
          <p class="small muted" style="margin-top:12px;">
            {capabilities.length} capabilities active · source: {actor.source}
          </p>
        {/if}
      </div>
    </section>

    <div class="grid stats">
      {#each Object.entries(counts ?? {}) as [key, value]}
        <div class="stat">
          <div class="value">{value}</div>
          <div class="label">{key.replace(/([A-Z])/g, ' $1').toLowerCase()}</div>
        </div>
      {/each}
    </div>

    <div class="section-title">
      <h2>Explore the demo</h2>
      <div class="tabs">
        {#each tabs as tab}
          <button
            class="tab"
            class:active={activeTab === tab.id}
            on:click={() => (activeTab = tab.id)}>{tab.label}</button
          >
        {/each}
      </div>
    </div>

    {#if activeTab === 'overview'}
      <div class="grid" style="grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));">
        <div class="card">
          <div class="eyebrow">01 · Setup</div>
          <h3>Configure without tech skills</h3>
          <p class="muted small" style="margin-top:8px;">
            A plain-language wizard will guide you through branding, clubs, categories, rules
            and formats. No files, no jargon — just questions and a live preview.
          </p>
        </div>
        <div class="card">
          <div class="eyebrow">02 · Run</div>
          <h3>Multiple mats, one truth</h3>
          <p class="muted small" style="margin-top:8px;">
            {liveCount} live contest{liveCount === 1 ? '' : 's'} in this snapshot. Officials score on phones; every
            confirmed event is committed and every display reads the same data.
          </p>
        </div>
        <div class="card">
          <div class="eyebrow">03 · Trust</div>
          <h3>Roles with boundaries</h3>
          <p class="muted small" style="margin-top:8px;">
            Runners, officials, club leaders, fighters and spectators each see and can change
            only what their authority allows. Every material action is audited.
          </p>
        </div>
        <div class="card">
          <div class="eyebrow">04 · Every screen</div>
          <h3>Dedicated live views</h3>
          <p class="muted small" style="margin-top:8px;">
            Switch views here without losing the navigation. Use Display control when you want a
            board on its own TV tab: public dashboard, scoreboard, brackets, tables/medals/kata
            and fight pools.
          </p>
          <div class="action-row">
            <a class="btn" href="/public">Public</a>
            <a class="btn" href="/brackets">Brackets</a>
            <a class="btn" href="/tables">Tables</a>
            <a class="btn" href="/pools">Pools</a>
            <a class="btn" href="/simulate">Simulate</a>
          </div>
        </div>
      </div>
    {:else if activeTab === 'accounts'}
      <div class="card">
        <div class="section-title" style="margin-top:0;">
          <div>
            <h2>Account switcher</h2>
            <p class="muted small">One click to see the app exactly as any demo user sees it.</p>
          </div>
          <input
            class="tab"
            style="cursor:text; color:var(--brand-ink);"
            placeholder="Search name, email or club…"
            bind:value={search}
          />
        </div>
        <div class="tabs" style="margin-bottom:14px;">
          {#each roles as role}
            <button
              class="tab"
              class:active={roleFilter === role}
              on:click={() => (roleFilter = role)}>{roleLabel[role]} ({roleCount(role)})</button
            >
          {/each}
        </div>
        <div class="accounts">
          {#each filteredAccounts as account (account.id)}
            <button
              class="account"
              class:active={actorId === account.id}
              title={`${account.name} (${account.roleLabel})`}
              aria-label={`${account.name} (${account.roleLabel})`}
              on:click={() => switchActor(account.id)}
            >
              <span class="avatar">{initials(account.name)}</span>
              <span style="min-width:0;">
                <span class="name" style="display:block; overflow:hidden; text-overflow:ellipsis;">{account.name}</span>
                <span class="meta" style="display:block; overflow:hidden; text-overflow:ellipsis;">
                  {account.clubName || account.roleLabel}
                </span>
              </span>
            </button>
          {:else}
            <p class="muted small">No accounts match this filter.</p>
          {/each}
        </div>
        {#if actor}
          <div style="margin-top:22px; border-top:1px solid var(--brand-border); padding-top:18px;">
            <div class="eyebrow">Permission check · {actor.roleLabel}</div>
            <div class="capabilities">
              {#each allActions as action}
                <span class="cap" class:on={capabilities.includes(action)} class:off={!capabilities.includes(action)}>
                  {capabilities.includes(action) ? '✓' : '✕'} {action.replaceAll('_', ' ')}
                </span>
              {/each}
            </div>
          </div>
        {/if}
      </div>
    {:else if activeTab === 'scoreboard'}
      <div class="card">
        <div class="section-title" style="margin-top:0;">
          <h2>Live fight display</h2>
          <select class="tab" style="color:var(--brand-ink);" bind:value={selectedFightId}>
            {#each fights as fight}
              <option value={fight.id}>{fight.mat} · {fight.category} · {fight.id}</option>
            {/each}
          </select>
        </div>
        {#if selectedFight}
          <div class="scoreboard">
            <div class="corner red">
              <div class="portrait">{initials(selectedFight.red.name || '—')}</div>
              <div>
                <div class="small muted">{$cornerConfig.red.label} · {selectedFight.red.belt || '—'}</div>
                <div style="font-weight:700; font-size:1.15rem;">{selectedFight.red.name || 'TBD'}</div>
                <div class="small" style="color:{selectedFight.red.clubColour};">{selectedFight.red.club || 'Independent'}</div>
              </div>
              <div class="score">{score(selectedFight, 'red')}</div>
            </div>
            <div style="text-align:center;">
              <div class="small muted">{selectedFight.mat} · {selectedFight.round}</div>
              <div class="clock">{new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</div>
              <span class="badge {selectedFight.state === 'live' ? 'fighter' : 'spectator'}">{selectedFight.state}</span>
            </div>
            <div class="corner blue">
              <div class="portrait">{initials(selectedFight.blue.name || '—')}</div>
              <div>
                <div class="small muted">{$cornerConfig.blue.label} · {selectedFight.blue.belt || '—'}</div>
                <div style="font-weight:700; font-size:1.15rem;">{selectedFight.blue.name || 'TBD'}</div>
                <div class="small" style="color:{selectedFight.blue.clubColour};">{selectedFight.blue.club || 'Independent'}</div>
              </div>
              <div class="score">{score(selectedFight, 'blue')}</div>
            </div>
          </div>
          <div style="margin-top:18px;">
            <h3>Tracked moves &amp; penalties</h3>
            <div class="capabilities">
              {#each selectedFight.events as event (event.seq)}
                <span class="cap" class:on={event.kind === 'score'}>
                  #{event.seq} · {event.kind}
                  {#if event.kind === 'score'}+{payload(event).points ?? 0} {payload(event).technique ?? ''}{/if}
                  {#if event.kind === 'penalty'}{String(payload(event).card ?? '')} · {String(payload(event).corner ?? '')}{/if}
                </span>
              {:else}
                <span class="muted small">No events recorded yet.</span>
              {/each}
            </div>
          </div>
        {/if}
      </div>
    {:else if activeTab === 'fights'}
      <div class="card">
        <h2>Fight ledger</h2>
        <p class="muted small">Every row is read from the committed database.</p>
        <div style="overflow-x:auto; margin-top:14px;">
          <table style="width:100%; border-collapse:collapse; font-size:.88rem;">
            <thead>
              <tr class="muted small" style="text-align:left; border-bottom:1px solid var(--brand-border);">
                <th style="padding:8px;">Fight</th><th>Mat</th><th>Category</th><th>{$cornerConfig.red.label}</th><th>{$cornerConfig.blue.label}</th><th>State</th><th>Result</th>
              </tr>
            </thead>
            <tbody>
              {#each fights as fight}
                <tr style="border-bottom:1px solid var(--brand-border);">
                  <td style="padding:8px;"><code>{fight.id}</code></td>
                  <td>{fight.mat}</td>
                  <td>{fight.category}</td>
                  <td>{fight.red.name}</td>
                  <td>{fight.blue.name}</td>
                  <td><span class="badge {fight.state === 'live' ? 'fighter' : 'spectator'}">{fight.state}</span></td>
                  <td>
                    {#if fight.outcome === 'win_red'}<span style="color:var(--win); font-weight:700;">W / L</span>
                    {:else if fight.outcome === 'draw'}<span style="color:var(--draw); font-weight:700;">D / D</span>
                    {:else}—{/if}
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      </div>
    {:else if activeTab === 'signup'}
      <div class="grid" style="grid-template-columns: 1.2fr 1fr;">
        <div class="card">
          <div class="eyebrow">Self-service sign-up</div>
          <h2>Create your account</h2>
          <p class="muted small" style="margin:6px 0 16px;">
            Metric or imperial — the backend stores canonical metric values. Compatible categories
            are computed from your stats; heavier divisions appear only with your consent.
          </p>
          <div class="grid" style="grid-template-columns: 1fr 1fr;">
            <label class="small">Full name<input class="tab" style="display:block; width:100%; margin-top:5px; color:var(--brand-ink); cursor:text;" bind:value={signupForm.name} /></label>
            <label class="small">Email<input class="tab" style="display:block; width:100%; margin-top:5px; color:var(--brand-ink); cursor:text;" bind:value={signupForm.email} /></label>
            <label class="small">Password (8+ characters)<input class="tab" type="password" style="display:block; width:100%; margin-top:5px; color:var(--brand-ink); cursor:text;" bind:value={signupForm.password} /></label>
            <label class="small">Phone<input class="tab" style="display:block; width:100%; margin-top:5px; color:var(--brand-ink); cursor:text;" bind:value={signupForm.phone} /></label>
            <label class="small">Date of birth<input class="tab" type="date" style="display:block; width:100%; margin-top:5px; color:var(--brand-ink);" bind:value={signupForm.dateOfBirth} /></label>
          </div>
          <div class="tabs" style="margin:16px 0 8px;">
            <button class="tab" class:active={signupForm.unitSystem === 'metric'} on:click={() => (signupForm.unitSystem = 'metric')}>Metric</button>
            <button class="tab" class:active={signupForm.unitSystem === 'imperial'} on:click={() => (signupForm.unitSystem = 'imperial')}>Imperial</button>
          </div>
          <div class="grid" style="grid-template-columns: repeat(3, 1fr);">
            {#if signupForm.unitSystem === 'metric'}
              <label class="small">Height (cm)<input class="tab" type="number" style="display:block; width:100%; margin-top:5px; color:var(--brand-ink);" bind:value={signupForm.heightCm} /></label>
              <label class="small">Weight (kg)<input class="tab" type="number" step="0.1" style="display:block; width:100%; margin-top:5px; color:var(--brand-ink);" bind:value={signupForm.weightKg} /></label>
            {:else}
              <label class="small">Height (ft)<input class="tab" type="number" style="display:block; width:100%; margin-top:5px; color:var(--brand-ink);" bind:value={signupForm.heightFt} /></label>
              <label class="small">Height (in)<input class="tab" type="number" style="display:block; width:100%; margin-top:5px; color:var(--brand-ink);" bind:value={signupForm.heightIn} /></label>
              <label class="small">Weight (lb)<input class="tab" type="number" style="display:block; width:100%; margin-top:5px; color:var(--brand-ink);" bind:value={signupForm.weightLb} /></label>
            {/if}
          </div>
          <div class="grid" style="grid-template-columns: 1fr 1fr; margin-top:14px;">
            <label class="small">Rank / grade
              <select class="tab" style="display:block; width:100%; margin-top:5px; color:var(--brand-ink);" bind:value={signupForm.beltRankId}>
                {#each beltRanks as rank}<option value={rank.id}>{rank.label}</option>{/each}
              </select>
            </label>
            <label class="small">Gender
              <select class="tab" style="display:block; width:100%; margin-top:5px; color:var(--brand-ink);" bind:value={signupForm.gender}>
                <option value="male">Male</option><option value="female">Female</option><option value="open">Open / prefer not to say</option>
              </select>
            </label>
            <label class="small">Account type
              <select class="tab" style="display:block; width:100%; margin-top:5px; color:var(--brand-ink);" bind:value={signupForm.accountType}>
                <option value="fighter">Fighter</option>
                <option value="official">Official / moderator</option>
                <option value="tournament_runner">Runner / organiser</option>
                <option value="club_leader">Club leader</option>
                <option value="spectator">Spectator</option>
              </select>
            </label>
            <label class="small">Club
              <select class="tab" style="display:block; width:100%; margin-top:5px; color:var(--brand-ink);" bind:value={signupForm.clubId}>
                {#each clubs as club}<option value={club.id}>{club.label}</option>{/each}
              </select>
            </label>
          </div>
          <label class="small" style="display:flex; align-items:center; gap:8px; margin-top:14px;">
            <input type="checkbox" bind:checked={signupForm.consentMoveUp} /> Allow a nearby heavier/older division if my exact one has no entrants
          </label>
          <div class="action-row">
            <button class="btn" on:click={previewSignup}>Check compatible categories</button>
            <button class="btn primary" on:click={submitSignup}>Create account</button>
          </div>
        </div>
        <div class="card">
          <div class="eyebrow">Computed from your stats</div>
          <h2>Compatible categories</h2>
          {#if eligible.length > 0}
            <div class="capabilities" style="flex-direction:column; align-items:stretch; margin-top:14px;">
              {#each eligible as item (item.categoryId)}
                <label class="cap" class:on={item.strict} style="display:flex; gap:9px; align-items:flex-start; cursor:pointer;">
                  <input type="checkbox" value={item.categoryId} bind:group={signupForm.categoryIds} />
                  <span><strong>{item.name}</strong><br /><span class="muted">{item.reason}{item.moveUp ? ' · needs consent' : ''}</span></span>
                </label>
              {/each}
            </div>
          {:else}
            <div class="placeholder" style="margin-top:14px;">Press “Check compatible categories” to see your eligible divisions.</div>
          {/if}
          {#if signupResult}
            <div style="margin-top:20px; border-top:1px solid var(--brand-border); padding-top:16px;">
              <h3>Account created</h3>
              <p class="small muted" style="margin-top:6px;">
                Status: <strong>{signupResult.status}</strong> · role: {signupResult.role}<br />
                Stored metric: {signupResult.heightCm.toFixed(1)} cm · {signupResult.weightKg.toFixed(1)} kg
              </p>
              {#if signupResult.roleRequestId}<p class="small" style="color:var(--draw); margin-top:8px;">Privileged role pending approval.</p>{/if}
            </div>
          {/if}
          <div style="margin-top:20px; border-top:1px solid var(--brand-border); padding-top:16px;">
            <h3>Role requests</h3>
            {#each roleRequests as request (request.ID)}
              <div style="display:flex; justify-content:space-between; align-items:center; gap:10px; padding:8px 0; border-bottom:1px solid var(--brand-border);">
                <span class="small">{request.Name} → <strong>{request.Role}</strong> <span class="badge spectator">{request.Status}</span></span>
                {#if request.Status === 'pending'}<button class="btn" on:click={() => approve(request.ID)}>Approve</button>{/if}
              </div>
            {:else}
              <p class="muted small">No role requests pending.</p>
            {/each}
          </div>
          <div style="margin-top:20px; border-top:1px solid var(--brand-border); padding-top:16px;">
            <h3>Event requirements</h3>
            <p class="muted small">These apply to every competitor at this event.</p>
            <div class="capabilities" style="margin-top:10px;">
              {#each eventDetails.protection ?? [] as requirement}<span class="cap on">✓ {requirement}</span>{/each}
            </div>
          </div>
        </div>
      </div>
    {:else if activeTab === 'pools'}
      <div class="card">
        <div class="section-title" style="margin-top:0;">
          <div>
            <h2>Automated pools</h2>
            <p class="muted small">Balanced by seed, mixed by club, generated from organiser constraints.</p>
          </div>
          <div style="display:flex; gap:8px; flex-wrap:wrap; align-items:center;">
            <select class="tab" style="color:var(--brand-ink);" bind:value={poolParams.categoryId}>
              <option value="">All categories</option>
              <option value="cat-adult-76">Adult -76 kg</option>
              <option value="cat-adult-82">Adult -82 kg</option>
              <option value="cat-adult-90">Adult +82 kg</option>
              <option value="cat-women-58">Women -58 kg</option>
              <option value="cat-youth-60">Youth -60 kg</option>
              <option value="cat-kata">Kata Open</option>
            </select>
            <label class="small">Target size<input class="tab" type="number" min="2" max="10" style="width:70px; color:var(--brand-ink);" bind:value={poolParams.targetPoolSize} /></label>
            <label class="small">Max pools<input class="tab" type="number" min="1" max="32" style="width:70px; color:var(--brand-ink);" bind:value={poolParams.maxPools} /></label>
            <label class="small">Seed<input class="tab" type="number" style="width:80px; color:var(--brand-ink);" bind:value={poolParams.seed} /></label>
            <label class="small" style="display:flex; align-items:center; gap:6px;"><input type="checkbox" bind:checked={poolParams.sameClubSeparation} /> Separate clubs</label>
            <button class="btn primary" disabled={generating} on:click={generatePools}>{generating ? 'Generating…' : 'Generate pools'}</button>
          </div>
        </div>
        {#if pools.length === 0}
          <div class="placeholder">No pools yet — choose constraints and press “Generate pools”.</div>
        {:else}
          <div class="grid" style="grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));">
            {#each pools as pool}
              <div class="stat" style="padding:18px;">
                <div style="display:flex; justify-content:space-between; align-items:center;">
                  <h3>{pool.name}</h3>
                  <span class="pill">{pool.members.length} fighters · max club {maxClubLoad(pool)}</span>
                </div>
                <div style="margin-top:12px; display:grid; gap:8px;">
                  {#each pool.members as member}
                    <div style="display:flex; align-items:center; gap:9px;">
                      <span class="avatar" style="width:30px; height:30px; border-radius:9px; background:{member.clubColour};">{initials(member.name)}</span>
                      <span style="flex:1; min-width:0;"><strong class="small">{member.name}</strong><br /><span class="muted small" style="font-size:.72rem;">{member.club || 'Independent'} · {member.belt}</span></span>
                      <span class="small muted">#{member.seed}</span>
                    </div>
                  {/each}
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    {/if}
  {/if}
</div>
