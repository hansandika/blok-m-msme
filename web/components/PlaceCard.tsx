"use client";

import Link from "next/link";
import type { Place } from "@/lib/types";
import { CATEGORY_COLOR, vibeLabel } from "@/lib/labels";
import { FavoriteButton } from "./FavoriteButton";

export function PlaceCard({
  place,
  selected,
  onSelect,
}: {
  place: Place;
  selected?: boolean;
  onSelect?: (slug: string) => void;
}) {
  return (
    <article
      className={`rounded-2xl border p-4 shadow-card transition ${
        selected ? "border-gold/50 bg-ink-soft" : "border-white/10 bg-ink-soft/70 hover:border-white/20"
      }`}
    >
      <button
        type="button"
        onClick={() => onSelect?.(place.slug)}
        className="w-full text-left"
      >
        <div className="flex items-start justify-between gap-3">
          <div>
            <p className="text-[11px] uppercase tracking-[0.18em] text-mist">
              {place.neighborhood} · {place.price}
            </p>
            <h2 className="mt-1 font-display text-xl leading-tight">{place.name}</h2>
          </div>
          <span
            className="mt-1 h-2.5 w-2.5 shrink-0 rounded-full"
            style={{ background: CATEGORY_COLOR[place.category] }}
            title={place.category}
          />
        </div>
        <p className="mt-2 line-clamp-2 text-sm leading-relaxed text-mist">{place.description}</p>
      </button>
      <div className="mt-3 flex flex-wrap items-center gap-1.5">
        <span className="rounded-full bg-white/5 px-2 py-0.5 text-[11px] capitalize text-paper/80">
          {place.category}
        </span>
        {place.openLate && (
          <span className="rounded-full bg-chili/20 px-2 py-0.5 text-[11px] text-[#f0b3a0]">
            Open late / larut
          </span>
        )}
        {place.vibes.slice(0, 3).map((v) => (
          <span key={v} className="rounded-full bg-white/5 px-2 py-0.5 text-[11px] text-mist">
            {vibeLabel(v)}
          </span>
        ))}
        <div className="ml-auto flex items-center gap-2">
          <FavoriteButton slug={place.slug} compact />
          <Link href={`/places/${place.slug}`} className="text-xs text-gold hover:underline">
            Detail
          </Link>
        </div>
      </div>
    </article>
  );
}
