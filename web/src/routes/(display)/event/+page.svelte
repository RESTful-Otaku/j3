<script lang="ts">
  import { onMount } from 'svelte';
  import { getJson } from '../../../lib/api';
  import { notifySuccess } from '../../../lib/toast';

  type Config = {
    event: {
      name: string; organiser: string; venue: string; costs: { label: string; amount: number; currency: string }[];
      refreshments: string; parking: string; hotel: { name: string; phone: string }[];
      evening: Record<string, string>; course: Record<string, any>; attendanceRule: string;
    };
    kit: { item: string; note: string }[];
    rules: { name: string; kind: string; ages: string; rules: string }[];
    staff: { mat: string; role: string; name: string }[];
    formats: { Discipline: string; Format: string; Knockout: string; Third: string; Order: string; DualBronze: boolean; RetryTies: boolean }[];
  };
  let config: Config | null = null;
  let attendance: { entrantId: string; name: string; attended: boolean; excluded: boolean }[] = [];
  let error = '';
  let tab: 'overview' | 'rules' | 'people' | 'course' = 'overview';

  onMount(async () => {
    try {
      config = await getJson<Config>('/api/v1/demo/event-config?eventId=ev-okami');
      attendance = (await getJson<{ attendance: typeof attendance }>('/api/v1/demo/attendance')).attendance;
    } catch (e) { error = e instanceof Error ? e.message : String(e); }
  });

  function ruleData(raw: string) {
    try { return JSON.parse(raw); } catch { return {}; }
  }
  function announce() { notifySuccess('Announcement sent to all connected displays'); }
</script>

<svelte:head><title>J3 · {config?.event.name ?? 'Event configuration'}</title></svelte:head>

<div class="display-shell">
  <header class="display-header">
    <div>
      <div class="event-mark"><span class="crest">J3</span>{config?.event.name ?? 'Tournament configuration'}</div>
      <div class="meta">Organiser: {config?.event.organiser ?? '—'} · {config?.event.venue ?? '—'}</div>
    </div>
    <button class="btn primary" on:click={announce}>📣 Announce to venue</button>
  </header>

  {#if error}<div class="card" style="border-color:var(--loss);">{error}</div>{/if}
  {#if !config}<div class="placeholder">Loading event configuration…</div>
  {:else}
    <div class="tabs" style="margin-bottom:18px;">
      {#each [['overview','Overview'],['rules','Rules & kit'],['people','Officials & attendance'],['course','Course & day']] as [id, label]}
        <button class="tab" class:active={tab === id} on:click={() => (tab = id as typeof tab)}>{label}</button>
      {/each}
    </div>

    {#if tab === 'overview'}
      <div class="grid" style="grid-template-columns:repeat(auto-fit,minmax(260px,1fr));">
        <div class="card slide-in"><div class="eyebrow">Costs (paid in advance)</div><h2>Entry fees</h2>
          {#each config.event.costs as cost}<div class="row"><strong>{cost.label}</strong><span>£{cost.amount}</span></div>{/each}
        </div>
        <div class="card slide-in"><div class="eyebrow">Venue</div><h2>{config.event.venue}</h2>
          <p class="muted small">{config.event.parking}</p>
          <h3 style="margin-top:12px;">Hotels</h3>
          {#each config.event.hotel as hotel}<div class="row"><strong>{hotel.name}</strong><span>{hotel.phone}</span></div>{/each}
          <div class="stat small" style="margin-top:12px;">Evening {config.event.evening.time} · {config.event.evening.venue}</div>
        </div>
        <div class="card slide-in"><div class="eyebrow">Refreshments</div><h2>Bring your own</h2><p class="muted">{config.event.refreshments}</p></div>
      </div>
    {:else if tab === 'rules'}
      <div class="card slide-in" style="margin-bottom:18px;">
        <div class="eyebrow">Progression & fairness</div><h2>Competition formats</h2>
        <div style="overflow-x:auto;">
          <table style="width:100%; border-collapse:collapse; margin-top:10px; font-size:.86rem;">
            <thead><tr class="muted small" style="text-align:left;"><th>Discipline</th><th>Format</th><th>Next stage</th><th>Third place</th><th>Order</th></tr></thead>
            <tbody>
              {#each config.formats ?? [] as format (format.Discipline)}
                <tr style="border-top:1px solid var(--brand-border);">
                  <td>{format.Discipline}</td><td>{format.Format}</td><td>{format.Knockout}</td>
                  <td>{format.Third}{format.DualBronze ? ' · dual bronze' : ''}{format.RetryTies ? ' · retry ties' : ''}</td>
                  <td>{format.Order}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      </div>
      <div class="grid" style="grid-template-columns:repeat(auto-fit,minmax(280px,1fr));">
        {#each config.rules as rule (rule.name)}
          <div class="card slide-in">
            <div class="eyebrow">{rule.kind} · {rule.ages}</div><h2>{rule.name}</h2>
            <div class="stat small" style="margin-top:10px;">
              {#each Object.entries(ruleData(rule.rules)) as [key, value]}
                <div><strong>{key.replaceAll('_',' ')}</strong>: {JSON.stringify(value)}</div>
              {/each}
            </div>
          </div>
        {/each}
        <div class="card slide-in"><div class="eyebrow">Required kit</div><h2>Equipment</h2>
          {#each config.kit as item}<div class="row"><strong>{item.item}</strong><span class="muted small">{item.note}</span></div>{/each}
        </div>
      </div>
    {:else if tab === 'people'}
      <div class="grid" style="grid-template-columns:repeat(auto-fit,minmax(280px,1fr));">
        <div class="card slide-in"><div class="eyebrow">Mat staffing</div><h2>Officials per mat</h2>
          {#each config.staff as member (member.mat + member.role)}<div class="row"><strong>{member.mat}</strong><span>{member.role} · {member.name}</span></div>{/each}
        </div>
        <div class="card slide-in"><div class="eyebrow">Course rule</div><h2>Eligibility</h2><p class="muted">{config.event.attendanceRule}</p></div>
        <div class="card slide-in"><div class="eyebrow">Attendance</div><h2>Morning course</h2>
          <p class="muted small">{attendance.filter((item) => item.attended).length} attended · {attendance.filter((item) => item.excluded).length} excluded</p>
          <div style="max-height:300px; overflow:auto; margin-top:10px;">
            {#each attendance.filter((item) => item.excluded) as item (item.entrantId)}
              <div class="row"><strong>{item.name}</strong><span class="badge spectator">excluded</span></div>
            {/each}
          </div>
        </div>
      </div>
    {:else}
      <div class="card slide-in"><div class="eyebrow">Day structure</div><h2>Morning course</h2>
        <p class="muted">Attendance is mandatory to compete in the afternoon.</p>
        {#each Object.entries(config.event.course.groups ?? {}) as [index, group]}<div class="stat small" style="margin-top:8px;">{group}</div>{/each}
        <h3 style="margin-top:18px;">Younger students (6-10)</h3>
        <p class="muted small">Dojo fitness challenges devised by sensei and organisers, with certificates and medals awarded during the course.</p>
      </div>
    {/if}
  {/if}
</div>

<style>
  .row { display:flex; justify-content:space-between; gap:12px; padding:8px 0; border-top:1px solid var(--brand-border); font-size:.9rem; }
</style>
