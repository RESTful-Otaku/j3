import { describe, expect, it } from 'vitest';
import { headings, renderMarkdown } from './markdown';

describe('documentation markdown', () => {
  it('renders headings, anchors, code, lists and tables', () => {
    const html = renderMarkdown([
      '# Organiser guide',
      '',
      'Run the **demo** with `bun run demo`.',
      '',
      '- First step',
      '- Second step',
      '',
      '```bash',
      'bun run smoke',
      '```',
      '',
      '| Role | Access |',
      '|---|---|',
      '| Runner | All |',
      '',
      '> Read the [API guide](/docs/api).',
    ].join('\n'));
    expect(html).toContain('<h1 id="organiser-guide">');
    expect(html).toContain('<strong>demo</strong>');
    expect(html).toContain('<code>bun run demo</code>');
    expect(html).toContain('<ul><li>First step</li><li>Second step</li></ul>');
    expect(html).toContain('<pre><code class="language-bash">bun run smoke</code></pre>');
    expect(html).toContain('<th>Role</th>');
    expect(html).toContain('<blockquote>');
    expect(html).toContain('href="/docs/api"');
  });

  it('escapes raw HTML and unsafe links', () => {
    const html = renderMarkdown('Hello <script>alert(1)</script> [click](javascript:alert(1))');
    expect(html).not.toContain('<script>');
    expect(html).toContain('&lt;script&gt;');
    expect(html).toContain('href="#"');
  });

  it('extracts heading metadata for the table of contents', () => {
    expect(headings('## Before you start\n\n### Setup\n')).toEqual([
      { id: 'before-you-start', text: 'Before you start', level: 2 },
      { id: 'setup', text: 'Setup', level: 3 },
    ]);
  });
});
