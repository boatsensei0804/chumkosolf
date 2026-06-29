"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Alert, Button, DatePicker, Input } from "antd";
import dayjs from "dayjs";
import { Controller, useForm, useWatch, type Control, type FieldPath } from "react-hook-form";
import type { ReactNode } from "react";

import { AddressCascade } from "@/shared/ui/AddressCascade";

import { myProfileFormSchema, type MyProfileFormValues } from "./formSchema";

type MyProfileFormProps = {
  defaultValues: MyProfileFormValues;
  onSubmit: (values: MyProfileFormValues) => void;
  isSubmitting: boolean;
  errorMessage?: string;
};

// เฉพาะ field ที่เป็น string (ตัด "address" ซึ่งเป็น object)
type FieldName = Exclude<FieldPath<MyProfileFormValues>, "address">;

function TextField(props: {
  control: Control<MyProfileFormValues>;
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
          <Input
            {...field}
            aria-label={label}
            placeholder={placeholder}
            maxLength={maxLength}
            status={error ? "error" : ""}
          />
        )}
      />
      {error && <p className="mt-1 text-sm text-red-500">{error}</p>}
    </div>
  );
}

export function MyProfileForm({
  defaultValues,
  onSubmit,
  isSubmitting,
  errorMessage,
}: MyProfileFormProps): ReactNode {
  const {
    control,
    handleSubmit,
    setValue,
    formState: { errors },
  } = useForm<MyProfileFormValues>({ resolver: zodResolver(myProfileFormSchema), defaultValues });

  const addr = useWatch({ control, name: "address" });

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-5" noValidate>
      {errorMessage && <Alert type="error" showIcon message={errorMessage} role="alert" />}

      <section>
        <h2 className="mb-4 border-l-[3px] border-brand pl-2.5 text-sm font-semibold uppercase tracking-wide text-slate-500">
          ข้อมูลส่วนตัว
        </h2>
        <div className="grid grid-cols-1 gap-x-4 gap-y-3 sm:grid-cols-2">
          <TextField control={control} name="prefix" label="คำนำหน้า" error={errors.prefix?.message} placeholder="เช่น นาย/นาง/นางสาว" />
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
          <TextField control={control} name="email" label="อีเมล" error={errors.email?.message} />
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
          บันทึกข้อมูลของฉัน
        </Button>
      </div>
    </form>
  );
}
