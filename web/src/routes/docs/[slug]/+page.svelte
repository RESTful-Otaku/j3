<script lang="ts">
  import { afterNavigate } from '$app/navigation';
  import { docs, docHref } from '../../../lib/docs';
  import { headings, renderMarkdown, slugify } from '../../../lib/markdown';
  import type { DocEntry } from '../../../lib/docs';
  import type { Heading } from '../../../lib/markdown';

  let all: DocEntry[] = [];
  let entry: DocEntry | undefined;
  let html = '';
  let toc: Heading[] = [];
  let error = '';
  let loading = true;
  let loadedSlug = '';
  let loadToken = 0;

  $: docTitle = entry?.title ?? 'Guide';
  $: position = entry ? all.findIndex((item) => item.slug === entry!.slug) : -1;
  $: previous = position > 0 ? all[position - 1] : undefined;
  $: next = position >= 0 && position < all.length - 1 ? all[position + 1] : undefined;

  function currentSlug() {
    return decodeURIComponent(window.location.pathname.split('/').filter(Boolean).pop() ?? '');
  }

  function scrollToHash() {
    const hash = window.location.hash.slice(1);
    if (!hash) return;
    requestAnimationFrame(() => {
      document.getElementById(decodeURIComponent(hash))?.scrollIntoView({ behavior: 'smooth', block: 'start' });
    });
  }

  async function loadDoc(slug: string) {
    const token = ++loadToken;
    loading = true;
    error = '';
    html = '';
    toc = [];
    entry = undefined;
    loadedSlug = slug;
    try {
      const data = await docs.manifest();
      if (token !== loadToken) return;
      all = data.docs;
      entry = all.find((item) => item.slug === slug);
      if (!entry) {
        error = 'That guide does not exist. Choose a page from the sidebar.';
        loading = false;
        return;
      }
      if (entry.path) {
        window.location.href = entry.path;
        return;
      }
      const source = await docs.markdown(entry.file ?? `${slug}.md`);
      if (token !== loadToken) return;
      // The page header already shows the document title; don't render it twice.
      html = renderMarkdown(source.replace(/^#\s+.*(?:\r?\n)+/, ''));
      toc = headings(source).filter((heading) => heading.level === 2 || heading.level === 3);
      loading = false;
      scrollToHash();
    } catch (cause) {
      if (token !== loadToken) return;
      error = String(cause);
      loading = false;
    }
  }

  // SvelteKit reuses this component for every /docs/[slug] URL, so onMount alone
  // would leave the previous guide on screen after a sidebar or inline link click.
  afterNavigate(() => {
    const slug = currentSlug();
    if (slug === loadedSlug) scrollToHash();
    else void loadDoc(slug);
  });
</script>

<svelte:head><title>J3 · {docTitle}</title></svelte:head>

<div class="doc-layout">
  <article class="prose">
    <div class="eyebrow">{entry?.group ?? 'Documentation'}</div>
    <h1 id={slugify(docTitle)}>{docTitle}</h1>
    {#if entry}<p class="lead">{entry.description}</p>{/if}
    {#if loading}<p class="muted">Loading guide…</p>{/if}
    {#if error}<div class="error">{error} <a href="/docs">Back to docs</a></div>{/if}
    {#if html}{@html html}{/if}
    <nav class="pagination" aria-label="More documentation">
      {#if previous}<a href={docHref(previous)}>← {previous.title}</a>{/if}
      {#if next}<a class="next" href={docHref(next)}>{next.title} →</a>{/if}
    </nav>
  </article>

  <nav class="toc" aria-label="On this page">
    <div class="eyebrow">On this page</div>
    {#each toc as heading}
      <a class:sub={heading.level === 3} href={`#${heading.id}`}>{heading.text}</a>
    {/each}
    <div class="toc-foot">
      <a href="/docs/api">🔌 Try the API</a>
      <a href="/openapi.yaml" target="_blank" rel="noreferrer">📄 OpenAPI contract ↗</a>
    </div>
  </nav>
</div>

<style>
  .doc-layout { display:grid; grid-template-columns:minmax(0,1fr) 230px; gap:40px; align-items:start; }
  .prose { min-width:0; }
  .eyebrow { color:var(--brand-accent); font-size:.7rem; letter-spacing:.14em; font-weight:900; text-transform:uppercase; }
  h1 { font-size:clamp(2rem,4vw,3rem); margin:10px 0 12px; scroll-margin-top:80px; }
  .lead { color:var(--brand-muted); font-size:1.05rem; line-height:1.6; margin-bottom:26px; }
  .muted { color:var(--brand-muted); }
  .error { padding:14px; border-radius:12px; background:rgba(255,85,51,.14); border:1px solid var(--loss); margin-bottom:18px; }
  .pagination { display:flex; justify-content:space-between; gap:12px; margin-top:42px; padding-top:18px; border-top:1px solid var(--brand-border); }
  .pagination a { color:var(--brand-primary); text-decoration:none; font-size:.85rem; font-weight:800; }
  .pagination .next { margin-left:auto; }
  .toc { position:sticky; top:78px; display:grid; gap:6px; padding-left:16px; border-left:1px solid var(--brand-border); max-height:calc(100vh - 110px); overflow:auto; }
  .toc a { color:var(--brand-muted); text-decoration:none; font-size:.78rem; line-height:1.4; }
  .toc a:hover { color:var(--brand-primary); }
  .toc a.sub { padding-left:12px; font-size:.74rem; }
  .toc-foot { display:grid; gap:6px; margin-top:14px; padding-top:12px; border-top:1px solid var(--brand-border); }
  .toc-foot a { color:var(--brand-ink); font-weight:700; }
  @media (max-width:1100px) { .doc-layout { grid-template-columns:1fr; } .toc { display:none; } }

  :global(.prose h2), :global(.prose h3), :global(.prose h4) { margin:30px 0 10px; font-size:1.25rem; scroll-margin-top:80px; }
  :global(.prose h3) { font-size:1.05rem; }
  :global(.prose p), :global(.prose li) { color:var(--brand-muted); line-height:1.7; font-size:.92rem; }
  :global(.prose p) { margin:0 0 14px; }
  :global(.prose ul), :global(.prose ol) { margin:0 0 16px; padding-left:22px; display:grid; gap:6px; }
  :global(.prose a) { color:var(--brand-primary); }
  :global(.prose code) { padding:2px 6px; border-radius:6px; background:var(--brand-surface-3); color:var(--brand-ink); font-family:ui-monospace,monospace; font-size:.82em; }
  :global(.prose pre) { overflow:auto; padding:14px; margin:0 0 18px; border-radius:12px; background:var(--brand-surface-3); border:1px solid var(--brand-border); }
  :global(.prose pre code) { padding:0; background:none; color:var(--brand-ink); font-size:.8rem; line-height:1.55; }
  :global(.prose blockquote) { margin:0 0 18px; padding:12px 16px; border-left:3px solid var(--brand-accent); background:var(--brand-surface-3); border-radius:0 10px 10px 0; }
  :global(.prose blockquote p) { margin:0; }
  :global(.prose table) { width:100%; border-collapse:collapse; margin:0 0 20px; font-size:.84rem; }
  :global(.prose th), :global(.prose td) { text-align:left; padding:9px 11px; border:1px solid var(--brand-border); color:var(--brand-muted); }
  :global(.prose th) { color:var(--brand-ink); background:var(--brand-surface-3); }
  :global(.prose hr) { border:0; border-top:1px solid var(--brand-border); margin:26px 0; }
</style>
