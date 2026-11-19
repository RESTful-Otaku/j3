<script lang="ts">
  import { onMount } from 'svelte';
  import {
    api,
    allActions,
    type Action,
    type DemoAccount,
    type DemoCounts,
    type DemoFight,
    type Role,
  } from '../lib/api';

  let ready = false;
  let error = '';
  let counts: DemoCounts | null = null;
  let accounts: DemoAccount[] = [];
  let fights: DemoFight[] = [];
  let actorId = 'u-runner';
  let actor: DemoAccount | null = null;
  let capabilities: Action[] = [];
  let activeTab = 'overview';
  let roleFilter: Role | 'all' = 'all';
  let search = '';
  let selectedFightId = '';
  let theme: 'dark' | 'light' = 'dark';

  const tabs = [
    { id: 'overview', label: 'Overview' },
    { id: 'accounts', label: 'Accounts & roles' },
    { id: 'scoreboard', label: 'Live scoreboard' },
    { id: 'fights', label: 'Fights' },
    { id: 'setup', label: 'Setup preview' },
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

  onMount(async () => {
    try {
      const [demo, accts, fs, view] = await Promise.all([
        api.demo(),
        api.accounts(),
        api.fights(),
        api.view(actorId),
      ]);
      counts = demo.counts;
      accounts = accts.accounts;
      fights = fs.fights;
      actor = view.actor;
      capabilities = view.actions;
      const live = fights.find((f) => f.state === 'live') ?? fights[0];
      selectedFightId = live?.id ?? '';
      ready = true;
    } catch (e) {
      error = e instanceof Error ? e.message : String(e);
    }
  });

  async function switchActor(id: string) {
    actorId = id;
    try {
      const view = await api.view(id);
      actor = view.actor;
      capabilities = view.actions;
    } catch (e) {
      error = e instanceof Error ? e.message : String(e);
    }
  }

  function initials(name: string) {
    return name
      .replace(/\s*\(.*\)$/, '')
      .split(' ')
      .slice(0, 2)
      .map((part) => part[0] ?? '')
      .join('')
      .toUpperCase();
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
    const matchesSearch = `${a.name} ${a.email} ${a.clubName}`
      .toLowerCase()
      .includes(search.toLowerCase());
    return matchesRole && matchesSearch;
  });
  $: selectedFight = fights.find((f) => f.id === selectedFightId);
  $: liveCount = fights.filter((f) => f.state === 'live').length;
</script>

<svelte:head>
  <title>J3 · Local QA Demo</title>
  <meta name="description" content="J3 local full-stack demo and QA dashboard" />
</svelte:head>

<div class="shell" data-theme={theme}>
  <header class="topbar">
    <div class="brand">
      <div class="brand-mark">J3</div>
      <div>
        <div class="brand-name">J Cubed</div>
        <div class="brand-sub">Tournament creator · local QA demo</div>
      </div>
    </div>
    <div style="display:flex; align-items:center; gap:10px; flex-wrap:wrap;">
      <span class="pill"><span class="dot" class:off={!!error}></span>{error ? 'Server offline' : 'Venue server running'}</span>
      <button class="tab" on:click={() => (theme = theme === 'dark' ? 'light' : 'dark')}>
        {theme === 'dark' ? '☀ Light' : '☾ Dark'}
      </button>
    </div>
  </header>

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
              on:click={() => switchActor(account.id)}
            >
              <span class="avatar">{initials(account.name)}</span>
              <span style="min-width:0;">
                <span class="name" style="display:block; overflow:hidden; text-overflow:ellipsis;">{account.name}</span>
                <span class="meta" style="display:block; overflow:hidden; text-overflow:ellipsis;">
                  {account.clubName || account.source}
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
                <div class="small muted">RED · {selectedFight.red.belt || '—'}</div>
                <div style="font-weight:700; font-size:1.15rem;">{selectedFight.red.name || 'TBD'}</div>
                <div class="small" style="color:{selectedFight.red.clubColour};">{selectedFight.red.club || 'Independent'}</div>
              </div>
              <div class="score">{score(selectedFight, 'red')}</div>
            </div>
            <div style="text-align:center;">
              <div class="small muted">{selectedFight.mat} · {selectedFight.round}</div>
              <div class="clock">03:00</div>
              <span class="badge {selectedFight.state === 'live' ? 'fighter' : 'spectator'}">{selectedFight.state}</span>
            </div>
            <div class="corner blue">
              <div class="portrait">{initials(selectedFight.blue.name || '—')}</div>
              <div>
                <div class="small muted">BLUE · {selectedFight.blue.belt || '—'}</div>
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
                <th style="padding:8px;">Fight</th><th>Mat</th><th>Category</th><th>Red</th><th>Blue</th><th>State</th><th>Result</th>
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
    {:else if activeTab === 'setup'}
      <div class="card">
        <div class="eyebrow">Coming in Phase 1</div>
        <h2>Guided event setup</h2>
        <p class="muted small" style="margin-top:8px;">
          The next slice makes this real: create an event, invite clubs, add fighters, choose
          categories and rules, generate a bracket and award medals — entirely through a
          plain-language wizard with a live preview.
        </p>
        <div class="grid" style="grid-template-columns:repeat(auto-fit,minmax(190px,1fr)); margin-top:18px;">
          {#each ['Brand your event', 'Invite clubs', 'Add fighters', 'Set categories', 'Pick rules', 'Build bracket'] as step, i}
            <div class="stat"><div class="value" style="font-size:1rem;">{i + 1}. {step}</div></div>
          {/each}
        </div>
      </div>
    {/if}
  {/if}
</div>
