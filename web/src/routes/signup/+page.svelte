<script lang="ts">
  import { goto } from '$app/navigation';
  import { api, getJson, type SignupRequest, type Eligibility } from '../../lib/api';
  import { notifySuccess, notifyError } from '../../lib/toast';

  let form: SignupRequest = {
    name: '', email: '', password: '', phone: '', dateOfBirth: '2010-01-01',
    unitSystem: 'metric', heightCm: 150, weightKg: 45,
    heightFt: 5, heightIn: 0, weightLb: 100,
    beltRankId: 'br-white', customBeltLabel: '', customBeltColour: '',
    guardianName: '', guardianContact: '', guardianConsent: false,
    gender: 'open', accountType: 'fighter', clubId: '',
    categoryIds: [], consentMoveUp: false,
  };
  let belts: { id: string; label: string; colour: string }[] = [];
  let clubs: { id: string; name: string }[] = [];
  let eligible: Eligibility[] = [];
  let submitted = false;
  let errorMessage = '';
  $: isMinor = (() => {
    const birth = new Date(form.dateOfBirth);
    if (Number.isNaN(birth.getTime())) return false;
    const now = new Date();
    let age = now.getFullYear() - birth.getFullYear();
    if (now.getMonth() < birth.getMonth() || (now.getMonth() === birth.getMonth() && now.getDate() < birth.getDate())) age--;
    return age < 16;
  })();

  api.belts().then((data) => (belts = data.belts)).catch(() => {});
  getJson<{ rows: { id: string; name: string }[] }>('/api/v1/query?table=clubs&limit=50')
    .then((data) => (clubs = data.rows)).catch(() => {});

  async function preview() {
    try { eligible = (await api.signupPreview(form)).eligible; }
    catch (error) { notifyError(error instanceof Error ? error.message : 'Could not check eligibility'); }
  }
  async function submit() {
    try {
      await api.signup(form);
      submitted = true;
      errorMessage = '';
      notifySuccess('Registration received');
    } catch (error) {
      errorMessage = error instanceof Error ? error.message : 'Registration failed';
      notifyError(errorMessage);
    }
  }
</script>

<svelte:head><title>J3 · Register</title></svelte:head>

<div class="shell">
  <header class="topbar"><div class="brand"><a class="brand-mark" href="/welcome" aria-label="Welcome">J3</a><div><div class="brand-name">Participant registration</div><div class="brand-sub">Okami Jiu Jitsu Challenger</div></div></div>
    <a class="btn" href="/login">Already registered? Sign in</a></header>

  {#if errorMessage}<div class="card" style="border-color:var(--loss); margin-bottom:16px;">{errorMessage}</div>{/if}
  {#if submitted}
    <div class="card success">
      <h2>✅ You're registered</h2>
      <p class="muted">Welcome, {form.name}. Your category suggestions are below. Sign in to see your schedule and mat calls.</p>
      <a class="btn primary" href="/login">Sign in</a>
    </div>
  {/if}

  <div class="grid" style="grid-template-columns:1.1fr .9fr;">
    <div class="card">
      <h2>Your details</h2>
      <div class="fields">
        <label>Full name<input bind:value={form.name} /></label>
        <label>Email<input type="email" bind:value={form.email} /></label>
        <label>Password (8+ characters)<input type="password" bind:value={form.password} /></label>
        <label>Phone<input type="tel" bind:value={form.phone} /></label>
        <label>Date of birth<input type="date" bind:value={form.dateOfBirth} /></label>
        <label>Unit system<select bind:value={form.unitSystem}><option value="metric">Metric</option><option value="imperial">Imperial</option></select></label>
        {#if form.unitSystem === 'metric'}
          <label>Height (cm)<input type="number" bind:value={form.heightCm} /></label>
          <label>Weight (kg)<input type="number" bind:value={form.weightKg} /></label>
        {:else}
          <label>Height (ft)<input type="number" bind:value={form.heightFt} /></label>
          <label>Height (in)<input type="number" bind:value={form.heightIn} /></label>
          <label>Weight (lb)<input type="number" bind:value={form.weightLb} /></label>
        {/if}
        <label>Belt grade<select bind:value={form.beltRankId}>{#each belts as belt}<option value={belt.id}>{belt.label}</option>{/each}</select></label>
        <label>Gender<select bind:value={form.gender}><option value="open">Open</option><option value="male">Male</option><option value="female">Female</option></select></label>
        <label>Club<select bind:value={form.clubId}><option value="">Independent</option>{#each clubs as club}<option value={club.id}>{club.name}</option>{/each}</select></label>
      </div>
      {#if isMinor}
        <div class="guardian">
          <div class="eyebrow">Safeguarding · under 16</div>
          <p class="muted small">A parent or guardian must register with you and consent to your participation.</p>
          <label class="field">Guardian name<input bind:value={form.guardianName} /></label>
          <label class="field" style="margin-top:8px;">Guardian email or phone<input bind:value={form.guardianContact} /></label>
          <label class="check" style="margin-top:8px;"><input type="checkbox" bind:checked={form.guardianConsent} /> I confirm my guardian consents to this registration</label>
        </div>
      {/if}
      <label class="check"><input type="checkbox" bind:checked={form.consentMoveUp} /> Allow promotion to a heavier/older division if needed</label>
      <div class="action-row">
        <button class="btn" on:click={preview}>Check my divisions</button>
        <button class="btn primary" on:click={submit}>Complete registration</button>
      </div>
    </div>

    <div class="card">
      <div class="eyebrow">Matched from your stats</div>
      <h2>Your suggested events</h2>
      <p class="muted small">Based on your age, weight, grade and club — no need to choose from a long list.</p>
      <div style="display:grid; gap:8px; margin-top:12px;">
        {#each eligible as item (item.categoryId)}
          <label class="option" class:full={item.full}>
            <input type="checkbox" value={item.categoryId} bind:group={form.categoryIds} disabled={item.full} />
            <span><strong>{item.name}</strong><br /><span class="muted small">{item.reason}{item.moveUp ? ' · promotion' : ''}{item.full ? ' · full' : ''}</span></span>
          </label>
        {:else}<div class="placeholder">Press “Check my divisions” to see your options.</div>{/each}
      </div>
      <div class="note" style="margin-top:16px;">
        <strong>Remember:</strong> all students must attend the morning course to compete in the afternoon.
      </div>
    </div>
  </div>
</div>

<style>
  .fields { display:grid; grid-template-columns:repeat(auto-fit,minmax(180px,1fr)); gap:10px; margin-top:14px; }
  .fields label { display:grid; gap:4px; font-size:.8rem; color:var(--brand-muted); }
  .fields input, .fields select { padding:9px; border-radius:9px; background:var(--brand-surface-3); color:var(--brand-ink); border:1px solid var(--brand-border); }
  .check { display:flex; gap:8px; align-items:center; margin-top:14px; font-size:.86rem; }
  .option { display:flex; gap:10px; align-items:center; padding:10px; border-radius:10px; background:var(--brand-surface-3); border:1px solid var(--brand-border); cursor:pointer; }
  .option.full { opacity:.5; cursor:not-allowed; }
  .note { padding:12px; border-radius:10px; background:rgba(245,158,11,.1); font-size:.85rem; }
  .success { border-color:var(--win); margin-bottom:18px; }
  .guardian { margin-top:14px; padding:14px; border-radius:12px; border:1px solid var(--brand-accent); background:rgba(245,158,11,.08); }
  @media (max-width: 820px) { .grid { grid-template-columns:1fr !important; } }
</style>
