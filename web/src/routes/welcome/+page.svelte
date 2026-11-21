<script lang="ts">
  import { goto } from '$app/navigation';
  import { getJson, type DemoCounts } from '../../lib/api';
  import { session } from '../../lib/auth';

  let counts: DemoCounts | null = null;
  onMountLoad();
  function onMountLoad() {
    getJson<{ counts: DemoCounts }>('/api/v1/demo')
      .then((data) => (counts = data.counts))
      .catch(() => {});
  }
</script>

<svelte:head><title>J3 · Okami Jiu Jitsu Challenger</title></svelte:head>

<div class="landing">
  <header class="landing-nav">
    <span class="brand"><span class="brand-mark">J3</span> J Cubed</span>
    <nav>
      <a href="/event">Event details</a>
      <a href="/public">Live dashboard</a>
      <a href="/login">Sign in</a>
      <a class="btn primary" href="/signup">Register now →</a>
    </nav>
  </header>

  <section class="hero">
    <div class="hero-copy">
      <span class="pill">21 November 2026 · Skegness</span>
      <h1>Okami Jiu Jitsu<br /><em>Challenger</em></h1>
      <p>A full day of kata, randori, ground fighting, continuous fighting and sparring — run smoothly on J3.</p>
      <div class="actions">
        <a class="btn primary" href="/signup">Register as a participant</a>
        <a class="btn" href="/login">I already have an account</a>
      </div>
      <div class="stats">
        <div><strong>{counts?.fighters ?? 40}</strong><span>fighters</span></div>
        <div><strong>{counts?.clubs ?? 4}</strong><span>clubs</span></div>
        <div><strong>2</strong><span>mats</span></div>
        <div><strong>{counts?.fights ?? 18}</strong><span>bouts</span></div>
      </div>
    </div>
    <div class="hero-card">
      <div class="eyebrow">Today's running order</div>
      {#each [['09:15','Registration'],['09:30','Course & drills'],['12:30','Open hand kata'],['13:30','Juniors & teens'],['15:30','Adults & sparring'],['19:00','Evening event']] as [time, label]}
        <div class="agenda"><strong>{time}</strong><span>{label}</span></div>
      {/each}
      <a class="btn wide" href="/event">Full event details ↗</a>
    </div>
  </section>

  <section class="features">
    <div class="feature"><span>🧩</span><h3>Smart categories</h3><p>Enter your stats once — the app suggests only the divisions and disciplines you qualify for.</p></div>
    <div class="feature"><span>🏆</span><h3>Live brackets</h3><p>Follow pools, eliminations and medals from any screen, in the hall or at home.</p></div>
    <div class="feature"><span>📣</span><h3>Never miss a bout</h3><p>Callouts, PA messages and next-up boards keep fighters and spectators informed.</p></div>
  </section>
</div>

<style>
  .landing { min-height:100vh; }
  .landing-nav { display:flex; align-items:center; justify-content:space-between; gap:20px; padding:18px 36px; border-bottom:1px solid var(--brand-border); flex-wrap:wrap; }
  .brand { display:flex; align-items:center; gap:10px; font-family:var(--font-display); font-weight:800; font-size:1.2rem; }
  .brand-mark { width:40px; height:40px; border-radius:12px; display:grid; place-items:center; background:linear-gradient(145deg,var(--brand-primary),var(--brand-accent)); color:#fff; }
  nav { display:flex; gap:18px; align-items:center; flex-wrap:wrap; }
  nav a { color:var(--brand-muted); text-decoration:none; font-size:.9rem; }
  nav a.btn { color:#fff; }
  .hero { display:grid; grid-template-columns:1.3fr 1fr; gap:40px; padding:64px 36px; align-items:center; }
  h1 { font-size:clamp(2.6rem,6vw,5.4rem); line-height:1; margin:14px 0; }
  h1 em { color:var(--brand-accent); font-style:normal; }
  .hero-copy p { color:var(--brand-muted); font-size:1.1rem; max-width:52ch; }
  .actions { display:flex; gap:12px; flex-wrap:wrap; margin-top:24px; }
  .stats { display:flex; gap:28px; margin-top:38px; flex-wrap:wrap; }
  .stats div { display:grid; }
  .stats strong { font-family:var(--font-display); font-size:2rem; }
  .stats span { color:var(--brand-muted); font-size:.8rem; }
  .hero-card { padding:24px; border-radius:18px; background:var(--brand-surface-2); border:1px solid var(--brand-border); box-shadow:var(--brand-shadow); }
  .agenda { display:flex; justify-content:space-between; gap:12px; padding:10px 0; border-top:1px solid var(--brand-border); }
  .agenda strong { color:var(--brand-accent); }
  .wide { width:100%; justify-content:center; margin-top:16px; }
  .features { display:grid; grid-template-columns:repeat(auto-fit,minmax(240px,1fr)); gap:18px; padding:0 36px 60px; }
  .feature { padding:22px; border-radius:16px; background:var(--brand-surface-2); border:1px solid var(--brand-border); }
  .feature span { font-size:2rem; }
  .feature h3 { margin:10px 0; }
  .feature p { color:var(--brand-muted); font-size:.9rem; }
  @media (max-width: 800px) { .hero { grid-template-columns:1fr; padding:32px 18px; } .landing-nav, .features { padding-left:18px; padding-right:18px; } }
</style>
