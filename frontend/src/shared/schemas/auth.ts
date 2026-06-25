import { z } from "zod";

import { userRoleSchema, workGroupCodeSchema } from "./enums";

// auth schemas — ต้องตรงกับ backend (internal/service/auth_service.go)
// login: POST /api/v1/auth/login → LoginResult

// คำขอ login: school_code + username + password (ตรงกับ loginRequest ฝั่ง Go)
export const loginRequestSchema = z.object({
  schoolCode: z.string().min(1, "กรุณากรอกรหัสโรงเรียน"),
  username: z.string().min(1, "กรุณากรอกชื่อผู้ใช้"),
  password: z.string().min(1, "กรุณากรอกรหัสผ่าน"),
});
export type LoginRequest = z.infer<typeof loginRequestSchema>;

// payload ที่ส่งจริงไป backend (snake_case ตาม JSON contract)
export type LoginPayload = {
  school_code: string;
  username: string;
  password: string;
};

// กลุ่มงานที่ผู้ใช้สังกัด (ตรงกับ domain.WorkGroupMembership)
export const workGroupMembershipSchema = z.object({
  work_group_id: z.string(),
  code: workGroupCodeSchema,
  name: z.string(),
  is_group_admin: z.boolean(),
});
export type WorkGroupMembership = z.infer<typeof workGroupMembershipSchema>;

// ข้อมูลผู้ใช้ที่ backend ส่งกลับ (ไม่มี password) — ตรงกับ service.UserInfo
export const userInfoSchema = z.object({
  id: z.string(),
  username: z.string(),
  role: userRoleSchema,
  school_id: z.string(),
  is_school_admin: z.boolean(),
  work_groups: z.array(workGroupMembershipSchema),
});
export type UserInfo = z.infer<typeof userInfoSchema>;

// ผลของ login/refresh — ตรงกับ service.LoginResult
export const loginResultSchema = z.object({
  access_token: z.string(),
  refresh_token: z.string(),
  token_type: z.string(),
  expires_in: z.number(),
  user: userInfoSchema,
});
export type LoginResult = z.infer<typeof loginResultSchema>;
