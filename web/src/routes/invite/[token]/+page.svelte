<script lang="ts">
  import { goto } from '$app/navigation';
  import { postJson } from '../../../lib/api';
  import { notifyError, notifySuccess } from '../../../lib/toast';

  let name = '';
  let password = '';
  let confirm = '';
  let busy = false;
  let error = '';
  let done = false;
  let accountEmail = '';

  $: token = typeof window === 'undefined'
    ? ''
    : decodeURIComponent(window.location.pathname.split('/').filter(Boolean).pop() ?? '');

  async function accept() {
    error = '';
    if (password.length < 8) {
      error = 'Use at least 8 characters for your password.';
      return;
    }
    if (password !== confirm) {
      error = 'The two passwords do not match.';
      return;
    }
    busy = true;
    try {
      const result = await postJson<{ email: string; role: string; status: string }>(
        '/api/v1/invites/accept',
        { token, name, password },
      );
      accountEmail = result.email;
      done = true;
      notifySuccess('Invitation accepted — you can now sign in');
    } catch (e) {
      error = e instanceof Error ? e.message : 'This invitation could not be accepted.';
      notifyError(error);
    } finally {
      busy = false;
    }
  }
</script>

<svelte:head><title>Accept invitation · J3</title></svelte:head>

<section class="invite-card">
  <div class="mark">🔑</div>
  <p class="eyebrow">J3 tournament access</p>
  <h1>Accept your invitation</h1>

  {#if done}
    <div class="success" role="status">
      <h2>You're in</h2>
      <p>The account <strong>{accountEmail}</strong> is active. Sign in with your new password.</p>
      <button class="btn primary" on:click={() => goto(`/login?email=${encodeURIComponent(accountEmail)}`)}>
        Go to sign in
      </button>
    </div>
  {:else}
    <p class="muted">Choose the name others will see and set a password for your account.</p>
    <form on:submit|preventDefault={accept}>
      <label class="field">
        <span>Your name</span>
        <input bind:value={name} autocomplete="name" placeholder="Full name" required />
      </label>
      <label class="field">
        <span>Password</span>
        <input type="password" bind:value={password} autocomplete="new-password" minlength="8" required />
      </label>
      <label class="field">
        <span>Confirm password</span>
        <input type="password" bind:value={confirm} autocomplete="new-password" minlength="8" required />
      </label>
      {#if error}<p class="error" role="alert">{error}</p>{/if}
      <button class="btn primary wide" type="submit" disabled={busy || !token}>
        {busy ? 'Activating…' : 'Accept invitation'}
      </button>
    </form>
    <p class="muted small">Invitation links expire. Ask the person who invited you for a new one if needed.</p>
  {/if}
</section>

<style>
  .invite-card {
    max-width: 30rem;
    margin: clamp(1rem, 6vh, 4rem) auto;
    padding: clamp(1.25rem, 4vw, 2.25rem);
    border-radius: 1.25rem;
    background: var(--panel, #fff);
    border: 1px solid var(--line, #d9e2ee);
    box-shadow: 0 18px 60px rgba(10, 25, 45, 0.12);
  }
  .mark { font-size: 2.5rem; }
  .eyebrow { text-transform: uppercase; letter-spacing: .14em; font-size: .72rem; color: var(--muted, #64748b); }
  h1 { margin: .2rem 0 .6rem; font-size: clamp(1.6rem, 5vw, 2.1rem); }
  h2 { margin: 0 0 .4rem; }
  form { display: grid; gap: .9rem; margin-top: 1.25rem; }
  .field { display: grid; gap: .3rem; }
  .field span { font-weight: 600; font-size: .88rem; }
  input {
    width: 100%; padding: .8rem .9rem; border-radius: .7rem;
    border: 1px solid var(--line, #cbd5e1); background: var(--input, #fff); color: inherit;
  }
  .btn { border: 0; border-radius: .75rem; padding: .85rem 1.1rem; font-weight: 700; cursor: pointer; }
  .btn.primary { background: var(--accent, #0284c7); color: #fff; }
  .btn.wide { width: 100%; margin-top: .4rem; }
  .btn:disabled { opacity: .6; cursor: not-allowed; }
  .error, .success { margin-top: 1rem; padding: .9rem 1rem; border-radius: .8rem; }
  .error { background: #fef2f2; color: #991b1b; }
  .success { background: #ecfdf5; color: #065f46; display: grid; gap: .5rem; justify-items: start; }
  .muted { color: var(--muted, #64748b); }
  .small { font-size: .82rem; }
</style>
