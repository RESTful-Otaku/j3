<script module>
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import Card from '../../lib/components/Card.svelte';
  const { Story } = defineMeta({ title: 'J3/Card', component: Card, tags: ['autodocs'] });
</script>
<Story name="Expanded" args={{ title: 'On the mats now', eyebrow: 'Live', expandable: true, expanded: true, preview: '1 contest' }}>
  <p class="muted small">Fighter A vs Fighter B · Mat 1 · Round 1</p>
</Story>
<Story name="Collapsed" args={{ title: 'Rules & kit', eyebrow: 'Event', expandable: true, expanded: false, preview: 'Minimal' }}>
  <p class="muted small">Hidden until expanded</p>
</Story>
<Story name="FormCard" args={{ title: 'Your details', eyebrow: 'Profile', expandable: false }}>
  <p class="muted small">Forms stay open</p>
</Story>
