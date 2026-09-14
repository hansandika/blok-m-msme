"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";
import { useSearchParams } from "next/navigation";
import { fetchPlaces, submitSuggestion } from "@/lib/api";
import { CATEGORIES } from "@/lib/labels";
import type { Place } from "@/lib/types";

export function ContributeForm() {
  const params = useSearchParams();
  const [kind, setKind] = useState<"new_place" | "correction">(
    params.get("kind") === "correction" ? "correction" : "new_place",
  );
  const [places, setPlaces] = useState<Place[]>([]);
  const [placeId, setPlaceId] = useState(params.get("placeId") ?? "");
  const [name, setName] = useState(params.get("name") ?? "");
  const [category, setCategory] = useState("");
  const [notes, setNotes] = useState("");
  const [lat, setLat] = useState("");
  const [lng, setLng] = useState("");
  const [status, setStatus] = useState<"idle" | "saving" | "ok" | "err">("idle");
  const [message, setMessage] = useState("");

  useEffect(() => {
    fetchPlaces()
      .then((d) => setPlaces(d.places))
      .catch(() => undefined);
  }, []);

  const selectedName = useMemo(
    () => places.find((p) => p.id === placeId)?.name,
    [places, placeId],
  );

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setStatus("saving");
    try {
      const latN = lat.trim() ? Number(lat) : NaN;
      const lngN = lng.trim() ? Number(lng) : NaN;
      const data = await submitSuggestion({
        kind,
        placeId: kind === "correction" ? placeId : undefined,
        name: kind === "new_place" ? name : selectedName || name,
        category: category || undefined,
        notes,
        lat: Number.isFinite(latN) ? latN : undefined,
        lng: Number.isFinite(lngN) ? lngN : undefined,
      });
      setStatus("ok");
      setMessage(data.message);
      setNotes("");
    } catch (err) {
      setStatus("err");
      setMessage(err instanceof Error ? err.message : "Failed");
    }
  }

  const field =
    "w-full rounded-2xl border border-white/10 bg-ink-soft px-4 py-3 text-sm outline-none ring-gold/40 focus:ring-2";

  return (
    <form onSubmit={onSubmit} className="max-w-xl space-y-4">
      <fieldset className="flex gap-2">
        <legend className="sr-only">Suggestion type</legend>
        {(
          [
            ["new_place", "Suggest a place"],
            ["correction", "Suggest a correction"],
          ] as const
        ).map(([id, label]) => (
          <button
            key={id}
            type="button"
            onClick={() => setKind(id)}
            className={`rounded-full px-3 py-1.5 text-sm ${
              kind === id ? "bg-gold text-ink" : "border border-white/10 text-mist"
            }`}
          >
            {label}
          </button>
        ))}
      </fieldset>

      {kind === "new_place" ? (
        <label className="block text-sm">
          Name
          <input className={`${field} mt-1`} value={name} onChange={(e) => setName(e.target.value)} required />
        </label>
      ) : (
        <label className="block text-sm">
          Existing place
          <select
            className={`${field} mt-1`}
            value={placeId}
            onChange={(e) => setPlaceId(e.target.value)}
            required
          >
            <option value="">Choose…</option>
            {places.map((p) => (
              <option key={p.id} value={p.id}>
                {p.name} — {p.neighborhood}
              </option>
            ))}
          </select>
        </label>
      )}

      <label className="block text-sm">
        Category (optional)
        <select className={`${field} mt-1`} value={category} onChange={(e) => setCategory(e.target.value)}>
          <option value="">—</option>
          {CATEGORIES.map((c) => (
            <option key={c.id} value={c.id}>
              {c.label}
            </option>
          ))}
        </select>
      </label>

      <label className="block text-sm">
        Notes
        <textarea
          className={`${field} mt-1 min-h-28`}
          value={notes}
          onChange={(e) => setNotes(e.target.value)}
          required
          placeholder="Hours wrong? New stall on Melawai 6? Quiet after 9?"
        />
      </label>

      <div className="grid grid-cols-2 gap-3">
        <label className="block text-sm">
          Lat (optional)
          <input className={`${field} mt-1`} value={lat} onChange={(e) => setLat(e.target.value)} />
        </label>
        <label className="block text-sm">
          Lng (optional)
          <input className={`${field} mt-1`} value={lng} onChange={(e) => setLng(e.target.value)} />
        </label>
      </div>

      <button
        type="submit"
        disabled={status === "saving"}
        className="rounded-full bg-gold px-5 py-2.5 text-sm font-medium text-ink disabled:opacity-60"
      >
        {status === "saving" ? "Sending…" : "Send to moderation queue"}
      </button>

      {status !== "idle" && status !== "saving" && (
        <p className={status === "ok" ? "text-sm text-leaf" : "text-sm text-chili"}>{message}</p>
      )}
      <p className="text-xs leading-relaxed text-mist">
        Submissions land in a pending queue. They do not write to live listings. Community edits stay
        reviewed — Google Places sync is a later hook, not this MVP.
      </p>
    </form>
  );
}
