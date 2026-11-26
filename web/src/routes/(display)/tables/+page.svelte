<script lang="ts">
  import { onMount } from 'svelte';
  import { initials } from '../../../lib/format';
  import { phoneLimit, phoneRows } from '../../../lib/responsive';
  import { api, type StandingRow, type KataPerformance, type MedalRow } from '../../../lib/api';
  import FighterPortrait from '../../../lib/components/FighterPortrait.svelte';
  import { replaySnapshot } from '../../../lib/demo-replay';
  import { onLiveChange } from '../../../lib/live-updates';

  let standings: StandingRow[] = [];
  let kataPerformances: KataPerformance[] = [];
  let medals: MedalRow[] = [];
  let tab: 'overall' | 'category' | 'medals' | 'kata' = 'overall';
  let category = '';
  let query = '';
  let group = '';
  let loadError = '';

  async function load() {
    try {
      const [ranking, kataData, medalData] = await Promise.all([api.standings(), api.kata(), api.medals()]);
      standings = ranking.standings;
      kataPerformances = kataData.performances;
      medals = medalData.medals;
      loadError = '';
    } catch {
      loadError = 'Live results are temporarily unavailable. Retrying automatically…';
    }
  }
  onMount(() => {
    void load();
    const handle = setInterval(() => void load(), 3000);
    const stopLive = onLiveChange(['standings', 'kata', 'medals', 'fights'], () => void load());
    return () => { clearInterval(handle); stopLive(); };
  });

  async function submitScore(id: string, judge: number, value: number) {
    kataPerformances = (await api.kataScore(id, { judge, score: value })).performances;
  }
  async function advanceKata() {
    kataPerformances = (await api.advanceKata(8)).performances;
  }

  $: replay = $replaySnapshot;
  $: displayStandings = replay.active ? (replay.standings as StandingRow[]) : standings;
  $: categories = Array.from(new Set(displayStandings.map((row) => row.category))).sort();
  $: groups = Array.from(new Set(displayStandings.map((row) => row.belt))).filter(Boolean).sort();
  $: filtered = displayStandings.filter((row) =>
    (!category || row.category === category) &&
    (!group || row.belt === group) &&
    `${row.name} ${row.club}`.toLowerCase().includes(query.toLowerCase()));
  $: displayMedals = replay.active
    ? replay.medals.map((medal) => ({
        id: medal.id, fighterId: medal.fighterId, name: medal.name, club: medal.club,
        clubColour: '#b91c1c', kind: medal.kind,
        categoryId: '', category: medal.category ?? '', disciplineId: '',
        discipline: medal.discipline ?? '', eventType: medal.eventType ?? '',
        awardKey: '', source: 'replay',
      }))
    : medals;
  $: medalRows = displayMedals.filter((medal) => !category || medal.category === category).slice(0, 40);
</script>

<svelte:head><title>J3 · Tables & rankings</title></svelte:head>

<div class="display-shell">
  {#if loadError}<div class="card" style="margin-bottom:14px; border-color:var(--loss);">{loadError}</div>{/if}
  <header class="display-header" data-guide="tables-header">
    <div><div class="event-mark"><span class="crest">J3</span>Tables & rankings</div><div class="meta">{replay.active ? `Demo replay · bout ${replay.index + 1}/${replay.total}` : 'Overall · category · medals · kata · live'}</div></div>
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
      <div style="display:flex; justify-content:space-between; gap:12px; align-items:center; flex-wrap:wrap;">
        <p class="muted small">Scores are dropped high/low and averaged; totals are authoritative.</p>
        <button class="btn" on:click={advanceKata}>⏭ Advance top 8 to next round</button>
      </div>
      <div style="display:grid; gap:12px; margin-top:14px;" use:phoneLimit={4}>
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
    <div class="card">
      <h2>Kata judged competition</h2>
      <p class="muted small">Panel of seven judges · highest and lowest discarded · mean of middle five.</p>
      <table style="width:100%; border-collapse:collapse; margin-top:14px;">
        <thead><tr class="muted small" style="text-align:left;">
          <th style="padding:8px;">Competitor</th><th>Approved kata</th>
          {#each Array(7) as _, i}<th>J{i + 1}</th>{/each}<th>Drop</th><th>Total</th><th>Rank</th>
        </tr></thead>
        <tbody use:phoneRows={5}>
          {#each [...kataPerformances].sort((a, b) => b.total - a.total) as performance, i (performance.id)}
            <tr style="border-top:1px solid var(--brand-border);">
              <td style="padding:8px;"><strong>{performance.name}</strong></td>
              <td>{performance.kataName || 'Open kata'}</td>
              {#each Array(performance.panelSize) as _, j}
                <td>{performance.scores.find((score) => score.judge === j + 1)?.score.toFixed(1) ?? '—'}</td>
              {/each}
              <td class="muted">{performance.panelSize >= 5 ? 'hi + lo' : 'all'}</td>
              <td><strong>{performance.total.toFixed(2)}</strong></td>
              <td>{i + 1}</td>
            </tr>
          {:else}
            <tr><td colspan="11" class="muted">No kata performances entered yet.</td></tr>
          {/each}
        </tbody>
      </table>
    </div>
  {:else if tab === 'medals'}
    <div class="grid" use:phoneLimit={6} style="grid-template-columns: repeat(auto-fit, minmax(min(100%,300px), 1fr));">
      {#each medalRows as row (row.id)}
        <div class="stat slide-in" style="border-left:5px solid {row.kind === 'gold' ? '#fbbf24' : row.kind === 'silver' ? '#cbd5e1' : '#d97706'};">
          <div class="value">{row.kind === 'gold' ? '🥇' : row.kind === 'silver' ? '🥈' : '🥉'}</div>
          <div><strong>{row.name}</strong></div>
          <div class="small muted">{row.club} · {row.category} · {row.discipline || row.eventType}</div>
          <div class="small" style="color:var(--brand-primary); text-transform:capitalize;">{row.kind} · {row.eventType.replaceAll('_', ' ')}</div>
        </div>
      {:else}
        <div class="card muted">No medals awarded yet.</div>
      {/each}
    </div>
  {:else}
    <div class="card">
      <h2>{tab === 'overall' ? 'Overall standings' : 'Category standings'}</h2>
      <table style="width:100%; border-collapse:collapse; margin-top:14px; font-size:clamp(.9rem,1.2vw,1.3rem);">
        <thead><tr class="muted small" style="text-align:left;"><th style="padding:8px;">#</th><th>Fighter</th><th>Category</th><th>Grade</th><th>Club</th><th>W</th><th>D</th><th>L</th><th>Pts</th></tr></thead>
        <tbody use:phoneRows={8}>
          {#each filtered as row, i (row.fighterId)}
            <tr style="border-top:1px solid var(--brand-border);">
              <td style="padding:8px;">{i + 1}</td>
              <td><FighterPortrait src={row.avatarPath} label={row.name} size="30px" fallback={initials(row.name)} clubColour={row.clubColour} /> <strong>{row.name}</strong></td>
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
</style>
