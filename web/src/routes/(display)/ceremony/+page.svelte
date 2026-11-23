<script lang="ts">
  import { onMount } from 'svelte';
  import { api, livePoll, broadcastPA } from '../../../lib/api';
  import { replaySnapshot } from '../../../lib/demo-replay';
  type Medal = { id: string; fighterId: string; name: string; club: string; clubColour: string; kind: string };
  let medals: Medal[] = [];
  let step = 0;

  onMount(() => livePoll(api.medals, 4000, (data) => (medals = data.medals as Medal[])));
  const podium = ['gold', 'silver', 'bronze'];
  $: replay = $replaySnapshot;
  $: displayMedals = replay.active
    ? replay.medals.map((medal) => ({
        id: medal.id, fighterId: medal.fighterId, name: medal.name,
        club: medal.club || 'Independent', clubColour: '#b91c1c', kind: medal.kind,
      }))
    : medals;
  $: grouped = podium.map((kind) => ({ kind, winners: displayMedals.filter((m) => m.kind === kind) }));
  $: activeKind = podium[step];
  function call() { broadcastPA('Medal ceremony at the podium — competitors please report', '', 8); }
  function next() {
    const nextStep = (step + 1) % podium.length;
    const kind = podium[nextStep];
    const winners = displayMedals.filter((m) => m.kind === kind);
    step = nextStep;
    broadcastPA(`${kind.toUpperCase()} medalists to the podium: ${winners.map((m) => m.name).join(', ') || 'to be announced'}`, '', 10);
  }
  function previous() { step = (step - 1 + podium.length) % podium.length; }
</script>

<svelte:head><title>J3 · Medal ceremony</title></svelte:head>

<div class="display-shell">
  <header class="display-header">
    <div><div class="event-mark"><span class="crest">🏆</span>Medal ceremony</div><div class="meta">{replay.active ? `Demo replay · bout ${replay.index + 1}/${replay.total}` : 'Podium call · live results'}</div></div>
    <div style="display:flex; gap:8px; flex-wrap:wrap;">
      <button class="btn" on:click={previous}>← Previous</button>
      <button class="btn" on:click={call}>📣 Call all</button>
      <button class="btn primary" on:click={next}>Next: {podium[(step + 1) % podium.length]} →</button>
    </div>
  </header>

  <div class="podium">
    {#each grouped as group, index}
      <div class="step {group.kind}" class:active={group.kind === activeKind} style="order:{index === 0 ? 2 : index === 1 ? 1 : 3}">
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
  .step.active { transform:translateY(-12px) scale(1.03); box-shadow:0 0 70px rgba(245,158,11,.45); transition:transform .3s cubic-bezier(.16,1,.3,1), box-shadow .3s; }
  .step.gold { min-height:280px; border-color:#fbbf24; box-shadow:0 0 60px rgba(251,191,36,.2); }
  .step.silver { min-height:230px; }
  .step.bronze { min-height:200px; }
  .medal-icon { font-size:3rem; }
  .winner { display:grid; gap:6px; justify-items:center; margin-top:14px; }
  .winner .avatar { width:60px; height:60px; border-radius:18px; display:grid; place-items:center; color:#fff; font-weight:800; }
</style>
