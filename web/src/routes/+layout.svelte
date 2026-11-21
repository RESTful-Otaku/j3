<script lang="ts">
  import '../app.css';
  import ToastHost from '../lib/components/ToastHost.svelte';
  import NavShell from '../lib/components/NavShell.svelte';
</script>

<NavShell />
<ToastHost />
<div class="app-content"><slot /></div>

<style>
  .app-content { margin-left:230px; }
  @media (max-width: 700px) { .app-content { margin-left:60px; } }
</style>
