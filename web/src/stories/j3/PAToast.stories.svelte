<script module>
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import PAToast from '../../lib/components/PAToast.svelte';
  const { Story } = defineMeta({ title: 'J3/PAToast', component: PAToast, tags: ['autodocs'] });
</script>
<Story name="Default">
  <PAToast />
  <button on:click={() => window.dispatchEvent(new CustomEvent('j3-pa', { detail: { message: 'Mateo Diaz to Mat 1', fighterId: 'mateo' } }))}>Fire PA</button>
</Story>
