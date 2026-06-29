import { ApiRequestError, apiRequest, apiUpload } from "@/lib/api/client";
import { getAccessToken } from "@/features/auth/storage";
import {
  recognizeResultSchema,
  reindexResultSchema,
  type RecognizeResult,
  type ReindexResult,
} from "@/shared/schemas/face";

function token(): string {
  const t = getAccessToken();
  if (!t) throw new ApiRequestError({ code: "NO_SESSION", message: "เซสชันหมดอายุ กรุณาเข้าสู่ระบบใหม่" });
  return t;
}

// สร้างฐานใบหน้าใหม่จากรูปนักเรียนทั้งหมด
export async function reindexFace(): Promise<ReindexResult> {
  return apiRequest("/face/reindex", reindexResultSchema, { method: "POST", token: token() });
}

// ส่งชุดเฟรมจากกล้อง (สำหรับตรวจ liveness) ไปจดจำ + บันทึกเช็คชื่อ
export async function recognizeFace(frames: Blob[]): Promise<RecognizeResult> {
  const form = new FormData();
  frames.forEach((f, i) => form.append("file", f, `frame-${i}.jpg`));
  return apiUpload("/face/recognize", recognizeResultSchema, form, { token: token() });
}
