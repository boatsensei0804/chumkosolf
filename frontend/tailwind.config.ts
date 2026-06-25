import type { Config } from "tailwindcss";

// design token กลาง — Government/Public Service palette (ui-ux-pro-max)
// navy + professional blue, high-contrast, data-dense
const config: Config = {
  content: ["./app/**/*.{ts,tsx}", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        brand: {
          DEFAULT: "#0369A1", // accent — ปุ่ม/link/active (professional blue)
          deep: "#075985", // hover/เข้ม
          bright: "#0EA5E9", // highlight
          navy: "#0F172A", // dark sidebar / surface เข้ม (primary)
          slate: "#334155", // secondary
        },
      },
      fontFamily: {
        sans: ["var(--font-sans)", "system-ui", "sans-serif"],
        mono: ["var(--font-mono)", "ui-monospace", "monospace"],
      },
    },
  },
  plugins: [],
};

export default config;
