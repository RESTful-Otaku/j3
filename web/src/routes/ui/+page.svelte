<script lang="ts">
  import ConfirmDialog from '../../lib/components/ConfirmDialog.svelte';
  import { notify, notifySuccess, notifyWarning, notifyError } from '../../lib/toast';
  import { setTheme, themeChoice } from '../../lib/theme';
  import { cornerConfig } from '../../lib/event-brand';

  let dialogOpen = false;
  let slider = 60;
  let toggle = true;
  let choice = 'one';
  let search = '';
</script>

<svelte:head><title>J3 · Component gallery</title></svelte:head>

<div class="shell">
  <header class="topbar">
    <div class="brand"><a class="brand-mark" href="/public" aria-label="Public dashboard">J3</a><div><div class="brand-name">Component gallery</div><div class="brand-sub">Storybook-style QA for every UI primitive</div></div></div>
    <a class="btn" href="/control">🖥️ Display control</a>
  </header>

  <div class="card">
    <div class="eyebrow">Design tokens</div><h2>Theme & typography</h2>
    <div class="tabs" style="margin-top:12px;">
      {#each ['dark','light','street'] as option}
        <button class="tab" class:active={$themeChoice === option} on:click={() => setTheme(option as 'dark' | 'light' | 'street')}>{option}</button>
      {/each}
    </div>
    <p style="margin-top:14px;">Heading example — clean, high contrast, readable at distance.</p>
    <p class="muted small">Muted supporting text for context and secondary detail.</p>
  </div>

  <div class="grid" style="grid-template-columns: repeat(auto-fit, minmax(280px,1fr)); margin-top:18px;">
    <div class="card">
      <div class="eyebrow">Actions</div><h2>Buttons</h2>
      <div class="action-row">
        <button class="btn primary">Primary</button>
        <button class="btn">Secondary</button>
        <button class="btn warn">Warning</button>
        <button class="btn" on:click={() => (dialogOpen = true)}>Open dialog</button>
      </div>
    </div>
    <div class="card">
      <div class="eyebrow">States</div><h2>Badges & outcomes</h2>
      <div class="capabilities" style="margin-top:12px;">
        <span class="badge runner">Runner</span><span class="badge official">Official</span>
        <span class="badge fighter">W · WP</span><span class="badge spectator">L · LP</span>
        <span class="badge fighter">W · WS</span><span class="badge spectator">D · DP</span>
      </div>
    </div>
    <div class="card">
      <div class="eyebrow">Notifications</div><h2>Toasts</h2>
      <div class="action-row">
        <button class="btn" on:click={() => notifySuccess('Saved successfully')}>Success</button>
        <button class="btn" on:click={() => notifyWarning('Check this value')}>Warning</button>
        <button class="btn" on:click={() => notifyError('Something failed')}>Error</button>
        <button class="btn" on:click={() => notify('Undo me', { undo: () => {} })}>With undo</button>
      </div>
    </div>
  </div>

  <div class="grid" style="grid-template-columns: repeat(auto-fit, minmax(280px,1fr)); margin-top:18px;">
    <div class="card">
      <div class="eyebrow">Inputs</div><h2>Widgets</h2>
      <label class="field" style="display:grid; gap:5px; margin-top:12px;">Search<input bind:value={search} placeholder="Type here…" /></label>
      <label class="small" style="display:flex; gap:8px; align-items:center; margin-top:12px;"><input type="checkbox" bind:checked={toggle} /> Checkbox</label>
      <div class="tabs" style="margin-top:10px;">
        <button class="tab" class:active={choice === 'one'} on:click={() => (choice = 'one')}>Option one</button>
        <button class="tab" class:active={choice === 'two'} on:click={() => (choice = 'two')}>Option two</button>
      </div>
      <label class="small" style="display:block; margin-top:14px;">Slider · {slider}<input type="range" min="0" max="100" bind:value={slider} /></label>
    </div>
    <div class="card">
      <div class="eyebrow">Data</div><h2>Table & charts</h2>
      <table style="width:100%; border-collapse:collapse; margin-top:12px; font-size:.86rem;">
        <thead><tr class="muted small" style="text-align:left;"><th>Name</th><th>W</th><th>Pts</th></tr></thead>
        <tbody>
          <tr style="border-top:1px solid var(--brand-border);"><td>Yuki Mori</td><td>3</td><td>9</td></tr>
          <tr style="border-top:1px solid var(--brand-border);"><td>Aisha Khan</td><td>2</td><td>6</td></tr>
        </tbody>
      </table>
      <div class="bar-row" style="margin-top:14px;"><span>Takedown</span><div class="bar"><div class="bar-fill" style="width:80%"></div></div><strong>12</strong></div>
      <div class="bar-row"><span>Submission</span><div class="bar"><div class="bar-fill alt" style="width:45%"></div></div><strong>7</strong></div>
    </div>
  </div>

  <div class="card" style="margin-top:18px;">
    <div class="eyebrow">Broadcast</div><h2>Scoreboard card</h2>
    <div class="scoreboard" style="margin-top:14px;">
      <div class="corner"><span class="portrait" style="width:76px;height:76px;">YM</span><div><div class="small muted">{$cornerConfig.red.label} · Blue belt</div><strong>Yuki Mori</strong></div><div class="score" style="margin-left:auto;">6</div></div>
      <div style="text-align:center;"><div class="clock">02:14</div><span class="badge fighter">LIVE</span></div>
      <div class="corner blue"><div class="score" style="margin-right:auto;">4</div><div><div class="small muted">{$cornerConfig.blue.label} · Purple belt</div><strong>Aisha Khan</strong></div><span class="portrait" style="width:76px;height:76px;">AK</span></div>
    </div>
  </div>
</div>

<ConfirmDialog
  open={dialogOpen}
  title="Confirm this action?"
  message="This is how important confirmations appear to runners and officials."
  confirmLabel="Confirm"
  onConfirm={() => notifySuccess('Confirmed')}
/>

<style>
  .field input { padding:9px 11px; border-radius:9px; border:1px solid var(--brand-border); background:var(--brand-surface-3); color:var(--brand-ink); }
  .bar-row { display:grid; grid-template-columns:90px 1fr 30px; align-items:center; gap:10px; font-size:.84rem; }
  .bar { height:12px; border-radius:999px; background:var(--brand-surface-3); overflow:hidden; }
  .bar-fill { height:100%; background:var(--brand-primary); border-radius:999px; }
  .bar-fill.alt { background:var(--win); }
</style>
