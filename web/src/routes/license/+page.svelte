<svelte:head><title>J3 · Licence & attribution</title></svelte:head>

<div class="shell">
  <header class="topbar">
    <div class="brand"><div class="brand-mark">J3</div><div><div class="brand-name">Licence & attribution</div><div class="brand-sub">White-label tournament platform</div></div></div>
    <a class="btn" href="/settings">⚙️ Settings</a>
  </header>

  <section class="card hero">
    <div class="eyebrow">Copyright & licence</div>
    <h1>J3 (J Cubed)</h1>
    <p class="muted">Copyright © 2026 Jonathan Harrison (RESTful_Otaku).</p>
    <p class="muted">Licensed under the PolyForm Noncommercial License 1.0.0.</p>
    <div class="actions">
      <a class="btn primary" href="/LICENSE" download>Read the full licence</a>
      <a class="btn" href="https://polyformproject.org/licenses/noncommercial/1.0.0" target="_blank" rel="noopener noreferrer">Official licence terms ↗</a>
    </div>
  </section>

  <div class="grid">
    <section class="card">
      <div class="eyebrow">You may</div>
      <h2>Use, share and adapt</h2>
      <p class="muted small">Use J3 for any non-commercial purpose, study how it works, modify it, share it and redistribute it. Keep the copyright and licence notices intact.</p>
    </section>
    <section class="card">
      <div class="eyebrow">You may not</div>
      <h2>Sell it or pass it off</h2>
      <p class="muted small">J3 cannot be sold. Do not remove the attribution or claim that you originally created the platform.</p>
    </section>
    <section class="card">
      <div class="eyebrow">Your instance</div>
      <h2>You operate it</h2>
      <p class="muted small">Hosting, backups, security, data protection and everything your users enter are your responsibility. J3 comes with no warranty or liability.</p>
    </section>
  </div>

  <p class="notice">Required Notice: Copyright 2026 Jonathan Harrison (RESTful_Otaku)</p>
</div>

<style>
  .hero { margin-top:18px; }
  h1 { font-family:var(--font-display); font-size:clamp(1.6rem,4vw,2.4rem); margin:6px 0; }
  h2 { font-size:1.05rem; margin:6px 0 8px; }
  .actions { display:flex; gap:8px; flex-wrap:wrap; margin-top:16px; }
  .grid { display:grid; grid-template-columns:repeat(auto-fit,minmax(240px,1fr)); gap:14px; margin-top:18px; }
  .notice { margin:18px 0; font-size:.8rem; color:var(--brand-muted); }
</style>
