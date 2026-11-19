import { describe, expect, it } from 'vitest';
import { allActions } from './api';

describe('RBAC capability catalog', () => {
  it('covers the nine capabilities used by the demo', () => {
    expect(allActions).toHaveLength(9);
    expect(allActions).toContain('score_fight');
    expect(allActions).toContain('manage_event');
  });
});
