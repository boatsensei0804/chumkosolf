import { ApiRequestError, apiRequest, apiUpload } from "@/lib/api/client";
import { getAccessToken } from "@/features/auth/storage";
import { messageSchema } from "@/shared/schemas/personnel";
import { studentPhotoListSchema, studentPhotoSchema, type StudentPhoto } from "@/shared/schemas/student";

function token(): string {
  const t = getAccessToken();
  if (!t) throw new ApiRequestError({ code: "NO_SESSION", message: "เซสชันหมดอายุ กรุณาเข้าสู่ระบบใหม่" });
  return t;
}

// รายการรูปนักเรียนทั้งหมด (signed URL; รูปโปรไฟล์มาก่อน)
export async function listStudentPhotos(studentId: string): Promise<StudentPhoto[]> {
  return apiRequest(`/students/${studentId}/photos`, studentPhotoListSchema, { token: token() });
}

export async function uploadStudentPhoto(studentId: string, file: File): Promise<StudentPhoto> {
  const form = new FormData();
  form.append("file", file);
  return apiUpload(`/students/${studentId}/photos`, studentPhotoSchema, form, { token: token() });
}

export async function setStudentPhotoPrimary(studentId: string, photoId: string): Promise<void> {
  await apiRequest(`/students/${studentId}/photos/${photoId}/primary`, messageSchema, {
    method: "PUT",
    token: token(),
  });
}

export async function deleteStudentPhoto(studentId: string, photoId: string): Promise<void> {
  await apiRequest(`/students/${studentId}/photos/${photoId}`, messageSchema, {
    method: "DELETE",
    token: token(),
  });
}
