<script lang="ts">
  import { onMount } from 'svelte';
  import { docs } from '../../../lib/docs';
  import type { ApiOperation } from '../../../lib/docs';

  let operations: ApiOperation[] = [];
  let group = 'All';
  let search = '';
  let selected: ApiOperation | undefined;
  let values: Record<string, unknown> = {};
  let bodyText = '';
  let baseUrl = '';
  let token = '';
  let response = '';
  let status = 0;
  let elapsed = 0;
  let busy = false;
  let message = '';
  let error = '';
  let contractLoaded = false;

  $: groups = ['All', ...Array.from(new Set(operations.map((operation) => operation.group)))];
  $: visible = operations.filter((operation) => {
    const matchesGroup = group === 'All' || operation.group === group;
    const needle = search.toLowerCase();
    const matchesSearch = !needle || `${operation.method} ${operation.path} ${operation.title} ${operation.description}`.toLowerCase().includes(needle);
    return matchesGroup && matchesSearch;
  });

  onMount(() => {
    void docs.operations()
      .then((list) => {
        operations = list;
        contractLoaded = true;
        const query = new URLSearchParams(window.location.search);
        const path = query.get('path');
        const method = query.get('method');
        choose(list.find((operation) => operation.path === path && (!method || operation.method === method)) ?? list[0]);
      })
      .catch((cause) => (error = String(cause)));
    try {
      const raw = localStorage.getItem('j3-session');
      if (raw) token = String(JSON.parse(raw).token ?? '');
    } catch { /* ignore a stale session payload */ }
  });

  function choose(operation?: ApiOperation) {
    selected = operation;
    values = {};
    for (const param of operation?.parameters ?? []) values[param.name] = param.value;
    bodyText = operation?.body ?? '';
    response = '';
    status = 0;
    elapsed = 0;
  }

  function buildUrl(): string {
    if (!selected) return '';
    let path = selected.path;
    for (const param of selected.parameters.filter((item) => item.in === 'path')) {
      path = path.replace(`{${param.name}}`, encodeURIComponent(String(values[param.name] ?? param.value ?? '')));
    }
    const query = selected.parameters
      .filter((item) => item.in === 'query' && values[item.name] !== '' && values[item.name] !== undefined)
      .map((item) => [item.name, String(values[item.name])])
      .filter(([, value]) => value.length > 0);
    const suffix = query.length ? `?${new URLSearchParams(query).toString()}` : '';
    return `${baseUrl.replace(/\/+$/, '')}${path}${suffix}`;
  }

  function note(text: string) {
    message = text;
    setTimeout(() => { if (message === text) message = ''; }, 2200);
  }

  async function signInDemo() {
    try {
      const result = await fetch(`${baseUrl.replace(/\/+$/, '')}/api/v1/auth/demo`, { method: 'POST' });
      if (!result.ok) throw new Error(`HTTP ${result.status}`);
      const data = await result.json() as { token?: string };
      token = data.token ?? '';
      note(token ? 'Demo runner token loaded' : 'No token returned');
    } catch (cause) {
      note(`Demo sign-in failed: ${String(cause)}`);
    }
  }

  async function tryIt() {
    if (!selected) return;
    if (selected.path === '/api/v1/events/stream') {
      response = 'Server-sent event stream. Open /display to watch live events, or subscribe with an EventSource client.';
      status = 0;
      return;
    }
    busy = true;
    response = '';
    const started = performance.now();
    try {
      const headers: Record<string, string> = { Accept: 'application/json' };
      if (selected.auth && token) headers.Authorization = `Bearer ${token}`;
      const hasBody = selected.method !== 'GET' && bodyText.trim().length > 0;
      if (hasBody) headers['Content-Type'] = 'application/json';
      const result = await fetch(buildUrl(), {
        method: selected.method,
        headers,
        body: hasBody ? bodyText : undefined,
      });
      elapsed = Math.round(performance.now() - started);
      status = result.status;
      const text = await result.text();
      try { response = JSON.stringify(JSON.parse(text), null, 2); } catch { response = text; }
    } catch (cause) {
      status = 0;
      response = cause instanceof Error ? cause.message : String(cause);
    } finally {
      busy = false;
    }
  }

  async function copyCurl() {
    if (!selected) return;
    const headers: string[] = [`'Accept: application/json'`];
    if (selected.auth && token) headers.push(`'Authorization: Bearer ${token}'`);
    if (selected.method !== 'GET' && bodyText.trim()) headers.push(`'Content-Type: application/json'`);
    const body = selected.method !== 'GET' && bodyText.trim() ? ` --data '${bodyText.replaceAll("'", "\\'")}'` : '';
    const url = buildUrl();
    const fullUrl = url.startsWith('http') ? url : `http://127.0.0.1:8787${url}`;
    const command = `curl -X ${selected.method} '${fullUrl}' ${headers.map((header) => `-H ${header}`).join(' ')}${body}`;
    try {
      await navigator.clipboard.writeText(command);
      note('cURL copied to clipboard');
    } catch {
      note('Clipboard unavailable in this browser');
    }
  }
