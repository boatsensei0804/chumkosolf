"use client";

import {
  useMutation,
  useQuery,
  useQueryClient,
  type UseMutationResult,
  type UseQueryResult,
} from "@tanstack/react-query";

import type { ApiRequestError } from "@/lib/api/client";
import type { StudentDetail } from "@/shared/schemas/student";
import type { Advisee, MyProfile, UpdateMyAdviseeBody, UpdateMyProfileBody } from "@/shared/schemas/me";

import { getMyAdvisee, getMyProfile, listMyAdvisees, updateMyAdvisee, updateMyProfile } from "./api";

// business logic ของ "ของฉัน" อยู่ในชั้น hook — component แค่ใช้ data/loading/error

const meKeys = {
  profile: ["me", "profile"] as const,
  advisees: ["me", "advisees"] as const,
  advisee: (id: string) => ["me", "advisee", id] as const,
};

export function useMyProfile(): UseQueryResult<MyProfile, ApiRequestError> {
  return useQuery({
    queryKey: meKeys.profile,
    queryFn: getMyProfile,
  });
}

export function useUpdateMyProfile(): UseMutationResult<void, ApiRequestError, UpdateMyProfileBody> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: updateMyProfile,
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: meKeys.profile });
    },
  });
}

export function useMyAdvisees(): UseQueryResult<Advisee[], ApiRequestError> {
  return useQuery({
    queryKey: meKeys.advisees,
    queryFn: listMyAdvisees,
  });
}

export function useMyAdvisee(studentId: string): UseQueryResult<StudentDetail, ApiRequestError> {
  return useQuery({
    queryKey: meKeys.advisee(studentId),
    queryFn: () => getMyAdvisee(studentId),
    enabled: studentId !== "",
  });
}

export function useUpdateMyAdvisee(
  studentId: string,
): UseMutationResult<void, ApiRequestError, UpdateMyAdviseeBody> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: UpdateMyAdviseeBody) => updateMyAdvisee(studentId, body),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: meKeys.advisees });
      void qc.invalidateQueries({ queryKey: meKeys.advisee(studentId) });
    },
  });
}
