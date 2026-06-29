"use client";

import {
  useMutation,
  useQuery,
  useQueryClient,
  type UseMutationResult,
  type UseQueryResult,
} from "@tanstack/react-query";

import type { ApiRequestError } from "@/lib/api/client";
import type {
  CreatePersonnelBody,
  PersonnelDetail,
  UpdatePersonnelBody,
} from "@/shared/schemas/personnel";

import {
  createPersonnel,
  deletePersonnel,
  getPersonnel,
  listPersonnel,
  updatePersonnel,
  type PersonnelListResult,
} from "./api";

// business logic ของ personnel อยู่ในชั้น hook — component แค่ใช้ data/loading/error

const personnelKeys = {
  all: ["personnel"] as const,
  list: (page: number, pageSize: number, search: string) => ["personnel", "list", page, pageSize, search] as const,
  detail: (id: string) => ["personnel", "detail", id] as const,
};

export function usePersonnelList(
  page: number,
  pageSize: number,
  search = "",
  enabled = true,
): UseQueryResult<PersonnelListResult, ApiRequestError> {
  return useQuery({
    queryKey: personnelKeys.list(page, pageSize, search),
    queryFn: () => listPersonnel(page, pageSize, search),
    enabled,
  });
}

export function usePersonnel(id: string): UseQueryResult<PersonnelDetail, ApiRequestError> {
  return useQuery({
    queryKey: personnelKeys.detail(id),
    queryFn: () => getPersonnel(id),
    enabled: id !== "",
  });
}

export function useCreatePersonnel(): UseMutationResult<string, ApiRequestError, CreatePersonnelBody> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createPersonnel,
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: personnelKeys.all });
    },
  });
}

export function useUpdatePersonnel(
  id: string,
): UseMutationResult<void, ApiRequestError, UpdatePersonnelBody> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: UpdatePersonnelBody) => updatePersonnel(id, body),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: personnelKeys.all });
    },
  });
}

export function useDeletePersonnel(): UseMutationResult<void, ApiRequestError, string> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deletePersonnel(id),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: personnelKeys.all });
    },
  });
}
