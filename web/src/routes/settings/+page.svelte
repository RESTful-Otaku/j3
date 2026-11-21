<script lang="ts">
  import { onMount } from 'svelte';
  import { api, type BeltOption } from '../../lib/api';
  import { LOCALES, translate } from '../../lib/i18n';
  import { notify, notifySuccess, notifyError } from '../../lib/toast';

  const KEY = 'j3-settings-v1';
  const locales = [
    { id: 'en', label: 'English' }, { id: 'fr', label: 'Français' },
    { id: 'de', label: 'Deutsch' }, { id: 'es', label: 'Español' },
    { id: 'pt', label: 'Português' }, { id: 'hi', label: 'हिन्दी' },
    { id: 'pa', label: 'ਪੰਜਾਬੀ' }, { id: 'ar', label: 'العربية' },
    { id: 'ja', label: '日本語' }, { id: 'ko', label: '한국어' },
    { id: 'zh', label: '中文' },
  ];
  type Settings = {
    name: string; email: string; phone: string; dateOfBirth: string;
    unitSystem: 'metric' | 'imperial'; heightCm: number; weightKg: number;
    heightFt: number; heightIn: number; weightLb: number;
    beltRankId: string; customBeltLabel: string; customBeltColour: string;
    clubId: string; locale: string; appearance: string;
    accountRole: string; clubColour: string; clubHighlight: string; photoPath: string;
  };
  const defaults: Settings = {
    name: '', email: '', phone: '', dateOfBirth: '', unitSystem: 'metric',
    heightCm: 175, weightKg: 75, heightFt: 5, heightIn: 9, weightLb: 165,
    beltRankId: 'br-white', customBeltLabel: '', customBeltColour: '#14b8a6',
    clubId: '', locale: 'en', appearance: 'system',
    accountRole: 'fighter', clubColour: '#2563eb', clubHighlight: '#93c5fd', photoPath: '',
  };
  let settings: Settings = { ...defaults };
  let belts: BeltOption[] = [];

  onMount(async () => {
    try { belts = (await api.belts()).belts; } catch { belts = []; }
    try {
      const saved = localStorage.getItem(KEY);
      if (saved) settings = { ...defaults, ...JSON.parse(saved) };
    } catch { /* ignore */ }
    applyTheme();
  });

  function applyTheme() {
    const choice = settings.appearance;
    const resolved = choice === 'system'
      ? (window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark')
      : choice;
    document.documentElement.dataset.theme = resolved;
    document.documentElement.lang = settings.locale;
    const localeInfo = LOCALES.find((item) => item.id === settings.locale);
    document.documentElement.dir = localeInfo?.dir ?? 'ltr';
  }
  function save() {
    localStorage.setItem(KEY, JSON.stringify(settings));
    applyTheme();
    notifySuccess('Account settings saved');
  }

  const mediaOptions: Record<string, string[]> = {
    fighter: ['photo'],
    official: ['photo'],
    club_leader: ['club', 'banner'],
    tournament_runner: ['logo', 'banner'],
  };
  async function uploadMedia(event: Event, kind: string) {
    const input = event.target as HTMLInputElement;
    const file = input.files?.[0];
    if (!file) return;
    try {
      const result = await api.upload(file, kind, settings.accountRole);
      if (kind === 'photo') settings.photoPath = result.url;
      else {
        settings.clubColour = result.colour;
        settings.clubHighlight = result.highlight;
      }
      save();
      notify(`${kind} uploaded · colours ${result.colour} / ${result.highlight}`);
    } catch (error) {
      notifyError(error instanceof Error ? error.message : 'Upload failed');
    }
  }
  $: selectedBelt = belts.find((belt) => belt.id === settings.beltRankId);
</script>

<svelte:head><title>J3 · Account settings</title></svelte:head>

<div class="shell">
  <header class="topbar">
    <div class="brand"><div class="brand-mark">J3</div><div><div class="brand-name">Account settings</div><div class="brand-sub">Your details, units, grade and preferences</div></div></div>
    <a class="btn" href="/">← Back to dashboard</a>
  </header>

  <div class="card" style="margin-bottom:18px;">
    <div class="eyebrow">Getting started</div><h2>Guided walkthrough</h2>
    <p class="muted small">Repeat the short guided tour for your account type at any time.</p>
    <button class="btn" style="margin-top:10px;" on:click={() => window.dispatchEvent(new CustomEvent('j3-start-tour', { detail: { role: settings.accountRole === 'tournament_runner' ? 'runner' : settings.accountRole === 'official' ? 'official' : settings.accountRole === 'club_leader' ? 'club_leader' : 'fighter' } }))}>▶ Restart guided tour</button>
  </div>

<div class="card">
    <h2>Profile</h2>
    <div class="grid" style="grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); margin-top:14px;">
      <label class="field">Full name<input bind:value={settings.name} /></label>
      <label class="field">Email<input type="email" bind:value={settings.email} /></label>
      <label class="field">Mobile phone<input type="tel" bind:value={settings.phone} /></label>
      <label class="field">Date of birth<input type="date" bind:value={settings.dateOfBirth} /></label>
      <label class="field">Club affiliation<select bind:value={settings.clubId}>
        <option value="">Independent / no club</option>
        <option value="club-riverside">Riverside Jiu Jitsu</option>
        <option value="club-hilltop">Hilltop Dojo</option>
        <option value="club-city">City Warriors</option>
        <option value="club-sakura">Sakura Jiu Jitsu Academy</option>
      </select></label>
    </div>

    <h2 style="margin-top:26px;">Stats & units</h2>
    <div class="tabs" style="margin-top:10px;">
      <button class="tab" class:active={settings.unitSystem === 'metric'} on:click={() => (settings.unitSystem = 'metric')}>Metric</button>
      <button class="tab" class:active={settings.unitSystem === 'imperial'} on:click={() => (settings.unitSystem = 'imperial')}>Imperial</button>
    </div>
    <div class="grid" style="grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); margin-top:12px;">
      {#if settings.unitSystem === 'metric'}
        <label class="field">Height (cm)<input type="number" bind:value={settings.heightCm} /></label>
        <label class="field">Weight (kg)<input type="number" step="0.1" bind:value={settings.weightKg} /></label>
      {:else}
        <label class="field">Height (ft)<input type="number" bind:value={settings.heightFt} /></label>
        <label class="field">Height (in)<input type="number" bind:value={settings.heightIn} /></label>
        <label class="field">Weight (lb)<input type="number" bind:value={settings.weightLb} /></label>
      {/if}
    </div>

    <h2 style="margin-top:26px;">Rank / grade belt</h2>
    <p class="muted small">Choose from the full Jiu Jitsu spectrum — each option shows its colour.</p>
    <div class="belt-row" style="margin-top:12px;">
      <div class="belt-swatch" style="background:{settings.beltRankId === '__custom__' ? settings.customBeltColour : selectedBelt?.colour || '#94a3b8'}"></div>
      <select bind:value={settings.beltRankId}>
        {#each belts as belt (belt.id)}
          <option value={belt.id}>{belt.label}</option>
        {/each}
        <option value="__custom__">✏️ Enter my own…</option>
      </select>
    </div>
    {#if settings.beltRankId === '__custom__'}
      <div class="grid" style="grid-template-columns: 100px 1fr; margin-top:10px;">
        <label class="field">Colour<input type="color" bind:value={settings.customBeltColour} /></label>
        <label class="field">Belt name<input bind:value={settings.customBeltLabel} placeholder="e.g. Red & White senior" /></label>
      </div>
    {/if}

    <h2 style="margin-top:26px;">Language & appearance</h2>
    <div class="grid" style="grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); margin-top:12px;">
      <label class="field">Language<select bind:value={settings.locale}>
        {#each locales as locale}<option value={locale.id}>{locale.label}</option>{/each}
      </select></label>
      <label class="field">Appearance<select bind:value={settings.appearance} on:change={applyTheme}>
        <option value="system">Follow system</option><option value="light">Light</option>
        <option value="dark">Dark</option><option value="street">Street Fighter</option>
      </select></label>
    </div>
    <div class="capabilities" style="margin-top:12px;">
      {#each ['dashboard','setup','scoreboard','brackets','tables','medals','upcoming','live'] as key}
        <span class="cap on">{translate(settings.locale, key as any)}</span>
      {/each}
    </div>
    <h2 style="margin-top:26px;">Media & branding</h2>
    <p class="muted small">Uploads are limited by account type. Main and highlight colours are detected from your image.</p>
    <label class="field" style="margin-top:10px;">Account type<select bind:value={settings.accountRole}>
      <option value="fighter">Fighter — photo</option>
      <option value="official">Official — photo</option>
      <option value="club_leader">Club leader — logo & banner</option>
      <option value="tournament_runner">Runner / organiser — logo & banner</option>
    </select></label>
    <div class="grid" style="grid-template-columns: repeat(auto-fit, minmax(210px, 1fr)); margin-top:12px;">
      {#each mediaOptions[settings.accountRole] ?? [] as kind}
        <label class="field">{kind === 'photo' ? 'Your photo' : kind === 'club' ? 'Club symbol / logo' : 'Banner'}
          <input type="file" accept="image/png,image/jpeg,image/webp,image/svg+xml" on:change={(e) => uploadMedia(e, kind)} />
        </label>
      {/each}
      <label class="field">Main colour<input type="color" bind:value={settings.clubColour} /></label>
      <label class="field">Highlight colour<input type="color" bind:value={settings.clubHighlight} /></label>
    </div>
    <div class="action-row"><button class="btn primary" on:click={save}>Save settings</button></div>
    <p class="muted small" style="margin-top:12px;">Locales: English, French, German, Spanish, Portuguese, Hindi, Punjabi, Arabic, Japanese, Korean and Mandarin.</p>
  </div>
</div>

<style>
  :global(.field) { display:grid; gap:5px; font-size:.8rem; color:var(--brand-muted); }
  :global(.field input), :global(.field select) { padding:9px 11px; border-radius:9px; border:1px solid var(--brand-border); background:var(--brand-surface-3); color:var(--brand-ink); width:100%; }
  .belt-row { display:flex; gap:12px; align-items:center; }
  .belt-swatch { width:54px; height:30px; border-radius:7px; border:2px solid rgba(255,255,255,.3); flex:none; }
  .belt-row select { flex:1; padding:9px 11px; border-radius:9px; border:1px solid var(--brand-border); background:var(--brand-surface-3); color:var(--brand-ink); }
</style>
