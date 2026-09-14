"use client";

import { useEffect, useMemo, useRef } from "react";
import { MapContainer, Marker, Popup, TileLayer, useMap } from "react-leaflet";
import L from "leaflet";
import "leaflet/dist/leaflet.css";
import type { Place } from "@/lib/types";
import { CATEGORY_COLOR } from "@/lib/labels";

const CENTER: [number, number] = [-6.2442, 106.7995];
const MAPBOX_TOKEN = process.env.NEXT_PUBLIC_MAPBOX_TOKEN;

function divIcon(place: Place, selected: boolean) {
  const color = CATEGORY_COLOR[place.category];
  const size = selected ? 18 : 12;
  return L.divIcon({
    className: "blokm-pin",
    html: `<span style="
      display:block;width:${size}px;height:${size}px;border-radius:999px;
      background:${color};border:2px solid ${selected ? "#f3ead8" : "rgba(18,20,16,0.85)"};
      box-shadow:0 0 0 ${selected ? 4 : 0}px rgba(212,160,23,0.35);
    "></span>`,
    iconSize: [size, size],
    iconAnchor: [size / 2, size / 2],
  });
}

function FitPlaces({ places, selected }: { places: Place[]; selected?: string | null }) {
  const map = useMap();
  const key = useMemo(
    () => places.map((p) => p.slug).join(",") + "|" + (selected ?? ""),
    [places, selected],
  );

  useEffect(() => {
    if (selected) {
      const hit = places.find((p) => p.slug === selected);
      if (hit) {
        map.flyTo([hit.lat, hit.lng], Math.max(map.getZoom(), 16), { duration: 0.6 });
        return;
      }
    }
    if (places.length === 0) {
      map.setView(CENTER, 15);
      return;
    }
    if (places.length === 1) {
      map.setView([places[0].lat, places[0].lng], 16);
      return;
    }
    const b = L.latLngBounds(places.map((p) => [p.lat, p.lng] as [number, number]));
    map.fitBounds(b.pad(0.18), { animate: true, maxZoom: 16 });
  }, [key, map, places, selected]);

  return null;
}

function LeafletMap({
  places,
  selected,
  onSelect,
}: {
  places: Place[];
  selected?: string | null;
  onSelect?: (slug: string) => void;
}) {
  return (
    <MapContainer
      center={CENTER}
      zoom={15}
      className="h-full w-full rounded-2xl"
      scrollWheelZoom
    >
      <TileLayer
        attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
        url="https://tile.openstreetmap.org/{z}/{x}/{y}.png"
      />
      <FitPlaces places={places} selected={selected} />
      {places.map((place) => (
        <Marker
          key={place.slug}
          position={[place.lat, place.lng]}
          icon={divIcon(place, place.slug === selected)}
          eventHandlers={{ click: () => onSelect?.(place.slug) }}
        >
          <Popup>
            <div className="min-w-40">
              <p className="text-[10px] uppercase tracking-wider text-mist">{place.neighborhood}</p>
              <p className="font-display text-base leading-tight">{place.name}</p>
              <p className="mt-1 text-xs text-mist">
                {place.category} · {place.price}
                {place.openLate ? " · larut" : ""}
              </p>
            </div>
          </Popup>
        </Marker>
      ))}
    </MapContainer>
  );
}

function MapboxMap({
  places,
  selected,
  onSelect,
}: {
  places: Place[];
  selected?: string | null;
  onSelect?: (slug: string) => void;
}) {
  const ref = useRef<HTMLDivElement>(null);
  const mapRef = useRef<import("mapbox-gl").Map | null>(null);
  const markers = useRef<import("mapbox-gl").Marker[]>([]);

  useEffect(() => {
    let cancelled = false;
    async function boot() {
      const mapboxgl = (await import("mapbox-gl")).default;
      await import("mapbox-gl/dist/mapbox-gl.css");
      if (!ref.current || cancelled) return;
      mapboxgl.accessToken = MAPBOX_TOKEN as string;
      const map = new mapboxgl.Map({
        container: ref.current,
        style: "mapbox://styles/mapbox/dark-v11",
        center: [CENTER[1], CENTER[0]],
        zoom: 14.4,
      });
      map.addControl(new mapboxgl.NavigationControl({ showCompass: false }), "top-right");
      mapRef.current = map;
    }
    void boot();
    return () => {
      cancelled = true;
      mapRef.current?.remove();
      mapRef.current = null;
    };
  }, []);

  useEffect(() => {
    const map = mapRef.current;
    if (!map) return;
    let cancelled = false;
    async function draw() {
      const mapboxgl = (await import("mapbox-gl")).default;
      if (cancelled || !mapRef.current) return;
      markers.current.forEach((m) => m.remove());
      markers.current = [];
      for (const place of places) {
        const el = document.createElement("button");
        el.type = "button";
        el.style.cssText = `
          width:${place.slug === selected ? 16 : 11}px;height:${place.slug === selected ? 16 : 11}px;
          border-radius:999px;border:2px solid #f3ead8;background:${CATEGORY_COLOR[place.category]};
          cursor:pointer;padding:0;
        `;
        el.title = place.name;
        el.addEventListener("click", () => onSelect?.(place.slug));
        const marker = new mapboxgl.Marker({ element: el }).setLngLat([place.lng, place.lat]).addTo(mapRef.current!);
        markers.current.push(marker);
      }
      if (selected) {
        const hit = places.find((p) => p.slug === selected);
        if (hit) mapRef.current.easeTo({ center: [hit.lng, hit.lat], zoom: 16 });
      }
    }
    if (map.isStyleLoaded()) void draw();
    else map.once("load", () => void draw());
    return () => {
      cancelled = true;
    };
  }, [places, selected, onSelect]);

  return <div ref={ref} className="h-full w-full rounded-2xl" />;
}

export default function PlaceMap(props: {
  places: Place[];
  selected?: string | null;
  onSelect?: (slug: string) => void;
}) {
  if (MAPBOX_TOKEN) {
    return <MapboxMap {...props} />;
  }
  return <LeafletMap {...props} />;
}
