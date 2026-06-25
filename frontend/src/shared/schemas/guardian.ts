import { z } from "zod";

import { addressSchema, type AddressData } from "./personnel";
import { guardianRelationshipSchema, type GuardianRelationship } from "./enums";

// schema ผู้ปกครอง + ความเชื่อมโยงนักเรียน-ผู้ปกครอง — ตรงกับ backend

export const relationshipLabel: Record<GuardianRelationship, string> = {
  father: "บิดา",
  mother: "มารดา",
  other: "อื่น ๆ",
};

export const guardianListItemSchema = z.object({
  id: z.string(),
  prefix: z.string(),
  first_name: z.string(),
  last_name: z.string(),
  national_id_masked: z.string(),
  phone: z.string(),
  created_at: z.string(),
});
export type GuardianListItem = z.infer<typeof guardianListItemSchema>;
export const guardianListSchema = z.array(guardianListItemSchema);

export const guardianDetailSchema = z.object({
  id: z.string(),
  prefix: z.string(),
  first_name: z.string(),
  last_name: z.string(),
  national_id_masked: z.string(),
  birth_date: z.string(),
  phone: z.string(),
  address: addressSchema,
  created_at: z.string(),
  updated_at: z.string(),
});
export type GuardianDetail = z.infer<typeof guardianDetailSchema>;

export type CreateGuardianBody = {
  national_id: string;
  prefix: string;
  first_name: string;
  last_name: string;
  birth_date: string;
  phone: string;
  address: AddressData;
};
export type UpdateGuardianBody = CreateGuardianBody;

// ความเชื่อมโยง (ผู้ปกครองของนักเรียน) — มาจาก GET /students/:id/guardians
export const studentGuardianSchema = z.object({
  id: z.string(),
  guardian_id: z.string(),
  relationship: z.string(),
  is_primary: z.boolean(),
  prefix: z.string(),
  first_name: z.string(),
  last_name: z.string(),
  phone: z.string(),
  national_id_masked: z.string(),
});
export type StudentGuardian = z.infer<typeof studentGuardianSchema>;
export const studentGuardianListSchema = z.array(studentGuardianSchema);

// เชื่อมผู้ปกครอง = สร้าง inline (ส่งข้อมูลผู้ปกครองเต็ม) + ความสัมพันธ์
export type LinkGuardianBody = {
  national_id: string;
  prefix: string;
  first_name: string;
  last_name: string;
  birth_date: string;
  phone: string;
  address: AddressData;
  relationship: GuardianRelationship;
  is_primary: boolean;
};

export { guardianRelationshipSchema };
