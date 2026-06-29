"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Alert, Button, DatePicker, Input, Select } from "antd";
import dayjs from "dayjs";
import { Controller, useForm, useWatch, type Control, type FieldPath, type Resolver } from "react-hook-form";
import type { ReactNode } from "react";

import { useClassList } from "@/features/classes/hooks";
import { studentStatusLabel, type StudentStatus } from "@/shared/schemas/student";
import { AddressCascade } from "@/shared/ui/AddressCascade";

import {
  createStudentFormSchema,
  editStudentFormSchema,
  type CreateStudentFormValues,
} from "./formSchema";

type Mode = "create" | "edit";
type FieldName = Exclude<FieldPath<CreateStudentFormValues>, "address">;

const statusOptions = (Object.keys(studentStatusLabel) as StudentStatus[]).map((s) => ({
  value: s,
  label: studentStatusLabel[s],
}));

function TextField(props: {
  control: Control<CreateStudentFormValues>;
  name: FieldName;
  label: string;
  error?: string;
  placeholder?: string;
  maxLength?: number;
  required?: boolean;
}): ReactNode {
  const { control, name, label, error, placeholder, maxLength, required } = props;
  return (
    <div>
      <label className="mb-1.5 block text-sm font-medium text-slate-700">
        {label}
        {required && <span className="text-red-500"> *</span>}
      </label>
      <Controller
        name={name}
        control={control}
        render={({ field }) => (
          <Input {...field} aria-label={label} placeholder={placeholder} maxLength={maxLength} status={error ? "error" : ""} />
        )}
      />
      {error && <p className="mt-1 text-sm text-red-500">{error}</p>}
    </div>
  );
}

export function StudentForm({
  mode,
  defaultValues,
  onSubmit,
  isSubmitting,
  errorMessage,
  submitLabel,
}: {
  mode: Mode;
  defaultValues: CreateStudentFormValues;
  onSubmit: (values: CreateStudentFormValues) => void;
  isSubmitting: boolean;
  errorMessage?: string;
  submitLabel: string;
}): ReactNode {
  const resolver = (
    mode === "create" ? zodResolver(createStudentFormSchema) : zodResolver(editStudentFormSchema)
  ) as Resolver<CreateStudentFormValues>;

  const {
    control,
    handleSubmit,
    setValue,
    formState: { errors },
  } = useForm<CreateStudentFormValues>({ resolver, defaultValues });

  const addr = useWatch({ control, name: "address" });
  const { data: classes, isLoading: loadingClasses } = useClassList();
  const classOptions = (classes ?? []).map((c) => ({ value: c.id, label: `${c.grade_level} ${c.room_name}` }));

  const isCreate = mode === "create";

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-5" noValidate>
      {errorMessage && <Alert type="error" showIcon message={errorMessage} role="alert" />}

      <section>
        <h2 className="mb-4 border-l-[3px] border-brand pl-2.5 text-sm font-semibold uppercase tracking-wide text-slate-500">
          ข้อมูลนักเรียน
        </h2>
        <div className="grid grid-cols-1 gap-x-4 gap-y-3 sm:grid-cols-2">
          <TextField control={control} name="studentCode" label="รหัสนักเรียน" required error={errors.studentCode?.message} />
          <TextField
            control={control}
            name="nationalId"
            label="เลขบัตรประชาชน"
            required={isCreate}
            maxLength={13}
            error={errors.nationalId?.message}
            placeholder={isCreate ? "13 หลัก" : "เว้นว่างหากไม่เปลี่ยน"}
          />
          <TextField control={control} name="prefix" label="คำนำหน้า" error={errors.prefix?.message} placeholder="เช่น ด.ช./ด.ญ./นาย/นางสาว" />
          <div className="hidden sm:block" />
          <TextField control={control} name="firstName" label="ชื่อ" required error={errors.firstName?.message} />
          <TextField control={control} name="lastName" label="นามสกุล" required error={errors.lastName?.message} />
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">วันเกิด</label>
            <Controller
              name="birthDate"
              control={control}
              render={({ field }) => (
                <DatePicker
                  className="w-full"
                  format="YYYY-MM-DD"
                  value={field.value ? dayjs(field.value) : null}
                  onChange={(d) => field.onChange(d ? d.format("YYYY-MM-DD") : "")}
                />
              )}
            />
          </div>
          <TextField control={control} name="phone" label="เบอร์โทร" maxLength={20} error={errors.phone?.message} />
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">
              สถานะ<span className="text-red-500"> *</span>
            </label>
            <Controller
              name="status"
              control={control}
              render={({ field }) => (
                <Select {...field} className="w-full" options={statusOptions} aria-label="สถานะ" />
              )}
            />
          </div>
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">ห้องเรียน (เทอมปัจจุบัน)</label>
            <Controller
              name="classId"
              control={control}
              render={({ field }) => (
                <Select
                  className="w-full"
                  showSearch
                  allowClear
                  loading={loadingClasses}
                  placeholder="ยังไม่จัดห้อง"
                  optionFilterProp="label"
                  value={field.value || undefined}
                  onChange={(v) => field.onChange(v ?? "")}
                  options={classOptions}
                  aria-label="ห้องเรียน"
                />
              )}
            />
          </div>
        </div>
      </section>

      <section>
        <h2 className="mb-4 border-l-[3px] border-brand pl-2.5 text-sm font-semibold uppercase tracking-wide text-slate-500">
          ที่อยู่
        </h2>
        <div className="grid grid-cols-1 gap-x-4 gap-y-3 sm:grid-cols-2 lg:grid-cols-3">
          <TextField control={control} name="address.houseNo" label="บ้านเลขที่" error={errors.address?.houseNo?.message} />
          <TextField control={control} name="address.moo" label="หมู่" error={errors.address?.moo?.message} />
          <TextField control={control} name="address.road" label="ถนน" error={errors.address?.road?.message} />
          <AddressCascade
            value={{
              province: addr?.province ?? "",
              district: addr?.district ?? "",
              subdistrict: addr?.subdistrict ?? "",
              postalCode: addr?.postalCode ?? "",
            }}
            onChange={(g) => {
              setValue("address.province", g.province, { shouldValidate: true });
              setValue("address.district", g.district, { shouldValidate: true });
              setValue("address.subdistrict", g.subdistrict, { shouldValidate: true });
              setValue("address.postalCode", g.postalCode, { shouldValidate: true });
            }}
          />
        </div>
      </section>

      <div className="flex justify-end border-t border-slate-100 pt-5">
        <Button type="primary" htmlType="submit" size="large" loading={isSubmitting}>
          {submitLabel}
        </Button>
      </div>
    </form>
  );
}
