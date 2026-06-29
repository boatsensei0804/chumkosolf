import { ApiRequestError, apiRequest, apiUpload } from "@/lib/api/client";
import { getAccessToken } from "@/features/auth/storage";
import { createdIdSchema, messageSchema } from "@/shared/schemas/personnel";
import {
  personnelWorkListSchema,
  workFileListSchema,
  type PersonnelWorkRecord,
  type WorkBody,
  type WorkFileRecord,
  type WorkFileType,
} from "@/shared/schemas/personnelWork";

function requireToken(): string {
  const token = getAccessToken();
  if (!token) {
    throw new ApiRequestError({ code: "NO_SESSION", message: "เซสชันหมดอายุ กรุณาเข้าสู่ระบบใหม่" });
  }
  return token;
}

// --- ผลงาน ---
export async function listWorks(personnelId: string): Promise<PersonnelWorkRecord[]> {
  return apiRequest(`/personnel/${personnelId}/works`, personnelWorkListSchema, {
    token: requireToken(),
  });
}

export async function createWork(personnelId: string, body: WorkBody): Promise<void> {
  await apiRequest(`/personnel/${personnelId}/works`, createdIdSchema, {
    method: "POST",
    body,
    token: requireToken(),
  });
}

export async function updateWork(personnelId: string, workId: string, body: WorkBody): Promise<void> {
  await apiRequest(`/personnel/${personnelId}/works/${workId}`, messageSchema, {
    method: "PUT",
    body,
    token: requireToken(),
  });
}

export async function deleteWork(personnelId: string, workId: string): Promise<void> {
  await apiRequest(`/personnel/${personnelId}/works/${workId}`, messageSchema, {
    method: "DELETE",
    token: requireToken(),
  });
}

// --- ไฟล์แนบ ---
export async function listWorkFiles(personnelId: string, workId: string): Promise<WorkFileRecord[]> {
  return apiRequest(`/personnel/${personnelId}/works/${workId}/files`, workFileListSchema, {
    token: requireToken(),
  });
}

export async function uploadWorkFile(
  personnelId: string,
  workId: string,
  fileType: WorkFileType,
  file: File,
): Promise<void> {
  const form = new FormData();
  form.append("file_type", fileType);
  form.append("file", file);
  await apiUpload(`/personnel/${personnelId}/works/${workId}/files`, createdIdSchema, form, {
    token: requireToken(),
  });
}

export async function deleteWorkFile(
  personnelId: string,
  workId: string,
  fileId: string,
): Promise<void> {
  await apiRequest(`/personnel/${personnelId}/works/${workId}/files/${fileId}`, messageSchema, {
    method: "DELETE",
    token: requireToken(),
  });
}
