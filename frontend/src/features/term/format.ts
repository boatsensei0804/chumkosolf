// formatTerm สร้างข้อความปี/เทอมจากค่า has_current (ใช้แสดงผลทั่วไป)
export function formatTerm(t: { has_current: boolean; academic_year: number; term: number }): string {
  if (!t.has_current) return "ยังไม่ได้กำหนดเทอมปัจจุบัน";
  return `ปีการศึกษา ${t.academic_year} · ภาคเรียนที่ ${t.term}`;
}
