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
  AssignWorkGroupBody,
  WorkGroup,
  WorkGroupMembership,
} from "@/shared/schemas/workGroup";

import {
  assignWorkGroup,
  listPersonnelWorkGroups,
  listWorkGroups,
  unassignWorkGroup,
} from "./workGroupApi";

const keys = {
  all: ["work-groups"] as const,
  personnel: (id: string) => ["personnel", id, "work-groups"] as const,
};

export function useWorkGroups(): UseQueryResult<WorkGroup[], ApiRequestError> {
  return useQuery({ queryKey: keys.all, queryFn: listWorkGroups });
}

export function usePersonnelWorkGroups(
  personnelId: string,
): UseQueryResult<WorkGroupMembership[], ApiRequestError> {
  return useQuery({
    queryKey: keys.personnel(personnelId),
    queryFn: () => listPersonnelWorkGroups(personnelId),
    enabled: personnelId !== "",
  });
}

export function useAssignWorkGroup(
  personnelId: string,
): UseMutationResult<void, ApiRequestError, AssignWorkGroupBody> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: AssignWorkGroupBody) => assignWorkGroup(personnelId, body),
    onSuccess: () => void qc.invalidateQueries({ queryKey: keys.personnel(personnelId) }),
  });
}

export function useUnassignWorkGroup(
  personnelId: string,
): UseMutationResult<void, ApiRequestError, string> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (workGroupId: string) => unassignWorkGroup(personnelId, workGroupId),
    onSuccess: () => void qc.invalidateQueries({ queryKey: keys.personnel(personnelId) }),
  });
}
