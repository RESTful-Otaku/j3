<script lang="ts">
  import { onMount } from 'svelte';
  import { api, livePoll, type DemoFight, type FightState } from '../../../lib/api';
  import { onLiveChange } from '../../../lib/live-updates';

  let fights: DemoFight[] = [];
  let state: FightState | null = null;
  onMount(() => {
    const stopFights = livePoll(api.fights, 2000, (data) => (fights = data.fights));
    const stopStream = api.eventStream((message) => { if (message.payload) state = message.payload; });
    const stopLive = onLiveChange(['fights'], () => void api.fights().then((data) => (fights = data.fights)).catch(() => {}));
    return () => { stopFights(); stopStream(); stopLive(); };
  });
  $: fight = fights.find((item) => item.state === 'live') ?? fights[0];
</script>

<svelte:head><title>J3 · Broadcast overlay</title></svelte:head>

<div class="overlay">
  <h1 class="sr-only">Broadcast overlay</h1>
  {#if fight}
    <div class="scorebug">
      <span class="corner red">{fight.red.name} <strong>{state?.redScore ?? 0}</strong></span>
      <span class="clock">{String(Math.floor((state?.seconds ?? 180) / 60)).padStart(2,'0')}:{String((state?.seconds ?? 180) % 60).padStart(2,'0')}</span>
      <span class="corner blue"><strong>{state?.blueScore ?? 0}</strong> {fight.blue.name}</span>
    </div>
    <div class="lower-third">{fight.mat} · {fight.category} · {fight.round}</div>
  {:else}
    <div class="lower-third">J3 · waiting for the first contest</div>
  {/if}
</div>

<style>
  .overlay { min-height:100vh; background:transparent; padding:28px; color:#fff; font-family:var(--font-display); }
  .sr-only { position:absolute; width:1px; height:1px; padding:0; margin:-1px; overflow:hidden; clip:rect(0,0,0,0); white-space:nowrap; border:0; }
  .scorebug { display:inline-flex; align-items:center; gap:18px; padding:12px 22px; border-radius:14px; background:rgba(8,12,24,.82); border:1px solid rgba(255,255,255,.2); box-shadow:0 12px 40px rgba(0,0,0,.5); font-size:clamp(1rem,2vw,2rem); }
  .corner.red { color:#fca5a5; } .corner.blue { color:#93c5fd; }
  .clock { font-weight:800; letter-spacing:.05em; }
  .lower-third { margin-top:12px; display:inline-block; padding:8px 16px; border-radius:10px; background:rgba(8,12,24,.8); font-size:clamp(.8rem,1.4vw,1.3rem); }
</style>
