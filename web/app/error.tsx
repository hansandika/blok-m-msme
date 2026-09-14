"use client";

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <div className="py-16">
      <h2 className="font-display text-3xl">Something on this page failed</h2>
      <p className="mt-3 max-w-lg text-sm text-mist">
        {error.message || "The API may be down. Start the Go server on :8080 and retry."}
      </p>
      <button
        type="button"
        onClick={reset}
        className="mt-6 rounded-full bg-gold px-4 py-2 text-sm text-ink"
      >
        Retry
      </button>
    </div>
  );
}
