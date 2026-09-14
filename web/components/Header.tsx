import Link from "next/link";

export function Header() {
  return (
    <header className="sticky top-0 z-30 border-b border-white/10 bg-night/85 backdrop-blur-md">
      <div className="mx-auto flex max-w-7xl items-center justify-between gap-4 px-4 py-3 sm:px-6">
        <Link href="/" className="group min-w-0">
          <p className="font-display text-[11px] uppercase tracking-[0.28em] text-gold">Kebayoran Baru</p>
          <h1 className="font-display text-2xl leading-none text-paper sm:text-[1.7rem]">
            Blok M <span className="text-gold">Lokal</span>
          </h1>
        </Link>
        <nav className="flex items-center gap-1 text-sm sm:gap-2">
          <Link
            href="/admin"
            className="rounded-full px-3 py-1.5 text-mist transition hover:bg-white/5 hover:text-paper"
          >
            Admin
          </Link>
          <Link
            href="/favorites"
            className="rounded-full px-3 py-1.5 text-mist transition hover:bg-white/5 hover:text-paper"
          >
            Favorites
          </Link>
          <Link
            href="/contribute"
            className="rounded-full bg-gold px-3 py-1.5 font-medium text-ink transition hover:bg-[#e0b12a]"
          >
            Suggest
          </Link>
        </nav>
      </div>
    </header>
  );
}
