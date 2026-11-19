<script lang="ts">
  import { onMount } from 'svelte';
  import { api, type CreateTournamentRequest } from '../../lib/api';
  import { notify, notifySuccess, notifyError, notifyWarning } from '../../lib/toast';

  type Discipline = {
    id: string; name: string; title: string; style: string; scoring: string;
    matchSeconds: number; rounds: number; requirements: string;
  };
  type Category = { id: string; name: string; kind: string; gender: string; min: number; max: number };
  type Invite = { id: string; email: string; phone: string; role: string; clubName: string };
  type Wizard = {
    step: number; name: string; title: string; venue: string; startDate: string; endDate: string;
    theme: string; logoPath: string; bannerPath: string; saveTemplate: boolean; templateName: string;
    disciplines: Discipline[]; categories: Category[];
    constraints: { cap: number; maxPools: number; autoPlace: boolean; separateClubs: boolean; genderPolicy: string };
    invites: Invite[];
  };

  const STORAGE_KEY = 'j3-wizard-v1';
  const steps = ['Event', 'Disciplines', 'Categories', 'People', 'Review'];
  const themePresets = ['classic', 'ocean', 'street'];
  const stylePresets = ['Round robin', 'Single elimination', 'Double elimination', 'Fight pools', 'Kata panel'];
  const scoringPresets = ['Ippon / Waza-ari', 'BJJ points', 'Kata panel 5–10', 'Custom'];
  const uid = () => Math.random().toString(36).slice(2, 9);

  function defaults(): Wizard {
    return {
      step: 0, name: '', title: '', venue: '', startDate: '', endDate: '', theme: 'classic',
      logoPath: '', bannerPath: '', saveTemplate: true, templateName: '',
      disciplines: [{ id: uid(), name: 'Fighting System', title: 'Main fighting division', style: 'Fight pools', scoring: 'Ippon / Waza-ari', matchSeconds: 180, rounds: 3, requirements: 'Gi inspection; 3 second control' }],
      categories: [
        { id: uid(), name: 'Adult -76 kg', kind: 'weight', gender: 'male', min: 0, max: 76 },
        { id: uid(), name: 'Youth U16', kind: 'age', gender: 'open', min: 14, max: 16 },
      ],
      constraints: { cap: 200, maxPools: 12, autoPlace: true, separateClubs: true, genderPolicy: 'optional' },
      invites: [],
    };
  }

  let wizard: Wizard = defaults();
  let ready = false;
  let errors: { step: number; message: string }[] = [];
  let bulkInvites = '';
  let dragIndex = -1;

  onMount(() => {
    try {
      const saved = localStorage.getItem(STORAGE_KEY);
      if (saved) wizard = { ...defaults(), ...JSON.parse(saved) };
    } catch { /* ignore corrupt drafts */ }
    ready = true;
    validate();
  });

  function persist() {
    if (ready) localStorage.setItem(STORAGE_KEY, JSON.stringify(wizard));
  }
  function touch() {
    wizard = { ...wizard };
    persist();
    validate();
  }
  function validate() {
    errors = [];
    if (!wizard.name.trim()) errors.push({ step: 0, message: 'Give your tournament a name.' });
    if (!wizard.startDate) errors.push({ step: 0, message: 'Choose a start date.' });
    if (wizard.disciplines.length === 0) errors.push({ step: 1, message: 'Add at least one discipline.' });
    if (wizard.disciplines.some((d) => !d.name.trim())) errors.push({ step: 1, message: 'Every discipline needs a name.' });
    if (wizard.categories.length === 0) errors.push({ step: 2, message: 'Add at least one category.' });
    if (wizard.constraints.cap < 2) errors.push({ step: 2, message: 'Participant cap must be at least 2.' });
    if (wizard.constraints.maxPools < 1) errors.push({ step: 2, message: 'Allow at least one fight pool.' });
    for (const invite of wizard.invites) {
      if (!invite.email.includes('@')) errors.push({ step: 3, message: `Invalid invite email: ${invite.email || '(empty)'}` });
    }
  }
  function go(step: number) {
    wizard.step = Math.max(0, Math.min(4, step));
    touch();
  }
  function addDiscipline() {
    wizard.disciplines = [...wizard.disciplines, {
      id: uid(), name: '', title: '', style: stylePresets[0], scoring: scoringPresets[0],
      matchSeconds: 180, rounds: 3, requirements: '',
    }];
    touch();
  }
  function removeDiscipline(index: number) {
    const removed = wizard.disciplines[index];
    wizard.disciplines = wizard.disciplines.filter((_, i) => i !== index);
    touch();
    notifySuccess('Discipline removed', () => {
      const list = [...wizard.disciplines];
      list.splice(index, 0, removed);
      wizard.disciplines = list;
      touch();
    });
  }
  function dragStart(index: number, event: DragEvent) {
    dragIndex = index;
    event.dataTransfer?.setData('text/plain', String(index));
  }
  function drop(index: number, event: DragEvent) {
    event.preventDefault();
    if (dragIndex < 0 || dragIndex === index) return;
    const list = [...wizard.disciplines];
    const [moved] = list.splice(dragIndex, 1);
    list.splice(index, 0, moved);
    wizard.disciplines = list;
    dragIndex = -1;
    touch();
    notifySuccess('Discipline order updated');
  }
  function addCategory() {
    wizard.categories = [...wizard.categories, { id: uid(), name: '', kind: 'weight', gender: 'open', min: 0, max: 100 }];
    touch();
  }
  function removeCategory(index: number) {
    wizard.categories = wizard.categories.filter((_, i) => i !== index);
    touch();
    notifySuccess('Category removed');
  }
  function parseBulkInvites() {
    const added: Invite[] = [];
    for (const line of bulkInvites.split('\n').map((l) => l.trim()).filter(Boolean)) {
      const [email, phone, role, clubName] = line.split(',').map((part) => part.trim());
      if (email) added.push({ id: uid(), email, phone: phone ?? '', role: role ?? 'fighter', clubName: clubName ?? '' });
    }
    if (added.length === 0) { notifyWarning('No invites found — use one per line: email, phone, role, club'); return; }
    wizard.invites = [...wizard.invites, ...added];
    bulkInvites = '';
    touch();
    notifySuccess(`${added.length} invitation${added.length === 1 ? '' : 's'} added`);
  }
  async function upload(event: Event, kind: 'logo' | 'banner') {
    const input = event.target as HTMLInputElement;
    const file = input.files?.[0];
    if (!file) return;
    try {
      const result = await api.upload(file, kind, 'tournament_runner');
      if (kind === 'logo') wizard.logoPath = result.url;
      else wizard.bannerPath = result.url;
      touch();
      notifySuccess(`${kind === 'logo' ? 'Logo' : 'Banner'} uploaded`);
    } catch (error) {
      notifyError(error instanceof Error ? error.message : 'Upload failed');
    }
  }
  async function launch() {
    validate();
    if (errors.length > 0) {
      notifyWarning(`${errors.length} issue${errors.length === 1 ? '' : 's'} to fix before launching`);
      go(errors[0].step);
      return;
    }
    const payload: CreateTournamentRequest = {
      name: wizard.name, title: wizard.title || wizard.name, venue: wizard.venue,
      startDate: wizard.startDate, endDate: wizard.endDate, theme: wizard.theme,
      logoPath: wizard.logoPath, bannerPath: wizard.bannerPath,
      saveTemplate: wizard.saveTemplate, templateName: wizard.templateName,
      disciplines: wizard.disciplines.map((d) => ({
        name: d.name, title: d.title, style: d.style, scoring: d.scoring,
        matchSeconds: Number(d.matchSeconds), rounds: Number(d.rounds),
        requirements: d.requirements.split(',').map((r) => r.trim()).filter(Boolean),
      })),
      categories: wizard.categories.map((c) => ({
        name: c.name, kind: c.kind, gender: c.gender, min: c.min, max: c.max,
      })),
      constraints: wizard.constraints,
      invites: wizard.invites,
    };
    try {
      const result = await api.createTournament(payload);
      notifySuccess(`Tournament created · ${result.disciplineCount} disciplines · ${result.categoryCount} categories · ${result.inviteCount} invites`);
      localStorage.removeItem(STORAGE_KEY);
      wizard = defaults();
      ready = false;
      go(0);
      ready = true;
    } catch (error) {
      notifyError(error instanceof Error ? error.message : 'Could not create tournament');
    }
  }
