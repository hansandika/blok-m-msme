import { Suspense } from "react";
import { ContributeForm } from "@/components/ContributeForm";

export default function ContributePage() {
  return (
    <div className="max-w-2xl">
      <p className="text-xs uppercase tracking-[0.22em] text-gold">Community · moderated</p>
      <h2 className="mt-2 font-display text-4xl">Suggest a place</h2>
      <p className="mt-3 text-sm leading-relaxed text-mist">
        Know a quiet izakaya off Melawai 6, or a Hub kiosk the seed missed? Send it in. Live map
        listings stay curated until someone reviews the queue.
      </p>
      <div className="mt-8">
        <Suspense fallback={<p className="text-sm text-mist">Loading form…</p>}>
          <ContributeForm />
        </Suspense>
      </div>
    </div>
  );
}
