import type { Category, Vibe } from "./types";

export const CATEGORIES: { id: Category; label: string; sub: string }[] = [
  { id: "japanese", label: "Japanese", sub: "Izakaya / ramen" },
  { id: "street", label: "Street", sub: "Kaki lima" },
  { id: "cafe", label: "Cafe", sub: "Kopi" },
  { id: "retail", label: "Retail", sub: "Toko / distro" },
  { id: "other", label: "Other", sub: "Lainnya" },
];

export const VIBES: { id: Vibe; label: string; idn: string }[] = [
  { id: "quiet", label: "Quiet", idn: "sepi" },
  { id: "lively", label: "Lively", idn: "rame" },
  { id: "hidden", label: "Hidden", idn: "tersembunyi" },
  { id: "hangout", label: "Hangout", idn: "nongkrong" },
  { id: "date-night", label: "Date night", idn: "kencan" },
  { id: "after-office", label: "After office", idn: "pulang kantor" },
  { id: "quick-bite", label: "Quick bite", idn: "cepat" },
  { id: "family", label: "Family", idn: "keluarga" },
];

export const PRICES = ["$", "$$", "$$$"] as const;

export const CATEGORY_COLOR: Record<Category, string> = {
  japanese: "#d4512a",
  street: "#d4a017",
  cafe: "#6f8f76",
  retail: "#8aa0d4",
  other: "#b7b0a4",
};

export function vibeLabel(id: string): string {
  return VIBES.find((v) => v.id === id)?.label ?? id;
}
