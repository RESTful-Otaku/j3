<script module>
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import FighterCard from '../../lib/components/FighterCard.svelte';
  const { Story } = defineMeta({ title: 'J3/FighterCard', component: FighterCard, tags: ['autodocs'] });
</script>
<Story name="Shrunk" args={{ name: 'Mateo Diaz', avatarPath: '/media/fighter_1.jpg', club: 'Okami Jiu Jitsu', clubLogo: '/media/club_logo_1.jpg', clubColour: '#b91c1c', belt: 'Blue', event: 'Okami 2027', category: 'Adult -76 kg', mat: 'Mat 1', weightKg: 74, heightCm: 178, score: 1, expanded: false }} />
<Story name="Expanded" args={{ name: 'Sakura Yamamoto', avatarPath: '/media/fighter_2.jpg', club: 'Skegness Dojo', clubLogo: '/media/club_logo_2.jpg', clubColour: '#2563eb', belt: 'Green', event: 'Okami 2027', category: 'Adult -76 kg', mat: 'Mat 1', weightKg: 63, heightCm: 165, score: 0, expanded: true }} />
<Story name="Unaffiliated" args={{ name: 'Independent', avatarPath: '', club: 'Independent', clubLogo: '', clubColour: '#64748b', expanded: true }} />
