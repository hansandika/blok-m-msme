import type { Metadata } from "next";
import { Fraunces, Figtree } from "next/font/google";
import "./globals.css";
import { Header } from "@/components/Header";

const display = Fraunces({
  subsets: ["latin"],
  variable: "--font-display",
});

const sans = Figtree({
  subsets: ["latin"],
  variable: "--font-sans",
});

export const metadata: Metadata = {
  title: "Blok M Lokal — MSME discovery",
  description:
    "Opinionated map of cafes, izakayas, street food and Hub tenants around Blok M, Melawai, M Bloc, and Blok M Hub. Not a mega-directory.",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className={`${display.variable} ${sans.variable} font-sans antialiased bg-night text-paper`}>
        <Header />
        <main className="mx-auto w-full max-w-7xl px-4 pb-16 pt-4 sm:px-6">{children}</main>
      </body>
    </html>
  );
}
