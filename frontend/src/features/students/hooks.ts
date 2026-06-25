"use client";

import {
  useMutation,
  useQuery,
  useQueryClient,
  type UseMutationResult,
  type UseQueryResult,
} from "@tanstack/react-query";

import type { ApiRequestError } from "@/lib/api/client";
import type { CreateStudentBody, StudentDetail, UpdateStudentBody } from "@/shared/schemas/student";
import type { LinkGuardianBody, StudentGuardian } from "@/shared/schemas/guardian";

import {
  createStudent,
  deleteStudent,
  getStudent,
  linkGuardian,
  listStudentGuardians,
  listStudents,
  unlinkGuardian,
  updateStudent,
  type StudentListResult,
} from "./api";

const keys = {
  all: ["students"] as const,
  list: (p: number, s: number) => ["students", "list", p, s] as const,
  detail: (id: string) => ["students", "detail", id] as const,
  guardians: (id: string) => ["students", id, "guardians"] as const,
};

export function useStudentList(
  page: number,
  pageSize: number,
  enabled = true,
): UseQueryResult<StudentListResult, ApiRequestError> {
  return useQuery({ queryKey: keys.list(page, pageSize), queryFn: () => listStudents(page, pageSize), enabled });
}

export function useStudent(id: string): UseQueryResult<StudentDetail, ApiRequestError> {
  return useQuery({ queryKey: keys.detail(id), queryFn: () => getStudent(id), enabled: id !== "" });
}

export function useCreateStudent(): UseMutationResult<string, ApiRequestError, CreateStudentBody> {
  const qc = useQueryClient();
  return useMutation({ mutationFn: createStudent, onSuccess: () => void qc.invalidateQueries({ queryKey: keys.all }) });
}

export function useUpdateStudent(id: string): UseMutationResult<void, ApiRequestError, UpdateStudentBody> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: UpdateStudentBody) => updateStudent(id, body),
    onSuccess: () => void qc.invalidateQueries({ queryKey: keys.all }),
  });
}

export function useDeleteStudent(): UseMutationResult<void, ApiRequestError, string> {
  const qc = useQueryClient();
  return useMutation({ mutationFn: deleteStudent, onSuccess: () => void qc.invalidateQueries({ queryKey: keys.all }) });
}

// --- guardian links ---
export function useStudentGuardians(studentId: string): UseQueryResult<StudentGuardian[], ApiRequestError> {
  return useQuery({
    queryKey: keys.guardians(studentId),
    queryFn: () => listStudentGuardians(studentId),
    enabled: studentId !== "",
  });
}

export function useLinkGuardian(studentId: string): UseMutationResult<void, ApiRequestError, LinkGuardianBody> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: LinkGuardianBody) => linkGuardian(studentId, body),
    onSuccess: () => void qc.invalidateQueries({ queryKey: keys.guardians(studentId) }),
  });
}

export function useUnlinkGuardian(studentId: string): UseMutationResult<void, ApiRequestError, string> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (linkId: string) => unlinkGuardian(studentId, linkId),
    onSuccess: () => void qc.invalidateQueries({ queryKey: keys.guardians(studentId) }),
  });
}
