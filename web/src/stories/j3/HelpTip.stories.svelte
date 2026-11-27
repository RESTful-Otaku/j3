<script module>
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import HelpTip from '../../lib/components/HelpTip.svelte';
  const { Story } = defineMeta({ title: 'J3/HelpTip', component: HelpTip, tags: ['autodocs'] });
</script>
<Story name="Default" args={{ text: 'Calendar-year age policy.' }} />
