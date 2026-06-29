"use client";

import { useQuery, type UseQueryResult } from "@tanstack/react-query";

import { ApiRequestError, apiRequest } from "@/lib/api/client";
import { getAccessToken } from "@/features/auth/storage";
import { dashboardSchema, type Dashboard } from "@/shared/schemas/dashboard";

export function useDashboard(): UseQueryResult<Dashboard, ApiRequestError> {
  return useQuery({
    queryKey: ["dashboard"],
    queryFn: () => {
      const token = getAccessToken();
      if (!token) throw new ApiRequestError({ code: "NO_SESSION", message: "เซสชันหมดอายุ กรุณาเข้าสู่ระบบใหม่" });
      return apiRequest("/dashboard", dashboardSchema, { token });
    },
  });
}
