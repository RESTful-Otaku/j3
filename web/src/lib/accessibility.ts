import { writable, get } from 'svelte/store';

export type FontScale = 'normal' | 'large' | 'x-large';
export type Contrast = 'normal' | 'high';
export type ColourMode = 'colour' | 'monochrome';
export type Motion = 'system' | 'reduced';
export type Accessibility = {
  fontScale: FontScale;
  contrast: Contrast;
  colourMode: ColourMode;
  dyslexic: boolean;
  motion: Motion;
};

const KEY = 'j3-accessibility';
export const accessibility = writable<Accessibility>({
  fontScale: 'normal', contrast: 'normal', colourMode: 'colour', dyslexic: false, motion: 'system',
});

function apply(a: Accessibility) {
  if (typeof document === 'undefined') return;
  const root = document.documentElement;
  root.dataset.fontScale = a.fontScale;
  root.dataset.contrast = a.contrast;
  root.dataset.colourMode = a.colourMode;
  root.dataset.dyslexic = a.dyslexic ? 'on' : 'off';
  root.dataset.motion = a.motion;
}

export function initAccessibility(): void {
  if (typeof localStorage === 'undefined') return;
  try {
    const saved = localStorage.getItem(KEY);
    if (saved) {
      const parsed = JSON.parse(saved) as Partial<Accessibility>;
      accessibility.set({ ...get(accessibility), ...parsed });
    }
  } catch { /* ignore */ }
  apply(get(accessibility));
  accessibility.subscribe((a) => {
    apply(a);
    try { localStorage.setItem(KEY, JSON.stringify(a)); } catch { /* ignore */ }
  });
}

export function setAccessibility(partial: Partial<Accessibility>): void {
  accessibility.update((a) => ({ ...a, ...partial }));
}
