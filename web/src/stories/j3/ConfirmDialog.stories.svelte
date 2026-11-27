<script module>
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import ConfirmDialog from '../../lib/components/ConfirmDialog.svelte';
  const { Story } = defineMeta({ title: 'J3/ConfirmDialog', component: ConfirmDialog, tags: ['autodocs'] });
</script>
<Story name="Danger" args={{ open: true, title: 'Rebuild bracket?', message: 'Existing slots replaced.', confirmLabel: 'Rebuild', danger: true }} />
<Story name="Confirm" args={{ open: true, title: 'Announce winner?', message: 'Calls winner.', confirmLabel: 'Announce' }} />
