import { z } from "zod";

import { addressSchema, type AddressData } from "./personnel";

// schema นักเรียน — ตรงกับ backend (service.StudentListItem / StudentDetail) · เลขบัตร masked เท่านั้น (PDPA: ข้อมูลเด็ก)

// สถานะนักเรียน — ตรงกับ backend domain
export const studentStatusSchema = z.enum(["studying", "resigned", "suspended"]);
export type StudentStatus = z.infer<typeof studentStatusSchema>;

export const studentStatusLabel: Record<StudentStatus, string> = {
  studying: "กำลังศึกษา",
  resigned: "ลาออก",
  suspended: "แขวนลอย",
};
export const studentStatusColor: Record<StudentStatus, string> = {
  studying: "success",
  resigned: "default",
  suspended: "warning",
};
export function isStudentStatus(v: string): v is StudentStatus {
  return v === "studying" || v === "resigned" || v === "suspended";
}

export const studentListItemSchema = z.object({
  id: z.string(),
  student_code: z.string(),
  status: z.string(),
  prefix: z.string(),
  first_name: z.string(),
  last_name: z.string(),
  national_id_masked: z.string(),
  phone: z.string(),
  created_at: z.string(),
});
export type StudentListItem = z.infer<typeof studentListItemSchema>;
export const studentListSchema = z.array(studentListItemSchema);

export const studentDetailSchema = z.object({
  id: z.string(),
  student_code: z.string(),
  status: z.string(),
  prefix: z.string(),
  first_name: z.string(),
  last_name: z.string(),
  national_id_masked: z.string(),
  birth_date: z.string(),
  phone: z.string(),
  address: addressSchema,
  photo_path: z.string(),
  current_class_id: z.string(),
  current_class_label: z.string(),
  current_enrollment_id: z.string(),
  created_at: z.string(),
  updated_at: z.string(),
});
export type StudentDetail = z.infer<typeof studentDetailSchema>;

// รูปนักเรียน (หลายรูปต่อคน; is_primary = รูปโปรไฟล์) — url เป็น signed URL หมดอายุ
export const studentPhotoSchema = z.object({
  id: z.string(),
  url: z.string(),
  is_primary: z.boolean(),
  created_at: z.string(),
});
export type StudentPhoto = z.infer<typeof studentPhotoSchema>;
export const studentPhotoListSchema = z.array(studentPhotoSchema);

export type StudentAddressBody = AddressData;

export type CreateStudentBody = {
  national_id: string;
  student_code: string;
  status: StudentStatus;
  prefix: string;
  first_name: string;
  last_name: string;
  birth_date: string;
  phone: string;
  address: StudentAddressBody;
};

export type UpdateStudentBody = CreateStudentBody;
