<script lang="ts">
  import { onMount } from 'svelte';
  import { api, livePoll, type StandingRow, type DemoFight, type KataPerformance } from '../../../lib/api';

  let standings: StandingRow[] = [];
  let fights: DemoFight[] = [];
  let kataPerformances: KataPerformance[] = [];
  let tab: 'overall' | 'category' | 'medals' | 'kata' = 'overall';
  let category = '';
  let query = '';
  let group = '';

  onMount(() => {
    const stopStandings = livePoll(api.standings, 3000, (data) => (standings = data.standings));
    const stopFights = livePoll(api.fights, 3000, (data) => (fights = data.fights));
    const stopKata = livePoll(api.kata, 5000, (data) => (kataPerformances = data.performances));
    return () => { stopStandings(); stopFights(); stopKata(); };
  });

  async function submitScore(id: string, judge: number, value: number) {
    kataPerformances = (await api.kataScore(id, { judge, score: value })).performances;
  }
  function initials(name: string) { return name.split(' ').slice(0, 2).map((p) => p[0] ?? '').join('').toUpperCase(); }
  function kataScores(index: number) {
    const base = 7.2 + ((index * 7) % 27) / 10;
    return Array.from({ length: 7 }, (_, i) => Number((base + ((i * 3 + index) % 9) / 10).toFixed(1)));
  }
  function kataTotal(index: number) {
    const scores = kataScores(index).sort((a, b) => a - b).slice(1, -1);
    return (scores.reduce((a, b) => a + b, 0) / scores.length).toFixed(2);
  }

  $: categories = Array.from(new Set(standings.map((row) => row.category))).sort();
  $: groups = Array.from(new Set(standings.map((row) => row.belt))).filter(Boolean).sort();
  $: filtered = standings.filter((row) =>
    (!category || row.category === category) &&
    (!group || row.belt === group) &&
    `${row.name} ${row.club}`.toLowerCase().includes(query.toLowerCase()));
  $: medalRows = filtered.slice().sort((a, b) => b.points - a.points).slice(0, 12);
  $: kataFights = fights.filter((f) => f.category.toLowerCase().includes('kata'));
</script>

<svelte:head><title>J3 · Tables & rankings</title></svelte:head>

