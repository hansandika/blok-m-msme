"use client";

import { useEffect, useMemo, useState } from "react";
import { FilterBar } from "@/components/FilterBar";
import { PlaceCard } from "@/components/PlaceCard";
import { PlaceMapClient } from "@/components/PlaceMapClient";
import { fetchPlaces } from "@/lib/api";
import type { Place, PlaceFilters } from "@/lib/types";

export function Explore() {
  const [filters, setFilters] = useState<PlaceFilters>({});
  const [places, setPlaces] = useState<Place[]>([]);
  const [count, setCount] = useState(0);
  const [selected, setSelected] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const debounced = useDebounced(filters, 180);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    fetchPlaces(debounced)
      .then((data) => {
        if (cancelled) return;
        setPlaces(data.places);
        setCount(data.count);
        setError(null);
        if (data.places.length && !data.places.some((p) => p.slug === selected)) {
          setSelected(null);
        }
      })
      .catch((err: Error) => {
        if (!cancelled) setError(err.message);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
    // selected is intentionally not a fetch dependency
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [debounced]);

  const summary = useMemo(() => {
    if (loading) return "Loading Melawai, Hub, M Bloc…";
    if (error) return "API unreachable — is the Go server running on :8080?";
    return `${count} place${count === 1 ? "" : "s"} inside the Blok M geofence`;
  }, [loading, error, count]);

  return (
    <div className="space-y-6">
      <section className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_minmax(0,1.05fr)]">
        <div>
          <p className="text-xs uppercase tracking-[0.22em] text-gold">Module 1 · Tonight later</p>
          <h2 className="mt-2 font-display text-4xl leading-[1.05] sm:text-5xl">
            Not another mega-directory.
          </h2>
          <p className="mt-3 max-w-xl text-sm leading-relaxed text-mist sm:text-base">
            Locals in South Jakarta already know Blok M is dense — ~500 Hub tenants plus Melawai
            Japanese plus kaki lima. Filter for a quiet izakaya, quick kopi, or something open past 11.
          </p>
        </div>
        <FilterBar filters={filters} onChange={setFilters} />
      </section>

      <p className="text-sm text-mist">{summary}</p>

      <section className="grid gap-4 lg:grid-cols-[minmax(0,0.95fr)_minmax(0,1.15fr)]">
        <div className="h-[42vh] min-h-[260px] overflow-hidden rounded-2xl border border-white/10 lg:h-[calc(100vh-12rem)] lg:min-h-[480px]">
          <PlaceMapClient places={places} selected={selected} onSelect={setSelected} />
        </div>
        <div className="grid max-h-[70vh] gap-3 overflow-y-auto pr-1 lg:max-h-[calc(100vh-12rem)]">
          {places.map((place) => (
            <PlaceCard
              key={place.slug}
              place={place}
              selected={place.slug === selected}
              onSelect={setSelected}
            />
          ))}
          {!loading && places.length === 0 && !error && (
            <p className="rounded-2xl border border-dashed border-white/15 p-8 text-sm text-mist">
              Nothing matches. Try “kopi”, “izakaya”, or clear the vibe chips.
            </p>
          )}
        </div>
      </section>
    </div>
  );
}

function useDebounced<T>(value: T, ms: number): T {
  const [v, setV] = useState(value);
  useEffect(() => {
    const t = setTimeout(() => setV(value), ms);
    return () => clearTimeout(t);
  }, [value, ms]);
  return v;
}
