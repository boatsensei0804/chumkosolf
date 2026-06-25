"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Alert, Button, Input } from "antd";
import { Controller, useForm } from "react-hook-form";
import type { ReactNode } from "react";

import { loginRequestSchema, type LoginRequest } from "@/shared/schemas/auth";

type LoginFormProps = {
  onSubmit: (values: LoginRequest) => void;
  isSubmitting: boolean;
  // ข้อความ error จาก backend (ภาษาไทย) แสดงเหนือฟอร์ม
  errorMessage?: string;
};

// LoginForm เป็น presentational form — logic การเรียก API/redirect อยู่ที่ page (useLogin)
export function LoginForm({
  onSubmit,
  isSubmitting,
  errorMessage,
}: LoginFormProps): ReactNode {
  const {
    control,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginRequest>({
    resolver: zodResolver(loginRequestSchema),
    defaultValues: { schoolCode: "CHUMKO", username: "", password: "" },
  });

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-4" noValidate>
      {errorMessage && (
        <Alert type="error" showIcon message={errorMessage} role="alert" />
      )}

      <div>
        <label htmlFor="schoolCode" className="mb-1.5 block text-sm font-medium text-slate-700">
          รหัสโรงเรียน
        </label>
        <Controller
          name="schoolCode"
          control={control}
          render={({ field }) => (
            <Input
              {...field}
              id="schoolCode"
              size="large"
              placeholder="เช่น CHUMKO"
              autoComplete="organization"
              status={errors.schoolCode ? "error" : ""}
            />
          )}
        />
        {errors.schoolCode && (
          <p className="mt-1 text-sm text-red-500">{errors.schoolCode.message}</p>
        )}
      </div>

      <div>
        <label htmlFor="username" className="mb-1.5 block text-sm font-medium text-slate-700">
          ชื่อผู้ใช้
        </label>
        <Controller
          name="username"
          control={control}
          render={({ field }) => (
            <Input
              {...field}
              id="username"
              size="large"
              placeholder="ชื่อผู้ใช้"
              autoComplete="username"
              status={errors.username ? "error" : ""}
            />
          )}
        />
        {errors.username && (
          <p className="mt-1 text-sm text-red-500">{errors.username.message}</p>
        )}
      </div>

      <div>
        <label htmlFor="password" className="mb-1.5 block text-sm font-medium text-slate-700">
          รหัสผ่าน
        </label>
        <Controller
          name="password"
          control={control}
          render={({ field }) => (
            <Input.Password
              {...field}
              id="password"
              size="large"
              placeholder="รหัสผ่าน"
              autoComplete="current-password"
              status={errors.password ? "error" : ""}
            />
          )}
        />
        {errors.password && (
          <p className="mt-1 text-sm text-red-500">{errors.password.message}</p>
        )}
      </div>

      <Button
        type="primary"
        htmlType="submit"
        size="large"
        loading={isSubmitting}
        block
        className="mt-2"
      >
        เข้าสู่ระบบ
      </Button>
    </form>
  );
}
