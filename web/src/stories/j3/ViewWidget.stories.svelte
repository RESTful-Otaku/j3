<script module>
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import ViewWidget from '../../lib/components/ViewWidget.svelte';
  const { Story } = defineMeta({ title: 'J3/ViewWidget', component: ViewWidget, tags: ['autodocs'] });
</script>
<Story name="Clock" args={{ type: 'clock', title: 'Venue time', span: 1 }} />
<Story name="Standings" args={{ type: 'standings', title: 'Leaderboard', span: 2 }} />
