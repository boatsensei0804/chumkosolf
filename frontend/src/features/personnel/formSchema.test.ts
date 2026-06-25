import { describe, expect, it } from "vitest";

import {
  createPersonnelFormSchema,
  editPersonnelFormSchema,
  emptyAddress,
  toCreateBody,
  toUpdateBody,
  type CreatePersonnelFormValues,
  type EditPersonnelFormValues,
} from "./formSchema";

function validCreate(): CreatePersonnelFormValues {
  return {
    username: "kru.somchai",
    password: "password123",
    role: "teacher",
    nationalId: "1234567890123",
    civilServantId: "",
    prefix: "นาย",
    firstName: "สมชาย",
    lastName: "ใจดี",
    birthDate: "1985-05-20",
    phone: "0812345678",
    email: "somchai@example.com",
    address: { ...emptyAddress, province: "กรุงเทพ" },
  };
}

describe("createPersonnelFormSchema", () => {
  it("ผ่านเมื่อข้อมูลครบถูกต้อง", () => {
    expect(createPersonnelFormSchema.safeParse(validCreate()).success).toBe(true);
  });

  it("เลขบัตรไม่ครบ 13 หลัก → ไม่ผ่าน พร้อมข้อความไทย", () => {
    const r = createPersonnelFormSchema.safeParse({ ...validCreate(), nationalId: "123" });
    expect(r.success).toBe(false);
    if (!r.success) {
      expect(r.error.issues[0]?.message).toBe("เลขบัตรประชาชนต้องเป็นตัวเลข 13 หลัก");
    }
  });

  it("รหัสผ่านสั้นเกินไป → ไม่ผ่าน", () => {
    expect(createPersonnelFormSchema.safeParse({ ...validCreate(), password: "123" }).success).toBe(false);
  });

  it("อีเมลผิดรูปแบบ → ไม่ผ่าน แต่อีเมลว่างได้", () => {
    expect(createPersonnelFormSchema.safeParse({ ...validCreate(), email: "bad" }).success).toBe(false);
    expect(createPersonnelFormSchema.safeParse({ ...validCreate(), email: "" }).success).toBe(true);
  });

  it("ชื่อ/นามสกุลว่าง → ไม่ผ่าน", () => {
    expect(createPersonnelFormSchema.safeParse({ ...validCreate(), firstName: "" }).success).toBe(false);
  });
});

describe("editPersonnelFormSchema", () => {
  it("เลขบัตรว่างได้ (ไม่เปลี่ยน)", () => {
    const v: EditPersonnelFormValues = {
      nationalId: "",
      civilServantId: "",
      prefix: "",
      firstName: "สมหญิง",
      lastName: "ใจงาม",
      birthDate: "",
      phone: "",
      email: "",
      address: emptyAddress,
    };
    expect(editPersonnelFormSchema.safeParse(v).success).toBe(true);
  });

  it("ถ้ากรอกเลขบัตรต้องครบ 13 หลัก", () => {
    const base = { civilServantId: "", prefix: "", firstName: "a", lastName: "b", birthDate: "", phone: "", email: "", address: emptyAddress };
    expect(editPersonnelFormSchema.safeParse({ ...base, nationalId: "999" }).success).toBe(false);
    expect(editPersonnelFormSchema.safeParse({ ...base, nationalId: "1234567890123" }).success).toBe(true);
  });
});

describe("body mappers", () => {
  it("toCreateBody แปลง camelCase → snake_case ครบ", () => {
    const body = toCreateBody(validCreate());
    expect(body).toMatchObject({
      username: "kru.somchai",
      national_id: "1234567890123",
      first_name: "สมชาย",
      last_name: "ใจดี",
      birth_date: "1985-05-20",
    });
    expect(body.address.province).toBe("กรุงเทพ");
  });

  it("toUpdateBody ไม่มี field บัญชี (username/password/role)", () => {
    const body = toUpdateBody({
      nationalId: "",
      civilServantId: "",
      prefix: "",
      firstName: "ก",
      lastName: "ข",
      birthDate: "",
      phone: "",
      email: "",
      address: emptyAddress,
    });
    expect(body).not.toHaveProperty("username");
    expect(body).not.toHaveProperty("password");
    expect(body.first_name).toBe("ก");
  });
});
