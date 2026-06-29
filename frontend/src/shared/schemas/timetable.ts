import { z } from "zod";

// ป้ายชื่อวัน (index = day_of_week, 1=จันทร์)
export const DAY_LABELS = ["", "จันทร์", "อังคาร", "พุธ", "พฤหัสบดี", "ศุกร์", "เสาร์", "อาทิตย์"] as const;

// --- มอบหมายการสอน — ตรงกับ service.TeachingAssignmentDTO ---
export const teachingAssignmentSchema = z.object({
  id: z.string(),
  personnel_id: z.string(),
  subject_id: z.string(),
  class_id: z.string(),
  teacher_name: z.string(),
  subject_code: z.string(),
  subject_name: z.string(),
  grade_level: z.string(),
  room_name: z.string(),
});
export type TeachingAssignment = z.infer<typeof teachingAssignmentSchema>;
export const teachingAssignmentListSchema = z.array(teachingAssignmentSchema);

export type AssignmentBody = {
  personnel_id: string;
  subject_id: string;
  class_id: string;
};

// --- ตั้งค่าคาบ — ตรงกับ service.TimetableConfigDTO ---
export const periodDefSchema = z.object({
  period_no: z.number(),
  label: z.string(),
  start_time: z.string(),
  end_time: z.string(),
  is_break: z.boolean(),
});
export type PeriodDef = z.infer<typeof periodDefSchema>;

export const timetableConfigSchema = z.object({
  days_per_week: z.number(),
  periods_per_day: z.number(),
  periods: z.array(periodDefSchema),
});
export type TimetableConfig = z.infer<typeof timetableConfigSchema>;

export type ConfigBody = {
  days_per_week: number;
  periods_per_day: number;
  periods: PeriodDef[];
};

// --- ช่องตารางสอน — ตรงกับ service.TimetableSlotDTO ---
export const timetableSlotSchema = z.object({
  id: z.string(),
  day_of_week: z.number(),
  period_no: z.number(),
  teaching_assignment_id: z.string(),
  subject_code: z.string(),
  subject_name: z.string(),
  teacher_name: z.string(),
});
export type TimetableSlot = z.infer<typeof timetableSlotSchema>;
export const timetableSlotListSchema = z.array(timetableSlotSchema);

export type SlotBody = {
  day_of_week: number;
  period_no: number;
  teaching_assignment_id: string;
};

// --- ครูว่างวันนี้ — ตรงกับ service.FreeTeachersDTO ---
export const teacherBriefSchema = z.object({
  id: z.string(),
  name: z.string(),
});
export type TeacherBrief = z.infer<typeof teacherBriefSchema>;

export const freePeriodSchema = z.object({
  period_no: z.number(),
  label: z.string(),
  free_teachers: z.array(teacherBriefSchema),
});
export type FreePeriod = z.infer<typeof freePeriodSchema>;

export const freeTeachersSchema = z.object({
  day: z.number(),
  periods: z.array(freePeriodSchema),
});
export type FreeTeachers = z.infer<typeof freeTeachersSchema>;

// --- overview เช็คชื่อรายวิชาของครู — ตรงกับ service.CheckinOverviewDTO ---
export const checkinSlotSchema = z.object({
  slot_id: z.string(),
  day_of_week: z.number(),
  period_no: z.number(),
  subject_code: z.string(),
  subject_name: z.string(),
  class_label: z.string(),
  date: z.string(),
  checked: z.boolean(),
});
export type CheckinSlot = z.infer<typeof checkinSlotSchema>;

export const checkinWeekSchema = z.object({
  index: z.number(),
  start: z.string(),
  end: z.string(),
});
export type CheckinWeek = z.infer<typeof checkinWeekSchema>;

export const checkinOverviewSchema = z.object({
  week_start: z.string(),
  slots: z.array(checkinSlotSchema),
  unchecked_this_week: z.number(),
  total_this_week: z.number(),
  has_week_stats: z.boolean(),
  incomplete_weeks: z.number(),
  total_weeks: z.number(),
  weeks: z.array(checkinWeekSchema),
  current_week_index: z.number(),
});
export type CheckinOverview = z.infer<typeof checkinOverviewSchema>;
