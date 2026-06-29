import { z } from "zod";

// schema สำหรับดู/ค้นหาห้องเรียน-นักเรียน (read-only, ข้อมูลพื้นฐาน) — ตรงกับ backend DirectoryService

export const directoryClassSchema = z.object({
  id: z.string(),
  grade_level: z.string(),
  room_name: z.string(),
  student_count: z.number(),
});
export type DirectoryClass = z.infer<typeof directoryClassSchema>;
export const directoryClassListSchema = z.array(directoryClassSchema);

export const directoryStudentSchema = z.object({
  student_id: z.string(),
  student_code: z.string(),
  prefix: z.string(),
  first_name: z.string(),
  last_name: z.string(),
});
export type DirectoryStudent = z.infer<typeof directoryStudentSchema>;
export const directoryStudentListSchema = z.array(directoryStudentSchema);

export const directoryStudentClassSchema = directoryStudentSchema.extend({
  class_label: z.string(),
});
export type DirectoryStudentClass = z.infer<typeof directoryStudentClassSchema>;
export const directoryStudentClassListSchema = z.array(directoryStudentClassSchema);
