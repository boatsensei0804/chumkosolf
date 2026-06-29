import { z } from "zod";

import { addressSchema, type AddressData } from "./personnel";

// ข้อมูลโรงเรียน — ตรงกับ backend service.SchoolDTO
export const schoolSchema = z.object({
  id: z.string(),
  name: z.string(),
  code: z.string(),
  address: addressSchema,
  phone: z.string(),
  email: z.string(),
  website: z.string(),
  director_name: z.string(),
  is_active: z.boolean(),
  attendance_late_after: z.string(),
  attendance_late_penalty: z.number(),
});
export type School = z.infer<typeof schoolSchema>;

export type UpdateSchoolBody = {
  name: string;
  address: AddressData;
  phone: string;
  email: string;
  website: string;
  director_name: string;
  attendance_late_after: string;
  attendance_late_penalty: number;
};
