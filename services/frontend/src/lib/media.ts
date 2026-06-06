const MEDIA_BASE = import.meta.env.VITE_MEDIA_URL ?? "http://localhost:9000/merch-products";

export function photoUrl(key?: string): string | undefined {
  if (!key) return undefined;
  return `${MEDIA_BASE}/${key}`;
}
