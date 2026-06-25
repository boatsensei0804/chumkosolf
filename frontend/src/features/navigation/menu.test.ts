import { describe, expect, it } from "vitest";

import type { UserInfo } from "@/shared/schemas/auth";
import type { WorkGroupCode } from "@/shared/schemas/enums";

import { menuItemsForUser } from "./menu";

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
      "personnel",
      "students",
      "attendance",
      "budget",
      "settings",
    ]);
  });

  it("is_school_admin เห็นทุกเมนูรวมตั้งค่าระบบ", () => {
    const user = makeUser({ role: "teacher", is_school_admin: true });
    expect(keys(user)).toContain("settings");
    expect(keys(user)).toContain("personnel");
  });

  it("ครูไม่มีกลุ่มงาน เห็นหน้าแรก + เช็คชื่อ (สิทธิ์ครู) แต่ไม่เห็นบุคคล/วิชาการ/ตั้งค่า", () => {
    const user = makeUser({ role: "teacher", work_groups: [] });
    const k = keys(user);
    expect(k).toContain("home");
    expect(k).toContain("attendance"); // allowRoles teacher
    expect(k).not.toContain("personnel");
    expect(k).not.toContain("students");
    expect(k).not.toContain("settings");
  });

  it("สมาชิกกลุ่มวิชาการ เห็นเมนูนักเรียน แต่ไม่เห็นบุคลากร", () => {
    const user = makeUser({ role: "teacher", work_groups: [workGroup("academic")] });
    const k = keys(user);
    expect(k).toContain("students");
    expect(k).not.toContain("personnel");
  });

  it("สมาชิกกลุ่มบุคคล เห็นเมนูบุคลากร แต่ไม่เห็นนักเรียน", () => {
    const user = makeUser({ role: "executive", work_groups: [workGroup("personnel")] });
    const k = keys(user);
    expect(k).toContain("personnel");
    expect(k).not.toContain("students");
  });

  it("ทุกคนต้องเห็นหน้าแรกเสมอ", () => {
    expect(keys(makeUser({ role: "student" }))).toContain("home");
  });
});
