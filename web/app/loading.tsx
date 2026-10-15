export default function Loading() {
  return (
    <div className="space-y-4" role="status" aria-live="polite">
      <span className="sr-only">Loading…</span>
      <div className="h-10 w-2/3 animate-pulse rounded-xl bg-white/5" />
      <div className="h-4 w-1/2 animate-pulse rounded bg-white/5" />
      <div className="grid gap-4 lg:grid-cols-2">
        <div className="h-[42vh] animate-pulse rounded-2xl bg-white/5" />
        <div className="space-y-3">
          {[0, 1, 2, 3].map((i) => (
            <div key={i} className="h-24 animate-pulse rounded-2xl bg-white/5" />
          ))}
        </div>
      </div>
    </div>
  );
}
