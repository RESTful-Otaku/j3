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
  let saving = false;
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
    saving = true;
    try {
      await api.signup(form);
      submitted = true;
      errorMessage = '';
      notifySuccess('Registration received');
    } catch (error) {
      errorMessage = error instanceof Error ? error.message : 'Registration failed';
      notifyError(errorMessage);
    } finally {
      saving = false;
    }
  }
</script>

<svelte:head><title>J3 · Register</title></svelte:head>

<div class="shell">
  <header class="topbar"><div class="brand"><a class="brand-mark" href="/welcome" aria-label="Welcome">J3</a><div><div class="brand-name">Participant registration</div><div class="brand-sub">Create your fighter profile</div></div></div>
    <a class="btn" href="/login">Already registered? Sign in</a></header>

  {#if errorMessage}
    <div class="card" style="border-color:var(--loss); margin-bottom:16px;">
      <span class="muted small" style="color:var(--loss);">{errorMessage}</span>
    </div>
  {/if}

  {#if submitted}
    <div class="card success">
      <h2>✅ You&rsquo;re registered</h2>
      <p class="muted">Welcome, {form.name}. Your category suggestions are below. Sign in to see your schedule and mat calls.</p>
      <a class="btn primary" href="/login">Sign in</a>
    </div>
  {/if}

  <form class="signup-form" on:submit|preventDefault={submit}>
    <section class="card">
      <h2>Your details</h2>
      <div class="fields">
        <label>Full name<input bind:value={form.name} required /></label>
        <label>Email<input type="email" bind:value={form.email} required /></label>
        <label>Password (8+ characters)<input type="password" bind:value={form.password} required /></label>
        <label>Phone<input type="tel" bind:value={form.phone} /></label>
        <label>Date of birth<input type="date" bind:value={form.dateOfBirth} required /></label>
        <label>Unit system
          <select bind:value={form.unitSystem}>
            <option value="metric">Metric</option>
            <option value="imperial">Imperial</option>
          </select>
        </label>
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
      <div class="action-row" style="margin-top:18px;">
        <button class="btn" on:click={preview} type="button">Check my divisions</button>
        <button class="btn primary" type="submit" disabled={saving}>{saving ? 'Registering…' : 'Complete registration'}</button>
      </div>
    </section>

    {#if eligible.length > 0}
      <section class="card" style="margin-top:14px;">
        <h2>Your suggested divisions</h2>
        <p class="muted small" style="margin:6px 0 10px;">Based on your age, weight, grade and club.</p>
        <div style="display:grid; gap:8px;">
          {#each eligible as item (item.categoryId)}
            <label class="option" class:full={item.full}>
              <input type="checkbox" value={item.categoryId} bind:group={form.categoryIds} disabled={item.full} />
              <span><strong>{item.name}</strong><br /><span class="muted small">{item.reason}{item.moveUp ? ' · promotion' : ''}{item.full ? ' · full' : ''}</span></span>
            </label>
          {/each}
        </div>
      </section>
    {:else if !submitted}
      <section class="card placeholder" style="margin-top:14px;">
        <p class="muted small">Press <strong>Check my divisions</strong> to see your eligible divisions.</p>
      </section>
    {/if}
  </form>
</div>

<style>
  .fields { display:grid; grid-template-columns:repeat(auto-fit,minmax(160px,1fr)); gap:8px; margin-top:12px; }
  .fields label { display:grid; gap:4px; font-size:.78rem; color:var(--brand-muted); margin:0; }
  .fields input, .fields select { padding:10px 12px; border-radius:10px; background:var(--brand-surface-3); color:var(--brand-ink); border:1px solid var(--brand-border); font-size:1rem; }
  .check { display:flex; gap:8px; align-items:center; margin-top:12px; font-size:.86rem; margin:0; }
  .option { display:flex; gap:10px; align-items:center; padding:10px; border-radius:10px; background:var(--brand-surface-3); border:1px solid var(--brand-border); cursor:pointer; }
  .option.full { opacity:.5; cursor:not-allowed; }
  .success { border-color:var(--win); margin-bottom:18px; }
  .guardian { margin-top:12px; padding:12px; border-radius:12px; border:1px solid var(--brand-accent); background:rgba(245,158,11,.08); }
  .placeholder { text-align:center; padding:16px; border-radius:10px; background:var(--brand-surface-2); }
  @media (max-width: 700px) {
    .signup-form .card { border-radius: 14px; padding: 14px; }
    .fields { grid-template-columns: 1fr; gap: 8px; }
    .btn { min-height: 46px; width: 100%; justify-content: center; font-size: 1rem; font-weight: 800; }
    .action-row { display: flex; gap: 8px; flex-direction: column; }
    .action-row .btn { width: 100%; }
  }
</style>
