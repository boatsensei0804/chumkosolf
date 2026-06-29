// เทอมทำงานที่ผู้ใช้เลือก (override) — เก็บใน localStorage แล้วส่งเป็น header X-Semester-Id
// ค่าว่าง = ใช้เทอมปัจจุบันตาม token (ไม่ override)
const KEY = "chumko.semesterOverride";

// safeStorage คืน localStorage ถ้าใช้ได้ (กัน env ที่ไม่มี window/localStorage เช่น SSR/test)
function safeStorage(): Storage | null {
  try {
    if (typeof window === "undefined" || !window.localStorage) return null;
    return window.localStorage;
  } catch {
    return null;
  }
}

export function getSelectedSemesterId(): string {
  return safeStorage()?.getItem(KEY) ?? "";
}

export function setSelectedSemesterId(id: string): void {
  const s = safeStorage();
  if (!s) return;
  if (id === "") s.removeItem(KEY);
  else s.setItem(KEY, id);
}
