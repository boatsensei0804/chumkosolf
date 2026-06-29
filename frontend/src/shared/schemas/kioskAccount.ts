import { z } from "zod";

// บัญชีเครื่องสแกนหน้า (role kiosk) — ตรงกับ backend KioskAccountDTO
export const kioskAccountSchema = z.object({
  id: z.string(),
  username: z.string(),
  is_active: z.boolean(),
  created_at: z.string(),
});
export type KioskAccount = z.infer<typeof kioskAccountSchema>;
export const kioskAccountListSchema = z.array(kioskAccountSchema);

export type CreateKioskAccountBody = { username: string; password: string };
