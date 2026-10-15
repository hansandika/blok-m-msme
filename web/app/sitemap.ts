import type { MetadataRoute } from "next";
import { SITE_URL, fetchPlaces } from "@/lib/api";

export const revalidate = 3600;

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const base: MetadataRoute.Sitemap = [
    { url: `${SITE_URL}/`, changeFrequency: "daily", priority: 1 },
    { url: `${SITE_URL}/contribute`, changeFrequency: "monthly", priority: 0.3 },
  ];
  try {
    const { places } = await fetchPlaces();
    return base.concat(
      places.map((p) => ({ url: `${SITE_URL}/places/${p.slug}`, changeFrequency: "weekly" as const, priority: 0.7 })),
    );
  } catch {
    return base;
  }
}
