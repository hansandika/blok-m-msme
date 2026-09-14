import Link from "next/link";
import { notFound } from "next/navigation";
import { FavoriteButton } from "@/components/FavoriteButton";
import { PlaceMapClient } from "@/components/PlaceMapClient";
import { fetchPlace } from "@/lib/api";
import { vibeLabel } from "@/lib/labels";

export const dynamic = "force-dynamic";

export default async function PlacePage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  let place;
  try {
    place = await fetchPlace(id);
  } catch (err) {
    if (err instanceof Error && err.message === "not-found") notFound();
    throw err;
  }

  return (
    <article className="grid gap-8 lg:grid-cols-[minmax(0,1fr)_minmax(0,1fr)]">
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
