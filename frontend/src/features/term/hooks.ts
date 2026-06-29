"use client";

import { useQuery, type UseQueryResult } from "@tanstack/react-query";

import { ApiRequestError, apiRequest } from "@/lib/api/client";
import { getAccessToken } from "@/features/auth/storage";
import { currentTermSchema, type CurrentTerm } from "@/shared/schemas/term";

async function getCurrentTerm(): Promise<CurrentTerm> {
  const token = getAccessToken();
  if (!token) throw new ApiRequestError({ code: "NO_SESSION", message: "เซสชันหมดอายุ กรุณาเข้าสู่ระบบใหม่" });
  return apiRequest("/current-term", currentTermSchema, { token });
}

export function useCurrentTerm(): UseQueryResult<CurrentTerm, ApiRequestError> {
  return useQuery({ queryKey: ["current-term"], queryFn: getCurrentTerm });
}
