<script lang="ts">
  import { onMount } from 'svelte';
  import { api, type BeltOption } from '../../lib/api';
  import { session, uiRole } from '../../lib/auth';
  import { LOCALES, translate } from '../../lib/i18n';
  import { notifySuccess, notifyError } from '../../lib/toast';
  import { setTheme, themeChoice, type ThemeChoice } from '../../lib/theme';

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
    clubId: string; locale: string; appearance: ThemeChoice;
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
  let saving = false;

  onMount(async () => {
    try { belts = (await api.belts()).belts; } catch { belts = []; }
    try {
      const saved = localStorage.getItem(KEY);
      if (saved) settings = { ...defaults, ...JSON.parse(saved) };
    } catch { /* ignore */ }
    settings.appearance = $themeChoice;
    applyLocale();
  });

  function applyLocale() {
    document.documentElement.lang = settings.locale;
    const localeInfo = LOCALES.find((item) => item.id === settings.locale);
    document.documentElement.dir = localeInfo?.dir ?? 'ltr';
  }
  function applyTheme() {
    setTheme(settings.appearance);
    applyLocale();
  }
  async function save() {
    saving = true;
    try {
      localStorage.setItem(KEY, JSON.stringify(settings));
      applyTheme();
      notifySuccess('Settings saved');
    } catch (error) {
      notifyError(error instanceof Error ? error.message : 'Save failed');
    } finally {
      saving = false;
    }
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
      notifySuccess(`${kind} uploaded`);
    } catch (error) {
      notifyError(error instanceof Error ? error.message : 'Upload failed');
    }
  }
  $: selectedBelt = belts.find((belt) => belt.id === settings.beltRankId);
</script>

<svelte:head><title>J3 · Account settings</title></svelte:head>

