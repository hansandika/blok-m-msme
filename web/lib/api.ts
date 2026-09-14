import type { Place, PlaceFilters, PlaceListResponse } from "./types";

export const API_URL =
  process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "") || "http://localhost:8080";

export function placesQuery(filters: PlaceFilters): string {
  const params = new URLSearchParams();
  if (filters.q) params.set("q", filters.q);
  if (filters.category) params.set("category", filters.category);
  if (filters.vibe) params.set("vibe", filters.vibe);
  if (filters.price) params.set("price", filters.price);
  if (filters.openLate) params.set("openLate", filters.openLate);
  return params.toString();
}

export async function fetchPlaces(filters: PlaceFilters = {}): Promise<PlaceListResponse> {
  const qs = placesQuery(filters);
  const res = await fetch(`${API_URL}/places${qs ? `?${qs}` : ""}`, {
    cache: "no-store",
  });
  if (!res.ok) {
    throw new Error(`Places request failed (${res.status})`);
  }
  return res.json();
}

export async function fetchPlace(id: string): Promise<Place> {
  const res = await fetch(`${API_URL}/places/${encodeURIComponent(id)}`, {
    cache: "no-store",
  });
  if (res.status === 404) {
    throw new Error("not-found");
  }
  if (!res.ok) {
    throw new Error(`Place request failed (${res.status})`);
  }
  return res.json();
}

export async function submitSuggestion(body: {
  kind: "new_place" | "correction";
  placeId?: string;
  name?: string;
  category?: string;
  notes: string;
  lat?: number;
  lng?: number;
}) {
  const res = await fetch(`${API_URL}/suggestions`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(data.error || "Could not submit");
  }
  return data as { message: string };
}

const ADMIN_TOKEN_KEY = "blokm-admin-token";

export function readAdminToken(): string {
  if (typeof window === "undefined") return "";
  return window.localStorage.getItem(ADMIN_TOKEN_KEY) ?? "";
}

export function writeAdminToken(token: string) {
  window.localStorage.setItem(ADMIN_TOKEN_KEY, token);
}

export function clearAdminToken() {
  window.localStorage.removeItem(ADMIN_TOKEN_KEY);
}

async function adminRequest<T>(path: string, token: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_URL}${path}`, {
    ...init,
    cache: "no-store",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
      "X-Admin-Token": token,
      ...(init?.headers ?? {}),
    },
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(data.error || `Request failed (${res.status})`);
  }
  return data as T;
}

export async function fetchAdminSuggestions(token: string, status = "") {
  const qs = status ? `?status=${encodeURIComponent(status)}` : "";
  return adminRequest<{ suggestions: import("./types").Suggestion[]; count: number }>(
    `/admin/suggestions${qs}`,
    token,
  );
}

export async function moderateSuggestion(
  token: string,
  id: string,
  action: "approve" | "reject" | "apply",
) {
  return adminRequest<{
    suggestion: import("./types").Suggestion;
    place?: import("./types").Place;
    alreadyApplied?: boolean;
    message?: string;
  }>(`/admin/suggestions/${encodeURIComponent(id)}/${action}`, token, { method: "POST" });
}
