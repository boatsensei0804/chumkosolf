import { ApiRequestError, apiRequest } from "@/lib/api/client";
import { getAccessToken } from "@/features/auth/storage";
import { createdIdSchema, messageSchema } from "@/shared/schemas/personnel";
import { subjectListSchema, type Subject, type SubjectBody } from "@/shared/schemas/subject";
import {
  teachingAssignmentListSchema,
  timetableConfigSchema,
  timetableSlotListSchema,
  type AssignmentBody,
  type ConfigBody,
  type SlotBody,
  type TeachingAssignment,
  type TimetableConfig,
  type TimetableSlot,
} from "@/shared/schemas/timetable";
import {
  attendanceRosterSchema,
  type AttendanceRosterEntry,
  type SaveAttendanceBody,
} from "@/shared/schemas/attendance";
import {
  checkinOverviewSchema,
  freeTeachersSchema,
  type CheckinOverview,
  type FreeTeachers,
} from "@/shared/schemas/timetable";

function token(): string {
  const t = getAccessToken();
  if (!t) throw new ApiRequestError({ code: "NO_SESSION", message: "เซสชันหมดอายุ กรุณาเข้าสู่ระบบใหม่" });
  return t;
}

// --- รายวิชา ---
export async function listSubjects(): Promise<Subject[]> {
  return apiRequest("/subjects", subjectListSchema, { token: token() });
}
export async function createSubject(body: SubjectBody): Promise<void> {
  await apiRequest("/subjects", createdIdSchema, { method: "POST", body, token: token() });
}
export async function updateSubject(id: string, body: SubjectBody): Promise<void> {
  await apiRequest(`/subjects/${id}`, messageSchema, { method: "PUT", body, token: token() });
}
export async function deleteSubject(id: string): Promise<void> {
  await apiRequest(`/subjects/${id}`, messageSchema, { method: "DELETE", token: token() });
}

// --- มอบหมายการสอน ---
export async function listAssignments(): Promise<TeachingAssignment[]> {
  return apiRequest("/teaching-assignments", teachingAssignmentListSchema, { token: token() });
}
export async function createAssignment(body: AssignmentBody): Promise<void> {
  await apiRequest("/teaching-assignments", createdIdSchema, { method: "POST", body, token: token() });
}
export async function deleteAssignment(id: string): Promise<void> {
  await apiRequest(`/teaching-assignments/${id}`, messageSchema, { method: "DELETE", token: token() });
}

// --- ตั้งค่าคาบ ---
export async function getConfig(): Promise<TimetableConfig> {
  return apiRequest("/timetable/config", timetableConfigSchema, { token: token() });
}
export async function saveConfig(body: ConfigBody): Promise<void> {
  await apiRequest("/timetable/config", messageSchema, { method: "PUT", body, token: token() });
}

// --- ช่องตารางสอน ---
export async function listSlots(classId: string): Promise<TimetableSlot[]> {
  return apiRequest(`/timetable/classes/${classId}`, timetableSlotListSchema, { token: token() });
}
export async function setSlot(classId: string, body: SlotBody): Promise<void> {
  await apiRequest(`/timetable/classes/${classId}/slots`, messageSchema, {
    method: "POST",
    body,
    token: token(),
  });
}
export async function clearSlot(classId: string, slotId: string): Promise<void> {
  await apiRequest(`/timetable/classes/${classId}/slots/${slotId}`, messageSchema, {
    method: "DELETE",
    token: token(),
  });
}

// --- ครูว่างวันนี้ ---
export async function getFreeTeachers(day: number): Promise<FreeTeachers> {
  return apiRequest(`/timetable/free-teachers?day=${day}`, freeTeachersSchema, { token: token() });
}

// --- เช็คชื่อรายวิชา ---
export async function getCheckinOverview(date: string): Promise<CheckinOverview> {
  return apiRequest(`/timetable/my-checkin?date=${encodeURIComponent(date)}`, checkinOverviewSchema, {
    token: token(),
  });
}

export async function getSubjectRoster(slotId: string, date: string): Promise<AttendanceRosterEntry[]> {
  return apiRequest(
    `/timetable/slots/${slotId}/attendance?date=${encodeURIComponent(date)}`,
    attendanceRosterSchema,
    { token: token() },
  );
}
export async function saveSubjectAttendance(slotId: string, body: SaveAttendanceBody): Promise<void> {
  await apiRequest(`/timetable/slots/${slotId}/attendance`, messageSchema, {
    method: "POST",
    body,
    token: token(),
  });
}
