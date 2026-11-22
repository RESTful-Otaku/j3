import { writable } from 'svelte/store';
import { getJson } from './api';

// White-label title: the event runner's configured event name replaces the
// platform placeholder everywhere the app shows a human-facing title.
export const eventTitle = writable('J3 Tournament');

let loaded = false;

export async function loadEventTitle(): Promise<void> {
  if (loaded) return;
  loaded = true;
  try {
    const data = await getJson<{ event?: { name?: string } }>('/api/v1/demo/event-config');
    const name = data.event?.name?.trim();
    if (name) eventTitle.set(name);
  } catch {
    // Keep the placeholder when the venue server and offline snapshot are both
    // unavailable; the rest of the read-only demo still degrades gracefully.
  }
}
