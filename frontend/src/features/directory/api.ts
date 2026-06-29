import { ApiRequestError, apiRequest } from "@/lib/api/client";
import { getAccessToken } from "@/features/auth/storage";
import {
  directoryClassListSchema,
  directoryStudentClassListSchema,
  directoryStudentListSchema,
  type DirectoryClass,
  type DirectoryStudent,
  type DirectoryStudentClass,
} from "@/shared/schemas/directory";

function token(): string {
  const t = getAccessToken();
  if (!t) throw new ApiRequestError({ code: "NO_SESSION", message: "เซสชันหมดอายุ กรุณาเข้าสู่ระบบใหม่" });
  return t;
}

export async function listDirectoryClasses(): Promise<DirectoryClass[]> {
  return apiRequest("/directory/classes", directoryClassListSchema, { token: token() });
}

export async function listDirectoryClassStudents(classId: string): Promise<DirectoryStudent[]> {
  return apiRequest(`/directory/classes/${classId}/students`, directoryStudentListSchema, { token: token() });
}

export async function searchDirectoryStudents(q: string): Promise<DirectoryStudentClass[]> {
  return apiRequest(`/directory/students?q=${encodeURIComponent(q)}`, directoryStudentClassListSchema, {
    token: token(),
  });
}
