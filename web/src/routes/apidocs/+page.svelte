<script lang="ts">
  import { onMount } from 'svelte';

  type Operation = {
    group: string;
    method: 'GET' | 'POST';
    path: string;
    title: string;
    description: string;
    auth: boolean;
    body?: string;
  };

  const operations: Operation[] = [
    { group: 'Authentication', method: 'POST', path: '/api/v1/auth/login', title: 'Begin sign-in', description: 'Password verification issues a six-digit MFA challenge. Google/Apple/passkey flows are demo-mode shims.', auth: false, body: '{"email":"tony@okami.j3","password":"demo","provider":"password"}' },
    { group: 'Authentication', method: 'POST', path: '/api/v1/auth/verify', title: 'Verify MFA', description: 'Exchange the six-digit code for a server-side session token.', auth: false, body: '{"email":"tony@okami.j3","code":"123456"}' },
    { group: 'Authentication', method: 'POST', path: '/api/v1/auth/demo', title: 'Demo runner session', description: 'QA-only shortcut, active when J3_DEMO=1.', auth: false },
    { group: 'Authentication', method: 'POST', path: '/api/v1/auth/reset/request', title: 'Request password reset', description: 'Always returns success to avoid account enumeration.', auth: false, body: '{"email":"tony@okami.j3"}' },
    { group: 'Authentication', method: 'POST', path: '/api/v1/auth/reset', title: 'Complete password reset', description: 'Consume a reset token and set a new password.', auth: false, body: '{"token":"reset-token","password":"newpassword123"}' },

    { group: 'Public boards', method: 'GET', path: '/api/v1/demo', title: 'Seeded data counts', description: 'Counts of clubs, fighters, fights, categories and more.', auth: false },
    { group: 'Public boards', method: 'GET', path: '/api/v1/demo/fights', title: 'Fight ledger', description: 'Fights with fighter, club, technique and penalty detail.', auth: false },
    { group: 'Public boards', method: 'GET', path: '/api/v1/demo/standings', title: 'Standings', description: 'Ordered points table with optional tie-break chain.', auth: false },
    { group: 'Public boards', method: 'GET', path: '/api/v1/demo/brackets', title: 'Bracket stages', description: 'Persisted stages and slots with results and history.', auth: false },
    { group: 'Public boards', method: 'GET', path: '/api/v1/demo/event-config', title: 'Event configuration', description: 'Costs, venue, kit, rules, formats and staffing.', auth: false },
    { group: 'Public boards', method: 'GET', path: '/api/v1/demo/stats', title: 'Spectator statistics', description: 'Top moves, club medals and fighter standings.', auth: false },
    { group: 'Public boards', method: 'GET', path: '/api/v1/events/stream', title: 'Live event stream', description: 'Server-sent events for scores, clocks and announcements.', auth: false },

    { group: 'Live scoring', method: 'GET', path: '/api/v1/fights/fight-01/state', title: 'Fight state', description: 'Authoritative state, derived clock, scores and committed events.', auth: true },
    { group: 'Live scoring', method: 'POST', path: '/api/v1/fights/fight-01/start', title: 'Start fight', description: 'Starts the absolute-time clock.', auth: true, body: '{"seconds":180,"round":1}' },
    { group: 'Live scoring', method: 'POST', path: '/api/v1/fights/fight-01/event', title: 'Log a technique', description: 'Records a score, penalty or hold against an assigned contest.', auth: true, body: '{"corner":"red","kind":"score","points":5,"technique":"Taoshi (Takedown)"}' },
    { group: 'Live scoring', method: 'POST', path: '/api/v1/fights/fight-01/pause', title: 'Pause timer', description: 'Referee stoppage; remaining time is stored.', auth: true },
    { group: 'Live scoring', method: 'POST', path: '/api/v1/fights/fight-01/undo', title: 'Undo last event', description: 'Removes the most recent committed scoring event.', auth: true },
    { group: 'Live scoring', method: 'POST', path: '/api/v1/fights/fight-01/submit', title: 'Confirm result', description: 'Locks the result and publishes it to every display.', auth: true, body: '{"outcome":"win_red"}' },

    { group: 'Tournament operations', method: 'POST', path: '/api/v1/signup/preview', title: 'Eligibility preview', description: 'Converts imperial units to metric and returns compatible divisions.', auth: false, body: '{"dateOfBirth":"1998-01-01","unitSystem":"metric","heightCm":175,"weightKg":74,"gender":"male"}' },
    { group: 'Tournament operations', method: 'POST', path: '/api/v1/signup', title: 'Register participant', description: 'Creates an account and entrant with guardian consent for under-16s.', auth: false, body: '{"name":"New Fighter","email":"new@example.com","password":"password123","dateOfBirth":"1998-01-01","unitSystem":"metric","heightCm":175,"weightKg":74,"accountType":"fighter"}' },
    { group: 'Tournament operations', method: 'POST', path: '/api/v1/demo/auto-run', title: 'Auto-generate tournament', description: 'Creates pools, brackets and official assignments from entrant data.', auth: true },
    { group: 'Tournament operations', method: 'POST', path: '/api/v1/demo/pools/generate', title: 'Generate pools', description: 'Balanced pools with same-club separation.', auth: true, body: '{"eventId":"ev-okami","targetPoolSize":4,"sameClubSeparation":true,"seed":42}' },
    { group: 'Tournament operations', method: 'POST', path: '/api/v1/demo/brackets/build', title: 'Build bracket', description: 'Generates stages and slots from current entrants.', auth: true, body: '{"eventId":"ev-okami","format":"Double elimination","seed":42,"thirdPlace":true}' },
    { group: 'Tournament operations', method: 'POST', path: '/api/v1/demo/medals/award', title: 'Award medals', description: 'Gold, silver and optional dual bronze.', auth: true, body: '{"gold":"f01","silver":"f02","bronze":["f03","f04"]}' },

    { group: 'Administration', method: 'GET', path: '/api/v1/demo/admin/users', title: 'List accounts', description: 'Runner-only account administration data.', auth: true },
    { group: 'Administration', method: 'GET', path: '/api/v1/demo/permissions', title: 'List RBAC permissions', description: 'Per-role create/read/update/delete matrix.', auth: true },
    { group: 'Administration', method: 'POST', path: '/api/v1/demo/permissions', title: 'Update permission', description: 'Changes one role/resource capability.', auth: true, body: '{"role":"official","resource":"fights","flag":"delete","value":false}' },
    { group: 'Administration', method: 'POST', path: '/api/v1/demo/announcements', title: 'Send announcement', description: 'Broadcasts via push/email/SMS channels.', auth: true, body: '{"title":"Fighters to Mat 1","channels":"push,email","audience":"fighters"}' },
    { group: 'Administration', method: 'GET', path: '/api/v1/query?table=fighters&limit=10', title: 'Query dataset', description: 'Allow-listed search, sort and pagination.', auth: true },
  ];

  let selected: Operation = operations[0];
  let group = 'All';
  let search = '';
  let bodyText = selected.body ?? '';
  let response = '';
  let status = 0;
  let busy = false;
  let spec = '';

  const groups = ['All', ...Array.from(new Set(operations.map((op) => op.group)))];
  $: visible = operations.filter((op) =>
    (group === 'All' || op.group === group) &&
    `${op.method} ${op.path} ${op.title}`.toLowerCase().includes(search.toLowerCase()));

  onMount(() => {
    fetch('/openapi.yaml').then((res) => res.text()).then((text) => (spec = text)).catch(() => {});
  });

  function choose(op: Operation) {
    selected = op;
    bodyText = op.body ?? '';
    response = '';
    status = 0;
  }

  function token() {
    const raw = localStorage.getItem('j3-session');
    if (!raw) return '';
    try { return JSON.parse(raw).token ?? ''; } catch { return ''; }
  }

  async function tryIt() {
    busy = true;
    response = '';
    try {
      const headers: Record<string, string> = { Accept: 'application/json' };
      if (selected.body) headers['Content-Type'] = 'application/json';
      const bearer = token();
      if (bearer) headers.Authorization = `Bearer ${bearer}`;
      const res = await fetch(selected.path, {
        method: selected.method,
        headers,
        body: selected.method === 'POST' && bodyText ? bodyText : undefined,
      });
      status = res.status;
      const text = await res.text();
      try { response = JSON.stringify(JSON.parse(text), null, 2); } catch { response = text; }
    } catch (error) {
      status = 0;
      response = error instanceof Error ? error.message : String(error);
    } finally {
      busy = false;
    }
  }

  async function copyCurl() {
    const bearer = token();
    const auth = bearer && selected.auth ? ` -H 'Authorization: Bearer ${bearer}'` : '';
    const data = selected.body ? ` -d '${bodyText.replaceAll("'", "\\'")}'` : '';
    const command = `curl -X ${selected.method}${auth}${data} 'http://127.0.0.1:8787${selected.path}'`;
    await navigator.clipboard.writeText(command);
  }
