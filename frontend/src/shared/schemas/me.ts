import { z } from "zod";

import { attendanceStatusSchema } from "./enums";
import { personnelDetailSchema, type AddressData } from "./personnel";

// schema ของ self-service ("ของฉัน") — ต้องตรงกับ backend service.MeService
// โปรไฟล์ของฉันใช้ shape เดียวกับ PersonnelDetail (เลขบัตร masked เสมอ — PDPA)
export const myProfileSchema = personnelDetailSchema;
export type MyProfile = z.infer<typeof myProfileSchema>;

// นักเรียนในห้องที่ปรึกษาของฉัน (today_status: "" = ยังไม่เช็คชื่อวันนี้)
export const adviseeSchema = z.object({
  student_id: z.string(),
  student_code: z.string(),
  prefix: z.string(),
  first_name: z.string(),
  last_name: z.string(),
  phone: z.string(),
  national_id_masked: z.string(),
  class_label: z.string(),
  today_status: z.union([attendanceStatusSchema, z.literal("")]),
});
export type Advisee = z.infer<typeof adviseeSchema>;

export const adviseeListSchema = z.array(adviseeSchema);

// payload แก้โปรไฟล์ตัวเอง (ไม่รวมเลขบัตร/role/username — เป็นงานกลุ่มบุคคล)
export type UpdateMyProfileBody = {
  prefix: string;
  first_name: string;
  last_name: string;
  birth_date: string;
  phone: string;
  email: string;
  address: AddressData;
};

// payload แก้ข้อมูลนักเรียนที่ปรึกษา (ครูที่ปรึกษาแก้ได้เฉพาะข้อมูลส่วนตัว/ที่อยู่
// ไม่รวมเลขบัตร/รหัสนักเรียน/สถานะ ซึ่งเป็นงานทะเบียนของกลุ่มวิชาการ)
export type UpdateMyAdviseeBody = {
  prefix: string;
  first_name: string;
  last_name: string;
  birth_date: string;
  phone: string;
  address: AddressData;
};
