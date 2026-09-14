import Link from "next/link";

export default function NotFound() {
  return (
    <div className="py-16">
      <h2 className="font-display text-4xl">Place not on the map</h2>
      <p className="mt-3 text-mist">It may have been renamed, or the seed never had it.</p>
      <Link href="/" className="mt-6 inline-block text-gold hover:underline">
        Back to Blok M
      </Link>
    </div>
  );
}
