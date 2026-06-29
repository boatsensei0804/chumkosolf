"use client";

import {
  useMutation,
  useQuery,
  useQueryClient,
  type UseMutationResult,
  type UseQueryResult,
} from "@tanstack/react-query";

import type { ApiRequestError } from "@/lib/api/client";
import type { Subject, SubjectBody } from "@/shared/schemas/subject";
import type {
  AssignmentBody,
  ConfigBody,
  SlotBody,
  TeachingAssignment,
  TimetableConfig,
  TimetableSlot,
} from "@/shared/schemas/timetable";
import type { AttendanceRosterEntry, SaveAttendanceBody } from "@/shared/schemas/attendance";
import type { CheckinOverview, FreeTeachers } from "@/shared/schemas/timetable";

import {
  clearSlot,
  createAssignment,
  createSubject,
  deleteAssignment,
  deleteSubject,
  getCheckinOverview,
  getConfig,
  getFreeTeachers,
  getSubjectRoster,
  listAssignments,
  listSlots,
  listSubjects,
  saveConfig,
  saveSubjectAttendance,
  setSlot,
  updateSubject,
} from "./api";

const keys = {
  subjects: ["subjects"] as const,
  assignments: ["teaching-assignments"] as const,
  config: ["timetable", "config"] as const,
  slots: (classId: string) => ["timetable", "slots", classId] as const,
  subjectRoster: (slotId: string, date: string) => ["subject-attendance", slotId, date] as const,
  checkinOverview: (date: string) => ["checkin-overview", date] as const,
  freeTeachers: (day: number) => ["free-teachers", day] as const,
};

// useFreeTeachers — รายชื่อครูว่างในแต่ละคาบของวันที่เลือก
export function useFreeTeachers(day: number): UseQueryResult<FreeTeachers, ApiRequestError> {
  return useQuery({
    queryKey: keys.freeTeachers(day),
    queryFn: () => getFreeTeachers(day),
    enabled: day >= 1 && day <= 7,
  });
}

// useCheckinOverview — กริดคาบที่ครูสอน + สถานะเช็ค/สรุปสัปดาห์ (ตามวันที่/สัปดาห์)
export function useCheckinOverview(date: string): UseQueryResult<CheckinOverview, ApiRequestError> {
  return useQuery({
    queryKey: keys.checkinOverview(date),
    queryFn: () => getCheckinOverview(date),
    enabled: date !== "",
  });
}

// --- รายวิชา ---
export function useSubjects(): UseQueryResult<Subject[], ApiRequestError> {
  return useQuery({ queryKey: keys.subjects, queryFn: listSubjects });
}
export function useCreateSubject(): UseMutationResult<void, ApiRequestError, SubjectBody> {
  const qc = useQueryClient();
  return useMutation({ mutationFn: createSubject, onSuccess: () => void qc.invalidateQueries({ queryKey: keys.subjects }) });
}
export function useUpdateSubject(id: string): UseMutationResult<void, ApiRequestError, SubjectBody> {
  const qc = useQueryClient();
  return useMutation({ mutationFn: (b: SubjectBody) => updateSubject(id, b), onSuccess: () => void qc.invalidateQueries({ queryKey: keys.subjects }) });
}
export function useDeleteSubject(): UseMutationResult<void, ApiRequestError, string> {
  const qc = useQueryClient();
  return useMutation({ mutationFn: deleteSubject, onSuccess: () => void qc.invalidateQueries({ queryKey: keys.subjects }) });
}

// --- มอบหมายการสอน ---
export function useAssignments(opts?: { enabled?: boolean }): UseQueryResult<TeachingAssignment[], ApiRequestError> {
  return useQuery({ queryKey: keys.assignments, queryFn: listAssignments, enabled: opts?.enabled ?? true });
}
export function useCreateAssignment(): UseMutationResult<void, ApiRequestError, AssignmentBody> {
  const qc = useQueryClient();
  return useMutation({ mutationFn: createAssignment, onSuccess: () => void qc.invalidateQueries({ queryKey: keys.assignments }) });
}
export function useDeleteAssignment(): UseMutationResult<void, ApiRequestError, string> {
  const qc = useQueryClient();
  return useMutation({ mutationFn: deleteAssignment, onSuccess: () => void qc.invalidateQueries({ queryKey: keys.assignments }) });
}

// --- ตั้งค่าคาบ ---
export function useTimetableConfig(): UseQueryResult<TimetableConfig, ApiRequestError> {
  return useQuery({ queryKey: keys.config, queryFn: getConfig });
}
export function useSaveConfig(): UseMutationResult<void, ApiRequestError, ConfigBody> {
  const qc = useQueryClient();
  return useMutation({ mutationFn: saveConfig, onSuccess: () => void qc.invalidateQueries({ queryKey: keys.config }) });
}

// --- ช่องตารางสอน ---
export function useSlots(classId: string): UseQueryResult<TimetableSlot[], ApiRequestError> {
  return useQuery({ queryKey: keys.slots(classId), queryFn: () => listSlots(classId), enabled: classId !== "" });
}
export function useSetSlot(classId: string): UseMutationResult<void, ApiRequestError, SlotBody> {
  const qc = useQueryClient();
  return useMutation({ mutationFn: (b: SlotBody) => setSlot(classId, b), onSuccess: () => void qc.invalidateQueries({ queryKey: keys.slots(classId) }) });
}
export function useClearSlot(classId: string): UseMutationResult<void, ApiRequestError, string> {
  const qc = useQueryClient();
  return useMutation({ mutationFn: (slotId: string) => clearSlot(classId, slotId), onSuccess: () => void qc.invalidateQueries({ queryKey: keys.slots(classId) }) });
}

// --- เช็คชื่อรายวิชา ---
export function useSubjectRoster(slotId: string, date: string): UseQueryResult<AttendanceRosterEntry[], ApiRequestError> {
  return useQuery({
    queryKey: keys.subjectRoster(slotId, date),
    queryFn: () => getSubjectRoster(slotId, date),
    enabled: slotId !== "" && date !== "",
  });
}
export function useSaveSubjectAttendance(slotId: string): UseMutationResult<void, ApiRequestError, SaveAttendanceBody> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (b: SaveAttendanceBody) => saveSubjectAttendance(slotId, b),
    onSuccess: (_d, b) => {
      void qc.invalidateQueries({ queryKey: keys.subjectRoster(slotId, b.date) });
      void qc.invalidateQueries({ queryKey: ["checkin-overview"] }); // อัปเดตสถานะในกริด
    },
  });
}
