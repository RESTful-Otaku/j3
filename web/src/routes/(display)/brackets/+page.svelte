<script lang="ts">
  import { onMount } from 'svelte';
  import { api, livePoll, type DemoFight, broadcastPA } from '../../../lib/api';
  import ConfirmDialog from '../../../lib/components/ConfirmDialog.svelte';

  let fights: DemoFight[] = [];
  let mode: 'elimination' | 'roundRobin' = 'elimination';
  let pendingPromotion: DemoFight | null = null;
  let confirmOpen = false;

  onMount(() => livePoll(api.fights, 2000, (data) => (fights = data.fights)));

  function initials(name: string) { return name.split(' ').slice(0, 2).map((p) => p[0] ?? '').join('').toUpperCase(); }
  function outcome(fight: DemoFight, corner: 'red' | 'blue') {
    if (!fight.outcome) return { code: '', win: false };
    if (fight.outcome === 'draw') return { code: 'DP', win: false };
    const winner = fight.outcome === 'win_red' ? 'red' : 'blue';
    const win = corner === winner;
    const submission = fight.events.some((e) => e.kind === 'submission');
    const technical = fight.outcome.startsWith('dq') || fight.state === 'forfeit';
    return { code: `${win ? 'W' : 'L'}${submission ? 'S' : technical ? 'T' : 'P'}`, win };
  }
  function cards(fight: DemoFight, corner: 'red' | 'blue') {
    return fight.events.filter((e) => e.kind === 'penalty' && String((e.payload as Record<string, unknown>)?.corner) === corner)
      .map((e) => String((e.payload as Record<string, unknown>)?.card ?? ''));
  }
  function roundGroups() {
    const groups: Record<string, DemoFight[]> = {};
    for (const fight of fights) {
      const key = fight.round || 'Round 1';
      (groups[key] ??= []).push(fight);
    }
    return Object.entries(groups).sort(([a], [b]) => a.localeCompare(b));
  }
  function promote(fight: DemoFight) {
    pendingPromotion = fight;
    confirmOpen = true;
  }
  function confirmPromotion() {
    if (pendingPromotion) {
      broadcastPA(`${pendingPromotion.red.name} advances from ${pendingPromotion.round}`, pendingPromotion.red.id, 5);
    }
  }

  $: groups = roundGroups();
</script>

<svelte:head><title>J3 · Brackets</title></svelte:head>

<div class="display-shell">
  <header class="display-header">
    <div><div class="event-mark"><span class="crest">J3</span>Brackets</div><div class="meta">Progression, results and history · live</div></div>
    <div class="tabs">
      <button class="tab" class:active={mode === 'elimination'} on:click={() => (mode = 'elimination')}>Elimination bracket</button>
      <button class="tab" class:active={mode === 'roundRobin'} on:click={() => (mode = 'roundRobin')}>Round robin table</button>
    </div>
  </header>

  {#if mode === 'elimination'}
    <div style="display:flex; gap:clamp(14px,2vw,40px); overflow-x:auto; padding-bottom:20px; align-items:flex-start;">
      {#each groups as [round, roundFights] (round)}
        <section style="min-width:clamp(280px,24vw,420px); flex:1;">
          <div class="eyebrow">{round}</div>
          <div style="display:grid; gap:14px; margin-top:12px;">
            {#each roundFights as fight (fight.id)}
              <div class="card slide-in" style="padding:14px;">
                <div class="small muted">{fight.mat} · {fight.category}</div>
                {#each [{ name: fight.red.name, colour: fight.red.clubColour, id: fight.red.id }, { name: fight.blue.name, colour: fight.blue.clubColour, id: fight.blue.id }] as fighter, i}
                  {@const corner = i === 0 ? 'red' : 'blue'}
                  {@const result = outcome(fight, corner)}
                  <div style="display:flex; align-items:center; gap:10px; padding:8px 0; border-top:1px solid var(--brand-border);">
                    <span class="portrait small-portrait" style="background:{fighter.colour};">{initials(fighter.name)}</span>
                    <span style="flex:1; font-weight:600;">{fighter.name || 'TBD'}</span>
                    {#each cards(fight, corner) as card}<span class="badge {card === 'red' ? 'official' : 'runner'}">{card === 'red' ? '🟥' : '🟨'}</span>{/each}
                    {#if result.code}<span class="badge {result.win ? 'fighter' : 'spectator'}">{result.code}</span>{/if}
                  </div>
                {/each}
                {#if fight.state !== 'complete'}
                  <button class="btn" style="width:100%; margin-top:8px;" on:click={() => promote(fight)}>Promote winner →</button>
                {:else}
                  <div class="small muted" style="margin-top:8px;">Result committed · history retained at this round</div>
                {/if}
              </div>
            {/each}
          </div>
        </section>
      {:else}
        <div class="placeholder">No brackets generated yet.</div>
      {/each}
    </div>
  {:else}
    <div class="card" style="overflow-x:auto;">
      <h2>Round robin — every pairing and its result</h2>
      <table style="width:100%; border-collapse:collapse; margin-top:12px;">
        <thead><tr class="muted small" style="text-align:left;"><th style="padding:8px;">Round</th><th>Mat</th><th>Red</th><th>Blue</th><th>Result</th><th>State</th></tr></thead>
        <tbody>
          {#each fights as fight (fight.id)}
            <tr style="border-top:1px solid var(--brand-border);">
              <td style="padding:8px;">{fight.round}</td><td>{fight.mat}</td>
              <td>{fight.red.name} <span class="badge fighter">{outcome(fight, 'red').code}</span></td>
              <td>{fight.blue.name} <span class="badge spectator">{outcome(fight, 'blue').code}</span></td>
              <td>{fight.outcome || '—'}</td><td>{fight.state}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

<ConfirmDialog
  open={confirmOpen}
  title="Promote winner?"
  message="This records the winner into the next round and keeps a copy in this round's history."
  confirmLabel="Promote & announce"
  onConfirm={confirmPromotion}
/>

<style>
  .small-portrait { width: 42px; height: 42px; border-radius: 12px; display: grid; place-items: center; color: #fff; font-weight: 800; font-size: .85rem; flex: none; }
</style>
