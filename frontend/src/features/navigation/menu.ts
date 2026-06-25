import type { UserRole, WorkGroupCode } from "@/shared/schemas/enums";
import type { UserInfo } from "@/shared/schemas/auth";

// นิยามเมนู dashboard + กฎสิทธิ์การมองเห็น (business logic แยกจาก component)
// สิทธิ์ 3 ชั้นตาม CLAUDE.md ข้อ 4.9:
//   1) School Admin (super_admin หรือ is_school_admin) เห็นทุกเมนู
//   2/3) Group Admin / Member เห็นตามกลุ่มงานที่สังกัด
//   สิทธิ์ครู (role) เห็นเมนูบาง action ได้แม้ไม่สังกัดกลุ่มงานนั้น

export type MenuItemConfig = {
  key: string;
  label: string;
  // path ของหน้า (ถ้ามีหน้าจริงแล้ว)
  path: string;
  // กลุ่มงานที่เข้าถึงเมนูนี้ได้ (ไม่ระบุ = ทุกคนที่ล็อกอินเห็นได้)
  requiredWorkGroup?: WorkGroupCode;
  // role ที่เห็นเมนูนี้ได้เพิ่มเติม (นอกเหนือกลุ่มงาน) เช่น ครูเห็นเมนูเช็คชื่อ
  allowRoles?: UserRole[];
  // เห็นเฉพาะ school admin เท่านั้น
  requireSchoolAdmin?: boolean;
  // false = ยังไม่มีหน้าจริงในเฟสนี้ (แสดงแบบ disabled "เร็ว ๆ นี้")
  available: boolean;
};

// เมนูทั้งหมด — Phase 1 มีเฉพาะ "หน้าแรก" ที่ใช้งานได้จริง
// เมนูกลุ่มงานอื่นแสดงเพื่อสะท้อนสิทธิ์ แต่ยัง disabled จนกว่าจะถึงเฟสนั้น
export const MENU_ITEMS: readonly MenuItemConfig[] = [
  { key: "home", label: "หน้าแรก", path: "/", available: true },
  {
    key: "personnel",
    label: "บุคลากร",
    path: "/personnel",
    requiredWorkGroup: "personnel",
    available: true,
  },
  {
    key: "students",
    label: "นักเรียนและผู้ปกครอง",
    path: "/students",
    requiredWorkGroup: "academic",
    available: true,
  },
  {
    key: "attendance",
    label: "เช็คชื่อและความประพฤติ",
    path: "/attendance",
    requiredWorkGroup: "general_affairs",
    allowRoles: ["teacher"],
    available: false,
  },
  {
    key: "budget",
    label: "งบประมาณและแผน",
    path: "/budget",
    requiredWorkGroup: "budget_plan",
    available: false,
  },
  {
    key: "settings",
    label: "ตั้งค่าระบบ",
    path: "/settings",
    requireSchoolAdmin: true,
    available: false,
  },
] as const;

// isSchoolAdmin: School Admin = super_admin หรือ is_school_admin (เห็นทุกกลุ่มงาน + ตั้งค่าระบบ)
export function isSchoolAdmin(user: UserInfo): boolean {
  return user.role === "super_admin" || user.is_school_admin;
}

// canSeeMenuItem ตัดสินว่าผู้ใช้เห็นเมนูนี้ไหม
export function canSeeMenuItem(item: MenuItemConfig, user: UserInfo): boolean {
  // School Admin เห็นทุกเมนู
  if (isSchoolAdmin(user)) return true;

  // เมนูเฉพาะ school admin → คนอื่นไม่เห็น
  if (item.requireSchoolAdmin) return false;

  // เมนูทั่วไป (ไม่ผูกกลุ่มงาน) เช่นหน้าแรก → เห็นได้ทุกคน
  if (!item.requiredWorkGroup) return true;

  // role ที่อนุญาตพิเศษ (เช่น ครูเห็นเมนูเช็คชื่อ)
  if (item.allowRoles?.includes(user.role)) return true;

  // สมาชิกกลุ่มงานที่ตรงกัน
  return user.work_groups.some((g) => g.code === item.requiredWorkGroup);
}

// menuItemsForUser คืนเมนูที่ผู้ใช้มีสิทธิ์เห็น (กรองแล้ว)
export function menuItemsForUser(user: UserInfo): MenuItemConfig[] {
  return MENU_ITEMS.filter((item) => canSeeMenuItem(item, user));
}
