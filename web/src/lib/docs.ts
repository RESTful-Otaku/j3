export type DocEntry = {
  slug: string;
  title: string;
  description: string;
  icon: string;
  group: string;
  file?: string;
  path?: string;
};

export type ApiParameter = {
  name: string;
  in: 'path' | 'query' | 'header';
  required: boolean;
  description: string;
  value: unknown;
};

export type ApiOperation = {
  group: string;
  method: string;
  path: string;
  operationId: string;
  title: string;
  description: string;
  auth: boolean;
  parameters: ApiParameter[];
  body: string;
};

export type SearchEntry = {
  title: string;
  section: string;
  text: string;
  url: string;
};

async function getJson<T>(path: string): Promise<T> {
  const response = await fetch(path, { cache: 'no-store' });
  if (!response.ok) throw new Error(`Could not load ${path} (${response.status})`);
  return response.json() as Promise<T>;
}

export const docs = {
  manifest: () => getJson<{ docs: DocEntry[] }>('/docs/manifest.json'),
  operations: () => getJson<ApiOperation[]>('/docs/operations.json'),
  search: () => getJson<SearchEntry[]>('/docs/search.json'),
  markdown: (file: string) =>
    fetch(`/docs/${encodeURIComponent(file)}`).then((response) => {
      if (!response.ok) throw new Error(`Guide not found (${response.status})`);
      return response.text();
    }),
};

export function docHref(doc: DocEntry): string {
  return doc.path ?? `/docs/${doc.slug}`;
}
