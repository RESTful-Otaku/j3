import { get } from 'svelte/store';
import { api } from './api';
import { appMode } from './demo-replay';

export type LiveFrame = {
  type: string;
  fightId?: string;
  payload?: Record<string, unknown>;
};

type Listener = { predicate: (frame: LiveFrame) => boolean; callback: () => void; timer?: number };

const listeners = new Set<Listener>();
let closeStream: (() => void) | null = null;

function ensureStream() {
  if (closeStream || typeof window === 'undefined') return;
  closeStream = api.eventStream((event) => {
    // Replay mode owns the visible state; live frames must never mutate it.
    if (get(appMode) === 'sim') return;
    const frame = event as unknown as LiveFrame;
    for (const listener of listeners) {
      if (!listener.predicate(frame)) continue;
      if (listener.timer !== undefined) window.clearTimeout(listener.timer);
      listener.timer = window.setTimeout(() => {
        listener.timer = undefined;
        listener.callback();
      }, 120);
    }
  });
}

export function onLiveFrame(
  predicate: (frame: LiveFrame) => boolean,
  callback: () => void,
): () => void {
  const listener: Listener = { predicate, callback };
  listeners.add(listener);
  ensureStream();
  return () => {
    if (listener.timer !== undefined && typeof window !== 'undefined') window.clearTimeout(listener.timer);
    listeners.delete(listener);
    if (listeners.size === 0 && closeStream) {
      closeStream();
      closeStream = null;
    }
  };
}

/** Refetch a dataset when one of the supplied server invalidation kinds changes. */
export function onLiveChange(kinds: string[], callback: () => void): () => void {
  return onLiveFrame(
    (frame) => frame.type === 'state.changed' && kinds.includes(String(frame.payload?.kind ?? '')),
    callback,
  );
}

/** Refetch when any known fight payload changes (scores, clock, outcomes). */
export function onFightChange(callback: () => void): () => void {
  return onLiveFrame((frame) => frame.type.startsWith('fight.') || frame.type === 'clock', callback);
}
