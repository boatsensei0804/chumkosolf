// ระบบสี accent ต่อหมวด (color-coding) — class คงที่เพื่อให้ Tailwind ไม่ purge
// ใช้กับ KPI cards, ไอคอนเมนู, หัวข้อ section ให้แต่ละหมวดมีสีประจำ จำง่าย/เด่นชัด

export type Accent = "blue" | "emerald" | "violet" | "amber" | "rose" | "slate";

type AccentClasses = {
  chip: string; // พื้น+สีไอคอน (chip วงกลม/สี่เหลี่ยมมน)
  strip: string; // แถบสีด้านบนการ์ด (bg)
  solid: string; // พื้นเต็มสี + ตัวขาว (icon เด่น)
};

export const ACCENTS: Record<Accent, AccentClasses> = {
  blue: { chip: "bg-sky-50 text-sky-600", strip: "bg-sky-500", solid: "bg-sky-600 text-white" },
  emerald: { chip: "bg-emerald-50 text-emerald-600", strip: "bg-emerald-500", solid: "bg-emerald-600 text-white" },
  violet: { chip: "bg-violet-50 text-violet-600", strip: "bg-violet-500", solid: "bg-violet-600 text-white" },
  amber: { chip: "bg-amber-50 text-amber-600", strip: "bg-amber-500", solid: "bg-amber-500 text-white" },
  rose: { chip: "bg-rose-50 text-rose-600", strip: "bg-rose-500", solid: "bg-rose-600 text-white" },
  slate: { chip: "bg-slate-100 text-slate-600", strip: "bg-slate-400", solid: "bg-slate-600 text-white" },
};

// สีประจำของแต่ละเมนู/หมวดงาน
export const MODULE_ACCENT: Record<string, Accent> = {
  home: "blue",
  personnel: "blue",
  students: "emerald",
  attendance: "violet",
  budget: "amber",
  settings: "slate",
};

export function moduleAccent(key: string): Accent {
  return MODULE_ACCENT[key] ?? "blue";
}
