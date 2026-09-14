export type Category = "japanese" | "street" | "cafe" | "retail" | "other";
export type Price = "$" | "$$" | "$$$";
export type Vibe =
  | "quiet"
  | "lively"
  | "hidden"
  | "hangout"
  | "date-night"
  | "after-office"
  | "quick-bite"
  | "family";

export type Place = {
  id: string;
  slug: string;
  name: string;
  category: Category;
  vibes: Vibe[];
  price: Price;
  openLate: boolean;
  lat: number;
  lng: number;
  neighborhood: string;
  address: string;
  hours?: string;
  description: string;
  sourceNote: string;
  aliases: string[];
};

export type PlaceListResponse = {
  places: Place[];
  count: number;
  geofence: { south: number; north: number; west: number; east: number };
};

export type PlaceFilters = {
  q?: string;
  category?: string;
  vibe?: string;
  price?: string;
  openLate?: string;
};