</script>

<svelte:head><title>J3 · API reference</title></svelte:head>

<div class="api-page">
  <header class="api-head">
    <div>
      <div class="eyebrow">Reference · OpenAPI 3.1</div>
      <h1>API explorer</h1>
      <p>Browse every venue endpoint and try it against the local server. Changes are real — use the demo database for experiments.</p>
    </div>
    <div class="head-actions">
      <a class="btn" href="/openapi.yaml" target="_blank" rel="noreferrer">📄 Raw contract ↗</a>
      <a class="btn" href="/docs/api-guide">📖 API guide</a>
    </div>
  </header>

  <div class="controls">
    <label>Server
      <input aria-label="API server base URL" bind:value={baseUrl} placeholder="Same origin (proxied)" />
    </label>
    <button class="btn" on:click={() => (baseUrl = 'http://127.0.0.1:8787')}>Use local :8787</button>
    <label class="grow">Bearer token
      <input aria-label="Bearer token" type="password" bind:value={token} placeholder="Paste a session token" />
    </label>
    <button class="btn" on:click={() => void signInDemo()}>🔑 Demo sign-in</button>
  </div>

  {#if error}<div class="error">Could not load the API index: {error}</div>{/if}
  {#if message}<div class="note" aria-live="polite">{message}</div>{/if}

  <div class="api-grid">
    <section class="panel endpoints" aria-label="Endpoints">
      <div class="panel-head">
        <strong>Endpoints</strong>
        <small>{contractLoaded ? `${operations.length} operations` : 'Loading…'}</small>
      </div>
      <input aria-label="Search endpoints" class="search" placeholder="Search paths and actions…" bind:value={search} />
      <div class="chips">
        {#each groups as item}
          <button class:active={group === item} on:click={() => (group = item)}>{item}</button>
        {/each}
      </div>
      <div class="list">
        {#each visible as operation (operation.method + operation.path)}
          <button class:active={selected === operation} on:click={() => choose(operation)}>
            <span class="method {operation.method.toLowerCase()}">{operation.method}</span>
            <span><strong>{operation.title}</strong><small>{operation.path}</small></span>
          </button>
        {:else}
          <p class="empty">No endpoints match “{search}”.</p>
        {/each}
      </div>
    </section>

    <section class="panel request" aria-label="Request">
      {#if selected}
        <div class="panel-head"><strong>{selected.title}</strong><small>{selected.group}</small></div>
        <div class="request-line"><span class="method {selected.method.toLowerCase()}">{selected.method}</span><code>{buildUrl()}</code></div>
        <p class="description">{selected.description || 'No description in the contract.'}</p>
        <p class="auth-note">Auth: {selected.auth ? 'bearer token required' : 'public endpoint'}</p>

        {#if selected.parameters.length}
          <h2>Parameters</h2>
          {#each selected.parameters as param}
            <label class="field">
              {param.name} <span class="pill">{param.in}{param.required ? ' · required' : ''}</span>
              <input
                aria-label={`${param.in} parameter ${param.name}`}
                bind:value={values[param.name]}
                placeholder={String(param.value ?? '')}
              />
              {#if param.description}<small>{param.description}</small>{/if}
            </label>
          {/each}
        {/if}

        {#if selected.method !== 'GET'}
          <h2>Request body</h2>
          <textarea aria-label="Request body" rows="10" bind:value={bodyText}></textarea>
        {/if}

        <div class="actions">
          <button class="btn primary" disabled={busy} on:click={() => void tryIt()}>{busy ? 'Sending…' : '▶ Try it'}</button>
          <button class="btn" on:click={() => void copyCurl()}>Copy cURL</button>
        </div>
      {:else}
        <p class="empty">Choose an endpoint to see its details.</p>
      {/if}
    </section>

    <section class="panel response-panel" aria-label="Response">
      <div class="panel-head"><strong>Response</strong><small>{elapsed ? `${elapsed} ms` : '—'}</small></div>
      <div class="status" class:ok={status >= 200 && status < 300} class:bad={status >= 400}>{status || '—'}</div>
      <pre>{response || 'Run a request to see the response here.'}</pre>
    </section>
  </div>
</div>

<style>
  .api-page { max-width:1800px; margin:0 auto; }
  .api-head { display:flex; justify-content:space-between; gap:18px; flex-wrap:wrap; margin-bottom:18px; }
  .eyebrow { color:var(--brand-accent); font-size:.7rem; letter-spacing:.14em; font-weight:900; text-transform:uppercase; }
  h1 { font-size:clamp(1.9rem,4vw,2.8rem); margin:8px 0; }
  .api-head p { max-width:70ch; color:var(--brand-muted); font-size:.9rem; line-height:1.6; }
  .head-actions { display:flex; gap:8px; align-items:flex-start; }
  .controls { display:flex; flex-wrap:wrap; align-items:end; gap:10px; padding:12px; margin-bottom:12px; border:1px solid var(--brand-border); border-radius:14px; background:var(--brand-surface-2); }
  .controls label { display:grid; gap:5px; font-size:.72rem; color:var(--brand-muted); }
  .controls .grow { flex:1; min-width:200px; }
  input, textarea { width:100%; padding:9px 10px; border-radius:9px; background:var(--brand-surface-3); border:1px solid var(--brand-border); color:var(--brand-ink); font:inherit; font-size:.8rem; }
  .api-grid { display:grid; grid-template-columns:minmax(240px,300px) minmax(0,1.2fr) minmax(260px,1fr); gap:12px; align-items:start; }
  .panel { min-width:0; padding:14px; border-radius:16px; border:1px solid var(--brand-border); background:var(--brand-surface-2); }
  .panel-head { display:flex; justify-content:space-between; gap:8px; align-items:baseline; margin-bottom:12px; }
  .panel-head small { color:var(--brand-muted); font-size:.72rem; }
  .search { margin-bottom:10px; }
  .chips { display:flex; flex-wrap:wrap; gap:5px; margin-bottom:10px; }
  .chips button { padding:5px 8px; border-radius:99px; border:1px solid var(--brand-border); background:var(--brand-surface-3); color:var(--brand-muted); font-size:.68rem; cursor:pointer; }
  .chips button.active { color:var(--brand-ink); border-color:var(--brand-primary); }
  .list { display:grid; gap:5px; max-height:65vh; overflow:auto; }
  .list button { display:flex; gap:8px; align-items:flex-start; text-align:left; padding:9px; border-radius:10px; border:1px solid transparent; background:var(--brand-surface-3); color:var(--brand-ink); cursor:pointer; }
  .list button.active { border-color:var(--brand-primary); }
  .list strong { display:block; font-size:.78rem; }
  .list small { display:block; margin-top:2px; font-size:.67rem; color:var(--brand-muted); word-break:break-all; }
  .method { flex:none; padding:3px 6px; border-radius:6px; font-size:.63rem; font-weight:900; }
  .method.get { background:rgba(59,130,246,.2); color:#93c5fd; }
  .method.post { background:rgba(255,122,0,.18); color:var(--orange); }
  .method.delete, .method.put, .method.patch { background:rgba(255,85,51,.18); color:var(--loss); }
  .request-line { display:flex; gap:8px; align-items:center; padding:10px; border-radius:10px; background:var(--brand-surface-3); overflow:auto; }
  .request-line code { font-size:.72rem; white-space:nowrap; }
  .description, .auth-note { color:var(--brand-muted); font-size:.8rem; line-height:1.55; }
  h2 { font-size:.9rem; margin:18px 0 8px; }
  .field { display:grid; gap:5px; margin-bottom:10px; font-size:.76rem; color:var(--brand-ink); }
  .field small { color:var(--brand-muted); }
  .pill { display:inline-block; padding:1px 6px; border-radius:99px; background:var(--brand-surface-3); color:var(--brand-muted); font-size:.62rem; }
  .actions { display:flex; gap:8px; margin-top:16px; flex-wrap:wrap; }
  .status { font-family:var(--font-display); font-size:2rem; font-weight:900; color:var(--brand-muted); }
  .status.ok { color:var(--win); }
  .status.bad { color:var(--loss); }
  .response-panel pre { min-height:340px; max-height:65vh; overflow:auto; padding:12px; border-radius:10px; background:var(--brand-surface-3); color:var(--brand-ink); font-family:ui-monospace,monospace; font-size:.72rem; white-space:pre-wrap; word-break:break-word; margin:8px 0 0; }
  .empty { color:var(--brand-muted); font-size:.82rem; }
  .error, .note { margin-bottom:10px; padding:10px 12px; border-radius:10px; font-size:.8rem; }
  .error { background:rgba(255,85,51,.14); border:1px solid var(--loss); }
  .note { background:var(--brand-surface-3); border:1px solid var(--brand-primary); }
  @media (max-width:1200px) { .api-grid { grid-template-columns:minmax(220px,280px) minmax(0,1fr); } .response-panel { grid-column:1 / -1; } }
  @media (max-width:760px) { .api-grid { grid-template-columns:1fr; } .response-panel { grid-column:auto; } }
</style>
