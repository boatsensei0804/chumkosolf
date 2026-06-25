"use client";

import { useMutation, type UseMutationResult } from "@tanstack/react-query";

import { ApiRequestError } from "@/lib/api/client";
import type { LoginRequest, LoginResult } from "@/shared/schemas/auth";

import { login } from "./api";

// useLogin ห่อ business logic ของการเรียก API login
// การเก็บ session ทำที่ AuthContext.setSession (ส่งผ่าน onSuccess) เพื่อให้ route guard เห็นสถานะทันที
// component แค่เรียก mutate แล้วใช้ isPending/error — ไม่มี logic ปนใน UI
export function useLogin(
  onSuccess?: (result: LoginResult) => void,
): UseMutationResult<LoginResult, ApiRequestError, LoginRequest> {
  return useMutation<LoginResult, ApiRequestError, LoginRequest>({
    mutationFn: login,
    onSuccess: (result) => {
      onSuccess?.(result);
    },
  });
}