<div class="shell settings-shell">
  <header class="topbar">
    <div class="brand"><div class="brand-mark">J3</div><div><div class="brand-name">Account settings</div><div class="brand-sub">Your profile and preferences</div></div></div>
    <a class="btn" href="/">← Back</a>
  </header>

  <form class="settings-form" on:submit|preventDefault={save}>
    <section class="card">
      <div class="eyebrow">Profile</div>
      <h2>Your details</h2>
      <div class="field-row">
        <label class="field">Full name
          <input bind:value={settings.name} required />
        </label>
        <label class="field">Email
          <input type="email" bind:value={settings.email} required />
        </label>
      </div>
      <div class="field-row">
        <label class="field">Mobile phone
          <input type="tel" bind:value={settings.phone} />
        </label>
        <label class="field">Club
          <select bind:value={settings.clubId}>
            <option value="">Independent / no club</option>
            <option value="club-riverside">Riverside Jiu Jitsu</option>
            <option value="club-hilltop">Hilltop Dojo</option>
            <option value="club-city">City Warriors</option>
            <option value="club-sakura">Sakura Jiu Jitsu Academy</option>
          </select>
        </label>
      </div>
    </section>

    <section class="card">
      <div class="eyebrow">Stats &amp; units</div>
      <h2>Height &amp; weight</h2>
      <div class="tabs" style="margin-top:8px;">
        <button class="tab" class:active={settings.unitSystem === 'metric'} on:click={() => (settings.unitSystem = 'metric')}>Metric</button>
        <button class="tab" class:active={settings.unitSystem === 'imperial'} on:click={() => (settings.unitSystem = 'imperial')}>Imperial</button>
      </div>
      <div class="field-row" style="margin-top:12px;">
        {#if settings.unitSystem === 'metric'}
          <label class="field">Height (cm)<input type="number" bind:value={settings.heightCm} /></label>
          <label class="field">Weight (kg)<input type="number" step="0.1" bind:value={settings.weightKg} /></label>
        {:else}
          <label class="field">Height (ft)<input type="number" bind:value={settings.heightFt} /></label>
          <label class="field">Weight (lb)<input type="number" bind:value={settings.weightLb} /></label>
        {/if}
      </div>
    </section>

    <section class="card">
      <div class="eyebrow">Rank / grade</div>
      <h2>Belt</h2>
      <div class="belt-row" style="margin-top:10px;">
        <div class="belt-swatch" style="background:{settings.beltRankId === '__custom__' ? settings.customBeltColour : selectedBelt?.colour || '#94a3b8'}"></div>
        <select bind:value={settings.beltRankId}>
          {#each belts as belt (belt.id)}
            <option value={belt.id}>{belt.label}</option>
          {/each}
          <option value="__custom__">✏️ Enter my own…</option>
        </select>
      </div>
      {#if settings.beltRankId === '__custom__'}
        <div class="field-row" style="margin-top:10px;">
          <label class="field">Colour
            <input type="color" bind:value={settings.customBeltColour} />
          </label>
          <label class="field">Belt name
            <input bind:value={settings.customBeltLabel} placeholder="e.g. Red & White senior" />
          </label>
        </div>
      {/if}
    </section>

    <section class="card">
      <div class="eyebrow">Appearance &amp; language</div>
      <h2>Preferences</h2>
      <div class="field-row">
        <label class="field">Language
          <select bind:value={settings.locale}>
            {#each locales as locale}<option value={locale.id}>{locale.label}</option>{/each}
          </select>
        </label>
        <label class="field">Appearance
          <select bind:value={settings.appearance} on:change={applyTheme}>
            <option value="system">Follow system</option>
            <option value="light">Light</option>
            <option value="dark">Dark gray</option>
            <option value="street">Street Fighter</option>
          </select>
        </label>
      </div>
    </section>

    <section class="card">
      <div class="eyebrow">Media &amp; branding</div>
      <h2>Uploads</h2>
      <p class="muted small" style="margin:6px 0 12px;">
        Main and highlight colours are detected from your image. Limited by account type.
      </p>
      <label class="field">Account type
        <select bind:value={settings.accountRole}>
          <option value="fighter">Fighter — photo</option>
          <option value="official">Official — photo</option>
          <option value="club_leader">Club leader — logo & banner</option>
          <option value="tournament_runner">Runner / organiser — logo & banner</option>
        </select>
      </label>
      <div class="field-row" style="margin-top:12px;">
        {#each mediaOptions[settings.accountRole] ?? [] as kind}
          <label class="field">{kind === 'photo' ? 'Your photo' : kind === 'club' ? 'Club symbol / logo' : 'Banner'}
            <input type="file" accept="image/png,image/jpeg,image/webp,image/svg+xml" on:change={(e) => uploadMedia(e, kind)} />
          </label>
        {/each}
        <label class="field">Main colour
          <input type="color" bind:value={settings.clubColour} />
        </label>
        <label class="field">Highlight colour
          <input type="color" bind:value={settings.clubHighlight} />
        </label>
      </div>
    </section>

    <div class="action-row" style="margin-top:24px; justify-content:flex-end;">
      <button class="btn primary" disabled={saving}>{saving ? 'Saving…' : 'Save settings'}</button>
    </div>
  </form>
</div>

<style>
  :global(.settings-shell) { max-width: 520px; margin: 0 auto; }
  :global(.settings-shell) .shell { padding-bottom: calc(92px + var(--safe-bottom)); }
  :global(.settings-shell) .eyebrow { font-size: .7rem; letter-spacing: .14em; text-transform: uppercase; }
  :global(.settings-shell) h2 { font-size: 1.15rem; margin: 4px 0 14px; letter-spacing: -0.01em; }
  :global(.settings-shell) .card { padding: 16px; gap: 0; }
  :global(.settings-shell) .card + .card { margin-top: 12px; }
  :global(.settings-shell) .field-row { display: grid; gap: 10px; grid-template-columns: 1fr; }
  :global(.settings-shell) .field { display: grid; gap: 5px; font-size: .72rem; color: var(--brand-muted); margin: 0; }
  :global(.settings-shell) .field input,
  :global(.settings-shell) .field select { padding: 10px 12px; border-radius: 10px; border: 1px solid var(--brand-border); background: var(--brand-surface-3); color: var(--brand-ink); width: 100%; font-size: 1rem; }
  :global(.settings-shell) .field input:focus { outline: 2px solid var(--brand-accent); outline-offset: 1px; }
  :global(.settings-shell) .belt-row { display: flex; gap: 12px; align-items: center; }
  :global(.settings-shell) .belt-swatch { width: 44px; height: 30px; border-radius: 8px; border: 2px solid rgba(255,255,255,.35); flex: none; }
  :global(.settings-shell) .belt-row select { flex: 1; padding: 10px 12px; border-radius: 10px; border: 1px solid var(--brand-border); background: var(--brand-surface-3); color: var(--brand-ink); }
  :global(.settings-shell) .tabs { gap: 8px; margin-top: 6px; }
  :global(.settings-shell) .tab { padding: 7px 14px; border-radius: 10px; border: 1px solid var(--brand-border); background: var(--brand-surface-3); color: var(--brand-ink); font-size: .82rem; font-weight: 700; }
  :global(.settings-shell) .tab.active { background: var(--brand-primary); color: var(--brand-on-primary); border-color: var(--brand-primary); }
  :global(.settings-shell) .btn { min-height: 46px; padding: 0 22px; font-size: 1rem; font-weight: 800; }
  :global(.settings-shell) .btn.primary { background: var(--brand-primary-fill); color: var(--brand-on-primary); border: 0; border-radius: 12px; box-shadow: 0 10px 20px color-mix(in srgb, var(--brand-primary) 45%, transparent); }
  :global(.settings-shell) .muted.small { font-size: .75rem; }
  @media (max-width: 700px) {
    :global(.settings-shell) { max-width: 100%; }
    :global(.settings-shell) .shell { padding: calc(10px + var(--safe-top)) 14px calc(96px + var(--safe-bottom)) 14px; }
    :global(.settings-shell) .card { padding: 14px; border-radius: 14px; }
    :global(.settings-shell) .card + .card { margin-top: 10px; }
    :global(.settings-shell) h2 { font-size: 1.05rem; margin: 3px 0 12px; }
    :global(.settings-shell) .field-row { gap: 8px; }
    :global(.settings-shell) .btn { min-height: 46px; width: 100%; justify-content: center; }
  }
</style>
