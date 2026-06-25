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
  AcademicStandingRecord,
  AdminPositionRecord,
  CreatePositionBody,
  StandingBody,
} from "@/shared/schemas/personnelExtra";

import {
  createPosition,
  createStanding,
  deletePosition,
  deleteStanding,
  listPositions,
  listStandings,
} from "./extraApi";

const keys = {
  positions: (id: string) => ["personnel", id, "positions"] as const,
  standings: (id: string) => ["personnel", id, "standings"] as const,
};

// --- ตำแหน่งบริหาร ---
export function usePositions(personnelId: string): UseQueryResult<AdminPositionRecord[], ApiRequestError> {
  return useQuery({
    queryKey: keys.positions(personnelId),
    queryFn: () => listPositions(personnelId),
    enabled: personnelId !== "",
  });
}

export function useCreatePosition(personnelId: string): UseMutationResult<void, ApiRequestError, CreatePositionBody> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreatePositionBody) => createPosition(personnelId, body),
    onSuccess: () => void qc.invalidateQueries({ queryKey: keys.positions(personnelId) }),
  });
}

export function useDeletePosition(personnelId: string): UseMutationResult<void, ApiRequestError, string> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (posId: string) => deletePosition(personnelId, posId),
    onSuccess: () => void qc.invalidateQueries({ queryKey: keys.positions(personnelId) }),
  });
}

// --- วิทยฐานะ ---
export function useStandings(personnelId: string): UseQueryResult<AcademicStandingRecord[], ApiRequestError> {
  return useQuery({
    queryKey: keys.standings(personnelId),
    queryFn: () => listStandings(personnelId),
    enabled: personnelId !== "",
  });
}

export function useCreateStanding(personnelId: string): UseMutationResult<void, ApiRequestError, StandingBody> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: StandingBody) => createStanding(personnelId, body),
    onSuccess: () => void qc.invalidateQueries({ queryKey: keys.standings(personnelId) }),
  });
}

export function useDeleteStanding(personnelId: string): UseMutationResult<void, ApiRequestError, string> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (sid: string) => deleteStanding(personnelId, sid),
    onSuccess: () => void qc.invalidateQueries({ queryKey: keys.standings(personnelId) }),
  });
}
