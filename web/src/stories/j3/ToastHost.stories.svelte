<script module>
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import ToastHost from '../../lib/components/ToastHost.svelte';
  import { notifySuccess, notifyError } from '../../lib/toast';
  const { Story } = defineMeta({ title: 'J3/ToastHost', component: ToastHost, tags: ['autodocs'] });
</script>
<Story name="Default">
  <button on:click={() => notifySuccess('Saved', () => {})}>Success + undo</button>
  <button on:click={() => notifyError('Failed')}>Error</button>
  <ToastHost />
</Story>
