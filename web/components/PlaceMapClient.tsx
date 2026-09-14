"use client";

import dynamic from "next/dynamic";
import type { Place } from "@/lib/types";

const PlaceMap = dynamic(() => import("@/components/PlaceMap"), {
  ssr: false,
  loading: () => <div className="h-full min-h-[240px] animate-pulse rounded-2xl bg-ink-soft" />,
});

export function PlaceMapClient(props: {
  places: Place[];
  selected?: string | null;
  onSelect?: (slug: string) => void;
}) {
  return (
    <div className="h-full min-h-[240px] w-full">
      <PlaceMap {...props} />
    </div>
  );
}
