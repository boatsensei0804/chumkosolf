import { z } from "zod";

// สถานะการเช็คชื่อเข้าเรียน — ต้องตรงกับ backend (domain + CHECK constraint)
export const attendanceStatusSchema = z.enum([
  "present",
  "absent",
  "late",
  "sick_leave",
  "personal_leave",
]);
export type AttendanceStatus = z.infer<typeof attendanceStatusSchema>;

export const attendanceStatusLabel: Record<AttendanceStatus, string> = {
  present: "มาเรียน",
  absent: "ขาด",
  late: "สาย",
  sick_leave: "ลาป่วย",
  personal_leave: "ลากิจ",
};

// สี antd Tag/Select ของแต่ละสถานะ
export const attendanceStatusColor: Record<AttendanceStatus, string> = {
  present: "success",
  absent: "error",
  late: "warning",
  sick_leave: "blue",
  personal_leave: "purple",
};

export function isAttendanceStatus(v: string): v is AttendanceStatus {
  return v === "present" || v === "absent" || v === "late" || v === "sick_leave" || v === "personal_leave";
}

// แถวรายชื่อ + สถานะของวันนั้น (status = "" คือยังไม่เช็ค) — ตรง service.AttendanceRosterDTO
export const attendanceRosterEntrySchema = z.object({
  student_id: z.string(),
  student_no: z.number().nullable(),
  student_code: z.string(),
  prefix: z.string(),
  first_name: z.string(),
  last_name: z.string(),
  status: z.string(),
  note: z.string(),
  daily_status: z.string(), // สถานะเช็คชื่อเข้าเรียนรายวัน (โชว์ "มาสาย" ในเช็คชื่อรายวิชา)
});
export type AttendanceRosterEntry = z.infer<typeof attendanceRosterEntrySchema>;
export const attendanceRosterSchema = z.array(attendanceRosterEntrySchema);

export type AttendanceMarkBody = {
  student_id: string;
  status: AttendanceStatus;
  note: string;
};

export type SaveAttendanceBody = {
  date: string;
  records: AttendanceMarkBody[];
};
