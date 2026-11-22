import { writable } from 'svelte/store';

/** Shared event name. Populated by lib/event-brand.ts from the venue API. */
export const eventTitle = writable('J3 Tournament');
