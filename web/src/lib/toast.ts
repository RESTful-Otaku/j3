export type ToastType = 'success' | 'error' | 'warning' | 'info';

export type ToastDetail = {
  id: string;
  message: string;
  type: ToastType;
  duration: number;
  undoable: boolean;
};

const undoHandlers = new Map<string, () => void>();

export function notify(
  message: string,
  options: { type?: ToastType; duration?: number; undo?: () => void } = {},
): string {
  const id = `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
  if (options.undo) undoHandlers.set(id, options.undo);
  window.dispatchEvent(new CustomEvent<ToastDetail>('j3-toast', {
    detail: {
      id,
      message,
      type: options.type ?? 'info',
      duration: options.duration ?? 6000,
      undoable: !!options.undo,
    },
  }));
  return id;
}

export const notifySuccess = (message: string, undo?: () => void) => notify(message, { type: 'success', undo });
export const notifyError = (message: string) => notify(message, { type: 'error', duration: 8000 });
export const notifyWarning = (message: string) => notify(message, { type: 'warning', duration: 8000 });

export function runUndo(id: string) {
  const handler = undoHandlers.get(id);
  if (handler) {
    handler();
    undoHandlers.delete(id);
  }
}
