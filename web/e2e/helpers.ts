// Shared E2E setup: guided tours are tested separately; disable their modal
// overlay so interaction tests exercise the underlying controls.
export function disableTours(page: import('@playwright/test').Page) {
  return page.addInitScript(() => {
    const role = localStorage.getItem('j3-role') || 'runner';
    localStorage.setItem('j3-tour-progress', JSON.stringify({ [role]: 99 }));
    localStorage.setItem('j3-tour-seen', '1');
  });
}
