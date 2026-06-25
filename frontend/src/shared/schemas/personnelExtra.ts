import { z } from "zod";

import { adminPositionSchema, type AdminPosition } from "./enums";

// schema ของ sub-resource บุคลากร — ต้องตรงกับ backend (service.AdminPositionDTO / AcademicStandingDTO)

// label ตำแหน่งบริหาร (ใช้ enum กลางจาก enums.ts)
export const positionLabel: Record<AdminPosition, string> = {
  director: "ผู้อำนวยการ",
  deputy_director: "รองผู้อำนวยการ",
};

// --- ตำแหน่งบริหาร ---
export const adminPositionRecordSchema = z.object({
  id: z.string(),
  position: z.string(),
  is_active: z.boolean(),
  appointed_at: z.string(),
  created_at: z.string(),
});
export type AdminPositionRecord = z.infer<typeof adminPositionRecordSchema>;
export const adminPositionListSchema = z.array(adminPositionRecordSchema);

export type CreatePositionBody = {
  position: AdminPosition;
  appointed_at: string;
};

// --- วิทยฐานะ ---
export const academicStandingRecordSchema = z.object({
  id: z.string(),
  standing: z.string(),
  effective_date: z.string(),
  is_current: z.boolean(),
  created_at: z.string(),
});
export type AcademicStandingRecord = z.infer<typeof academicStandingRecordSchema>;
export const academicStandingListSchema = z.array(academicStandingRecordSchema);

export type StandingBody = {
  standing: string;
  effective_date: string;
  is_current: boolean;
};

// re-export enum สำหรับใช้ใน form
export { adminPositionSchema };
