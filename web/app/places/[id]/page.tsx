import type { Metadata } from "next";
import Link from "next/link";
import { cache } from "react";
import { notFound } from "next/navigation";
import { FavoriteButton } from "@/components/FavoriteButton";
import { PlaceMapClient } from "@/components/PlaceMapClient";
import { ShareActions } from "@/components/ShareActions";
import { SITE_URL, fetchPlace } from "@/lib/api";
import type { Place } from "@/lib/types";
import { vibeLabel } from "@/lib/labels";

export const revalidate = 300;

const loadPlace = cache(async (id: string): Promise<Place | null> => {
  try {
    return await fetchPlace(id);
  } catch (err) {
    if (err instanceof Error && err.message === "not-found") return null;
    throw err;
  }
});

export async function generateMetadata({ params }: { params: Promise<{ id: string }> }): Promise<Metadata> {
  const { id } = await params;
  const place = await loadPlace(id).catch(() => null);
  if (!place) return { title: "Place not found · Blok M Lokal" };
  const title = `${place.name} — ${place.neighborhood} · Blok M Lokal`;
  const description = place.description.slice(0, 160);
  const url = `${SITE_URL}/places/${place.slug}`;
  return {
    title,
    description,
    alternates: { canonical: url },
    openGraph: { title, description, url, siteName: "Blok M Lokal", type: "website", locale: "en_ID" },
    twitter: { card: "summary", title, description },
  };
}

function jsonLd(place: Place) {
  const type = place.category === "retail" ? "Store" : place.category === "cafe" ? "CafeOrCoffeeShop" : "Restaurant";
  return {
    "@context": "https://schema.org",
    "@type": type,
    name: place.name,
    description: place.description,
    url: `${SITE_URL}/places/${place.slug}`,
    priceRange: place.price,
    address: {
      "@type": "PostalAddress",
      streetAddress: place.address,
      addressLocality: "Jakarta Selatan",
      addressRegion: "DKI Jakarta",
      addressCountry: "ID",
    },
    geo: { "@type": "GeoCoordinates", latitude: place.lat, longitude: place.lng },
  };
}

export default async function PlacePage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const place = await loadPlace(id);
  if (!place) notFound();

  return (
    <article className="grid gap-8 lg:grid-cols-[minmax(0,1fr)_minmax(0,1fr)]">
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd(place)).replace(/</g, "\\u003c") }}
      />
      <div>
        <Link href="/" className="text-sm text-gold hover:underline">
          ← Map
        </Link>
        <p className="mt-4 text-xs uppercase tracking-[0.2em] text-mist">
          {place.neighborhood} · {place.price}
        </p>
        <h2 className="mt-2 font-display text-4xl sm:text-5xl">{place.name}</h2>
        <div className="mt-4 flex flex-wrap gap-2">
          <span className="rounded-full bg-white/10 px-3 py-1 text-xs capitalize">{place.category}</span>
          {place.openLate && (
            <span className="rounded-full bg-chili/20 px-3 py-1 text-xs text-[#f0b3a0]">Open late / larut</span>
          )}
          {place.vibes.map((v) => (
            <span key={v} className="rounded-full bg-white/5 px-3 py-1 text-xs text-mist">
              {vibeLabel(v)}
            </span>
          ))}
        </div>
        <p className="mt-6 max-w-xl text-base leading-relaxed text-paper/90">{place.description}</p>
        <dl className="mt-8 space-y-3 text-sm">
          <div>
            <dt className="text-mist">Approx. address</dt>
            <dd>{place.address}</dd>
          </div>
          {place.hours && (
            <div>
              <dt className="text-mist">Hours (if known)</dt>
              <dd>{place.hours}</dd>
            </div>
          )}
          <div>
            <dt className="text-mist">Source</dt>
            <dd className="text-mist">{place.sourceNote}</dd>
          </div>
        </dl>
        <div className="mt-8 flex flex-wrap gap-3">
          <FavoriteButton slug={place.slug} />
          <ShareActions name={place.name} lat={place.lat} lng={place.lng} />
          <Link
            href={`/contribute?kind=correction&placeId=${place.id}&name=${encodeURIComponent(place.name)}`}
            className="rounded-full border border-white/15 px-3 py-1 text-xs text-mist hover:text-paper"
          >
            Suggest a correction
          </Link>
        </div>
      </div>
      <div className="h-[360px] overflow-hidden rounded-2xl border border-white/10 lg:h-[520px]">
        <PlaceMapClient places={[place]} selected={place.slug} />
      </div>
    </article>
  );
}
