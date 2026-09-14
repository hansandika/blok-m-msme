import { AdminQueue } from "@/components/AdminQueue";

export default function AdminPage() {
  return (
    <div className="max-w-3xl">
      <p className="text-xs uppercase tracking-[0.22em] text-gold">Moderation · token gated</p>
      <h2 className="mt-2 font-display text-4xl">Suggestion queue</h2>
      <p className="mt-3 max-w-2xl text-sm leading-relaxed text-mist">
        Approve or reject community reports, then apply approved rows onto live places. Rejected
        suggestions never mutate the map. Applying twice is safe (idempotent).
      </p>
      <div className="mt-8">
        <AdminQueue />
      </div>
    </div>
  );
}
