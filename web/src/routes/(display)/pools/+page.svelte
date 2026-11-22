<script lang="ts">
  import { onMount } from 'svelte';
  import { initials } from '../../../lib/format';
  import { api, livePoll, type PoolView, type DemoFight } from '../../../lib/api';
  import { notifySuccess, notifyError } from '../../../lib/toast';
  import ConfirmDialog from '../../../lib/components/ConfirmDialog.svelte';

  let pools: PoolView[] = [];
  let fights: DemoFight[] = [];
  let confirmOpen = false;
  let generating = false;

  onMount(() => {
    const stopPools = livePoll(api.pools, 4000, (data) => (pools = data.pools));
    const stopFights = livePoll(api.fights, 2000, (data) => (fights = data.fights));
    return () => { stopPools(); stopFights(); };
  });

  async function generate() {
    generating = true;
    try {
      const result = await api.generatePools({
        eventId: '', categoryId: '', targetPoolSize: 4, maxPools: 12,
        sameClubSeparation: true, seed: 42,
      });
      pools = result.pools;
      notifySuccess(`${result.pools.length} balanced pools generated`);
    } catch (error) {
      notifyError(error instanceof Error ? error.message : 'Pool generation failed');
    } finally { generating = false; }
  }
  function resultCode(fight: DemoFight) {
    if (fight.outcome === 'draw') return 'DP';
    return fight.outcome === 'win_red' ? `W ${fight.red.name}` : `W ${fight.blue.name}`;
  }

  $: live = fights.filter((f) => f.state === 'live');
  $: finished = fights.filter((f) => f.state === 'complete');
  $: upcoming = fights.filter((f) => f.state === 'pending');
</script>

<svelte:head><title>J3 · Fight pools</title></svelte:head>

<div class="display-shell">
  <header class="display-header" data-guide="pools-header">
    <div><div class="event-mark"><span class="crest">J3</span>Fight pools</div><div class="meta">Draw review · {pools.length} pools · {fights.length} contests</div></div>
    <button class="btn" on:click={() => (confirmOpen = true)} disabled={generating}>{generating ? 'Generating…' : '↻ Regenerate pools'}</button>
  </header>

  <div class="grid" style="grid-template-columns: repeat(auto-fit, minmax(min(100%,300px), 1fr));">
    {#each pools as pool (pool.id)}
      <div class="card slide-in">
        <div style="display:flex; justify-content:space-between;"><h3>{pool.name}</h3><span class="pill">{pool.members.length} fighters</span></div>
        <div style="display:grid; gap:8px; margin-top:12px;">
          {#each pool.members as member (member.id)}
            <div style="display:flex; align-items:center; gap:10px;">
              <span class="portrait tiny" style="background:{member.clubColour};">{initials(member.name)}</span>
              <span style="flex:1;"><strong>{member.name}</strong><br /><span class="muted small">{member.club || 'Independent'} · {member.belt}</span></span>
              <span class="small muted">#{member.seed}</span>
            </div>
          {/each}
        </div>
      </div>
    {:else}
      <div class="placeholder">No pools generated yet.</div>
    {/each}
  </div>

  <div class="section-title"><h2>Ongoing fights</h2><span class="pill"><span class="dot"></span>{live.length} live</span></div>
  <div class="grid" style="grid-template-columns: repeat(auto-fit, minmax(min(100%,320px), 1fr));">
    {#each live as fight (fight.id)}
      <div class="card flash" style="border-color:var(--brand-primary);">
        <div class="eyebrow">{fight.mat} · {fight.category}</div>
        <strong>{fight.red.name} vs {fight.blue.name}</strong>
      </div>
    {:else}<div class="placeholder">Nothing live right now.</div>{/each}
  </div>

  <div class="grid" style="grid-template-columns:1fr 1fr; margin-top:26px;">
    <div class="card"><h2>Finished</h2>
      {#each finished as fight (fight.id)}<div class="stat small" style="margin-top:8px;">{fight.red.name} vs {fight.blue.name} · <strong>{resultCode(fight)}</strong></div>{/each}
    </div>
    <div class="card"><h2>Upcoming</h2>
      {#each upcoming as fight (fight.id)}<div class="stat small" style="margin-top:8px;">{fight.mat} · {fight.red.name} vs {fight.blue.name}</div>{/each}
    </div>
  </div>
</div>

<ConfirmDialog
  bind:open={confirmOpen}
  title="Regenerate pools?"
  message="Existing pool assignments will be replaced. Fighters and results remain untouched."
  confirmLabel="Regenerate"
  danger={true}
  onConfirm={generate}
/>

<style>
  .portrait.tiny { width:34px; height:34px; border-radius:9px; display:grid; place-items:center; color:#fff; font-size:.74rem; font-weight:800; }
</style>
