import { z } from "zod";

// ข้อมูลสรุปหน้าแรก — ตรงกับ backend service.DashboardDTO
export const dashboardAttendanceSchema = z.object({
  present: z.number(),
  late: z.number(),
  absent: z.number(),
  sick_leave: z.number(),
  personal_leave: z.number(),
  unchecked: z.number(),
  total: z.number(),
});
export type DashboardAttendance = z.infer<typeof dashboardAttendanceSchema>;

export const dashboardSlotSchema = z.object({
  day_of_week: z.number(),
  period_no: z.number(),
  subject_code: z.string(),
  subject_name: z.string(),
  class_label: z.string(),
});
export type DashboardSlot = z.infer<typeof dashboardSlotSchema>;

export const dashboardSchema = z.object({
  is_advisor: z.boolean(),
  advisee_count: z.number(),
  today: z.string(),
  today_weekday: z.number(),
  attendance: dashboardAttendanceSchema,
  slots: z.array(dashboardSlotSchema),
});
export type Dashboard = z.infer<typeof dashboardSchema>;
