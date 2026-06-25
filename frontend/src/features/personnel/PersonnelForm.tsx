"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Alert, Button, DatePicker, Input, Select } from "antd";
import dayjs from "dayjs";
import { Controller, useForm, type Control, type FieldPath, type Resolver } from "react-hook-form";
import type { ReactNode } from "react";

import { personnelRoleLabel } from "@/shared/schemas/personnel";

import {
  createPersonnelFormSchema,
  editPersonnelFormSchema,
  type CreatePersonnelFormValues,
} from "./formSchema";

type Mode = "create" | "edit";

type PersonnelFormProps = {
  mode: Mode;
  defaultValues: CreatePersonnelFormValues;
  onSubmit: (values: CreatePersonnelFormValues) => void;
  isSubmitting: boolean;
  errorMessage?: string;
  submitLabel: string;
};

// เฉพาะ field ที่เป็น string (ตัด path "address" ซึ่งเป็น object ออก) เพื่อให้ value เป็น string เสมอ
type FieldName = Exclude<FieldPath<CreatePersonnelFormValues>, "address">;

// TextField: antd Input ห่อด้วย Controller + error ไทยใต้ field (ฟอร์มนี้ใช้ type เดียวคงที่)
function TextField(props: {
  control: Control<CreatePersonnelFormValues>;
  name: FieldName;
  label: string;
  error?: string;
  placeholder?: string;
  password?: boolean;
  maxLength?: number;
  required?: boolean;
}): ReactNode {
  const { control, name, label, error, placeholder, password, maxLength, required } = props;
  return (
    <div>
      <label className="mb-1.5 block text-sm font-medium text-slate-700">
        {label}
        {required && <span className="text-red-500"> *</span>}
      </label>
      <Controller
        name={name}
        control={control}
        render={({ field }) =>
          password ? (
            <Input.Password
              {...field}
              aria-label={label}
              placeholder={placeholder}
              maxLength={maxLength}
              status={error ? "error" : ""}
            />
          ) : (
            <Input
              {...field}
              aria-label={label}
              placeholder={placeholder}
              maxLength={maxLength}
              status={error ? "error" : ""}
            />
          )
        }
      />
      {error && <p className="mt-1 text-sm text-red-500">{error}</p>}
    </div>
  );
}

export function PersonnelForm({
  mode,
  defaultValues,
  onSubmit,
  isSubmitting,
  errorMessage,
  submitLabel,
}: PersonnelFormProps): ReactNode {
  // เลือก schema ตามโหมด — assertion เพราะ edit schema เป็น subset ของ create (account fields ผ่านได้)
  const resolver = (
    mode === "create"
      ? zodResolver(createPersonnelFormSchema)
      : zodResolver(editPersonnelFormSchema)
  ) as Resolver<CreatePersonnelFormValues>;

  const {
    control,
    handleSubmit,
    formState: { errors },
  } = useForm<CreatePersonnelFormValues>({ resolver, defaultValues });

  const isCreate = mode === "create";

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-6" noValidate>
      {errorMessage && <Alert type="error" showIcon message={errorMessage} role="alert" />}

      {isCreate && (
        <section>
          <h2 className="mb-4 border-l-[3px] border-brand pl-2.5 text-sm font-semibold uppercase tracking-wide text-slate-500">บัญชีผู้ใช้</h2>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <TextField
              control={control}
              name="username"
              label="ชื่อผู้ใช้"
              required
              error={errors.username?.message}
              placeholder="ชื่อสำหรับเข้าสู่ระบบ"
            />
            <TextField
              control={control}
              name="password"
              label="รหัสผ่าน"
              required
              password
              error={errors.password?.message}
              placeholder="อย่างน้อย 8 ตัวอักษร"
            />
            <div>
              <label className="mb-1 block text-sm text-gray-700">
                ตำแหน่ง<span className="text-red-500"> *</span>
              </label>
              <Controller
                name="role"
                control={control}
                render={({ field }) => (
                  <Select
                    {...field}
                    className="w-full"
                    status={errors.role ? "error" : ""}
                    options={[
                      { value: "teacher", label: personnelRoleLabel.teacher },
                      { value: "executive", label: personnelRoleLabel.executive },
                    ]}
                  />
                )}
              />
              {errors.role && <p className="mt-1 text-sm text-red-500">{errors.role.message}</p>}
            </div>
          </div>
        </section>
      )}

      <section>
        <h2 className="mb-4 border-l-[3px] border-brand pl-2.5 text-sm font-semibold uppercase tracking-wide text-slate-500">ข้อมูลส่วนตัว</h2>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <TextField control={control} name="prefix" label="คำนำหน้า" error={errors.prefix?.message} placeholder="เช่น นาย/นาง/นางสาว" />
          <div className="hidden sm:block" />
          <TextField control={control} name="firstName" label="ชื่อ" required error={errors.firstName?.message} />
          <TextField control={control} name="lastName" label="นามสกุล" required error={errors.lastName?.message} />
          <TextField
            control={control}
            name="nationalId"
            label="เลขบัตรประชาชน"
            required={isCreate}
            maxLength={13}
            error={errors.nationalId?.message}
            placeholder={isCreate ? "13 หลัก" : "เว้นว่างหากไม่เปลี่ยน"}
          />
          <TextField control={control} name="civilServantId" label="เลขบัตรประจำตัวราชการ" error={errors.civilServantId?.message} placeholder={isCreate ? "" : "เว้นว่างหากไม่เปลี่ยน"} />
          <div>
            <label className="mb-1 block text-sm text-gray-700">วันเกิด</label>
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
        <h2 className="mb-4 border-l-[3px] border-brand pl-2.5 text-sm font-semibold uppercase tracking-wide text-slate-500">ที่อยู่</h2>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <TextField control={control} name="address.houseNo" label="บ้านเลขที่" error={errors.address?.houseNo?.message} />
          <TextField control={control} name="address.moo" label="หมู่" error={errors.address?.moo?.message} />
          <TextField control={control} name="address.road" label="ถนน" error={errors.address?.road?.message} />
          <TextField control={control} name="address.subdistrict" label="ตำบล/แขวง" error={errors.address?.subdistrict?.message} />
          <TextField control={control} name="address.district" label="อำเภอ/เขต" error={errors.address?.district?.message} />
          <TextField control={control} name="address.province" label="จังหวัด" error={errors.address?.province?.message} />
          <TextField control={control} name="address.postalCode" label="รหัสไปรษณีย์" maxLength={10} error={errors.address?.postalCode?.message} />
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
