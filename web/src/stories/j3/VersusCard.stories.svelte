<script module>
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import VersusCard from '../../lib/components/VersusCard.svelte';
  const { Story } = defineMeta({ title: 'J3/VersusCard', component: VersusCard, tags: ['autodocs'] });
</script>
<Story name="Shrunk" args={{ event: 'Okami 2027', category: 'Adult -76 kg', mat: 'Mat 1', redScore: 1, blueScore: 0, state: 'Round 1 · live', expanded: false,
  red: { name: 'Mateo Diaz', avatarPath: '/media/fighter_1.jpg', club: 'Okami', clubLogo: '/media/club_logo_1.jpg', clubColour: '#b91c1c', belt: 'Blue', weightKg: 74 },
  blue: { name: 'Sakura Yamamoto', avatarPath: '/media/fighter_2.jpg', club: 'Skegness', clubLogo: '/media/club_logo_2.jpg', clubColour: '#2563eb', belt: 'Green', weightKg: 63 } }} />
<Story name="Expanded" args={{ event: 'Okami 2027', category: 'Adult -76 kg', mat: 'Mat 1', redScore: 1, blueScore: 0, expanded: true,
  red: { name: 'Mateo Diaz', avatarPath: '/media/fighter_1.jpg', club: 'Okami', clubLogo: '/media/club_logo_1.jpg', clubColour: '#b91c1c', belt: 'Blue', weightKg: 74 },
  blue: { name: 'Sakura Yamamoto', avatarPath: '/media/fighter_2.jpg', club: 'Skegness', clubLogo: '/media/club_logo_2.jpg', clubColour: '#2563eb', belt: 'Green', weightKg: 63 } }} />
