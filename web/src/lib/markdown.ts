/**
 * Small, dependency-free renderer for J3's own documentation markdown.
 *
 * Docs are trusted files shipped with the app, but everything is still escaped
 * before being inserted with {@html} so a copied snippet can never inject HTML.
 */

export type Heading = { id: string; text: string; level: number };

export function slugify(text: string): string {
  return text
    .toLowerCase()
    .replace(/<[^>]*>/g, '')
    .replace(/[`*_]/g, '')
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '');
}

function escapeHtml(value: string): string {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}

function inline(value: string): string {
  const escaped = escapeHtml(value);
  return escaped
    .replace(/`([^`]+)`/g, '<code>$1</code>')
    .replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
    .replace(/_([^_]+)_/g, '<em>$1</em>')
    .replace(/\[([^\]]+)\]\(([^)]+)\)/g, (_match, label: string, href: string) => {
      const safe = /^(https?:\/\/|\/|#)/.test(href) ? href : '#';
      const external = safe.startsWith('http') ? ' target="_blank" rel="noreferrer"' : '';
      return `<a href="${escapeHtml(safe)}"${external}>${label}</a>`;
    });
}

function isBlockStart(line: string): boolean {
  return (
    /^#{1,4}\s/.test(line) ||
    /^```/.test(line) ||
    /^\s*[-*]\s/.test(line) ||
    /^\s*\d+\.\s/.test(line) ||
    /^>\s?/.test(line) ||
    /^\|.*\|/.test(line) ||
    /^---+\s*$/.test(line)
  );
}

export function headings(source: string): Heading[] {
  const found: Heading[] = [];
  for (const line of source.replace(/\r\n/g, '\n').split('\n')) {
    const match = /^(#{1,4})\s+(.+)$/.exec(line);
    if (!match) continue;
    const text = match[2].replace(/`/g, '');
    found.push({ level: match[1].length, text, id: slugify(text) });
  }
  return found;
}

export function renderMarkdown(source: string): string {
  const lines = source.replace(/\r\n/g, '\n').split('\n');
  const out: string[] = [];
  let i = 0;

  while (i < lines.length) {
    const line = lines[i];
    if (!line.trim()) {
      i += 1;
      continue;
    }

    if (/^```/.test(line)) {
      const language = line.slice(3).trim().replace(/[^a-z0-9+-]/gi, '');
      const code: string[] = [];
      i += 1;
      while (i < lines.length && !/^```/.test(lines[i])) {
        code.push(lines[i]);
        i += 1;
      }
      i += 1;
      out.push(`<pre><code${language ? ` class="language-${language}"` : ''}>${escapeHtml(code.join('\n'))}</code></pre>`);
      continue;
    }

    const heading = /^(#{1,4})\s+(.+)$/.exec(line);
    if (heading) {
      const text = heading[2].replace(/`/g, '');
      const level = heading[1].length;
      out.push(`<h${level} id="${slugify(text)}">${inline(heading[2])}</h${level}>`);
      i += 1;
      continue;
    }

    if (/^---+\s*$/.test(line.trim())) {
      out.push('<hr />');
      i += 1;
      continue;
    }

    if (line.startsWith('|') && /^\|?[\s:|-]+\|?\s*$/.test(lines[i + 1] ?? '')) {
      const cells = (row: string) => row.split('|').map((cell) => cell.trim()).filter(Boolean);
      const head = cells(line);
      i += 2;
      const rows: string[][] = [];
      while (i < lines.length && lines[i].startsWith('|')) {
        rows.push(cells(lines[i]));
        i += 1;
      }
      out.push(
        `<table><thead><tr>${head.map((cell) => `<th>${inline(cell)}</th>`).join('')}</tr></thead><tbody>` +
          rows.map((row) => `<tr>${row.map((cell) => `<td>${inline(cell)}</td>`).join('')}</tr>`).join('') +
          '</tbody></table>'
      );
      continue;
    }

    if (/^>\s?/.test(line)) {
      const quote: string[] = [];
      while (i < lines.length && /^>\s?/.test(lines[i])) {
        quote.push(lines[i].replace(/^>\s?/, ''));
        i += 1;
      }
      out.push(`<blockquote>${quote.map((part) => `<p>${inline(part)}</p>`).join('')}</blockquote>`);
      continue;
    }

    const listMatch = /^(\s*)([-*]|\d+\.)\s+(.+)$/.exec(line);
    if (listMatch) {
      const ordered = /\d/.test(listMatch[2]);
      const items: string[] = [];
      while (i < lines.length) {
        const item = /^(\s*)([-*]|\d+\.)\s+(.+)$/.exec(lines[i]);
        if (!item) break;
        items.push(`<li>${inline(item[3])}</li>`);
        i += 1;
      }
      const tag = ordered ? 'ol' : 'ul';
      out.push(`<${tag}>${items.join('')}</${tag}>`);
      continue;
    }

    const paragraph: string[] = [];
    while (i < lines.length && lines[i].trim() && !isBlockStart(lines[i])) {
      paragraph.push(lines[i].trim());
      i += 1;
    }
    if (paragraph.length) out.push(`<p>${inline(paragraph.join(' '))}</p>`);
    else i += 1;
  }

  return out.join('\n');
}
