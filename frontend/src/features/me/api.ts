import { ApiRequestError, apiRequest } from "@/lib/api/client";
import { getAccessToken } from "@/features/auth/storage";
import { messageSchema } from "@/shared/schemas/personnel";
import { studentDetailSchema, type StudentDetail } from "@/shared/schemas/student";
import {
  adviseeListSchema,
  myProfileSchema,
  type Advisee,
  type MyProfile,
  type UpdateMyAdviseeBody,
  type UpdateMyProfileBody,
} from "@/shared/schemas/me";

function requireToken(): string {
  const token = getAccessToken();
  if (!token) {
    throw new ApiRequestError({ code: "NO_SESSION", message: "เซสชันหมดอายุ กรุณาเข้าสู่ระบบใหม่" });
  }
  return token;
}

export async function getMyProfile(): Promise<MyProfile> {
  return apiRequest("/me/profile", myProfileSchema, { token: requireToken() });
}

export async function updateMyProfile(body: UpdateMyProfileBody): Promise<void> {
  await apiRequest("/me/profile", messageSchema, {
    method: "PUT",
    body,
    token: requireToken(),
  });
}

export async function listMyAdvisees(): Promise<Advisee[]> {
  return apiRequest("/me/advisees", adviseeListSchema, { token: requireToken() });
}

export async function getMyAdvisee(studentId: string): Promise<StudentDetail> {
  return apiRequest(`/me/advisees/${studentId}`, studentDetailSchema, { token: requireToken() });
}

export async function updateMyAdvisee(studentId: string, body: UpdateMyAdviseeBody): Promise<void> {
  await apiRequest(`/me/advisees/${studentId}`, messageSchema, {
    method: "PUT",
    body,
    token: requireToken(),
  });
}
