<script lang="ts">
  import { goto } from '$app/navigation';
  import { login, passkey, verify, session } from '../../lib/auth';
  import { notifyError, notifySuccess } from '../../lib/toast';

  let email = 'tony@okami.j3';
  let password = 'demo';
  let step: 'method' | 'mfa' = 'method';
  let demoCode = '';
  let code = '';
  let busy = false;

  if (session()) goto('/');

  async function start(provider: string) {
    if (!email.trim()) { notifyError('Enter your email first'); return; }
    busy = true;
    try {
      const result = provider === 'passkey' ? await passkey(email) : await login(email, provider, password);
      demoCode = result.demoCode;
      code = result.demoCode;
      step = 'mfa';
      notifySuccess('MFA code sent');
    } catch (error) {
      notifyError(error instanceof Error ? error.message : 'Sign-in failed');
    } finally { busy = false; }
  }

  async function confirm() {
    busy = true;
    try {
      const result = await verify(email, code);
      notifySuccess(`Welcome, ${result.email}`);
      await goto('/');
    } catch (error) {
      notifyError(error instanceof Error ? error.message : 'Verification failed');
    } finally { busy = false; }
  }

  function logout() { localStorage.removeItem('j3-session'); notifySuccess('Signed out'); }
</script>

<svelte:head><title>J3 · Sign in</title></svelte:head>

<div class="auth-shell">
  <div class="auth-card">
    <div class="brand"><span class="brand-mark">J3</span><div><h1>J Cubed</h1><p class="muted small">Tournament creator & operator</p></div></div>

    {#if step === 'method'}
      <label class="field">Email<input type="email" bind:value={email} /></label>
      <label class="field">Password<input type="password" bind:value={password} /></label>
      <button class="btn primary wide" disabled={busy} on:click={() => start('password')}>Sign in with password</button>
      <div class="divider"><span>or continue with</span></div>
      <div class="providers">
        <button class="btn provider" disabled={busy} on:click={() => start('google')}><span class="g">G</span> Google</button>
        <button class="btn provider" disabled={busy} on:click={() => start('apple')}><span class="a"></span> Apple</button>
      </div>
      <button class="btn wide" disabled={busy} on:click={() => start('passkey')}>🔑 Passkey / biometrics</button>
      <p class="muted small">Demo mode: Google/Apple/passkey issue an MFA code shown on screen so the full flow can be tested without external credentials.</p>
    {:else}
      <h2>Two-factor verification</h2>
      <p class="muted small">Enter the 6-digit code for {email}.</p>
      {#if demoCode}<div class="demo-code">Demo code: <strong>{demoCode}</strong></div>{/if}
      <label class="field">6-digit code<input inputmode="numeric" maxlength="6" bind:value={code} /></label>
      <button class="btn primary wide" disabled={busy || code.length !== 6} on:click={confirm}>Verify & sign in</button>
      <button class="btn wide" on:click={() => (step = 'method')}>← Back</button>
    {/if}
    {#if session()}<button class="btn wide" on:click={logout}>Sign out</button>{/if}
  </div>
</div>

<style>
  .auth-shell { min-height:100vh; display:grid; place-items:center; padding:20px; }
  .auth-card { width:min(420px, 100%); padding:28px; border-radius:20px; background:var(--brand-surface-2); border:1px solid var(--brand-border); box-shadow:var(--brand-shadow); display:grid; gap:12px; }
  .brand { display:flex; gap:12px; align-items:center; margin-bottom:10px; }
  h1 { font-size:1.5rem; }
  .field { display:grid; gap:5px; font-size:.82rem; color:var(--brand-muted); }
  .field input { padding:11px; border-radius:10px; background:var(--brand-surface-3); border:1px solid var(--brand-border); color:var(--brand-ink); }
  .wide { width:100%; justify-content:center; }
  .providers { display:grid; grid-template-columns:1fr 1fr; gap:10px; }
  .provider { justify-content:center; }
  .g { color:#4285f4; font-weight:800; font-size:1.2rem; }
  .a { width:14px; height:16px; background:var(--brand-ink); clip-path:polygon(50% 0,65% 24%,100% 24%,78% 58%,88% 100%,50% 78%,12% 100%,22% 58%,0 24%,35% 24%); }
  .divider { display:flex; align-items:center; gap:10px; color:var(--brand-muted); font-size:.76rem; }
  .divider::before, .divider::after { content:''; flex:1; height:1px; background:var(--brand-border); }
  .demo-code { padding:10px; border-radius:9px; background:rgba(245,158,11,.12); font-size:.9rem; }
</style>
