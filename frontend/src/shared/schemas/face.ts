import { z } from "zod";

// schema ระบบสแกนหน้าเข้าเรียน — ตรงกับ backend FaceService

// ผลการสร้างฐานใบหน้าใหม่ (reindex)
export const reindexResultSchema = z.object({
  enrolled: z.number(),
  skipped: z.number(),
  total: z.number(),
});
export type ReindexResult = z.infer<typeof reindexResultSchema>;

// ผลการจดจำใบหน้า + บันทึกเช็คชื่อ
export const recognizeResultSchema = z.object({
  matched: z.boolean(),
  student_id: z.string(),
  student_code: z.string(),
  full_name: z.string(),
  class_label: z.string(),
  score: z.number(),
  marked: z.boolean(),
  already_marked: z.boolean(),
  status: z.string(),
  penalty_applied: z.number(),
  liveness_passed: z.boolean(),
  reason: z.string(),
});
export type RecognizeResult = z.infer<typeof recognizeResultSchema>;
