"use client";

import {
  useMutation,
  useQuery,
  useQueryClient,
  type UseMutationResult,
  type UseQueryResult,
} from "@tanstack/react-query";

import type { ApiRequestError } from "@/lib/api/client";
import type { AttendanceRosterEntry, SaveAttendanceBody } from "@/shared/schemas/attendance";
import type { BehaviorBody, BehaviorSummary } from "@/shared/schemas/behavior";

import {
  addBehavior,
  deleteBehavior,
  getBehavior,
  getRoster,
  saveAttendance,
} from "./api";

const keys = {
  roster: (classId: string, date: string) => ["attendance", classId, date] as const,
  behavior: (studentId: string) => ["behavior", studentId] as const,
};

// --- เช็คชื่อ ---
export function useRoster(
  classId: string,
  date: string,
): UseQueryResult<AttendanceRosterEntry[], ApiRequestError> {
  return useQuery({
    queryKey: keys.roster(classId, date),
    queryFn: () => getRoster(classId, date),
    enabled: classId !== "" && date !== "",
  });
}

export function useSaveAttendance(
  classId: string,
): UseMutationResult<void, ApiRequestError, SaveAttendanceBody> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: SaveAttendanceBody) => saveAttendance(classId, body),
    onSuccess: (_data, body) =>
      void qc.invalidateQueries({ queryKey: keys.roster(classId, body.date) }),
  });
}

// --- ความประพฤติ ---
export function useBehavior(studentId: string): UseQueryResult<BehaviorSummary, ApiRequestError> {
  return useQuery({
    queryKey: keys.behavior(studentId),
    queryFn: () => getBehavior(studentId),
    enabled: studentId !== "",
  });
}

export function useAddBehavior(
  studentId: string,
): UseMutationResult<void, ApiRequestError, BehaviorBody> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: BehaviorBody) => addBehavior(studentId, body),
    onSuccess: () => void qc.invalidateQueries({ queryKey: keys.behavior(studentId) }),
  });
}

export function useDeleteBehavior(
  studentId: string,
): UseMutationResult<void, ApiRequestError, string> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (recordId: string) => deleteBehavior(studentId, recordId),
    onSuccess: () => void qc.invalidateQueries({ queryKey: keys.behavior(studentId) }),
  });
}
