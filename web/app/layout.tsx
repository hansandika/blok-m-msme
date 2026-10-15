import type { Metadata } from "next";
import { Fraunces, Figtree } from "next/font/google";
import "./globals.css";
import { Header } from "@/components/Header";
import { SITE_URL } from "@/lib/api";

const display = Fraunces({
  subsets: ["latin"],
  variable: "--font-display",
});

const sans = Figtree({
  subsets: ["latin"],
  variable: "--font-sans",
});

export const metadata: Metadata = {
  metadataBase: new URL(SITE_URL),
  title: "Blok M Lokal — MSME discovery",
  openGraph: {
    title: "Blok M Lokal — kopi, izakaya & kaki lima around Blok M",
    description: "Hand-curated map of local cafes, izakayas, street food and Hub tenants in Blok M, Jakarta.",
    siteName: "Blok M Lokal",
    type: "website",
    locale: "en_ID",
  },
  description:
    "Opinionated map of cafes, izakayas, street food and Hub tenants around Blok M, Melawai, M Bloc, and Blok M Hub. Not a mega-directory.",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className={`${display.variable} ${sans.variable} font-sans antialiased bg-night text-paper`}>
        <a
          href="#main"
          className="sr-only focus:not-sr-only focus:fixed focus:left-4 focus:top-4 focus:z-50 focus:rounded-full focus:bg-gold focus:px-4 focus:py-2 focus:text-ink"
        >
          Skip to content
        </a>
        <Header />
        <main id="main" className="mx-auto w-full max-w-7xl px-4 pb-16 pt-4 sm:px-6">{children}</main>
      </body>
    </html>
  );
}
