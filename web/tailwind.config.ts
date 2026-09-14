import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./app/**/*.{ts,tsx}", "./components/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        ink: {
          DEFAULT: "#121410",
          soft: "#1c211c",
          mid: "#2a322b",
        },
        paper: "#f3ead8",
        mist: "#b7b0a4",
        gold: "#d4a017",
        chili: "#d4512a",
        leaf: "#6f8f76",
        night: "#0e100d",
      },
      fontFamily: {
        display: ["var(--font-display)", "Georgia", "serif"],
        sans: ["var(--font-sans)", "system-ui", "sans-serif"],
      },
      boxShadow: {
        card: "0 18px 40px -24px rgba(0,0,0,0.55)",
      },
    },
  },
  plugins: [],
};

export default config;
