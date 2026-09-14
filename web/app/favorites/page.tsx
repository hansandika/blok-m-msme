"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { PlaceCard } from "@/components/PlaceCard";
import { fetchPlaces } from "@/lib/api";
import { listFavorites } from "@/lib/favorites";
import type { Place } from "@/lib/types";

export default function FavoritesPage() {
  const [places, setPlaces] = useState<Place[]>([]);
  const [favs, setFavs] = useState<string[]>([]);

  useEffect(() => {
    const sync = () => setFavs(listFavorites());
    sync();
    window.addEventListener("blokm:favorites", sync);
    return () => window.removeEventListener("blokm:favorites", sync);
  }, []);

  useEffect(() => {
    fetchPlaces()
      .then((d) => setPlaces(d.places))
      .catch(() => undefined);
  }, []);

  const saved = places.filter((p) => favs.includes(p.slug));

  return (
    <div>
      <p className="text-xs uppercase tracking-[0.22em] text-gold">Local only</p>
      <h2 className="mt-2 font-display text-4xl">Favorites</h2>
      <p className="mt-3 max-w-xl text-sm text-mist">
        Stored in this browser via localStorage. No account for the MVP — magic-link auth can wait.
      </p>
      <div className="mt-8 grid gap-3 md:grid-cols-2">
        {saved.map((place) => (
          <PlaceCard key={place.slug} place={place} />
        ))}
      </div>
      {saved.length === 0 && (
        <p className="mt-8 rounded-2xl border border-dashed border-white/15 p-8 text-sm text-mist">
          Nothing pinned yet.{" "}
          <Link href="/" className="text-gold hover:underline">
            Browse the map
          </Link>{" "}
          and tap the star.
        </p>
      )}
    </div>
  );
}
