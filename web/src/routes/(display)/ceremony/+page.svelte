<script lang="ts">
  import { onMount } from 'svelte';
  import { api, livePoll, broadcastPA } from '../../../lib/api';
  type Medal = { id: string; fighterId: string; name: string; club: string; clubColour: string; kind: string };
  let medals: Medal[] = [];

  onMount(() => livePoll(api.medals, 4000, (data) => (medals = data.medals as Medal[])));
  const podium = ['gold', 'silver', 'bronze'];
  $: grouped = podium.map((kind) => ({ kind, winners: medals.filter((m) => m.kind === kind) }));
  function call() { broadcastPA('Medal ceremony at the podium — competitors please report', '', 8); }
</script>

<svelte:head><title>J3 · Medal ceremony</title></svelte:head>

<div class="display-shell">
  <header class="display-header">
    <div><div class="event-mark"><span class="crest">🏆</span>Medal ceremony</div><div class="meta">Podium call · live results</div></div>
    <button class="btn primary" on:click={call}>📣 Call competitors to podium</button>
  </header>

  <div class="podium">
    {#each grouped as group, index}
      <div class="step {group.kind}" style="order:{index === 0 ? 2 : index === 1 ? 1 : 3}">
        <div class="medal-icon">{group.kind === 'gold' ? '🥇' : group.kind === 'silver' ? '🥈' : '🥉'}</div>
        <div class="eyebrow">{group.kind}</div>
        {#each group.winners as winner (winner.id)}
          <div class="winner celebrate">
            <span class="avatar" style="background:{winner.clubColour}">{winner.name.split(' ').map((p) => p[0]).join('')}</span>
            <strong>{winner.name}</strong>
            <span class="muted">{winner.club}</span>
          </div>
        {:else}<span class="muted small">To be decided</span>{/each}
      </div>
    {/each}
  </div>
</div>

<style>
  .podium { display:flex; gap:clamp(10px,2vw,30px); align-items:flex-end; justify-content:center; flex-wrap:wrap; margin-top:40px; }
  .step { flex:1; min-width:220px; max-width:420px; padding:20px; border-radius:18px 18px 0 0; background:var(--brand-surface-2); border:1px solid var(--brand-border); text-align:center; }
  .step.gold { min-height:280px; border-color:#fbbf24; box-shadow:0 0 60px rgba(251,191,36,.2); }
  .step.silver { min-height:230px; }
  .step.bronze { min-height:200px; }
  .medal-icon { font-size:3rem; }
  .winner { display:grid; gap:6px; justify-items:center; margin-top:14px; }
  .winner .avatar { width:60px; height:60px; border-radius:18px; display:grid; place-items:center; color:#fff; font-weight:800; }
</style>
