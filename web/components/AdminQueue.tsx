"use client";

import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import Link from "next/link";
import {
  clearAdminToken,
  fetchAdminSuggestions,
  moderateSuggestion,
  readAdminToken,
  writeAdminToken,
} from "@/lib/api";
import type { Suggestion } from "@/lib/types";

const FILTERS = ["", "pending", "approved", "rejected", "applied"] as const;

export function AdminQueue() {
  const [tokenInput, setTokenInput] = useState("");
  const [token, setToken] = useState("");
  const [status, setStatus] = useState<(typeof FILTERS)[number]>("pending");
  const [items, setItems] = useState<Suggestion[]>([]);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState<string | null>(null);
  const [flash, setFlash] = useState("");

  const load = useCallback(
    async (tok: string, filter: string) => {
      setError("");
      const data = await fetchAdminSuggestions(tok, filter);
      setItems(data.suggestions);
    },
    [],
  );

  useEffect(() => {
    const saved = readAdminToken();
    if (!saved) return;
    setToken(saved);
    setTokenInput(saved);
    load(saved, status).catch((err: Error) => setError(err.message));
  }, [load, status]);

  function onUnlock(e: FormEvent) {
    e.preventDefault();
    const next = tokenInput.trim();
    writeAdminToken(next);
    setToken(next);
    load(next, status).catch((err: Error) => setError(err.message));
  }

  function lock() {
    clearAdminToken();
    setToken("");
    setItems([]);
    setFlash("");
  }

  async function act(id: string, action: "approve" | "reject" | "apply") {
    setBusy(`${id}:${action}`);
    setFlash("");
    setError("");
    try {
      const res = await moderateSuggestion(token, id, action);
      setFlash(res.message || `${action} ok`);
      await load(token, status);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed");
    } finally {
      setBusy(null);
    }
  }

  const counts = useMemo(() => ({ n: items.length }), [items.length]);

  if (!token) {
    return (
      <form onSubmit={onUnlock} className="max-w-md space-y-4">
        <label className="block text-sm">
          Admin token
          <input
            className="mt-1 w-full rounded-2xl border border-white/10 bg-ink-soft px-4 py-3 text-sm outline-none ring-gold/40 focus:ring-2"
            value={tokenInput}
            onChange={(e) => setTokenInput(e.target.value)}
            placeholder="ADMIN_TOKEN (default demo: blokm-demo)"
            autoComplete="off"
            required
          />
        </label>
        <button type="submit" className="rounded-full bg-gold px-5 py-2.5 text-sm font-medium text-ink">
          Unlock queue
        </button>
        <p className="text-xs leading-relaxed text-mist">
          Token is stored in this browser&apos;s localStorage for the demo. Match{" "}
          <code className="text-paper/80">ADMIN_TOKEN</code> on the API (
          <code className="text-paper/80">make api</code> defaults to <code className="text-paper/80">blokm-demo</code>
          ).
        </p>
        {error && <p className="text-sm text-chili">{error}</p>}
      </form>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center gap-2">
        {FILTERS.map((id) => (
          <button
            key={id || "all"}
            type="button"
            onClick={() => setStatus(id)}
            className={`rounded-full px-3 py-1.5 text-sm ${
              status === id ? "bg-gold text-ink" : "border border-white/10 text-mist"
            }`}
          >
            {id || "all"}
          </button>
        ))}
        <button type="button" onClick={lock} className="ml-auto text-xs text-mist hover:text-paper">
          Lock
        </button>
      </div>

      {flash && <p className="text-sm text-leaf">{flash}</p>}
      {error && <p className="text-sm text-chili">{error}</p>}
      <p className="text-xs text-mist">{counts.n} in this filter</p>

      <ul className="space-y-3">
        {items.map((s) => {
          const applied = Boolean(s.appliedAt);
          return (
            <li key={s.id} className="rounded-2xl border border-white/10 bg-ink-soft/80 p-4">
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <p className="text-[11px] uppercase tracking-[0.18em] text-mist">
                    {s.kind.replace("_", " ")} · {applied ? "applied" : s.status}
                  </p>
                  <h3 className="mt-1 font-display text-xl">{s.name || "Correction"}</h3>
                </div>
                <div className="flex flex-wrap gap-2">
                  <button
                    type="button"
                    disabled={applied || s.status === "approved" || busy !== null}
                    onClick={() => act(s.id, "approve")}
                    className="rounded-full border border-white/15 px-3 py-1 text-xs disabled:opacity-40"
                  >
                    Approve
                  </button>
                  <button
                    type="button"
                    disabled={applied || s.status === "rejected" || busy !== null}
                    onClick={() => act(s.id, "reject")}
                    className="rounded-full border border-white/15 px-3 py-1 text-xs disabled:opacity-40"
                  >
                    Reject
                  </button>
                  <button
                    type="button"
                    disabled={s.status !== "approved" || busy !== null}
                    onClick={() => act(s.id, "apply")}
                    className="rounded-full bg-gold px-3 py-1 text-xs font-medium text-ink disabled:opacity-40"
                  >
                    {applied ? "Applied" : busy === `${s.id}:apply` ? "Applying…" : "Apply"}
                  </button>
                </div>
              </div>
              <p className="mt-3 whitespace-pre-wrap text-sm leading-relaxed text-paper/90">{s.notes}</p>
              <dl className="mt-3 grid gap-1 text-xs text-mist sm:grid-cols-2">
                <div>id {s.id}</div>
                {s.placeId && (
                  <div>
                    place{" "}
                    <Link href={`/places/${s.placeId}`} className="text-gold hover:underline">
                      {s.placeId}
                    </Link>
                  </div>
                )}
                {s.appliedPlaceId && (
                  <div>
                    live{" "}
                    <Link href={`/places/${s.appliedPlaceId}`} className="text-gold hover:underline">
                      {s.appliedPlaceId}
                    </Link>
                  </div>
                )}
                {s.category && <div>category {s.category}</div>}
                {s.lat != null && s.lng != null && (
                  <div>
                    pin {s.lat.toFixed(5)}, {s.lng.toFixed(5)}
                  </div>
                )}
              </dl>
            </li>
          );
        })}
      </ul>
      {items.length === 0 && (
        <p className="rounded-2xl border border-dashed border-white/15 p-8 text-sm text-mist">
          Nothing in this filter. Public contribute lands here as pending — it never writes live listings.
        </p>
      )}
    </div>
  );
}
