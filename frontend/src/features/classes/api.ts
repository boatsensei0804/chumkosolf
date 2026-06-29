import { ApiRequestError, apiRequest } from "@/lib/api/client";
import { getAccessToken } from "@/features/auth/storage";
import { createdIdSchema, messageSchema } from "@/shared/schemas/personnel";
import {
  classAdvisorListSchema,
  classDetailSchema,
  classListSchema,
  enrollmentListSchema,
  type AddAdvisorBody,
  type ClassAdvisor,
  type ClassBody,
  type ClassDetail,
  type ClassListItem,
  type EnrollBody,
  type Enrollment,
} from "@/shared/schemas/class";

function token(): string {
  const t = getAccessToken();
  if (!t) throw new ApiRequestError({ code: "NO_SESSION", message: "เซสชันหมดอายุ กรุณาเข้าสู่ระบบใหม่" });
  return t;
}

export async function listClasses(): Promise<ClassListItem[]> {
  return apiRequest("/classes", classListSchema, { token: token() });
}
export async function getClass(id: string): Promise<ClassDetail> {
  return apiRequest(`/classes/${id}`, classDetailSchema, { token: token() });
}
export async function createClass(body: ClassBody): Promise<string> {
  const { id } = await apiRequest("/classes", createdIdSchema, { method: "POST", body, token: token() });
  return id;
}
export async function updateClass(id: string, body: ClassBody): Promise<void> {
  await apiRequest(`/classes/${id}`, messageSchema, { method: "PUT", body, token: token() });
}
export async function deleteClass(id: string): Promise<void> {
  await apiRequest(`/classes/${id}`, messageSchema, { method: "DELETE", token: token() });
}

// ครูที่ปรึกษา
export async function listAdvisors(classId: string): Promise<ClassAdvisor[]> {
  return apiRequest(`/classes/${classId}/advisors`, classAdvisorListSchema, { token: token() });
}
export async function addAdvisor(classId: string, body: AddAdvisorBody): Promise<void> {
  await apiRequest(`/classes/${classId}/advisors`, messageSchema, { method: "POST", body, token: token() });
}
export async function removeAdvisor(classId: string, advisorId: string): Promise<void> {
  await apiRequest(`/classes/${classId}/advisors/${advisorId}`, messageSchema, { method: "DELETE", token: token() });
}

// นักเรียนในห้อง
export async function listEnrollments(classId: string): Promise<Enrollment[]> {
  return apiRequest(`/classes/${classId}/students`, enrollmentListSchema, { token: token() });
}
export async function enrollStudent(classId: string, body: EnrollBody): Promise<void> {
  await apiRequest(`/classes/${classId}/students`, messageSchema, { method: "POST", body, token: token() });
}
export async function enrollStudentsBulk(classId: string, studentIds: string[]): Promise<void> {
  await apiRequest(`/classes/${classId}/students/bulk`, messageSchema, {
    method: "POST",
    body: { student_ids: studentIds },
    token: token(),
  });
}
export async function removeEnrollment(classId: string, enrollmentId: string): Promise<void> {
  await apiRequest(`/classes/${classId}/students/${enrollmentId}`, messageSchema, { method: "DELETE", token: token() });
}
