# Storybook (Svelte)

Full **Storybook 10** for SvelteKit (`@storybook/sveltekit`, `@storybook/addon-svelte-csf`, `addon-a11y`, `addon-docs`, `addon-vitest`, `addon-svelte-csf`).

```
bun run storybook            # http://localhost:6006
bun run storybook:build      # static build/web/storybook-static
bun run component:qa         # Playwright /ui component gallery
bun run component:storybook
# web: bun run test:storybook  (Vitest browser project 'storybook'; needs Playwright chromium)
```

## Stories
`web/src/stories/*` examples + `web/src/stories/j3/*.stories.svelte` for J3:
- `Card`, `ConfirmDialog`, `FighterPortrait`, `Pagination`, `HelpTip`, `ToastHost`, `PAToast`, `ViewWidget`, `DemoReplay`.
`@storybook/addon-svelte-csf` `defineMeta` + `<Story>`. `app.css` imported in `.storybook/preview.ts`. a11y `test:'todo'` (set `'error'` for CI fail). 
No auth needed. External venue APIs are best-effort; standalone components don't require `j3d`. 

## Adding stories for everything
Add `*.stories.svelte` for any component/widget/view/content. Keep title `J3/<Name>`. Test: `bun run storybook:build`, `svelte-check`, `test`, `component:qa`. 

## Fix note
Storybook inserts a `const dirname` shadowing `node:path` `dirname`; use `path.dirname` / `rootDir=process.cwd()`, and `path.join(rootDir,...)`. Unit tests run `vitest run --project unit`; Storybook Vitest is separate `project 'storybook'` so CI unit tests don't launch browser without Playwright installed.
