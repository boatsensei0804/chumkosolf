import { z } from "zod";

import {
  personnelRoleSchema,
  type AddressData,
  type CreatePersonnelBody,
  type UpdatePersonnelBody,
} from "@/shared/schemas/personnel";

// form schema (camelCase + ข้อความไทย) → infer type → ใช้กับ react-hook-form
// validation ฝั่ง frontend ไม่ทดแทน backend (มีทั้งสองฝั่ง)

// อีเมลปล่อยว่างได้ แต่ถ้ากรอกต้องถูกรูปแบบ
const optionalEmail = z
  .string()
  .refine((v) => v === "" || z.string().email().safeParse(v).success, {
    message: "อีเมลไม่ถูกต้อง",
  });

const addressFormSchema = z.object({
  houseNo: z.string(),
  moo: z.string(),
  road: z.string(),
  subdistrict: z.string(),
  district: z.string(),
  province: z.string(),
  postalCode: z.string(),
});
export type AddressFormValues = z.infer<typeof addressFormSchema>;

// ฟิลด์โปรไฟล์ที่ใช้ร่วมกันทั้ง create/edit
const profileShape = {
  prefix: z.string(),
  firstName: z.string().min(1, "กรุณากรอกชื่อ"),
  lastName: z.string().min(1, "กรุณากรอกนามสกุล"),
  birthDate: z.string(),
  phone: z.string(),
  email: optionalEmail,
  civilServantId: z.string(),
  address: addressFormSchema,
};

export const createPersonnelFormSchema = z.object({
  username: z.string().min(3, "ชื่อผู้ใช้อย่างน้อย 3 ตัวอักษร"),
  password: z.string().min(8, "รหัสผ่านอย่างน้อย 8 ตัวอักษร"),
  role: personnelRoleSchema,
  nationalId: z.string().regex(/^[0-9]{13}$/, "เลขบัตรประชาชนต้องเป็นตัวเลข 13 หลัก"),
  ...profileShape,
});
export type CreatePersonnelFormValues = z.infer<typeof createPersonnelFormSchema>;

export const editPersonnelFormSchema = z.object({
  // edit: ปล่อยว่าง = ไม่เปลี่ยนเลขบัตร
  nationalId: z
    .string()
    .refine((v) => v === "" || /^[0-9]{13}$/.test(v), {
      message: "เลขบัตรประชาชนต้องเป็นตัวเลข 13 หลัก",
    }),
  ...profileShape,
});
export type EditPersonnelFormValues = z.infer<typeof editPersonnelFormSchema>;

function mapAddress(a: AddressFormValues): AddressData {
  return {
    house_no: a.houseNo,
    moo: a.moo,
    road: a.road,
    subdistrict: a.subdistrict,
    district: a.district,
    province: a.province,
    postal_code: a.postalCode,
  };
}

export function toCreateBody(v: CreatePersonnelFormValues): CreatePersonnelBody {
  return {
    username: v.username,
    password: v.password,
    role: v.role,
    national_id: v.nationalId,
    civil_servant_id: v.civilServantId,
    prefix: v.prefix,
    first_name: v.firstName,
    last_name: v.lastName,
    birth_date: v.birthDate,
    phone: v.phone,
    email: v.email,
    address: mapAddress(v.address),
  };
}

export function toUpdateBody(v: EditPersonnelFormValues): UpdatePersonnelBody {
  return {
    national_id: v.nationalId,
    civil_servant_id: v.civilServantId,
    prefix: v.prefix,
    first_name: v.firstName,
    last_name: v.lastName,
    birth_date: v.birthDate,
    phone: v.phone,
    email: v.email,
    address: mapAddress(v.address),
  };
}

// ค่าเริ่มต้นว่างของ address (ใช้ใน defaultValues)
export const emptyAddress: AddressFormValues = {
  houseNo: "",
  moo: "",
  road: "",
  subdistrict: "",
  district: "",
  province: "",
  postalCode: "",
};
