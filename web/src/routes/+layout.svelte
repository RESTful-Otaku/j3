<script lang="ts">
  import { onMount } from 'svelte';
  import '../app.css';
  import ToastHost from '../lib/components/ToastHost.svelte';
  import NavShell from '../lib/components/NavShell.svelte';
  import TourOverlay from '../lib/components/TourOverlay.svelte';
  import { afterNavigate, goto } from '$app/navigation';
  import { ensureQaFirstRun, syncRoleFromSession, uiRole } from '../lib/auth';
  import { setRole } from '../lib/nav';
  import { apiUrl } from '../lib/api';
  import { notifySuccess } from '../lib/toast';
  import { canAccessPath, currentRole, isPhoneWidth, phoneHome } from '../lib/nav';
  import { labelTableRows } from '../lib/responsive';
  import { loadEventBrand } from '../lib/event-brand';
  import { initTheme } from '../lib/theme';

  let mobile = false;
  let phone = false;
  let ready = false;
  let online = true;
  let docsRoute = false;

  function syncRoute() {
    const path = window.location.pathname;
    docsRoute = path.startsWith('/docs') || path === '/apidocs';
  }

  function guard() {
    if (typeof window === 'undefined') return;
    syncRoleFromSession();
    const role = currentRole();
    const path = window.location.pathname;
    if (phone && path === '/') {
      void goto(phoneHome(role));
      return;
    }
    if (!canAccessPath(role, path, phone)) {
      void goto(phone ? phoneHome(role) : '/');
    }
  }

  afterNavigate(() => { syncRoute(); labelTableRows(); guard(); });
  onMount(() => {
    const sync = () => {
      mobile = window.innerWidth <= 820;
      phone = isPhoneWidth(window.innerWidth);
    };
    sync();
    syncRoute();

    const onResize = () => { sync(); guard(); };
    const onPasswordRequired = () => { void goto('/login?password=1'); };
    window.addEventListener('resize', onResize);
    window.addEventListener('j3-session-changed', guard);
    window.addEventListener('j3-password-change-required', onPasswordRequired);

    initTheme();
    void loadEventBrand();
    // QA mobile builds start with seeded data and an admin identity. Hold the
    // first render until that is ready so nobody sees a transient logged-out
    // screen on a phone handed to a demo audience.
    void ensureQaFirstRun().then((qaSession) => {
      if (qaSession) {
        setRole(uiRole(qaSession.role));
        notifySuccess('QA demo ready · signed in as Admin');
      }
      guard();
      ready = true;
    }).catch(() => {
      guard();
      ready = true;
    });
    const healthTimer = setInterval(async () => {
      try {
        const response = await fetch(apiUrl('/healthz'), { cache: 'no-store' });
        online = response.ok;
      } catch {
        online = false;
      }
    }, 10000);
    const labelTimer = setInterval(() => labelTableRows(), 1500);
    return () => {
      window.removeEventListener('resize', onResize);
      window.removeEventListener('j3-session-changed', guard);
      window.removeEventListener('j3-password-change-required', onPasswordRequired);
      clearInterval(healthTimer);
      clearInterval(labelTimer);
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
  .app-content.mobile { padding-top:78px; padding-bottom:72px; }
  .link-status { position:fixed; right:12px; bottom:12px; z-index:700; display:flex; align-items:center; gap:6px; padding:6px 10px; border-radius:999px; background:var(--brand-surface-2); border:1px solid var(--brand-border); color:var(--brand-muted); font-size:.72rem; }
  .link-status span { width:8px; height:8px; border-radius:50%; background:var(--win); }
  .link-status span.down { background:var(--loss); }
</style>
