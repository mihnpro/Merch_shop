import { getAccessToken } from "./tokens";

export function getRole(): string | null {
  const token = getAccessToken();
  if (!token) return null;

  const parts = token.split(".");
  if (parts.length !== 3) return null;

  try {
    const payload = parts[1].replace(/-/g, "+").replace(/_/g, "/");
    const json = atob(payload);
    const claims = JSON.parse(json) as { role?: string };
    return claims.role ?? null;
  } catch {
    return null;
  }
}

export function isAdmin(): boolean {
  return getRole() === "admin";
}