</script>

<svelte:head><title>J3 · API Explorer</title></svelte:head>

<div class="shell">
  <header class="topbar">
    <div class="brand"><a class="brand-mark" href="/" aria-label="Dashboard">J3</a><div><div class="brand-name">API explorer</div><div class="brand-sub">Developer & admin console · OpenAPI 3.1</div></div></div>
    <div style="display:flex; gap:8px; flex-wrap:wrap;">
      <a class="btn" href="/openapi.yaml" target="_blank" rel="noreferrer">📄 Raw OpenAPI ↗</a>
      <a class="btn" href="/docs/API.md" target="_blank" rel="noreferrer">📖 Guide ↗</a>
    </div>
  </header>

  <div class="grid" style="grid-template-columns: 300px 1fr 1fr; align-items:start;">
    <div class="card">
      <div class="eyebrow">Endpoints</div>
      <input class="tab" style="width:100%; color:var(--brand-ink); cursor:text;" placeholder="Search endpoints…" bind:value={search} />
      <div class="tabs" style="margin:10px 0;">
        {#each groups as g}<button class="tab" class:active={group === g} on:click={() => (group = g)}>{g}</button>{/each}
      </div>
      <div class="endpoint-list">
        {#each visible as op (op.method + op.path)}
          <button class="endpoint" class:active={selected === op} on:click={() => choose(op)}>
            <span class="method {op.method.toLowerCase()}">{op.method}</span>
            <span><strong>{op.title}</strong><small>{op.path}</small></span>
          </button>
        {/each}
      </div>
    </div>

    <div class="card">
      <div class="eyebrow">{selected.group}</div>
      <h2>{selected.title}</h2>
      <p class="muted small">{selected.description}</p>
      <div class="request-line"><span class="method {selected.method.toLowerCase()}">{selected.method}</span><code>{selected.path}</code></div>
      <p class="small muted" style="margin-top:10px;">Auth: {selected.auth ? 'bearer token (runner/official)' : 'public'}</p>
      {#if selected.method === 'POST'}
        <label class="field" style="margin-top:14px;">Request body<textarea rows="10" bind:value={bodyText}></textarea></label>
      {/if}
      <div class="action-row" style="margin-top:14px;">
        <button class="btn primary" disabled={busy} on:click={tryIt}>{busy ? 'Sending…' : '▶ Try it'}</button>
        <button class="btn" on:click={copyCurl}>Copy cURL</button>
      </div>
    </div>

    <div class="card">
      <div class="eyebrow">Response</div>
      <div class="status" class:ok={status >= 200 && status < 300} class:bad={status >= 400}>{status || '—'}</div>
      <pre class="response">{response || 'Run a request to see the response here.'}</pre>
      {#if spec}<p class="muted small">OpenAPI contract loaded · {spec.length.toLocaleString()} chars</p>{/if}
    </div>
  </div>
</div>

<style>
  .grid { grid-template-columns: 300px minmax(0,1fr) minmax(0,1fr); }
  .endpoint-list { display:grid; gap:4px; max-height:65vh; overflow:auto; margin-top:10px; }
  .endpoint { display:flex; gap:8px; align-items:flex-start; text-align:left; width:100%; padding:9px; border-radius:9px; border:1px solid var(--brand-border); background:var(--brand-surface-3); color:var(--brand-ink); cursor:pointer; }
  .endpoint.active { border-color:var(--brand-primary); }
  .endpoint small { display:block; color:var(--brand-muted); font-size:.72rem; word-break:break-all; }
  .method { padding:3px 6px; border-radius:6px; font-size:.66rem; font-weight:900; flex:none; }
  .method.get { background:rgba(59,130,246,.2); color:#93c5fd; }
  .method.post { background:rgba(255,122,0,.18); color:var(--orange); }
  .request-line { display:flex; gap:8px; align-items:center; padding:10px; border-radius:10px; background:var(--brand-surface-3); margin-top:12px; }
  .field { display:grid; gap:5px; font-size:.8rem; color:var(--brand-muted); }
  textarea { padding:10px; border-radius:9px; background:var(--brand-surface-3); color:var(--brand-ink); border:1px solid var(--brand-border); font-family:ui-monospace,monospace; font-size:.78rem; }
  .status { font-family:var(--font-display); font-size:2rem; font-weight:900; margin:8px 0; color:var(--brand-muted); }
  .status.ok { color:var(--win); } .status.bad { color:var(--loss); }
  .response { min-height:300px; max-height:60vh; overflow:auto; padding:12px; border-radius:10px; background:var(--brand-surface); color:var(--brand-ink); font-family:ui-monospace,monospace; font-size:.76rem; white-space:pre-wrap; word-break:break-word; }
  @media (max-width: 1100px) { .grid { grid-template-columns:1fr !important; } }
</style>
