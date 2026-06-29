"use client";

import {
  useMutation,
  useQuery,
  useQueryClient,
  type UseMutationResult,
  type UseQueryResult,
} from "@tanstack/react-query";

import { ApiRequestError, apiRequest } from "@/lib/api/client";
import { getAccessToken } from "@/features/auth/storage";
import { messageSchema } from "@/shared/schemas/personnel";
import { schoolSchema, type School, type UpdateSchoolBody } from "@/shared/schemas/school";

function token(): string {
  const t = getAccessToken();
  if (!t) throw new ApiRequestError({ code: "NO_SESSION", message: "เซสชันหมดอายุ กรุณาเข้าสู่ระบบใหม่" });
  return t;
}

export function useSchool(): UseQueryResult<School, ApiRequestError> {
  return useQuery({
    queryKey: ["school"],
    queryFn: () => apiRequest("/school", schoolSchema, { token: token() }),
  });
}

export function useUpdateSchool(): UseMutationResult<void, ApiRequestError, UpdateSchoolBody> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: UpdateSchoolBody) =>
      apiRequest("/school", messageSchema, { method: "PUT", body, token: token() }).then(() => undefined),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ["school"] }),
  });
}
