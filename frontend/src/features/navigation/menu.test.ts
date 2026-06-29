import { describe, expect, it } from "vitest";

import type { UserInfo } from "@/shared/schemas/auth";
import type { WorkGroupCode } from "@/shared/schemas/enums";

import { canAccessPath, menuItemForPath, menuItemsForUser } from "./menu";

function makeUser(overrides: Partial<UserInfo> = {}): UserInfo {
  return {
    id: "u1",
    username: "user",
    role: "teacher",
    school_id: "s1",
    is_school_admin: false,
    work_groups: [],
    ...overrides,
  };
}

function workGroup(code: WorkGroupCode): UserInfo["work_groups"][number] {
  return { work_group_id: `wg-${code}`, code, name: code, is_group_admin: false };
}

function keys(user: UserInfo): string[] {
  return menuItemsForUser(user).map((i) => i.key);
}

describe("menuItemsForUser", () => {
  it("super_admin เห็นทุกเมนู", () => {
    const user = makeUser({ role: "super_admin" });
    expect(keys(user)).toEqual([
      "home",
      "my_advisees",
      "my_profile",
      "people",
      "classes",
      "timetable",
      "kiosk",
      "subject_attendance",
      "attendance",
      "settings",
    ]);
  });

  it("is_school_admin เห็นทุกเมนูรวมตั้งค่าระบบ", () => {
    const user = makeUser({ role: "teacher", is_school_admin: true });
    expect(keys(user)).toContain("settings");
    expect(keys(user)).toContain("people");
  });

  it("ครูไม่มีกลุ่มงาน เห็นหน้าแรก + เช็คชื่อ (สิทธิ์ครู) แต่ไม่เห็นบุคคล/วิชาการ/ตั้งค่า", () => {
    const user = makeUser({ role: "teacher", work_groups: [] });
    const k = keys(user);
    expect(k).toContain("home");
    expect(k).toContain("subject_attendance"); // เมนู "เช็คชื่อ" — allowRoles teacher
    expect(k).not.toContain("personnel");
    expect(k).not.toContain("students");
    expect(k).not.toContain("settings");
  });

  it("สมาชิกกลุ่มวิชาการ เห็นเมนูข้อมูลบุคคล + สแกนหน้า", () => {
    const user = makeUser({ role: "teacher", work_groups: [workGroup("academic")] });
    const k = keys(user);
    expect(k).toContain("people");
    expect(k).toContain("kiosk");
  });

  it("ครูทั่วไป (ไม่สังกัดกลุ่ม) ไม่เห็นเมนูสแกนหน้า", () => {
    expect(keys(makeUser({ role: "teacher", work_groups: [] }))).not.toContain("kiosk");
  });

  it("บัญชี kiosk เห็นแค่หน้าสแกน เข้าหน้าอื่นไม่ได้", () => {
    const k = makeUser({ role: "kiosk", work_groups: [] });
    expect(keys(k)).toContain("kiosk");
    expect(keys(k)).not.toContain("students");
    expect(canAccessPath(k, "/kiosk")).toBe(true);
    expect(canAccessPath(k, "/students")).toBe(false);
  });

  it("สมาชิกกลุ่มบุคคล เห็นเมนูข้อมูลบุคคล", () => {
    const user = makeUser({ role: "executive", work_groups: [workGroup("personnel")] });
    const k = keys(user);
    expect(k).toContain("people");
    expect(canAccessPath(user, "/personnel")).toBe(true); // เข้าหน้าบุคลากรได้ (สิทธิ์กลุ่มบุคคล)
    expect(canAccessPath(user, "/students")).toBe(false); // แต่เข้าหน้านักเรียนไม่ได้
  });

  it("ทุกคนต้องเห็นหน้าแรกเสมอ", () => {
    expect(keys(makeUser({ role: "student" }))).toContain("home");
  });

  it("ครู/ผู้บริหาร เห็นหน้าส่วนตัว (ของฉัน) แต่นักเรียนไม่เห็น", () => {
    const teacher = keys(makeUser({ role: "teacher", work_groups: [] }));
    expect(teacher).toContain("my_advisees");
    expect(teacher).toContain("my_profile");
    expect(keys(makeUser({ role: "executive" }))).toContain("my_profile");
    const student = keys(makeUser({ role: "student" }));
    expect(student).not.toContain("my_advisees");
    expect(student).not.toContain("my_profile");
  });
});

