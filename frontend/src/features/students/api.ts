import { ApiRequestError, apiRequest, apiRequestWithMeta } from "@/lib/api/client";
import { getAccessToken } from "@/features/auth/storage";
import { createdIdSchema, messageSchema } from "@/shared/schemas/personnel";
import {
  studentDetailSchema,
  studentListSchema,
  type CreateStudentBody,
  type StudentDetail,
  type StudentListItem,
  type UpdateStudentBody,
} from "@/shared/schemas/student";
import {
  studentGuardianListSchema,
  type LinkGuardianBody,
  type StudentGuardian,
} from "@/shared/schemas/guardian";

function token(): string {
  const t = getAccessToken();
  if (!t) throw new ApiRequestError({ code: "NO_SESSION", message: "เซสชันหมดอายุ กรุณาเข้าสู่ระบบใหม่" });
  return t;
}

export type StudentListResult = { items: StudentListItem[]; total: number };

export async function listStudents(page: number, pageSize: number, search = ""): Promise<StudentListResult> {
  const qs = `page=${page}&page_size=${pageSize}${search ? `&q=${encodeURIComponent(search)}` : ""}`;
  const { data, meta } = await apiRequestWithMeta(`/students?${qs}`, studentListSchema, { token: token() });
  return { items: data, total: meta?.total ?? data.length };
}

export async function getStudent(id: string): Promise<StudentDetail> {
  return apiRequest(`/students/${id}`, studentDetailSchema, { token: token() });
}

export async function createStudent(body: CreateStudentBody): Promise<string> {
  const { id } = await apiRequest("/students", createdIdSchema, { method: "POST", body, token: token() });
  return id;
}

export async function updateStudent(id: string, body: UpdateStudentBody): Promise<void> {
  await apiRequest(`/students/${id}`, messageSchema, { method: "PUT", body, token: token() });
}

export async function deleteStudent(id: string): Promise<void> {
  await apiRequest(`/students/${id}`, messageSchema, { method: "DELETE", token: token() });
}

// --- ผู้ปกครองของนักเรียน ---
export async function listStudentGuardians(studentId: string): Promise<StudentGuardian[]> {
  return apiRequest(`/students/${studentId}/guardians`, studentGuardianListSchema, { token: token() });
}

export async function linkGuardian(studentId: string, body: LinkGuardianBody): Promise<void> {
  await apiRequest(`/students/${studentId}/guardians`, messageSchema, { method: "POST", body, token: token() });
}

export async function unlinkGuardian(studentId: string, linkId: string): Promise<void> {
  await apiRequest(`/students/${studentId}/guardians/${linkId}`, messageSchema, {
    method: "DELETE",
    token: token(),
  });
}
