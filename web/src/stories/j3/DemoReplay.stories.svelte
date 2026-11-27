<script module>
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import DemoReplay from '../../lib/components/DemoReplay.svelte';
  const { Story } = defineMeta({ title: 'J3/DemoReplay', component: DemoReplay, tags: ['autodocs'] });
</script>
<Story name="Sim"><DemoReplay /></Story>
