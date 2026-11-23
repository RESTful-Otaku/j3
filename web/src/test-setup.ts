// Minimal browser surface shared by the Node-based unit tests. Vitest 5 no
// longer ships DOM globals by default, and only a handful of stores use them.
class TestWindow extends EventTarget {
  location = { search: '' };

  history = {
    replaceState: (_state: unknown, _title: string, url: string) => {
      this.location.search = url.includes('?') ? `?${url.split('?')[1]}` : '';
    },
  };
}

const storage = new Map<string, string>();
const localStorage = {
  getItem: (key: string) => storage.get(key) ?? null,
  setItem: (key: string, value: string) => { storage.set(key, String(value)); },
  removeItem: (key: string) => { storage.delete(key); },
  clear: () => storage.clear(),
};

Object.defineProperty(globalThis, 'window', {
  value: new TestWindow(), configurable: true, writable: true,
});
Object.defineProperty(globalThis, 'localStorage', {
  value: localStorage, configurable: true, writable: true,
});

if (typeof globalThis.CustomEvent === 'undefined') {
  class TestCustomEvent<T> extends Event {
    detail: T;
    constructor(type: string, init?: { detail?: T }) {
      super(type);
      this.detail = init?.detail as T;
    }
  }
  Object.defineProperty(globalThis, 'CustomEvent', {
    value: TestCustomEvent, configurable: true, writable: true,
  });
}
