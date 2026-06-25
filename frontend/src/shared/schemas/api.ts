import { z } from "zod";

// response envelope มาตรฐานจาก backend: { success, data, error, meta }
// ตรงกับ httputil.APIResponse ฝั่ง Go

export const apiErrorSchema = z.object({
  code: z.string(),
  message: z.string(),
});
export type ApiError = z.infer<typeof apiErrorSchema>;

export const apiMetaSchema = z.object({
  page: z.number().optional(),
  total: z.number().optional(),
});
export type ApiMeta = z.infer<typeof apiMetaSchema>;

// สร้าง schema ของ response สำหรับ data ชนิดใด ๆ
export function apiResponseSchema<T extends z.ZodTypeAny>(dataSchema: T) {
  return z.object({
    success: z.boolean(),
    data: dataSchema.nullable(),
    error: apiErrorSchema.nullable(),
    meta: apiMetaSchema.nullish(),
  });
}

export type ApiResponse<T> = {
  success: boolean;
  data: T | null;
  error: ApiError | null;
  meta?: ApiMeta | null;
};
