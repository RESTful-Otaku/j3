<script lang="ts">
  import { onMount } from 'svelte';
  import { getJson, postJson } from '../../lib/api';
  import { notifySuccess, notifyError } from '../../lib/toast';

  type Permission = { role: string; resource: string; canView: boolean; canCreate: boolean; canEdit: boolean; canDelete: boolean };
  let permissions: Permission[] = [];

  async function load() {
    try {
      permissions = (await getJson<{ permissions: Permission[] }>('/api/v1/demo/permissions?init=1')).permissions;
    } catch { permissions = []; }
  }
  onMount(load);

  async function toggle(permission: Permission, flag: 'view' | 'create' | 'edit' | 'delete', value: boolean) {
    try {
      await postJson('/api/v1/demo/permissions', {
        role: permission.role, resource: permission.resource, flag, value,
      });
      notifySuccess(`${permission.role} · ${permission.resource} · ${flag} ${value ? 'enabled' : 'disabled'}`);
      await load();
    } catch (error) {
      notifyError(error instanceof Error ? error.message : 'Could not update permission');
    }
  }
</script>

<svelte:head><title>J3 · Access control</title></svelte:head>

<div class="shell">
  <header class="topbar">
    <div class="brand"><a class="brand-mark" href="/" aria-label="Dashboard">J3</a><div><div class="brand-name">Access control</div><div class="brand-sub">Runner-only · who can view, create, edit and delete</div></div></div>
    <a class="btn" href="/login">🔐 Sign in / out</a>
  </header>

  <div class="card" style="overflow-x:auto;">
    <table style="width:100%; border-collapse:collapse;">
      <thead><tr class="muted small" style="text-align:left;">
        <th style="padding:10px;">Role</th><th>Area</th><th>View</th><th>Create</th><th>Edit</th><th>Delete</th>
      </tr></thead>
      <tbody>
        {#each permissions as permission (permission.role + permission.resource)}
          <tr style="border-top:1px solid var(--brand-border);">
            <td style="padding:10px; text-transform:capitalize;">{permission.role.replaceAll('_',' ')}</td>
            <td>{permission.resource}</td>
            {#each [['view', permission.canView],['create', permission.canCreate],['edit', permission.canEdit],['delete', permission.canDelete]] as [flag, value]}
              <td><label class="switch">
                <input type="checkbox" checked={Boolean(value)} on:change={(event) => toggle(permission, flag as any, (event.target as HTMLInputElement).checked)} />
                <span class="slider"></span>
              </label></td>
            {/each}
          </tr>
        {/each}
      </tbody>
    </table>
    <p class="muted small" style="margin-top:14px;">Changes are audit-logged with the acting user and timestamp.</p>
  </div>
</div>

<style>
  .switch { position:relative; display:inline-block; width:42px; height:24px; }
  .switch input { opacity:0; width:0; height:0; }
  .slider { position:absolute; inset:0; background:var(--brand-surface-3); border-radius:999px; cursor:pointer; transition:.2s; border:1px solid var(--brand-border); }
  .slider::before { content:''; position:absolute; width:18px; height:18px; left:2px; bottom:2px; background:var(--brand-muted); border-radius:50%; transition:.2s; }
  input:checked + .slider { background:var(--brand-primary); }
  input:checked + .slider::before { transform:translateX(18px); background:#fff; }
</style>