describe("menuItemForPath", () => {
  it("map sub-route ไปยังเมนูที่จำเพาะที่สุด", () => {
    expect(menuItemForPath("/personnel/new")?.key).toBe("personnel");
    expect(menuItemForPath("/students/123/edit")?.key).toBe("students");
    expect(menuItemForPath("/subject-attendance")?.key).toBe("subject_attendance");
  });

  it("path '/' → หน้าแรก", () => {
    expect(menuItemForPath("/")?.key).toBe("home");
  });

  it("path ที่ไม่รู้จัก → undefined", () => {
    expect(menuItemForPath("/unknown-page")).toBeUndefined();
  });
});

describe("canAccessPath", () => {
  it("ครูไม่มีกลุ่มงาน เข้าหน้านักเรียน/บุคลากร/ตั้งค่าตรง ๆ ทาง URL ไม่ได้", () => {
    const user = makeUser({ role: "teacher", work_groups: [] });
    expect(canAccessPath(user, "/students")).toBe(false);
    expect(canAccessPath(user, "/students/123/edit")).toBe(false);
    expect(canAccessPath(user, "/personnel")).toBe(false);
    expect(canAccessPath(user, "/settings")).toBe(false);
  });

  it("ครูเข้าหน้าแรก + ตารางสอน + เช็คชื่อได้ (สิทธิ์ครู)", () => {
    const user = makeUser({ role: "teacher", work_groups: [] });
    expect(canAccessPath(user, "/")).toBe(true);
    expect(canAccessPath(user, "/timetable")).toBe(true);
    expect(canAccessPath(user, "/subject-attendance")).toBe(true);
  });

  it("สมาชิกกลุ่มวิชาการเข้าหน้านักเรียนได้ แต่บุคลากรไม่ได้", () => {
    const user = makeUser({ role: "teacher", work_groups: [workGroup("academic")] });
    expect(canAccessPath(user, "/students")).toBe(true);
    expect(canAccessPath(user, "/personnel")).toBe(false);
  });

  it("school admin เข้าได้ทุกหน้า", () => {
    const user = makeUser({ role: "teacher", is_school_admin: true });
    expect(canAccessPath(user, "/personnel")).toBe(true);
    expect(canAccessPath(user, "/settings")).toBe(true);
  });

  it("หน้าที่ไม่มี mapping → ปล่อยผ่าน (auth ครอบอยู่แล้ว)", () => {
    const user = makeUser({ role: "teacher", work_groups: [] });
    expect(canAccessPath(user, "/profile")).toBe(true);
  });

  it("ครูเข้าหน้าส่วนตัวของฉันได้ แต่นักเรียนเข้าไม่ได้", () => {
    const teacher = makeUser({ role: "teacher", work_groups: [] });
    expect(canAccessPath(teacher, "/my-advisees")).toBe(true);
    expect(canAccessPath(teacher, "/my-profile")).toBe(true);
    const student = makeUser({ role: "student" });
    expect(canAccessPath(student, "/my-advisees")).toBe(false);
    expect(canAccessPath(student, "/my-profile")).toBe(false);
  });

  it("ครูเข้าหน้าห้องเรียน (ดู+ค้นหานักเรียน) ได้ แต่นักเรียนเข้าไม่ได้", () => {
    expect(canAccessPath(makeUser({ role: "teacher", work_groups: [] }), "/classes")).toBe(true);
    expect(canAccessPath(makeUser({ role: "student" }), "/classes")).toBe(false);
  });
});
