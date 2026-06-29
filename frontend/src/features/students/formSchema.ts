import { z } from "zod";

import { addressFormSchema } from "@/features/personnel/formSchema";
import type { AddressData } from "@/shared/schemas/personnel";
import { studentStatusSchema, type CreateStudentBody, type UpdateStudentBody } from "@/shared/schemas/student";

// form schema นักเรียน (camelCase + ข้อความไทย)
const profileShape = {
  studentCode: z.string().min(1, "กรุณากรอกรหัสนักเรียน"),
  status: studentStatusSchema,
  // classId: ห้องของเทอมปัจจุบัน (ไม่บังคับ; "" = ยังไม่จัดห้อง) — บันทึกผ่าน enrollment แยก
  classId: z.string(),
  prefix: z.string(),
  firstName: z.string().min(1, "กรุณากรอกชื่อ"),
  lastName: z.string().min(1, "กรุณากรอกนามสกุล"),
  birthDate: z.string(),
  phone: z.string(),
  address: addressFormSchema,
};

export const createStudentFormSchema = z.object({
  nationalId: z.string().regex(/^[0-9]{13}$/, "เลขบัตรประชาชนต้องเป็นตัวเลข 13 หลัก"),
  ...profileShape,
});
export type CreateStudentFormValues = z.infer<typeof createStudentFormSchema>;

export const editStudentFormSchema = z.object({
  nationalId: z.string().refine((v) => v === "" || /^[0-9]{13}$/.test(v), {
    message: "เลขบัตรประชาชนต้องเป็นตัวเลข 13 หลัก",
  }),
  ...profileShape,
});

function mapAddress(a: CreateStudentFormValues["address"]): AddressData {
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

export function toStudentBody(v: CreateStudentFormValues): CreateStudentBody & UpdateStudentBody {
  return {
    national_id: v.nationalId,
    student_code: v.studentCode,
    status: v.status,
    prefix: v.prefix,
    first_name: v.firstName,
    last_name: v.lastName,
    birth_date: v.birthDate,
    phone: v.phone,
    address: mapAddress(v.address),
  };
}
