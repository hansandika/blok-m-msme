const KEY = "blokm:favorites";

function read(): string[] {
  if (typeof window === "undefined") return [];
  try {
    const raw = localStorage.getItem(KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw);
    return Array.isArray(parsed) ? parsed.filter((x) => typeof x === "string") : [];
  } catch {
    return [];
  }
}

export function listFavorites(): string[] {
  return read();
}

export function isFavorite(slug: string): boolean {
  return read().includes(slug);
}

export function toggleFavorite(slug: string): string[] {
  const current = new Set(read());
  if (current.has(slug)) current.delete(slug);
  else current.add(slug);
  const next = Array.from(current);
  localStorage.setItem(KEY, JSON.stringify(next));
  window.dispatchEvent(new Event("blokm:favorites"));
  return next;
}
