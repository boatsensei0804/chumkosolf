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
  AddAdvisorBody,
  ClassAdvisor,
  ClassBody,
  ClassDetail,
  ClassListItem,
  EnrollBody,
  Enrollment,
} from "@/shared/schemas/class";

import {
  addAdvisor,
  createClass,
  deleteClass,
  enrollStudent,
  enrollStudentsBulk,
  getClass,
  listAdvisors,
  listClasses,
  listEnrollments,
  removeAdvisor,
  removeEnrollment,
  updateClass,
} from "./api";

const keys = {
  all: ["classes"] as const,
  detail: (id: string) => ["classes", "detail", id] as const,
  advisors: (id: string) => ["classes", id, "advisors"] as const,
  students: (id: string) => ["classes", id, "students"] as const,
};

export function useClassList(): UseQueryResult<ClassListItem[], ApiRequestError> {
  return useQuery({ queryKey: keys.all, queryFn: listClasses });
}
export function useClass(id: string): UseQueryResult<ClassDetail, ApiRequestError> {
  return useQuery({ queryKey: keys.detail(id), queryFn: () => getClass(id), enabled: id !== "" });
}
export function useCreateClass(): UseMutationResult<string, ApiRequestError, ClassBody> {
  const qc = useQueryClient();
  return useMutation({ mutationFn: createClass, onSuccess: () => void qc.invalidateQueries({ queryKey: keys.all }) });
}
export function useUpdateClass(id: string): UseMutationResult<void, ApiRequestError, ClassBody> {
  const qc = useQueryClient();
  return useMutation({ mutationFn: (b: ClassBody) => updateClass(id, b), onSuccess: () => void qc.invalidateQueries({ queryKey: keys.all }) });
}
export function useDeleteClass(): UseMutationResult<void, ApiRequestError, string> {
  const qc = useQueryClient();
  return useMutation({ mutationFn: deleteClass, onSuccess: () => void qc.invalidateQueries({ queryKey: keys.all }) });
}

export function useAdvisors(classId: string): UseQueryResult<ClassAdvisor[], ApiRequestError> {
  return useQuery({ queryKey: keys.advisors(classId), queryFn: () => listAdvisors(classId), enabled: classId !== "" });
}
export function useAddAdvisor(classId: string): UseMutationResult<void, ApiRequestError, AddAdvisorBody> {
  const qc = useQueryClient();
  return useMutation({ mutationFn: (b: AddAdvisorBody) => addAdvisor(classId, b), onSuccess: () => void qc.invalidateQueries({ queryKey: keys.advisors(classId) }) });
}
export function useRemoveAdvisor(classId: string): UseMutationResult<void, ApiRequestError, string> {
  const qc = useQueryClient();
  return useMutation({ mutationFn: (advisorId: string) => removeAdvisor(classId, advisorId), onSuccess: () => void qc.invalidateQueries({ queryKey: keys.advisors(classId) }) });
}

export function useEnrollments(classId: string): UseQueryResult<Enrollment[], ApiRequestError> {
  return useQuery({ queryKey: keys.students(classId), queryFn: () => listEnrollments(classId), enabled: classId !== "" });
}
export function useEnrollStudent(classId: string): UseMutationResult<void, ApiRequestError, EnrollBody> {
  const qc = useQueryClient();
  return useMutation({ mutationFn: (b: EnrollBody) => enrollStudent(classId, b), onSuccess: () => void qc.invalidateQueries({ queryKey: keys.students(classId) }) });
}
export function useEnrollStudentsBulk(classId: string): UseMutationResult<void, ApiRequestError, string[]> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (ids: string[]) => enrollStudentsBulk(classId, ids),
    onSuccess: () => void qc.invalidateQueries({ queryKey: keys.students(classId) }),
  });
}
export function useRemoveEnrollment(classId: string): UseMutationResult<void, ApiRequestError, string> {
  const qc = useQueryClient();
  return useMutation({ mutationFn: (enrollmentId: string) => removeEnrollment(classId, enrollmentId), onSuccess: () => void qc.invalidateQueries({ queryKey: keys.students(classId) }) });
}
