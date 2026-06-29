import { describe, expect, it } from "vitest";

import { emptyAddress } from "@/features/personnel/formSchema";

import {
  createStudentFormSchema,
  editStudentFormSchema,
  toStudentBody,
  type CreateStudentFormValues,
} from "./formSchema";

function valid(): CreateStudentFormValues {
  return {
    nationalId: "1234567890123",
    studentCode: "S001",
    status: "studying",
    classId: "",
    prefix: "ด.ช.",
    firstName: "กิตติ",
    lastName: "เรียนดี",
    birthDate: "2012-05-01",
    phone: "0810000001",
    address: { ...emptyAddress, province: "กรุงเทพฯ" },
  };
}

describe("createStudentFormSchema", () => {
  it("ผ่านเมื่อข้อมูลครบ", () => {
    expect(createStudentFormSchema.safeParse(valid()).success).toBe(true);
  });
  it("รหัสนักเรียนว่าง → ไม่ผ่าน", () => {
    expect(createStudentFormSchema.safeParse({ ...valid(), studentCode: "" }).success).toBe(false);
  });
  it("เลขบัตรไม่ครบ 13 หลัก → ข้อความไทย", () => {
    const r = createStudentFormSchema.safeParse({ ...valid(), nationalId: "123" });
    expect(r.success).toBe(false);
    if (!r.success) expect(r.error.issues[0]?.message).toBe("เลขบัตรประชาชนต้องเป็นตัวเลข 13 หลัก");
  });
});

describe("editStudentFormSchema", () => {
  it("เลขบัตรว่างได้ (ไม่เปลี่ยน)", () => {
    expect(editStudentFormSchema.safeParse({ ...valid(), nationalId: "" }).success).toBe(true);
  });
});

describe("toStudentBody", () => {
  it("แปลง camelCase → snake_case", () => {
    const b = toStudentBody(valid());
    expect(b).toMatchObject({ national_id: "1234567890123", student_code: "S001", first_name: "กิตติ", birth_date: "2012-05-01" });
    expect(b.address.province).toBe("กรุงเทพฯ");
  });
});
