import { z } from "zod";

// schema ห้องเรียน + ครูที่ปรึกษา + นักเรียนในห้อง — ตรงกับ backend (รายเทอม)

export const classListItemSchema = z.object({
  id: z.string(),
  grade_level: z.string(),
  room_name: z.string(),
  student_count: z.number(),
  advisor_count: z.number(),
});
export type ClassListItem = z.infer<typeof classListItemSchema>;
export const classListSchema = z.array(classListItemSchema);

export const classDetailSchema = z.object({
  id: z.string(),
  grade_level: z.string(),
  room_name: z.string(),
  semester_id: z.string(),
  created_at: z.string(),
});
export type ClassDetail = z.infer<typeof classDetailSchema>;

export type ClassBody = { grade_level: string; room_name: string };

export const classAdvisorSchema = z.object({
  id: z.string(),
  personnel_id: z.string(),
  prefix: z.string(),
  first_name: z.string(),
  last_name: z.string(),
});
export type ClassAdvisor = z.infer<typeof classAdvisorSchema>;
export const classAdvisorListSchema = z.array(classAdvisorSchema);

export type AddAdvisorBody = { personnel_id: string };

export const enrollmentSchema = z.object({
  id: z.string(),
  student_id: z.string(),
  student_no: z.number().nullable(),
  student_code: z.string(),
  prefix: z.string(),
  first_name: z.string(),
  last_name: z.string(),
});
export type Enrollment = z.infer<typeof enrollmentSchema>;
export const enrollmentListSchema = z.array(enrollmentSchema);

export type EnrollBody = { student_id: string; student_no?: number };
