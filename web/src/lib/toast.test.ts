// @vitest-environment happy-dom
import { describe, expect, it, vi } from 'vitest';
import { notify, notifySuccess, runUndo, type ToastDetail } from './toast';

describe('notification system', () => {
  it('broadcasts a toast with the requested type and duration', () => {
    const listener = vi.fn();
    window.addEventListener('j3-toast', listener);
    const id = notify('Saved', { type: 'success', duration: 9000 });
    expect(listener).toHaveBeenCalledOnce();
    const detail = (listener.mock.calls[0][0] as CustomEvent<ToastDetail>).detail;
    expect(detail.id).toBe(id);
    expect(detail.type).toBe('success');
    expect(detail.duration).toBe(9000);
    window.removeEventListener('j3-toast', listener);
  });

  it('runs the undo callback exactly once', () => {
    const undo = vi.fn();
    const id = notifySuccess('Deleted', undo);
    runUndo(id);
    runUndo(id);
    expect(undo).toHaveBeenCalledOnce();
  });

  it('does not offer undo when no handler was supplied', () => {
    const listener = vi.fn();
    window.addEventListener('j3-toast', listener);
    const id = notify('Information');
    const detail = (listener.mock.calls[0][0] as CustomEvent<ToastDetail>).detail;
    expect(detail.id).toBe(id);
    expect(detail.undoable).toBe(false);
    expect(() => runUndo(id)).not.toThrow();
    window.removeEventListener('j3-toast', listener);
  });
});
