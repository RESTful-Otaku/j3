<script lang="ts">
  import { onMount } from 'svelte';
  import { getJson, postJson } from '../../lib/api';
  import { notifySuccess, notifyError } from '../../lib/toast';

  type Announcement = { id: string; title: string; body: string; channels: string; audience: string; createdAt: string };
  type Suggestion = { Kind: string; Title: string; Reason: string };
  let announcements: Announcement[] = [];
  let suggestions: Suggestion[] = [];
  let title = '';
  let body = '';
  let channels = 'push,email,sms';
  let audience = 'all';

  async function load() {
    try {
      announcements = (await getJson<{ announcements: Announcement[] }>('/api/v1/demo/announcements')).announcements;
      suggestions = (await getJson<{ suggestions: Suggestion[] }>('/api/v1/demo/suggestions')).suggestions;
    } catch { /* keep empty */ }
  }
  onMount(load);

  async function send() {
    if (!title.trim()) return;
    try {
      await postJson('/api/v1/demo/announcements', { eventId: 'ev-okami', title, body, channels, audience });
      notifySuccess(`Announcement sent via ${channels}`);
      title = ''; body = '';
      await load();
    } catch (error) { notifyError(error instanceof Error ? error.message : 'Send failed'); }
  }
</script>

<svelte:head><title>J3 · Notifications</title></svelte:head>

<div class="shell">
  <header class="topbar"><div class="brand"><a class="brand-mark" href="/public" aria-label="Public dashboard">J3</a><div><div class="brand-name">Notifications & suggestions</div><div class="brand-sub">Push, email, SMS and automatic setup ideas</div></div></div></header>

  <div class="grid" style="grid-template-columns:1fr .9fr;">
    <div class="card">
      <div class="eyebrow">Broadcast</div><h2>Send a message</h2>
      <label class="field">Title<input bind:value={title} placeholder="Fighters to Mat 1" /></label>
      <label class="field" style="margin-top:10px;">Message<textarea rows="3" bind:value={body}></textarea></label>
      <div class="grid" style="grid-template-columns:1fr 1fr; margin-top:10px;">
        <label class="field">Channels<select bind:value={channels}><option value="push">Push only</option><option value="push,email">Push + email</option><option value="push,email,sms">Push + email + SMS</option></select></label>
        <label class="field">Audience<select bind:value={audience}><option value="all">Everyone</option><option value="fighters">Fighters</option><option value="officials">Officials</option><option value="clubs">Clubs</option></select></label>
      </div>
      <button class="btn primary" style="margin-top:14px;" on:click={send}>Send announcement</button>
      <p class="muted small" style="margin-top:10px;">Delivery providers require credentials; demo mode records and verifies the payload.</p>
    </div>

    <div class="card">
      <div class="eyebrow">Automatic suggestions</div><h2>Recommended setup</h2>
      <div style="display:grid; gap:8px; margin-top:12px;">
        {#each suggestions as suggestion (suggestion.Kind + suggestion.Title)}
          <div class="stat small"><strong>{suggestion.Title}</strong><div class="muted">{suggestion.Reason}</div></div>
        {:else}<span class="muted">No suggestions yet.</span>{/each}
      </div>
    </div>
  </div>

  <div class="card" style="margin-top:18px;">
    <h2>Recent announcements</h2>
    {#each announcements as announcement (announcement.id)}
      <div class="row"><strong>{announcement.title}</strong><span class="muted small">{announcement.channels} · {announcement.audience} · {announcement.createdAt}</span></div>
    {:else}<span class="muted">Nothing sent yet.</span>{/each}
  </div>
</div>

<style>
  .field { display:grid; gap:4px; font-size:.82rem; color:var(--brand-muted); }
  .field input, .field textarea, .field select { padding:9px; border-radius:9px; background:var(--brand-surface-3); border:1px solid var(--brand-border); color:var(--brand-ink); }
  .row { display:flex; justify-content:space-between; gap:10px; padding:9px 0; border-top:1px solid var(--brand-border); }
  @media (max-width: 820px) { .grid { grid-template-columns:1fr !important; } }
</style>
