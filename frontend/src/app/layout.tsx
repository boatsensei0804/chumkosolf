import type { Metadata } from "next";
import { Fira_Code, Sarabun } from "next/font/google";
import type { ReactNode } from "react";

import { Providers } from "./providers";
import "./globals.css";

// ฟอนต์เนื้อหา/หัวข้อ: Sarabun (รองรับไทย) — ฟอนต์ตัวเลข/โค้ด: Fira Code (display เชิง technical)
const sarabun = Sarabun({
  subsets: ["thai", "latin"],
  weight: ["300", "400", "500", "600", "700"],
  variable: "--font-sans",
  display: "swap",
});

const firaCode = Fira_Code({
  subsets: ["latin"],
  weight: ["400", "500", "600"],
  variable: "--font-mono",
  display: "swap",
});

export const metadata: Metadata = {
  title: "chumkosoft",
  description: "ระบบบริหารจัดการโรงเรียน",
};

export default function RootLayout({
  children,
}: {
  children: ReactNode;
}): ReactNode {
  return (
    <html lang="th" className={`${sarabun.variable} ${firaCode.variable}`}>
      <body>
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
