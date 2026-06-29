import { z } from "zod";

// ปี/เทอมปัจจุบัน — ตรงกับ backend service.CurrentTermDTO
export const currentTermSchema = z.object({
  has_current: z.boolean(),
  semester_id: z.string(),
  academic_year: z.number(),
  term: z.number(),
});
export type CurrentTerm = z.infer<typeof currentTermSchema>;
