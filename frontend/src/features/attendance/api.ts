import { ApiRequestError, apiRequest } from "@/lib/api/client";
import { getAccessToken } from "@/features/auth/storage";
import { createdIdSchema, messageSchema } from "@/shared/schemas/personnel";
import {
  attendanceRosterSchema,
  type AttendanceRosterEntry,
  type SaveAttendanceBody,
} from "@/shared/schemas/attendance";
import {
  behaviorSummarySchema,
  type BehaviorBody,
  type BehaviorSummary,
} from "@/shared/schemas/behavior";

function token(): string {
  const t = getAccessToken();
  if (!t) throw new ApiRequestError({ code: "NO_SESSION", message: "เซสชันหมดอายุ กรุณาเข้าสู่ระบบใหม่" });
  return t;
}

// --- เช็คชื่อเข้าเรียน ---
export async function getRoster(classId: string, date: string): Promise<AttendanceRosterEntry[]> {
  return apiRequest(
    `/classes/${classId}/attendance?date=${encodeURIComponent(date)}`,
    attendanceRosterSchema,
    { token: token() },
  );
}

export async function saveAttendance(classId: string, body: SaveAttendanceBody): Promise<void> {
  await apiRequest(`/classes/${classId}/attendance`, messageSchema, {
    method: "POST",
    body,
    token: token(),
  });
}

// --- คะแนนความประพฤติ ---
export async function getBehavior(studentId: string): Promise<BehaviorSummary> {
  return apiRequest(`/students/${studentId}/behavior`, behaviorSummarySchema, { token: token() });
}

export async function addBehavior(studentId: string, body: BehaviorBody): Promise<void> {
  await apiRequest(`/students/${studentId}/behavior`, createdIdSchema, {
    method: "POST",
    body,
    token: token(),
  });
}

export async function deleteBehavior(studentId: string, recordId: string): Promise<void> {
  await apiRequest(`/students/${studentId}/behavior/${recordId}`, messageSchema, {
    method: "DELETE",
    token: token(),
  });
}