<div class="display-shell">
  <header class="display-header">
    <div><div class="event-mark"><span class="crest">J3</span>Tables & rankings</div><div class="meta">Overall · category · medals · kata · live</div></div>
    <div class="tabs">
      {#each [['overall','Overall'],['category','Category'],['medals','Medals'],['kata','Kata panel']] as [id, label]}
        <button class="tab" class:active={tab === id} on:click={() => (tab = id as typeof tab)}>{label}</button>
      {/each}
    </div>
  </header>

  <div class="card" style="margin-bottom:18px; display:flex; gap:12px; flex-wrap:wrap; align-items:center;">
    <input class="tab" style="color:var(--brand-ink); cursor:text;" placeholder="Search fighter or club…" bind:value={query} />
    <select class="tab" style="color:var(--brand-ink);" bind:value={category}>
      <option value="">All categories</option>
      {#each categories as cat}<option value={cat}>{cat}</option>{/each}
    </select>
    <select class="tab" style="color:var(--brand-ink);" bind:value={group}>
      <option value="">All grades / groups</option>
      {#each groups as grade}<option value={grade}>{grade}</option>{/each}
    </select>
    <span class="muted small">Filter by age bracket, weight, height and discipline via category selection.</span>
  </div>

  {#if tab === 'kata'}
    <div class="card" style="margin-bottom:18px;">
      <h2>Panel judging</h2>
      <p class="muted small">Scores are dropped high/low and averaged; totals are authoritative.</p>
      <div style="display:grid; gap:12px; margin-top:14px;">
        {#each kataPerformances as performance (performance.id)}
          <div class="rule-card" style="padding:12px; border-radius:10px; border:1px solid var(--brand-border); background:var(--brand-surface-3);">
            <div style="display:flex; justify-content:space-between; gap:12px; flex-wrap:wrap;">
              <strong>{performance.name}</strong>
              <span class="badge {performance.status === 'complete' ? 'fighter' : 'spectator'}">R{performance.roundNumber} · {performance.kataType} · {performance.kataName || 'Open kata'} · {performance.status} · {performance.total.toFixed(2)}</span>
            </div>
            <div style="display:flex; gap:8px; flex-wrap:wrap; margin-top:10px;">
              {#each Array(performance.panelSize) as _, i}
                {@const judge = i + 1}
                {@const score = performance.scores.find((s) => s.judge === judge)?.score ?? ''}
                <label class="small">J{i + 1}
                  <input type="number" min="0" max="10" step="0.1" value={score}
                    on:change={(e) => submitScore(performance.id, judge, Number((e.target as HTMLInputElement).value))} />
                </label>
              {/each}
            </div>
          </div>
        {:else}<span class="muted small">No kata performances recorded yet.</span>{/each}
      </div>
    </div>
    <div class="card" style="overflow-x:auto;">
      <h2>Kata judged competition</h2>
      <p class="muted small">Panel of seven judges · highest and lowest discarded · mean of middle five.</p>
      <table style="width:100%; border-collapse:collapse; margin-top:14px;">
        <thead><tr class="muted small" style="text-align:left;">
          <th style="padding:8px;">Competitor</th><th>Approved kata</th>
          {#each Array(7) as _, i}<th>J{i + 1}</th>{/each}<th>Drop</th><th>Total</th><th>Rank</th>
        </tr></thead>
        <tbody>
          {#each kataFights as fight, i (fight.id)}
            {@const name = i % 2 === 0 ? fight.red.name : fight.blue.name}
            <tr style="border-top:1px solid var(--brand-border);">
              <td style="padding:8px;"><strong>{name}</strong></td>
              <td>{['Seiyunchin', 'Kanku-dai', 'Gekusai-dai', 'Pinan sono ichi'][i % 4]}</td>
              {#each kataScores(i) as score}<td>{score.toFixed(1)}</td>{/each}
              <td class="muted">hi + lo</td><td><strong>{kataTotal(i)}</strong></td>
              <td>{i + 1}</td>
            </tr>
          {:else}
            <tr><td colspan="11" class="muted">No kata performances entered yet.</td></tr>
          {/each}
        </tbody>
      </table>
    </div>
  {:else if tab === 'medals'}
    <div class="grid" style="grid-template-columns: repeat(auto-fit, minmax(min(100%,300px), 1fr));">
      {#each medalRows as row, i (row.fighterId)}
        <div class="stat slide-in" style="border-left:5px solid {i === 0 ? '#fbbf24' : i === 1 ? '#cbd5e1' : i === 2 ? '#d97706' : 'var(--brand-primary)'};">
          <div class="value">{i === 0 ? '🥇' : i === 1 ? '🥈' : i === 2 ? '🥉' : '#' + (i + 1)}</div>
          <div><strong>{row.name}</strong></div>
          <div class="small muted">{row.club} · {row.category} · {row.points} pts</div>
        </div>
      {/each}
    </div>
  {:else}
    <div class="card" style="overflow-x:auto;">
      <h2>{tab === 'overall' ? 'Overall standings' : 'Category standings'}</h2>
      <table style="width:100%; border-collapse:collapse; margin-top:14px; font-size:clamp(.9rem,1.2vw,1.3rem);">
        <thead><tr class="muted small" style="text-align:left;"><th style="padding:8px;">#</th><th>Fighter</th><th>Category</th><th>Grade</th><th>Club</th><th>W</th><th>D</th><th>L</th><th>Pts</th></tr></thead>
        <tbody>
          {#each filtered as row, i (row.fighterId)}
            <tr style="border-top:1px solid var(--brand-border);">
              <td style="padding:8px;">{i + 1}</td>
              <td><span class="portrait tiny" style="background:{row.clubColour};">{initials(row.name)}</span> <strong>{row.name}</strong></td>
              <td>{row.category}</td><td>{row.belt || '—'}</td>
              <td><span style="color:{row.clubColour};">●</span> {row.club}</td>
              <td style="color:var(--win);">{row.wins}</td><td style="color:var(--draw);">{row.draws}</td>
              <td style="color:var(--loss);">{row.losses}</td><td><strong>{row.points}</strong></td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

<style>
  .portrait.tiny { display:inline-grid; place-items:center; width:30px; height:30px; border-radius:9px; color:#fff; font-size:.72rem; font-weight:800; vertical-align:middle; margin-right:6px; }
</style>
