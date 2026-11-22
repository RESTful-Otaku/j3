import { isPhoneWidth } from './nav';

// Tables become stacked labelled cards on phones. Derive the labels from the
// header row so every existing table works without editing each route.
export function labelTableRows(root: ParentNode = document) {
  root.querySelectorAll('table').forEach((table) => {
    const headers = Array.from(table.querySelectorAll('thead th'))
      .map((cell) => (cell.textContent ?? '').trim());
    if (!headers.length) return;
    table.querySelectorAll('tbody tr').forEach((row) => {
      Array.from(row.children).forEach((cell, index) => {
        if (cell instanceof HTMLElement && headers[index]) cell.dataset.label = headers[index];
      });
    });
  });
}

type LimitState = {
  limit: number;
  expanded: boolean;
  button: HTMLButtonElement | null;
  items: HTMLElement[];
};

function createLimiter(
  node: HTMLElement,
  limit: number,
  collect: () => HTMLElement[],
  insertAfter: () => HTMLElement | null,
) {
  const state: LimitState = { limit, expanded: false, button: null, items: [] };

  function apply() {
    const phone = isPhoneWidth();
    state.items = collect().filter((element) => !element.classList.contains('phone-limit-more'));
    state.items.forEach((element, index) => {
      element.classList.toggle('phone-hidden', phone && !state.expanded && index >= state.limit);
    });
    const extra = phone ? Math.max(0, state.items.length - state.limit) : 0;
    if (extra > 0) {
      if (!state.button) {
        state.button = document.createElement('button');
        state.button.type = 'button';
        state.button.className = 'phone-limit-more';
        state.button.addEventListener('click', () => {
          state.expanded = !state.expanded;
          apply();
        });
        const anchor = insertAfter();
        if (anchor) anchor.after(state.button); else node.after(state.button);
      }
      state.button.textContent = state.expanded ? 'Show less' : `Show ${extra} more`;
    } else {
      state.button?.remove();
      state.button = null;
    }
  }

  const observer = new MutationObserver(apply);
  observer.observe(node, { childList: true, subtree: false });
  window.addEventListener('resize', apply);
  apply();

  return {
    update(next: number) { state.limit = next; apply(); },
    destroy() {
      observer.disconnect();
      window.removeEventListener('resize', apply);
      state.button?.remove();
      state.items.forEach((element) => element.classList.remove('phone-hidden'));
    },
  };
}

// Caps long card grids/lists on phones with an explicit "show more" control.
export function phoneLimit(node: HTMLElement, limit = 8) {
  return createLimiter(
    node,
    limit,
    () => Array.from(node.children).filter((child): child is HTMLElement => child instanceof HTMLElement),
    () => null,
  );
}

// Same behaviour for table bodies; the control sits after the whole table.
export function phoneRows(node: HTMLTableSectionElement, limit = 8) {
  return createLimiter(
    node,
    limit,
    () => Array.from(node.rows).filter((row): row is HTMLTableRowElement => row instanceof HTMLTableRowElement),
    () => node.parentElement,
  );
}
