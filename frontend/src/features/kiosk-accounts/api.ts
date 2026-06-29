import { ApiRequestError, apiRequest } from "@/lib/api/client";
import { getAccessToken } from "@/features/auth/storage";
import { createdIdSchema, messageSchema } from "@/shared/schemas/personnel";
import {
  kioskAccountListSchema,
  type CreateKioskAccountBody,
  type KioskAccount,
} from "@/shared/schemas/kioskAccount";

function token(): string {
  const t = getAccessToken();
  if (!t) throw new ApiRequestError({ code: "NO_SESSION", message: "เซสชันหมดอายุ กรุณาเข้าสู่ระบบใหม่" });
  return t;
}

export async function listKioskAccounts(): Promise<KioskAccount[]> {
  return apiRequest("/kiosk-accounts", kioskAccountListSchema, { token: token() });
}

export async function createKioskAccount(body: CreateKioskAccountBody): Promise<void> {
  await apiRequest("/kiosk-accounts", createdIdSchema, { method: "POST", body, token: token() });
}

export async function deleteKioskAccount(id: string): Promise<void> {
  await apiRequest(`/kiosk-accounts/${id}`, messageSchema, { method: "DELETE", token: token() });
}
