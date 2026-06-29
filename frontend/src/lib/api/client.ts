import { z } from "zod";

import { apiResponseSchema, type ApiError, type ApiMeta } from "@/shared/schemas/api";
import { getSelectedSemesterId } from "./termOverride";

// applyTermOverride ใส่ header เทอมทำงานที่ผู้ใช้เลือก (ถ้ามี) — backend จะ validate กับโรงเรียนใน token
function applyTermOverride(headers: Record<string, string>): void {
  const sem = getSelectedSemesterId();
  if (sem !== "") headers["X-Semester-Id"] = sem;
}

// base URL ของ backend (ตั้งผ่าน NEXT_PUBLIC_API_BASE_URL ใน .env)
const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080/api/v1";

// ApiRequestError พก code + ข้อความไทยจาก backend เพื่อให้ UI แสดงได้ตรง
export class ApiRequestError extends Error {
  readonly code: string;

  constructor(error: ApiError) {
    super(error.message);
    this.name = "ApiRequestError";
    this.code = error.code;
  }
}

type RequestOptions = {
  method?: "GET" | "POST" | "PUT" | "PATCH" | "DELETE";
  body?: unknown;
  token?: string;
  signal?: AbortSignal;
};

// apiRequestWithMeta ยิง request แล้ว parse envelope ด้วย zod; คืนทั้ง data + meta (สำหรับ pagination)
// โยน ApiRequestError ถ้า network ล้ม, รูปแบบผิด, หรือ backend ตอบ error
export async function apiRequestWithMeta<T extends z.ZodTypeAny>(
  path: string,
  dataSchema: T,
  options: RequestOptions = {},
): Promise<{ data: z.infer<T>; meta: ApiMeta | null }> {
  const { method = "GET", body, token, signal } = options;

  const headers: Record<string, string> = { "Content-Type": "application/json" };
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }
  applyTermOverride(headers);

  let response: Response;
  try {
    response = await fetch(`${API_BASE_URL}${path}`, {
      method,
      headers,
      body: body === undefined ? undefined : JSON.stringify(body),
      signal,
    });
  } catch {
    throw new ApiRequestError({
      code: "NETWORK_ERROR",
      message: "เชื่อมต่อเซิร์ฟเวอร์ไม่สำเร็จ กรุณาลองใหม่อีกครั้ง",
    });
  }

  const raw: unknown = await response.json().catch(() => null);
  const parsed = apiResponseSchema(dataSchema).safeParse(raw);

  if (!parsed.success) {
    throw new ApiRequestError({
      code: "INVALID_RESPONSE",
      message: "เซิร์ฟเวอร์ตอบกลับในรูปแบบที่ไม่รองรับ",
    });
  }

  const envelope = parsed.data;
  if (!envelope.success || envelope.data === null) {
    throw new ApiRequestError(
      envelope.error ?? { code: "UNKNOWN_ERROR", message: "เกิดข้อผิดพลาดที่ไม่ทราบสาเหตุ" },
    );
  }

  return { data: envelope.data, meta: envelope.meta ?? null };
}

// apiRequest คืนเฉพาะ data ที่ผ่าน validate (กรณีไม่ต้องใช้ meta)
export async function apiRequest<T extends z.ZodTypeAny>(
  path: string,
  dataSchema: T,
  options: RequestOptions = {},
): Promise<z.infer<T>> {
  const { data } = await apiRequestWithMeta(path, dataSchema, options);
  return data;
}

// apiUpload ส่งไฟล์แบบ multipart/form-data (เช่นไฟล์แนบผลงานครู)
// ไม่ตั้ง Content-Type เอง ปล่อยให้เบราว์เซอร์ใส่ boundary; parse envelope ด้วย zod เหมือน apiRequest
export async function apiUpload<T extends z.ZodTypeAny>(
  path: string,
  dataSchema: T,
  formData: FormData,
  options: { token?: string; signal?: AbortSignal } = {},
): Promise<z.infer<T>> {
  const { token, signal } = options;

  const headers: Record<string, string> = {};
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }
  applyTermOverride(headers);

  let response: Response;
  try {
    response = await fetch(`${API_BASE_URL}${path}`, {
      method: "POST",
      headers,
      body: formData,
      signal,
    });
  } catch {
    throw new ApiRequestError({
      code: "NETWORK_ERROR",
      message: "เชื่อมต่อเซิร์ฟเวอร์ไม่สำเร็จ กรุณาลองใหม่อีกครั้ง",
    });
  }

  const raw: unknown = await response.json().catch(() => null);
  const parsed = apiResponseSchema(dataSchema).safeParse(raw);

  if (!parsed.success) {
    throw new ApiRequestError({
      code: "INVALID_RESPONSE",
      message: "เซิร์ฟเวอร์ตอบกลับในรูปแบบที่ไม่รองรับ",
    });
  }

  const envelope = parsed.data;
  if (!envelope.success || envelope.data === null) {
    throw new ApiRequestError(
      envelope.error ?? { code: "UNKNOWN_ERROR", message: "เกิดข้อผิดพลาดที่ไม่ทราบสาเหตุ" },
    );
  }

  return envelope.data;
}
