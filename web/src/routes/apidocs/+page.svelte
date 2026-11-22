<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';

  onMount(() => {
    void goto('/docs/api', { replaceState: true });
  });
</script>

<svelte:head><title>J3 · API explorer</title></svelte:head>

<p class="redirect">Opening the API explorer in the documentation site… <a href="/docs/api">Continue</a></p>

<style>
  .redirect { padding:24px; color:var(--brand-muted); }
  a { color:var(--brand-primary); }
</style>
