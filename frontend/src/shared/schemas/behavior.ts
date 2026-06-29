import { z } from "zod";

// คะแนนความประพฤติ — ต้องตรงกับ backend (service.BehaviorRecordDTO / BehaviorSummaryDTO)
export const behaviorRecordSchema = z.object({
  id: z.string(),
  points: z.number(),
  reason: z.string(),
  occurred_at: z.string(),
  created_at: z.string(),
});
export type BehaviorRecord = z.infer<typeof behaviorRecordSchema>;

export const behaviorSummarySchema = z.object({
  starting_score: z.number(),
  current_score: z.number(),
  records: z.array(behaviorRecordSchema),
});
export type BehaviorSummary = z.infer<typeof behaviorSummarySchema>;

export type BehaviorBody = {
  points: number;
  reason: string;
  occurred_at: string;
};
