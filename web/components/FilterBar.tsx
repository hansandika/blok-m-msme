"use client";

import { CATEGORIES, PRICES, VIBES } from "@/lib/labels";
import type { PlaceFilters } from "@/lib/types";

export function FilterBar({
  filters,
  onChange,
}: {
  filters: PlaceFilters;
  onChange: (next: PlaceFilters) => void;
}) {
  const set = (patch: PlaceFilters) => onChange({ ...filters, ...patch });

  const chip = (active: boolean) =>
    `shrink-0 rounded-full border px-3 py-1.5 text-xs transition ${
      active
        ? "border-gold bg-gold text-ink"
        : "border-white/10 bg-ink-soft text-mist hover:border-white/25 hover:text-paper"
    }`;

  return (
    <div className="space-y-3">
      <label className="block">
        <span className="sr-only">Search</span>
        <input
          value={filters.q ?? ""}
          onChange={(e) => set({ q: e.target.value })}
          placeholder="Cari izakaya, kopi, seblak, quiet…"
          className="w-full rounded-2xl border border-white/10 bg-ink-soft px-4 py-3 text-sm text-paper outline-none ring-gold/40 placeholder:text-mist/70 focus:ring-2"
        />
      </label>

      <div className="flex gap-2 overflow-x-auto no-scrollbar pb-1">
        <button type="button" className={chip(!filters.category)} onClick={() => set({ category: undefined })}>
          All
        </button>
        {CATEGORIES.map((c) => (
          <button
            key={c.id}
            type="button"
            className={chip(filters.category === c.id)}
            onClick={() => set({ category: filters.category === c.id ? undefined : c.id })}
            title={c.sub}
          >
            {c.label}
          </button>
        ))}
      </div>

      <div className="flex gap-2 overflow-x-auto no-scrollbar pb-1">
        {VIBES.map((v) => (
          <button
            key={v.id}
            type="button"
            className={chip(filters.vibe === v.id)}
            onClick={() => set({ vibe: filters.vibe === v.id ? undefined : v.id })}
            title={v.idn}
          >
            {v.label}
          </button>
        ))}
      </div>

      <div className="flex flex-wrap items-center gap-2">
        {PRICES.map((p) => (
          <button
            key={p}
            type="button"
            className={chip(filters.price === p)}
            onClick={() => set({ price: filters.price === p ? undefined : p })}
          >
            {p}
          </button>
        ))}
        <button
          type="button"
          className={chip(filters.openLate === "true")}
          onClick={() => set({ openLate: filters.openLate === "true" ? undefined : "true" })}
        >
          Open late / larut
        </button>
      </div>
    </div>
  );
}
