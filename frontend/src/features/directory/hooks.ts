"use client";

import { useQuery, type UseQueryResult } from "@tanstack/react-query";

import type { ApiRequestError } from "@/lib/api/client";
import type { DirectoryClass, DirectoryStudent, DirectoryStudentClass } from "@/shared/schemas/directory";

import { listDirectoryClassStudents, listDirectoryClasses, searchDirectoryStudents } from "./api";

const keys = {
  classes: ["directory", "classes"] as const,
  classStudents: (classId: string) => ["directory", "classStudents", classId] as const,
  search: (q: string) => ["directory", "search", q] as const,
};

export function useDirectoryClasses(): UseQueryResult<DirectoryClass[], ApiRequestError> {
  return useQuery({ queryKey: keys.classes, queryFn: listDirectoryClasses });
}

export function useDirectoryClassStudents(classId: string): UseQueryResult<DirectoryStudent[], ApiRequestError> {
  return useQuery({
    queryKey: keys.classStudents(classId),
    queryFn: () => listDirectoryClassStudents(classId),
    enabled: classId !== "",
  });
}

export function useDirectoryStudentSearch(q: string): UseQueryResult<DirectoryStudentClass[], ApiRequestError> {
  return useQuery({
    queryKey: keys.search(q),
    queryFn: () => searchDirectoryStudents(q),
    enabled: q.trim().length > 0,
  });
}
