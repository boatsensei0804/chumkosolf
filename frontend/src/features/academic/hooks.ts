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
import { createdIdSchema, messageSchema } from "@/shared/schemas/personnel";
import {
  academicYearListSchema,
  semesterListSchema,
  type AcademicYear,
  type CreateSemesterBody,
  type CreateYearBody,
  type Semester,
} from "@/shared/schemas/academic";

function token(): string {
  const t = getAccessToken();
  if (!t) throw new ApiRequestError({ code: "NO_SESSION", message: "เซสชันหมดอายุ กรุณาเข้าสู่ระบบใหม่" });
  return t;
}

const keys = { years: ["academic", "years"] as const, semesters: ["academic", "semesters"] as const };

// --- ปีการศึกษา ---
export function useYears(): UseQueryResult<AcademicYear[], ApiRequestError> {
  return useQuery({
    queryKey: keys.years,
    queryFn: () => apiRequest("/academic/years", academicYearListSchema, { token: token() }),
  });
}
export function useCreateYear(): UseMutationResult<void, ApiRequestError, CreateYearBody> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateYearBody) =>
      apiRequest("/academic/years", createdIdSchema, { method: "POST", body, token: token() }).then(() => undefined),
    onSuccess: () => void qc.invalidateQueries({ queryKey: keys.years }),
  });
}
export function useSetCurrentYear(): UseMutationResult<void, ApiRequestError, string> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      apiRequest(`/academic/years/${id}/current`, messageSchema, { method: "POST", token: token() }).then(() => undefined),
    onSuccess: () => void qc.invalidateQueries({ queryKey: keys.years }),
  });
}

// --- ภาคเรียน ---
export function useSemesters(): UseQueryResult<Semester[], ApiRequestError> {
  return useQuery({
    queryKey: keys.semesters,
    queryFn: () => apiRequest("/academic/semesters", semesterListSchema, { token: token() }),
  });
}
export function useCreateSemester(): UseMutationResult<void, ApiRequestError, CreateSemesterBody> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateSemesterBody) =>
      apiRequest("/academic/semesters", createdIdSchema, { method: "POST", body, token: token() }).then(() => undefined),
    onSuccess: () => void qc.invalidateQueries({ queryKey: keys.semesters }),
  });
}
export function useSetCurrentSemester(): UseMutationResult<void, ApiRequestError, string> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) =>
      apiRequest(`/academic/semesters/${id}/current`, messageSchema, { method: "POST", token: token() }).then(() => undefined),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: keys.semesters });
      void qc.invalidateQueries({ queryKey: ["current-term"] });
    },
  });
}
