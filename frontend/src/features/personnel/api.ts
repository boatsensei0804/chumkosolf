import { ApiRequestError, apiRequest, apiRequestWithMeta } from "@/lib/api/client";
import { getAccessToken } from "@/features/auth/storage";
import {
  createdIdSchema,
  messageSchema,
  personnelDetailSchema,
  personnelListSchema,
  type CreatePersonnelBody,
  type PersonnelDetail,
  type PersonnelListItem,
  type UpdatePersonnelBody,
} from "@/shared/schemas/personnel";

// ดึง access token จาก session — โยน error ที่สื่อความหมายถ้าไม่มี (ให้ผู้ใช้ล็อกอินใหม่)
function requireToken(): string {
  const token = getAccessToken();
  if (!token) {
    throw new ApiRequestError({ code: "NO_SESSION", message: "เซสชันหมดอายุ กรุณาเข้าสู่ระบบใหม่" });
  }
  return token;
}

export type PersonnelListResult = {
  items: PersonnelListItem[];
  total: number;
};

export async function listPersonnel(page: number, pageSize: number, search = ""): Promise<PersonnelListResult> {
  const qs = `page=${page}&page_size=${pageSize}${search ? `&q=${encodeURIComponent(search)}` : ""}`;
  const { data, meta } = await apiRequestWithMeta(`/personnel?${qs}`, personnelListSchema, { token: requireToken() });
  return { items: data, total: meta?.total ?? data.length };
}

export async function getPersonnel(id: string): Promise<PersonnelDetail> {
  return apiRequest(`/personnel/${id}`, personnelDetailSchema, { token: requireToken() });
}

export async function createPersonnel(body: CreatePersonnelBody): Promise<string> {
  const { id } = await apiRequest("/personnel", createdIdSchema, {
    method: "POST",
    body,
    token: requireToken(),
  });
  return id;
}

export async function updatePersonnel(id: string, body: UpdatePersonnelBody): Promise<void> {
  await apiRequest(`/personnel/${id}`, messageSchema, {
    method: "PUT",
    body,
    token: requireToken(),
  });
}

export async function deletePersonnel(id: string): Promise<void> {
  await apiRequest(`/personnel/${id}`, messageSchema, {
    method: "DELETE",
    token: requireToken(),
  });
}
