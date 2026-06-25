import { z } from "zod";

import { addressSchema, type AddressData } from "./personnel";

// schema นักเรียน — ตรงกับ backend (service.StudentListItem / StudentDetail) · เลขบัตร masked เท่านั้น (PDPA: ข้อมูลเด็ก)

export const studentListItemSchema = z.object({
  id: z.string(),
  student_code: z.string(),
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
  prefix: z.string(),
  first_name: z.string(),
  last_name: z.string(),
  national_id_masked: z.string(),
  birth_date: z.string(),
  phone: z.string(),
  address: addressSchema,
  photo_path: z.string(),
  created_at: z.string(),
  updated_at: z.string(),
});
export type StudentDetail = z.infer<typeof studentDetailSchema>;

export type StudentAddressBody = AddressData;

export type CreateStudentBody = {
  national_id: string;
  student_code: string;
  prefix: string;
  first_name: string;
  last_name: string;
  birth_date: string;
  phone: string;
  address: StudentAddressBody;
};

export type UpdateStudentBody = CreateStudentBody;
