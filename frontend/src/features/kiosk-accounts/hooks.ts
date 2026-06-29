"use client";

import {
  useMutation,
  useQuery,
  useQueryClient,
  type UseMutationResult,
  type UseQueryResult,
} from "@tanstack/react-query";

import type { ApiRequestError } from "@/lib/api/client";
import type { CreateKioskAccountBody, KioskAccount } from "@/shared/schemas/kioskAccount";

import { createKioskAccount, deleteKioskAccount, listKioskAccounts } from "./api";

const key = ["kiosk-accounts"] as const;

export function useKioskAccounts(): UseQueryResult<KioskAccount[], ApiRequestError> {
  return useQuery({ queryKey: key, queryFn: listKioskAccounts });
}

export function useCreateKioskAccount(): UseMutationResult<void, ApiRequestError, CreateKioskAccountBody> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createKioskAccount,
    onSuccess: () => void qc.invalidateQueries({ queryKey: key }),
  });
}

export function useDeleteKioskAccount(): UseMutationResult<void, ApiRequestError, string> {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: deleteKioskAccount,
    onSuccess: () => void qc.invalidateQueries({ queryKey: key }),
  });
}
