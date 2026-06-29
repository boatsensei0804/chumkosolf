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
  // เห็นได้ถ้าอยู่ในกลุ่มงานใดกลุ่มหนึ่งในนี้ (สำหรับหน้ารวมหลายกลุ่ม เช่น "ข้อมูลบุคคล")
  requiredWorkGroupsAny?: WorkGroupCode[];
  // true = ไม่แสดงใน sidebar แต่ยังใช้ map เส้นทางสำหรับ route guard (เช่น /personnel ที่ถูกรวมไป /people)
  hidden?: boolean;
  // role ที่เห็นเมนูนี้ได้เพิ่มเติม (นอกเหนือกลุ่มงาน) เช่น ครูเห็นเมนูเช็คชื่อ
  allowRoles?: UserRole[];
  // เห็นเฉพาะ role เหล่านี้ (สำหรับหน้าส่วนตัว เช่น "ของฉัน" — ไม่ผูกกลุ่มงาน)
  visibleToRoles?: UserRole[];
  // true = แสดงใน dropdown ไอคอนบัญชี (มุมขวาบน) แทน sidebar — แต่ยังคุมสิทธิ์ด้วย route guard
  inAccountMenu?: boolean;
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
    key: "my_advisees",
    label: "นักเรียนที่ปรึกษาของฉัน",
    path: "/my-advisees",
    visibleToRoles: ["teacher", "executive"], // หน้าส่วนตัวของบุคลากร (ครู/ผู้บริหาร)
    available: true,
  },
  {
    key: "my_profile",
    label: "ข้อมูลของฉัน",
    path: "/my-profile",
    visibleToRoles: ["teacher", "executive"],
    inAccountMenu: true, // แสดงใต้ไอคอนบัญชี ไม่ใช่ sidebar
    available: true,
  },
  {
    key: "people",
    label: "ข้อมูลบุคคล",
    path: "/people",
    requiredWorkGroupsAny: ["personnel", "academic"], // กลุ่มบุคคล (บุคลากร) หรือกลุ่มวิชาการ (นักเรียน) — แท็บกรองตามสิทธิ์
    available: true,
  },
  // คงไว้แบบซ่อน เพื่อ map เส้นทาง /personnel, /students (+ sub-route) ให้ route guard ตรวจสิทธิ์รายกลุ่ม
  {
    key: "personnel",
    label: "บุคลากร",
    path: "/personnel",
    requiredWorkGroup: "personnel",
    hidden: true,
    available: true,
  },
  {
    key: "students",
    label: "นักเรียนและผู้ปกครอง",
    path: "/students",
    requiredWorkGroup: "academic",
    hidden: true,
    available: true,
  },
  {
    key: "classes",
    label: "ห้องเรียน",
    path: "/classes",
    requiredWorkGroup: "academic", // วิชาการ/แอดมิน จัดการได้
    allowRoles: ["teacher", "executive"], // ครู/ผู้บริหารอื่น ดู+ค้นหานักเรียนได้ (read-only)
    available: true,
  },
  {
    key: "timetable",
    label: "ตารางสอน",
    path: "/timetable",
    requiredWorkGroup: "academic",
    allowRoles: ["teacher"], // ครูดูตารางสอนได้ (จัดการเฉพาะกลุ่มวิชาการ)
    available: true,
  },
  {
    key: "kiosk",
    label: "สแกนหน้าเข้าเรียน",
    path: "/kiosk",
    requiredWorkGroup: "academic", // กลุ่มวิชาการ/แอดมิน
    allowRoles: ["kiosk"], // บัญชีเครื่องสแกนหน้า เข้าได้เฉพาะหน้านี้ (จัดการบัญชี kiosk อยู่ในแท็บของหน้านี้สำหรับแอดมิน)
    available: true,
  },
  {
    key: "subject_attendance",
    label: "เช็คชื่อ",
    path: "/subject-attendance",
    requiredWorkGroup: "general_affairs",
    allowRoles: ["teacher"],
    available: true,
  },
  {
    key: "attendance",
    label: "คะแนนความประพฤติ",
    path: "/attendance",
    requiredWorkGroup: "general_affairs",
    available: true,
  },
  {
    key: "settings",
    label: "ตั้งค่าระบบ",
    path: "/settings",
    requireSchoolAdmin: true,
    available: true,
  },
] as const;

