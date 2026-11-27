<script module>
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import FighterPortrait from '../../lib/components/FighterPortrait.svelte';
  const { Story } = defineMeta({ title: 'J3/FighterPortrait', component: FighterPortrait, tags: ['autodocs'] });
</script>
<Story name="Photo" args={{ src: '/media/fighter_1.jpg', label: 'Mateo Diaz', size: '48px', fallback: 'MD', clubColour: '#b91c1c' }} />
<Story name="Initials" args={{ src: '', label: 'Sakura Yamamoto', size: '48px', fallback: 'SY', clubColour: '#f8fafc' }} />
