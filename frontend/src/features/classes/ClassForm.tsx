"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Alert, Button, Input } from "antd";
import { Controller, useForm } from "react-hook-form";
import type { ReactNode } from "react";
import { z } from "zod";

export const classFormSchema = z.object({
  gradeLevel: z.string().min(1, "กรุณากรอกระดับชั้น"),
  roomName: z.string().min(1, "กรุณากรอกห้อง"),
});
export type ClassFormValues = z.infer<typeof classFormSchema>;

export function ClassForm({
  defaultValues,
  onSubmit,
  isSubmitting,
  errorMessage,
  submitLabel,
}: {
  defaultValues: ClassFormValues;
  onSubmit: (values: ClassFormValues) => void;
  isSubmitting: boolean;
  errorMessage?: string;
  submitLabel: string;
}): ReactNode {
  const {
    control,
    handleSubmit,
    formState: { errors },
  } = useForm<ClassFormValues>({ resolver: zodResolver(classFormSchema), defaultValues });

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-5" noValidate>
      {errorMessage && <Alert type="error" showIcon message={errorMessage} role="alert" />}
      <div className="grid grid-cols-1 gap-x-4 gap-y-3 sm:grid-cols-2">
        <div>
          <label className="mb-1.5 block text-sm font-medium text-slate-700">
            ระดับชั้น<span className="text-red-500"> *</span>
          </label>
          <Controller
            name="gradeLevel"
            control={control}
            render={({ field }) => <Input {...field} aria-label="ระดับชั้น" placeholder="เช่น ม.1, ป.6" status={errors.gradeLevel ? "error" : ""} />}
          />
          {errors.gradeLevel && <p className="mt-1 text-sm text-red-500">{errors.gradeLevel.message}</p>}
        </div>
        <div>
          <label className="mb-1.5 block text-sm font-medium text-slate-700">
            ห้อง<span className="text-red-500"> *</span>
          </label>
          <Controller
            name="roomName"
            control={control}
            render={({ field }) => <Input {...field} aria-label="ห้อง" placeholder="เช่น 1/1, 1/2" status={errors.roomName ? "error" : ""} />}
          />
          {errors.roomName && <p className="mt-1 text-sm text-red-500">{errors.roomName.message}</p>}
        </div>
      </div>
      <div className="flex justify-end border-t border-slate-100 pt-5">
        <Button type="primary" htmlType="submit" size="large" loading={isSubmitting}>
          {submitLabel}
        </Button>
      </div>
    </form>
  );
}
