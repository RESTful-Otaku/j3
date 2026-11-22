<script lang="ts">
  import { onMount } from 'svelte';
  import '../app.css';
  import ToastHost from '../lib/components/ToastHost.svelte';
  import NavShell from '../lib/components/NavShell.svelte';
  import TourOverlay from '../lib/components/TourOverlay.svelte';
  import { afterNavigate, goto } from '$app/navigation';
  import { syncRoleFromSession } from '../lib/auth';
  import { apiUrl } from '../lib/api';
  import { canAccessPath, currentRole } from '../lib/nav';

  let mobile = false;
  let ready = false;
  let online = true;
  let docsRoute = false;

  function syncRoute() {
    const path = window.location.pathname;
    docsRoute = path.startsWith('/docs') || path === '/apidocs';
  }
  afterNavigate(syncRoute);
  onMount(() => {
    const sync = () => (mobile = window.innerWidth <= 820);
    sync();
    syncRoute();

    const guard = () => {
      syncRoleFromSession();
      if (!canAccessPath(currentRole(), window.location.pathname)) void goto('/');
    };
    guard();
    window.addEventListener('resize', sync);
    window.addEventListener('j3-session-changed', guard);
    ready = true;
    const healthTimer = setInterval(async () => {
      try {
        const response = await fetch(apiUrl('/healthz'), { cache: 'no-store' });
        online = response.ok;
      } catch {
        online = false;
      }
    }, 10000);
    return () => {
      window.removeEventListener('resize', sync);
      window.removeEventListener('j3-session-changed', guard);
      clearInterval(healthTimer);
    };
  });
</script>

{#if !docsRoute}<NavShell />{/if}
<div class="link-status" aria-live="polite"><span class:down={!online}></span>{online ? 'Live link' : 'Offline'}</div>
<ToastHost />
{#if !docsRoute}<TourOverlay />{/if}
<div class:mobile class:docs={docsRoute} class="app-content">
  {#if ready}<slot />{/if}
</div>

<style>
  .app-content { margin-left:var(--nav-width, 244px); transition:margin-left .2s ease; min-height:100vh; }
  .app-content.docs { margin-left:0; }
  .app-content.mobile { padding-top:60px; padding-bottom:72px; }
  .link-status { position:fixed; right:12px; bottom:12px; z-index:700; display:flex; align-items:center; gap:6px; padding:6px 10px; border-radius:999px; background:var(--brand-surface-2); border:1px solid var(--brand-border); color:var(--brand-muted); font-size:.72rem; }
  .link-status span { width:8px; height:8px; border-radius:50%; background:var(--win); }
  .link-status span.down { background:var(--loss); }
</style>
