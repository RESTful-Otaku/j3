export function paginate<T>(items: T[], page: number, pageSize: number): T[] {
  const size = Math.max(1, Math.floor(pageSize));
  const start = Math.max(0, (Math.floor(page) - 1) * size);
  return items.slice(start, start + size);
}
export function pageCount(total: number, pageSize: number): number {
  return Math.max(1, Math.ceil(total / Math.max(1, Math.floor(pageSize))));
}
export function clampPage(page: number, total: number, pageSize: number): number {
  const count = pageCount(total, pageSize);
  return Math.min(Math.max(1, Math.floor(page)), count);
}
