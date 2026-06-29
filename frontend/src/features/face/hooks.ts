"use client";

import { useMutation, type UseMutationResult } from "@tanstack/react-query";

import type { ApiRequestError } from "@/lib/api/client";
import type { RecognizeResult, ReindexResult } from "@/shared/schemas/face";

import { recognizeFace, reindexFace } from "./api";

export function useReindexFace(): UseMutationResult<ReindexResult, ApiRequestError, void> {
  return useMutation({ mutationFn: () => reindexFace() });
}

export function useRecognizeFace(): UseMutationResult<RecognizeResult, ApiRequestError, Blob[]> {
  return useMutation({ mutationFn: (frames: Blob[]) => recognizeFace(frames) });
}
