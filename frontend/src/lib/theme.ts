import type { ThemeConfig } from "antd";

// antd theme token — พาเลต #2563EB ตาม ui-ux-pro-max (Data-Dense Dashboard)
// ตรงกับ tailwind.config.ts และ design-system/MASTER.md
export const antdTheme: ThemeConfig = {
  token: {
    colorPrimary: "#0369A1",
    colorInfo: "#0369A1",
    colorLink: "#0369A1",
    colorLinkHover: "#0EA5E9",
    colorSuccess: "#059669",
    colorError: "#DC2626",
    colorBgLayout: "#f8fafc",
    colorBorderSecondary: "#e2e8f0",
    colorTextHeading: "#020617",
    borderRadius: 8,
    controlHeight: 36,
    fontFamily: "var(--font-sans), system-ui, sans-serif",
    fontSize: 14,
  },
  components: {
    Button: { fontWeight: 500, primaryShadow: "none", defaultShadow: "none" },
    Card: { boxShadowTertiary: "0 1px 2px rgba(15,23,42,0.04)" },
    Table: { headerBg: "#f8fafc", headerColor: "#475569", borderColor: "#e4ecfc" },
    Tag: { borderRadiusSM: 6 },
    Modal: { borderRadiusLG: 14 },
  },
};
