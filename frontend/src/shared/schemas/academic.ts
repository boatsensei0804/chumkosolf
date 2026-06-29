import { z } from "zod";

// ปีการศึกษา — ตรงกับ backend service.AcademicYearDTO
export const academicYearSchema = z.object({
  id: z.string(),
  year: z.number(),
  is_current: z.boolean(),
});
export type AcademicYear = z.infer<typeof academicYearSchema>;
export const academicYearListSchema = z.array(academicYearSchema);

// ภาคเรียน — ตรงกับ backend service.SemesterDTO
export const semesterSchema = z.object({
  id: z.string(),
  academic_year_id: z.string(),
  year: z.number(),
  term: z.number(),
  start_date: z.string(),
  end_date: z.string(),
  is_current: z.boolean(),
});
export type Semester = z.infer<typeof semesterSchema>;
export const semesterListSchema = z.array(semesterSchema);

export type CreateYearBody = { year: number };
export type CreateSemesterBody = {
  academic_year_id: string;
  term: number;
  start_date: string;
  end_date: string;
};

// semesterLabel ป้ายชื่อภาคเรียนแบบไทย
export function semesterLabel(s: { year: number; term: number }): string {
  return `ปีการศึกษา ${s.year} ภาคเรียนที่ ${s.term}`;
}
