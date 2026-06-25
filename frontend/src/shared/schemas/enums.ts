import { z } from "zod";

// enum ที่ใช้ร่วมกับ backend — ต้องตรงเป๊ะกับ CHECK constraint ใน migration
// เก็บที่เดียวที่นี่ ห้ามนิยามซ้ำในที่อื่น

export const attendanceStatusSchema = z.enum([
  "present",
  "absent",
  "late",
  "sick_leave",
  "personal_leave",
]);
export type AttendanceStatus = z.infer<typeof attendanceStatusSchema>;

export const guardianRelationshipSchema = z.enum(["father", "mother", "other"]);
export type GuardianRelationship = z.infer<typeof guardianRelationshipSchema>;

export const adminPositionSchema = z.enum(["director", "deputy_director"]);
export type AdminPosition = z.infer<typeof adminPositionSchema>;

export const userRoleSchema = z.enum([
  "super_admin",
  "teacher",
  "executive",
  "student",
]);
export type UserRole = z.infer<typeof userRoleSchema>;

export const workGroupCodeSchema = z.enum([
  "personnel",
  "general_affairs",
  "academic",
  "budget_plan",
]);
export type WorkGroupCode = z.infer<typeof workGroupCodeSchema>;

// label ภาษาไทยสำหรับแสดงผล
export const attendanceStatusLabel: Record<AttendanceStatus, string> = {
  present: "มาเรียน",
  absent: "ขาด",
  late: "สาย",
  sick_leave: "ลาป่วย",
  personal_leave: "ลากิจ",
};

export const workGroupLabel: Record<WorkGroupCode, string> = {
  personnel: "กลุ่มงานบุคคล",
  general_affairs: "กลุ่มงานบริหารทั่วไป",
  academic: "กลุ่มงานวิชาการ",
  budget_plan: "กลุ่มงานงบประมาณและแผน",
};