// isSchoolAdmin: School Admin = super_admin หรือ is_school_admin (เห็นทุกกลุ่มงาน + ตั้งค่าระบบ)
export function isSchoolAdmin(user: UserInfo): boolean {
  return user.role === "super_admin" || user.is_school_admin;
}

// canManageTimetable: จัดการตารางสอน/มอบหมาย/รายวิชา ได้เฉพาะ school admin หรือสมาชิกกลุ่มวิชาการ
export function canManageTimetable(user: UserInfo): boolean {
  return isSchoolAdmin(user) || user.work_groups.some((g) => g.code === "academic");
}

// canSeeMenuItem ตัดสินว่าผู้ใช้เห็นเมนูนี้ไหม
export function canSeeMenuItem(item: MenuItemConfig, user: UserInfo): boolean {
  // School Admin เห็นทุกเมนู
  if (isSchoolAdmin(user)) return true;

  // เมนูเฉพาะ school admin → คนอื่นไม่เห็น
  if (item.requireSchoolAdmin) return false;

  // หน้าส่วนตัว (จำกัดตาม role เช่น "ของฉัน") → เห็นเฉพาะ role ที่ระบุ
  if (item.visibleToRoles) return item.visibleToRoles.includes(user.role);

  // role ที่อนุญาตพิเศษ (เช่น ครูเห็นเมนูเช็คชื่อ)
  if (item.allowRoles?.includes(user.role)) return true;

  // หน้ารวมหลายกลุ่ม (เช่น ข้อมูลบุคคล) → อยู่กลุ่มใดกลุ่มหนึ่งก็เห็น
  if (item.requiredWorkGroupsAny) {
    return user.work_groups.some((g) => item.requiredWorkGroupsAny!.includes(g.code));
  }

  // เมนูทั่วไป (ไม่ผูกกลุ่มงาน) เช่นหน้าแรก → เห็นได้ทุกคน
  if (!item.requiredWorkGroup) return true;

  // สมาชิกกลุ่มงานที่ตรงกัน
  return user.work_groups.some((g) => g.code === item.requiredWorkGroup);
}

// menuItemsForUser คืนเมนูที่ผู้ใช้มีสิทธิ์เห็น (กรองแล้ว)
export function menuItemsForUser(user: UserInfo): MenuItemConfig[] {
  return MENU_ITEMS.filter((item) => !item.hidden && canSeeMenuItem(item, user));
}

// menuItemForPath หา config ของเมนูที่ตรงกับ path ปัจจุบัน
// เลือก prefix ที่จำเพาะที่สุด (path ยาวสุด) เพื่อให้ sub-route เช่น /personnel/new map กับ "บุคลากร"
export function menuItemForPath(pathname: string): MenuItemConfig | undefined {
  if (pathname === "/") return MENU_ITEMS.find((i) => i.key === "home");
  return MENU_ITEMS.filter(
    (i) => i.path !== "/" && (pathname === i.path || pathname.startsWith(`${i.path}/`)),
  ).sort((a, b) => b.path.length - a.path.length)[0];
}

// canAccessPath ตัดสินว่าผู้ใช้เปิดหน้านี้ได้ไหม (ใช้กฎเดียวกับการมองเห็นเมนู)
// ใช้กันการเข้าถึงหน้าตรง ๆ ทาง URL แม้เมนูจะซ่อนลิงก์ไว้แล้ว
export function canAccessPath(user: UserInfo, pathname: string): boolean {
  const item = menuItemForPath(pathname);
  if (!item) return true; // ไม่มี mapping (หน้าไม่รู้จัก) → ปล่อยผ่าน (AuthGuard ครอบ auth อยู่แล้ว)
  return canSeeMenuItem(item, user);
}
