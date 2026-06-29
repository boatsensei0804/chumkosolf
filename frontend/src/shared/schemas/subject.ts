import { z } from "zod";

// รายวิชา — ตรงกับ backend service.SubjectDTO
export const subjectSchema = z.object({
  id: z.string(),
  subject_code: z.string(),
  name: z.string(),
  credit: z.number().nullable(),
});
export type Subject = z.infer<typeof subjectSchema>;
export const subjectListSchema = z.array(subjectSchema);

export type SubjectBody = {
  subject_code: string;
  name: string;
  credit: number | null;
};
