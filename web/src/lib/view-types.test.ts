import { afterEach, describe, expect, it } from 'vitest';
import {
  clampSize, insertionIndex, loadViews, saveViews, sizeFromDelta,
} from './view-types';

function stubStorage() {
  const data = new Map<string, string>();
  const storage: Storage = {
    get length() { return data.size; },
    clear: () => data.clear(),
    getItem: (key) => data.get(key) ?? null,
    key: (index) => [...data.keys()][index] ?? null,
    removeItem: (key) => { data.delete(key); },
    setItem: (key, value) => { data.set(key, String(value)); },
  };
  Object.defineProperty(globalThis, 'localStorage', { value: storage, configurable: true });
  return storage;
}

describe('view sizing helpers', () => {
  it('clamps to the three supported grid sizes', () => {
    expect(clampSize(-5)).toBe(1);
    expect(clampSize(1.6)).toBe(2);
    expect(clampSize(99)).toBe(3);
  });

  it('snaps pointer deltas to grid tracks including the gap', () => {
    expect(sizeFromDelta(1, 120, 240, 14)).toBe(1);
    expect(sizeFromDelta(1, 260, 240, 14)).toBe(2);
    expect(sizeFromDelta(2, 508, 240, 14)).toBe(3);
    expect(sizeFromDelta(3, -508, 240, 14)).toBe(1);
  });

  it('finds the nearest insertion point in a grid', () => {
    const rects = [
      { x: 0, y: 0, width: 100, height: 100 },
      { x: 120, y: 0, width: 100, height: 100 },
      { x: 0, y: 120, width: 100, height: 100 },
    ];
    expect(insertionIndex({ x: 20, y: 20 }, rects)).toBe(0);
    expect(insertionIndex({ x: 210, y: 50 }, rects)).toBe(2);
    expect(insertionIndex({ x: 50, y: 190 }, rects)).toBe(3);
    expect(insertionIndex({ x: 10, y: 10 }, [])).toBe(0);
  });
});

describe('saved views', () => {
  afterEach(() => { stubStorage().clear(); });

  it('round-trips saved views', () => {
    stubStorage();
    const views = [{
      name: 'Mat wall',
      updatedAt: 'today',
      widgets: [{ id: 'a', type: 'clock' as const, span: 2 as const, rows: 3 as const }],
    }];
    saveViews(views);
    expect(loadViews()).toEqual(views);
  });

  it('recovers from corrupt storage and malformed widgets', () => {
    const storage = stubStorage();
    storage.setItem('j3-custom-views', 'not json');
    expect(loadViews()).toEqual([]);

    storage.setItem('j3-custom-views', JSON.stringify([
      null,
      { name: 'Broken', widgets: 'no' },
      { name: 'Mixed', updatedAt: 'now', widgets: [
        null,
        { id: 'skip', type: 'nope', span: 9, rows: -2 },
        { id: 'keep', type: 'medals', span: 9, rows: 0 },
      ] },
    ]));
    expect(loadViews()).toEqual([{
      name: 'Mixed', updatedAt: 'now',
      widgets: [{ id: 'keep', type: 'medals', span: 3, rows: 1 }],
    }]);
  });
});