</script>

<svelte:head><title>J3 · Tournament setup</title></svelte:head>

<div class="shell">
  <header class="topbar">
    <div class="brand"><div class="brand-mark">J3</div><div><div class="brand-name">Create a tournament</div><div class="brand-sub">Five quick steps · everything saved as you go</div></div></div>
    <a class="btn" href="/control" target="_blank" rel="noreferrer">🖥️ Displays ↗</a>
  </header>

  <div class="stepper">
    {#each steps as label, index}
      <button class="step" class:active={wizard.step === index} class:done={wizard.step > index} on:click={() => go(index)}>
        <span class="num">{wizard.step > index ? '✓' : index + 1}</span><span>{label}</span>
      </button>
    {/each}
  </div>

  <div class="card setup-card" on:input={persist} on:change={touch}>
    {#if wizard.step === 0}
      <div class="eyebrow">Step 1 of 5</div><h2>Event & identity</h2>
      <p class="muted small">Give the event a name and look. Pick a theme or upload your own branding.</p>
      <div class="grid" style="grid-template-columns: repeat(auto-fit, minmax(240px, 1fr)); margin-top:18px;">
        <label class="field">Tournament name<input bind:value={wizard.name} placeholder="Autumn Challenge 2026" /></label>
        <label class="field">Display title<input bind:value={wizard.title} placeholder="Main Hall · Mats 1–4" /></label>
        <label class="field">Venue<input bind:value={wizard.venue} placeholder="Riverside Sports Centre" /></label>
        <label class="field">Start date<input type="date" bind:value={wizard.startDate} /></label>
        <label class="field">End date<input type="date" bind:value={wizard.endDate} /></label>
      </div>
      <h3 style="margin-top:22px;">Theme</h3>
      <div class="tabs" style="margin-top:10px;">
        {#each themePresets as theme}
          <button class="tab" class:active={wizard.theme === theme} on:click={() => { wizard.theme = theme; touch(); }}>
            {theme === 'classic' ? '🔵 Classic' : theme === 'ocean' ? '🌊 Ocean' : '⚡ Street Fighter'}
          </button>
        {/each}
      </div>
      <div class="grid" style="grid-template-columns:1fr 1fr; margin-top:18px;">
        <label class="field">Organiser logo<input type="file" accept="image/*" on:change={(e) => upload(e, 'logo')} />{#if wizard.logoPath}<span class="muted small">✓ uploaded</span>{/if}</label>
        <label class="field">Event banner<input type="file" accept="image/*" on:change={(e) => upload(e, 'banner')} />{#if wizard.bannerPath}<span class="muted small">✓ uploaded</span>{/if}</label>
      </div>
    {:else if wizard.step === 1}
      <div class="eyebrow">Step 2 of 5</div><h2>Disciplines & rules</h2>
      <p class="muted small">Each discipline has its own scoring, timers, rounds and requirements. Drag cards to reorder.</p>
      <div style="display:grid; gap:12px; margin-top:18px;">
        {#each wizard.disciplines as discipline, i (discipline.id)}
          <div class="rule-card" role="listitem" aria-label="Discipline {i + 1}, draggable" draggable="true" on:dragstart={(e) => dragStart(i, e)} on:dragover={(e) => e.preventDefault()} on:drop={(e) => drop(i, e)}>
            <div style="display:flex; justify-content:space-between; align-items:center;">
              <span class="badge fighter">#{i + 1} · drag to reorder</span>
              <button class="tab" on:click={() => removeDiscipline(i)}>Remove</button>
            </div>
            <div class="grid" style="grid-template-columns: repeat(auto-fit, minmax(190px, 1fr)); margin-top:12px;">
              <label class="field">Discipline name<input bind:value={discipline.name} placeholder="Fighting System" /></label>
              <label class="field">Event title<input bind:value={discipline.title} placeholder="Main fighting division" /></label>
              <label class="field">Tournament style
                <select bind:value={discipline.style}>{#each stylePresets as style}<option value={style}>{style}</option>{/each}</select>
              </label>
              <label class="field">Point system
                <select bind:value={discipline.scoring}>{#each scoringPresets as scoring}<option value={scoring}>{scoring}</option>{/each}</select>
              </label>
              <label class="field">Match length
                <select bind:value={discipline.matchSeconds}><option value={90}>90 seconds</option><option value={180}>3 minutes</option><option value={300}>5 minutes</option><option value={480}>8 minutes</option></select>
              </label>
              <label class="field">Rounds <strong>{discipline.rounds}</strong>
                <input type="range" min="1" max="7" bind:value={discipline.rounds} />
              </label>
            </div>
            <label class="field" style="margin-top:12px;">Requirements / constraints<input bind:value={discipline.requirements} placeholder="Gi inspection, 3s control, no jumping guard…" /></label>
          </div>
        {/each}
      </div>
      <button class="btn" style="margin-top:14px;" on:click={addDiscipline}>+ Add discipline</button>
    {:else if wizard.step === 2}
      <div class="eyebrow">Step 3 of 5</div><h2>Categories & constraints</h2>
      <p class="muted small">Fighters are placed automatically into compatible categories using their stats.</p>
      <div style="display:grid; gap:10px; margin-top:18px;">
        {#each wizard.categories as category, i (category.id)}
          <div class="rule-card" style="display:flex; gap:10px; align-items:end; flex-wrap:wrap;">
            <label class="field" style="flex:2;">Category name<input bind:value={category.name} /></label>
            <label class="field">Type<select bind:value={category.kind}><option value="weight">Weight</option><option value="height">Height</option><option value="age">Age</option><option value="belt">Rank / belt</option><option value="combined">Combined</option></select></label>
            <label class="field">Gender<select bind:value={category.gender}><option value="open">Open</option><option value="male">Male</option><option value="female">Female</option></select></label>
            <label class="field">From<input type="number" bind:value={category.min} /></label>
            <label class="field">To<input type="number" bind:value={category.max} /></label>
            <button class="tab" on:click={() => removeCategory(i)}>Remove</button>
          </div>
        {/each}
      </div>
      <button class="btn" style="margin-top:12px;" on:click={addCategory}>+ Add category</button>
      <h3 style="margin-top:24px;">Automation limits</h3>
      <div class="grid" style="grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); margin-top:12px;">
        <label class="field">Participant cap <strong>{wizard.constraints.cap}</strong><input type="range" min="2" max="1000" step="2" bind:value={wizard.constraints.cap} /></label>
        <label class="field">Maximum fight pools <strong>{wizard.constraints.maxPools}</strong><input type="range" min="1" max="32" bind:value={wizard.constraints.maxPools} /></label>
        <label class="field">Gender policy<select bind:value={wizard.constraints.genderPolicy}><option value="optional">Optional / open</option><option value="split">Separate male & female</option><option value="strict">Strict, no transfers</option></select></label>
      </div>
      <div class="capabilities" style="margin-top:16px;">
        <label class="cap" class:on={wizard.constraints.autoPlace}><input type="checkbox" bind:checked={wizard.constraints.autoPlace} /> Automatically place fighters into pools</label>
        <label class="cap" class:on={wizard.constraints.separateClubs}><input type="checkbox" bind:checked={wizard.constraints.separateClubs} /> Spread fighters from the same club</label>
      </div>
    {:else if wizard.step === 3}
      <div class="eyebrow">Step 4 of 5</div><h2>People & invitations</h2>
      <p class="muted small">Invite runners, officials, club leaders and participants in bulk. Fighters can also sign up themselves.</p>
      <label class="field" style="margin-top:18px;">Bulk invites — one per line: <code>email, phone, role, club</code>
        <textarea rows="6" bind:value={bulkInvites} placeholder="judge@club.com, +447700900001, official, Riverside&#10;fighter@club.com, , fighter, Hilltop"></textarea>
      </label>
      <button class="btn primary" style="margin-top:12px;" on:click={parseBulkInvites}>Add invitations</button>
      <h3 style="margin-top:22px;">Invited ({wizard.invites.length})</h3>
      <div style="display:grid; gap:8px; margin-top:10px;">
        {#each wizard.invites as invite, i (invite.id)}
          <div class="rule-card" style="display:flex; gap:10px; align-items:center; flex-wrap:wrap;">
            <strong>{invite.email}</strong><span class="muted small">{invite.phone || 'no phone'}</span>
            <span class="badge {invite.role === 'official' ? 'official' : invite.role === 'club_leader' ? 'fighter' : 'spectator'}">{invite.role}</span>
            <span class="muted small">{invite.clubName || 'independent'}</span>
            <button class="tab" style="margin-left:auto;" on:click={() => { wizard.invites = wizard.invites.filter((_, idx) => idx !== i); touch(); notifySuccess('Invite removed', () => { const list = [...wizard.invites]; list.splice(i, 0, invite); wizard.invites = list; touch(); }); }}>Remove</button>
          </div>
        {:else}<span class="muted small">No invitations yet.</span>{/each}
      </div>
    {:else}
      <div class="eyebrow">Step 5 of 5</div><h2>Review & launch</h2>
      {#if errors.length > 0}
        <div class="warning-panel">
          <strong>{errors.length} thing{errors.length === 1 ? '' : 's'} need attention before launch:</strong>
          {#each errors as error}<button class="issue" on:click={() => go(error.step)}>⚠ {error.message} <span>Go to step {error.step + 1} →</span></button>{/each}
        </div>
      {:else}
        <div class="success-panel">✅ Everything checks out. The tournament is ready to create.</div>
      {/if}
      <div class="grid" style="grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); margin-top:18px;">
        <div class="stat"><div class="value">{wizard.name || '—'}</div><div class="label">Tournament</div></div>
        <div class="stat"><div class="value">{wizard.disciplines.length}</div><div class="label">Disciplines</div></div>
        <div class="stat"><div class="value">{wizard.categories.length}</div><div class="label">Categories</div></div>
        <div class="stat"><div class="value">{wizard.invites.length}</div><div class="label">Invitations</div></div>
        <div class="stat"><div class="value">{wizard.constraints.cap}</div><div class="label">Participant cap</div></div>
        <div class="stat"><div class="value">{wizard.constraints.maxPools}</div><div class="label">Max pools</div></div>
      </div>
      <div class="capabilities" style="margin-top:18px;">
        <label class="cap" class:on={wizard.saveTemplate}><input type="checkbox" bind:checked={wizard.saveTemplate} /> Save this setup as a reusable template</label>
      </div>
      {#if wizard.saveTemplate}<label class="field" style="margin-top:12px;">Template name<input bind:value={wizard.templateName} placeholder="{wizard.name || 'My tournament'} template" /></label>{/if}
    {/if}

    <div class="wizard-actions">
      <button class="btn" disabled={wizard.step === 0} on:click={() => go(wizard.step - 1)}>← Back</button>
      <span class="muted small">Draft saved automatically</span>
      {#if wizard.step < 4}
        <button class="btn primary" on:click={() => go(wizard.step + 1)}>Continue →</button>
      {:else}
        <button class="btn primary" on:click={launch}>🚀 Launch tournament</button>
      {/if}
    </div>
  </div>
</div>

<style>
  .stepper { display: flex; gap: 8px; flex-wrap: wrap; margin-bottom: 18px; }
  .step { display:flex; align-items:center; gap:8px; padding:9px 14px; border-radius:999px; border:1px solid var(--brand-border); background:var(--brand-surface-2); color:var(--brand-muted); cursor:pointer; }
  .step .num { width:22px; height:22px; border-radius:50%; display:grid; place-items:center; background:var(--brand-surface-3); font-size:.75rem; font-weight:800; }
  .step.active { color:var(--brand-ink); border-color:var(--brand-primary); }
  .step.active .num { background:var(--brand-primary); color:#fff; }
  .step.done .num { background:var(--win); color:#fff; }
  .setup-card { min-height: 480px; }
  :global(.field) { display:grid; gap:5px; font-size:.8rem; color:var(--brand-muted); }
  :global(.field input), :global(.field select), :global(.field textarea) { padding:9px 11px; border-radius:9px; border:1px solid var(--brand-border); background:var(--brand-surface-3); color:var(--brand-ink); font-size:.9rem; width:100%; }
  .rule-card { padding:14px; border-radius:12px; border:1px solid var(--brand-border); background:var(--brand-surface-3); }
  .wizard-actions { display:flex; align-items:center; gap:14px; margin-top:26px; padding-top:18px; border-top:1px solid var(--brand-border); }
  .wizard-actions .btn.primary { margin-left:auto; }
  .warning-panel { border:1px solid rgba(245,158,11,.5); background:rgba(245,158,11,.08); border-radius:12px; padding:16px; display:grid; gap:8px; }
  .success-panel { border:1px solid rgba(34,197,94,.5); background:rgba(34,197,94,.08); border-radius:12px; padding:16px; }
  .issue { display:flex; justify-content:space-between; gap:10px; text-align:left; background:transparent; border:none; color:var(--brand-ink); cursor:pointer; font-size:.9rem; padding:6px 0; }
  .issue span { color:var(--brand-accent); }
</style>
