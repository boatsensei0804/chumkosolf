import { apiRequest } from "@/lib/api/client";
import {
  loginResultSchema,
  type LoginRequest,
  type LoginResult,
} from "@/shared/schemas/auth";

// login เรียก POST /auth/login — แปลง form values (camelCase) เป็น payload (snake_case)
export async function login(values: LoginRequest): Promise<LoginResult> {
  return apiRequest("/auth/login", loginResultSchema, {
    method: "POST",
    body: {
      school_code: values.schoolCode,
      username: values.username,
      password: values.password,
    },
  });
}
