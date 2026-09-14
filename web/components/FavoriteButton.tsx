"use client";

import { useEffect, useState } from "react";
import { isFavorite, toggleFavorite } from "@/lib/favorites";

export function FavoriteButton({ slug, compact = false }: { slug: string; compact?: boolean }) {
  const [on, setOn] = useState(false);

  useEffect(() => {
    const sync = () => setOn(isFavorite(slug));
    sync();
    window.addEventListener("blokm:favorites", sync);
    window.addEventListener("storage", sync);
    return () => {
      window.removeEventListener("blokm:favorites", sync);
      window.removeEventListener("storage", sync);
    };
  }, [slug]);

  return (
    <button
      type="button"
      onClick={(e) => {
        e.preventDefault();
        e.stopPropagation();
        setOn(toggleFavorite(slug).includes(slug));
      }}
      aria-pressed={on}
      className={`inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-xs transition ${
        on
          ? "border-gold/40 bg-gold/15 text-gold"
          : "border-white/10 bg-white/5 text-mist hover:text-paper"
      }`}
    >
      <span aria-hidden>{on ? "★" : "☆"}</span>
      {!compact && (on ? "Saved" : "Save")}
    </button>
  );
}
