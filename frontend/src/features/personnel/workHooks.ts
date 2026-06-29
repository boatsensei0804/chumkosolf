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
  PersonnelWorkRecord,
  WorkBody,
  WorkFileRecord,
  WorkFileType,
} from "@/shared/schemas/personnelWork";

import {
  createWork,
  deleteWork,
  deleteWorkFile,
  listWorkFiles,
  listWorks,
  updateWork,
  uploadWorkFile,
} from "./workApi";

const keys = {
  works: (id: string) => ["personnel", id, "works"] as const,
  files: (id: string, workId: string) => ["personnel", id, "works", workId, "files"] as const,
};

// --- ผลงาน ---
export function useWorks(personnelId: string): UseQueryResult<PersonnelWorkRecord[], ApiRequestError> {
  return useQuery({
    queryKey: keys.works(personnelId),
    queryFn: () => listWorks(personnelId),
    enabled: personnelId !== "",
  });
}

export function useCreateWork(personnelId: string): UseMutationResult<void, ApiRequestError, WorkBody> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: WorkBody) => createWork(personnelId, body),
    onSuccess: () => void qc.invalidateQueries({ queryKey: keys.works(personnelId) }),
  });
}

export function useUpdateWork(
  personnelId: string,
): UseMutationResult<void, ApiRequestError, { workId: string; body: WorkBody }> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ workId, body }: { workId: string; body: WorkBody }) =>
      updateWork(personnelId, workId, body),
    onSuccess: () => void qc.invalidateQueries({ queryKey: keys.works(personnelId) }),
  });
}

export function useDeleteWork(personnelId: string): UseMutationResult<void, ApiRequestError, string> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (workId: string) => deleteWork(personnelId, workId),
    onSuccess: () => void qc.invalidateQueries({ queryKey: keys.works(personnelId) }),
  });
}

// --- ไฟล์แนบ ---
export function useWorkFiles(
  personnelId: string,
  workId: string,
  enabled: boolean,
): UseQueryResult<WorkFileRecord[], ApiRequestError> {
  return useQuery({
    queryKey: keys.files(personnelId, workId),
    queryFn: () => listWorkFiles(personnelId, workId),
    enabled: enabled && personnelId !== "" && workId !== "",
  });
}

export function useUploadWorkFile(
  personnelId: string,
  workId: string,
): UseMutationResult<void, ApiRequestError, { fileType: WorkFileType; file: File }> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ fileType, file }: { fileType: WorkFileType; file: File }) =>
      uploadWorkFile(personnelId, workId, fileType, file),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: keys.files(personnelId, workId) });
      void qc.invalidateQueries({ queryKey: keys.works(personnelId) });
    },
  });
}

export function useDeleteWorkFile(
  personnelId: string,
  workId: string,
): UseMutationResult<void, ApiRequestError, string> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (fileId: string) => deleteWorkFile(personnelId, workId, fileId),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: keys.files(personnelId, workId) });
      void qc.invalidateQueries({ queryKey: keys.works(personnelId) });
    },
  });
}
