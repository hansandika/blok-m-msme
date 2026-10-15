"use client";

import { useState } from "react";

export function ShareActions({ name, lat, lng }: { name: string; lat: number; lng: number }) {
  const [copied, setCopied] = useState(false);

  async function share() {
    const url = window.location.href;
    const data = { title: `${name} · Blok M Lokal`, text: `${name} — Blok M`, url };
    try {
      if (navigator.share) {
        await navigator.share(data);
        return;
      }
      await navigator.clipboard.writeText(url);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      /* user cancelled */
    }
  }

  const directions = `https://www.google.com/maps/dir/?api=1&destination=${lat},${lng}`;

  return (
    <>
      <button
        type="button"
        onClick={share}
        className="rounded-full border border-white/15 px-3 py-1 text-xs text-paper hover:bg-white/5"
        aria-live="polite"
      >
        {copied ? "Link copied ✓" : "Share / Bagikan"}
      </button>
      <a
        href={directions}
        target="_blank"
        rel="noopener noreferrer"
        className="rounded-full border border-white/15 px-3 py-1 text-xs text-paper hover:bg-white/5"
      >
        Directions / Rute ↗
      </a>
    </>
  );
}
