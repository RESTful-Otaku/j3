<script lang="ts">
  import { onMount } from 'svelte';
  import '../app.css';
  import ToastHost from '../lib/components/ToastHost.svelte';
  import NavShell from '../lib/components/NavShell.svelte';
  import TourOverlay from '../lib/components/TourOverlay.svelte';
  import { goto } from '$app/navigation';
  import { currentRole, itemsForRole } from '../lib/nav';

  let mobile = false;
  onMount(() => {
    const sync = () => (mobile = window.innerWidth <= 820);
    sync();
    window.addEventListener('resize', sync);

    const path = window.location.pathname;
    const open = new Set(['/', '/login', '/signup', '/welcome', '/public', '/event', '/brackets', '/tables', '/pools', '/details', '/simulate', '/view', '/ui']);
    if (!open.has(path)) {
      const allowed = itemsForRole(currentRole()).some((item) => item.href === path);
      if (!allowed) void goto('/');
    }
    return () => window.removeEventListener('resize', sync);
  });
</script>

<NavShell />
<ToastHost />
<TourOverlay />
<div class:mobile class="app-content"><slot /></div>

<style>
  .app-content { margin-left:228px; transition:margin-left .2s ease; min-height:100vh; }
  .app-content.mobile { margin-left:0; padding-top:60px; }
</style>
