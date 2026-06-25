import { ApiRequestError, apiRequest } from "@/lib/api/client";
import { getAccessToken } from "@/features/auth/storage";
import { createdIdSchema, messageSchema } from "@/shared/schemas/personnel";
import {
  academicStandingListSchema,
  adminPositionListSchema,
  type AcademicStandingRecord,
  type AdminPositionRecord,
  type CreatePositionBody,
  type StandingBody,
} from "@/shared/schemas/personnelExtra";

function requireToken(): string {
  const token = getAccessToken();
  if (!token) {
    throw new ApiRequestError({ code: "NO_SESSION", message: "เซสชันหมดอายุ กรุณาเข้าสู่ระบบใหม่" });
  }
  return token;
}

// --- ตำแหน่งบริหาร ---
export async function listPositions(personnelId: string): Promise<AdminPositionRecord[]> {
  return apiRequest(`/personnel/${personnelId}/positions`, adminPositionListSchema, {
    token: requireToken(),
  });
}

export async function createPosition(personnelId: string, body: CreatePositionBody): Promise<void> {
  await apiRequest(`/personnel/${personnelId}/positions`, createdIdSchema, {
    method: "POST",
    body,
    token: requireToken(),
  });
}

export async function deletePosition(personnelId: string, posId: string): Promise<void> {
  await apiRequest(`/personnel/${personnelId}/positions/${posId}`, messageSchema, {
    method: "DELETE",
    token: requireToken(),
  });
}

// --- วิทยฐานะ ---
export async function listStandings(personnelId: string): Promise<AcademicStandingRecord[]> {
  return apiRequest(`/personnel/${personnelId}/standings`, academicStandingListSchema, {
    token: requireToken(),
  });
}

export async function createStanding(personnelId: string, body: StandingBody): Promise<void> {
  await apiRequest(`/personnel/${personnelId}/standings`, createdIdSchema, {
    method: "POST",
    body,
    token: requireToken(),
  });
}

export async function deleteStanding(personnelId: string, sid: string): Promise<void> {
  await apiRequest(`/personnel/${personnelId}/standings/${sid}`, messageSchema, {
    method: "DELETE",
    token: requireToken(),
  });
}
