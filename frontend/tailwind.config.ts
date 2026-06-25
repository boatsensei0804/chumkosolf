import type { Config } from "tailwindcss";

// design token กลาง (ตรงกับ antd ConfigProvider ใน src/lib/theme.ts และ design-system/MASTER.md)
// พาเลตฟ้า-ขาว #2563EB ตาม ui-ux-pro-max (Data-Dense Dashboard) + accent เขียว + status colors
const config: Config = {
  content: ["./app/**/*.{ts,tsx}", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        brand: {
          DEFAULT: "#2563EB", // primary — ปุ่ม/link/active
          deep: "#1E40AF", // ปลาย gradient / จุดเน้นสุด
          bright: "#3B82F6", // secondary / hover
          accent: "#059669", // accent เขียว (CTA รอง, success เน้น)
          cyan: "#00D4EB", // decorative (วงตกแต่ง hero)
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
