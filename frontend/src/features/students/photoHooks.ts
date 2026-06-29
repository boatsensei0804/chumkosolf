"use client";

import {
  useMutation,
  useQuery,
  useQueryClient,
  type UseMutationResult,
  type UseQueryResult,
} from "@tanstack/react-query";

import type { ApiRequestError } from "@/lib/api/client";
import type { StudentPhoto } from "@/shared/schemas/student";

import { deleteStudentPhoto, listStudentPhotos, setStudentPhotoPrimary, uploadStudentPhoto } from "./photoApi";

const photosKey = (studentId: string): readonly string[] => ["students", "photos", studentId];

export function useStudentPhotos(studentId: string): UseQueryResult<StudentPhoto[], ApiRequestError> {
  return useQuery({
    queryKey: photosKey(studentId),
    queryFn: () => listStudentPhotos(studentId),
    enabled: studentId !== "",
  });
}

export function useUploadStudentPhoto(studentId: string): UseMutationResult<StudentPhoto, ApiRequestError, File> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (file: File) => uploadStudentPhoto(studentId, file),
    onSuccess: () => void qc.invalidateQueries({ queryKey: photosKey(studentId) }),
  });
}

export function useSetStudentPhotoPrimary(studentId: string): UseMutationResult<void, ApiRequestError, string> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (photoId: string) => setStudentPhotoPrimary(studentId, photoId),
    onSuccess: () => void qc.invalidateQueries({ queryKey: photosKey(studentId) }),
  });
}

export function useDeleteStudentPhoto(studentId: string): UseMutationResult<void, ApiRequestError, string> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (photoId: string) => deleteStudentPhoto(studentId, photoId),
    onSuccess: () => void qc.invalidateQueries({ queryKey: photosKey(studentId) }),
  });
}
