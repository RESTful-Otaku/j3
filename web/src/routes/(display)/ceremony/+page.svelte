<script lang="ts">
  import { onMount } from 'svelte';
  import { api, livePoll, broadcastPA } from '../../../lib/api';
  import { replaySnapshot } from '../../../lib/demo-replay';
  import { onLiveChange } from '../../../lib/live-updates';
  type Medal = {
    id: string; fighterId: string; name: string; club: string; clubColour: string;
    kind: string; category: string; discipline: string; eventType: string; source: string;
  };
  let medals: Medal[] = [];
  let step = 0;
  let divisionIndex = 0;

  onMount(() => {
    const load = () => api.medals().then((data) => (medals = data.medals as Medal[])).catch(() => {});
    void load();
    const stopPoll = livePoll(api.medals, 4000, (data) => (medals = data.medals as Medal[]));
    const stopLive = onLiveChange(['medals', 'fights', 'kata'], () => void load());
    return () => { stopPoll(); stopLive(); };
  });
  const podium = ['gold', 'silver', 'bronze'];
  $: replay = $replaySnapshot;
  $: displayMedals = replay.active
    ? replay.medals.map((medal) => ({
        id: medal.id, fighterId: medal.fighterId, name: medal.name,
        club: medal.club || 'Independent', clubColour: '#b91c1c', kind: medal.kind,
        category: medal.category || '', discipline: medal.discipline || medal.eventType || '',
        eventType: medal.eventType || '', source: 'replay',
      }))
    : medals;
  $: divisions = Array.from(new Map(displayMedals.map((medal) => {
    const key = `${medal.category}|${medal.discipline}`;
    return [key, { key, label: [medal.discipline, medal.category].filter(Boolean).join(' · ') || 'Overall' }];
  })).values()).sort((a, b) => a.label.localeCompare(b.label));
  $: activeDivision = divisions[divisionIndex % Math.max(1, divisions.length)] ?? null;
  $: divisionMedals = activeDivision
    ? displayMedals.filter((medal) => `${medal.category}|${medal.discipline}` === activeDivision.key)
    : displayMedals;
  $: grouped = podium.map((kind) => {
    const seen = new Set<string>();
    const winners = divisionMedals.filter((medal) => {
      if (medal.kind !== kind || seen.has(medal.fighterId)) return false;
      seen.add(medal.fighterId);
      return true;
    });
    return { kind, winners };
  });
  $: activeKind = podium[step];
  function call() { broadcastPA('Medal ceremony at the podium — competitors please report', '', 8); }
  function next() {
    const kind = podium[step];
    const winners = grouped.find((group) => group.kind === kind)?.winners ?? [];
    if (step < podium.length - 1) {
      step += 1;
    } else {
      step = 0;
      divisionIndex = (divisionIndex + 1) % Math.max(1, divisions.length);
    }
    broadcastPA(`${activeDivision?.label ?? 'Event'} · ${kind.toUpperCase()} medalists to the podium: ${winners.map((m) => m.name).join(', ') || 'to be announced'}`, '', 10);
  }
  function previous() { step = (step - 1 + podium.length) % podium.length; }
</script>

<svelte:head><title>J3 · Medal ceremony</title></svelte:head>

<div class="display-shell">
  <header class="display-header">
    <div><div class="event-mark"><span class="crest">🏆</span>Medal ceremony</div><div class="meta">{replay.active ? `Demo replay · bout ${replay.index + 1}/${replay.total}` : 'Podium call · live results'} · {activeDivision?.label ?? 'Overall'}</div></div>
    <div style="display:flex; gap:8px; flex-wrap:wrap;">
      <button class="btn" on:click={previous}>← Previous</button>
      <button class="btn" on:click={call}>📣 Call all</button>
      <button class="btn primary" on:click={next}>Next: {podium[(step + 1) % podium.length]} →</button>
    </div>
  </header>

  {#if divisions.length > 1}
    <div class="divisions" aria-label="Medal divisions">
      {#each divisions as division, index (division.key)}
        <button class:active={index === divisionIndex} on:click={() => { divisionIndex = index; step = 0; }}>{division.label}</button>
      {/each}
    </div>
  {/if}

  <div class="podium">
    {#each grouped as group, index}
      <div class="step {group.kind}" class:active={group.kind === activeKind} style="order:{index === 0 ? 2 : index === 1 ? 1 : 3}">
        <div class="medal-icon">{group.kind === 'gold' ? '🥇' : group.kind === 'silver' ? '🥈' : '🥉'}</div>
        <div class="eyebrow">{group.kind}</div>
        {#each group.winners as winner (winner.id)}
          <div class="winner celebrate">
            <span class="avatar" style="background:{winner.clubColour}">{winner.name.split(' ').map((p) => p[0]).join('')}</span>
            <strong>{winner.name}</strong>
            <span class="muted">{winner.club}{winner.eventType ? ` · ${winner.eventType.replaceAll('_', ' ')}` : ''}</span>
          </div>
        {:else}<span class="muted small">To be decided</span>{/each}
      </div>
    {/each}
  </div>
</div>

<style>
  .divisions { display:flex; gap:6px; flex-wrap:wrap; justify-content:center; margin-top:18px; }
  .divisions button { padding:6px 11px; border-radius:999px; border:1px solid var(--brand-border); background:var(--brand-surface-2); color:var(--brand-muted); cursor:pointer; font-size:.76rem; font-weight:700; }
  .divisions button.active { background:var(--brand-primary-fill); color:var(--brand-on-primary); border-color:transparent; }
  .podium { display:flex; gap:clamp(10px,2vw,30px); align-items:flex-end; justify-content:center; flex-wrap:wrap; margin-top:24px; }
  .step { flex:1; min-width:220px; max-width:420px; padding:20px; border-radius:18px 18px 0 0; background:var(--brand-surface-2); border:1px solid var(--brand-border); text-align:center; }
  .step.active { transform:translateY(-12px) scale(1.03); box-shadow:0 0 70px rgba(245,158,11,.45); transition:transform .3s cubic-bezier(.16,1,.3,1), box-shadow .3s; }
  .step.gold { min-height:280px; border-color:#fbbf24; box-shadow:0 0 60px rgba(251,191,36,.2); }
  .step.silver { min-height:230px; }
  .step.bronze { min-height:200px; }
  .medal-icon { font-size:3rem; }
  .winner { display:grid; gap:6px; justify-items:center; margin-top:14px; }
  .winner .avatar { width:60px; height:60px; border-radius:18px; display:grid; place-items:center; color:#fff; font-weight:800; }
</style>
