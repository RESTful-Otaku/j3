<script lang="ts">
  import { onMount } from 'svelte';
  import { api, livePoll, type DemoFight, type PoolView, broadcastPA } from '../../../lib/api';
  import { notifySuccess } from '../../../lib/toast';
  import ConfirmDialog from '../../../lib/components/ConfirmDialog.svelte';

  type Slot = { id: string; name: string; colour: string };
  type BracketNode = { id: string; label: string; a: Slot | null; b: Slot | null; winner: string };

  let fights: DemoFight[] = [];
  let pools: PoolView[] = [];
  let bracket: BracketNode[][] = [];
  let mode: 'bracket' | 'elimination' | 'roundRobin' = 'bracket';
  let pendingPromotion: DemoFight | null = null;
  let confirmOpen = false;
  let dragging: { round: number; node: number; corner: 'a' | 'b' } | null = null;

  onMount(() => {
    const stopFights = livePoll(api.fights, 2000, (data) => {
      fights = data.fights;
      if (bracket.length > 0) buildBracket();
    });
    const stopPools = livePoll(api.pools, 4000, (data) => {
      pools = data.pools;
      if (bracket.length === 0) buildBracket();
    });
    return () => { stopFights(); stopPools(); };
  });

  function toSlot(member: { id: string; name: string; clubColour: string }): Slot {
    return { id: member.id, name: member.name, colour: member.clubColour };
  }
  function winnerFor(a: Slot | null, b: Slot | null) {
    const fight = fights.find((f) =>
      (f.red.id === a?.id && f.blue.id === b?.id) || (f.red.id === b?.id && f.blue.id === a?.id));
    if (fight?.outcome === 'win_red') return a?.id ?? '';
    if (fight?.outcome === 'win_blue') return b?.id ?? '';
    if (fight?.outcome === 'draw') return 'draw';
    return a?.id ?? b?.id ?? '';
  }
  function node(id: string, label: string, a: Slot | null, b: Slot | null): BracketNode {
    return { id, label, a, b, winner: winnerFor(a, b) };
  }
  function buildBracket() {
    const first: BracketNode[] = [];
    pools.forEach((pool, poolIndex) => {
      for (let i = 0; i + 1 < pool.members.length; i += 2) {
        first.push(node(`r0-${poolIndex}-${i}`, `${pool.name} · ${i / 2 + 1}`,
          toSlot(pool.members[i]), toSlot(pool.members[i + 1])));
      }
    });
    const advance = (sources: BracketNode[]) => sources.map((source) =>
      source.winner === source.a?.id ? source.a : source.b);
    const semiWinners = advance(first);
    const semis: BracketNode[] = [];
    for (let i = 0; i + 1 < semiWinners.length && semis.length < 2; i += 2) {
      semis.push(node(`r1-${i}`, `Semi ${i / 2 + 1}`, semiWinners[i], semiWinners[i + 1]));
    }
    const finalists = advance(semis);
    const losers = semis.map((s) => s.winner === s.a?.id ? s.b : s.a);
    const last: BracketNode[] = [
      node('final', 'Final', finalists[0] ?? null, finalists[1] ?? null),
      node('third', 'Bronze decider', losers[0] ?? null, losers[1] ?? null),
    ];
    bracket = [first, semis, last];
  }
  function dropSlot(round: number, nodeIndex: number, corner: 'a' | 'b') {
    if (!dragging) return;
    const source = bracket[dragging.round]?.[dragging.node];
    const target = bracket[round]?.[nodeIndex];
    if (!source || !target) return;
    const from = dragging.corner === 'a' ? source.a : source.b;
    const to = corner === 'a' ? target.a : target.b;
    if (dragging.round === round && dragging.node === nodeIndex && dragging.corner === corner) return;
    if (dragging.corner === 'a') source.a = to; else source.b = to;
    if (corner === 'a') target.a = from; else target.b = from;
    source.winner = winnerFor(source.a, source.b);
    target.winner = winnerFor(target.a, target.b);
    bracket = [...bracket];
    dragging = null;
    notifySuccess('Bracket slot updated', () => buildBracket());
  }
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
