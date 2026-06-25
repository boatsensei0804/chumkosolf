# chumko-platform — Design System (MASTER)

Source of truth สำหรับ UI ทั้งระบบ ทุกหน้า/ทุก component ต้องอิงไฟล์นี้

**ที่มาของรสนิยม/ดีไซน์:** ยึด skill **`ui-ux-pro-max`** + **`frontend-design`** เป็นหลัก (แทนกฎ design เดิมใน `frontend-form-page`) — ทิศทาง: Data-Dense Dashboard
**กฎวิศวกรรม/ความปลอดภัย** (TypeScript non-any, zod, **PDPA mask**, **tenant school_id scope**, business-logic-แยก, responsive) ยังคงตาม `frontend/CLAUDE.md` + skill เดิม — non-negotiable

**โทนรวม:** สะอาด เป็นทางการ น่าเชื่อถือ สำหรับครู/ผู้บริหารโรงเรียนไทย — ฟ้า-ขาว, อ่านง่าย, ไม่หวือหวา แต่ "ตั้งใจ" (ไม่ใช่ default antd ดิบ) ใช้ความเด่นที่จุดเดียว (hero/แบรนด์) ที่เหลือเงียบ

Stack คงที่ (ห้ามเปลี่ยน): **Next.js + Ant Design + Tailwind + react-hook-form + zod**, TypeScript non-any

---

## 1. สี (Color tokens)

ทิศทางจาก **ui-ux-pro-max → Government/Public Service** (navy + professional blue, high-contrast, data-dense) กำหนดเป็น token กลางที่ `tailwind.config.ts` (`brand.*`) และ `src/lib/theme.ts` (antd) — **ห้าม hardcode hex ใน component** (ค่า brand.* ตรงกับ Tailwind sky/slate เป๊ะ เพื่อความเข้ากัน)

| Token | HEX | = Tailwind | ใช้กับ |
|-------|-----|-----------|--------|
| `brand.DEFAULT` | `#0369A1` | sky-700 | primary: ปุ่ม, link, active, ไอคอนเน้น |
| `brand.deep` | `#075985` | sky-800 | hover/เข้ม, ปลาย gradient |
| `brand.bright` | `#0EA5E9` | sky-500 | highlight |
| `brand.navy` | `#0F172A` | slate-900 | **dark sidebar / พื้นผิวเข้ม** |
| `brand.slate` | `#334155` | slate-700 | secondary |
| white | `#FFFFFF` | — | พื้นการ์ด/พื้นผิว |

**Neutral:** พื้นแอป `slate-50 (#f8fafc)` · foreground `slate-900 (#0f172a)` · รอง `slate-500` · จาง `slate-400` · เส้น `slate-200`/`slate-100` (border รอง `#e2e8f0`)

**Accent color-coding ต่อหมวด** (`src/shared/ui/accent.ts` — class คงที่กัน purge): blue/emerald/violet/amber/rose/slate · map หมวด: บุคลากร=blue, นักเรียน=emerald, เช็คชื่อ=violet, งบประมาณ=amber, ตั้งค่า=slate · ใช้กับ KPI strip, ไอคอนเมนู, หัวข้อ SectionCard ให้แต่ละหมวดมีสีประจำ (เด่น/จำง่าย)

**Semantic / status:** success `#059669` · error `#DC2626` · warning amber — ใช้กับ badge/แถวเพื่อ scan ง่าย

**กฎคอนทราสต์:** เนื้อหาใช้ `slate-800/900`, `brand` เป็น action/accent, **ตัวอักษรบนพื้น navy/gradient ใช้ขาว/slate-300** (sidebar), ทุกคู่ ≥ 4.5:1
> ⚠️ แก้ `tailwind.config.ts` (custom color) ต้อง **restart dev server** — HMR ไม่ reload config ทำให้ custom `brand-*` ไม่ generate (พื้น navy หาย ตัวขาวบนขาว)

---

## 2. ตัวอักษร (Typography)

- ฟอนต์เนื้อหา/หัวข้อ: **Sarabun** (`--font-sans`, โหลดผ่าน `next/font`) — รองรับไทย เหมาะครู/ผู้บริหาร
- ฟอนต์ตัวเลข/โค้ด: **Fira Code** (`--font-mono`) — ใช้กับเลขบัตร/เบอร์/วันที่/KPI ผ่าน class `.num` (mono + tabular-nums) ให้ดู technical และกันการขยับ
- Scale: `text-xs 12 · sm 14 · base 16 · lg 18 · xl 20 · 2xl 24 · 3xl 30`
- น้ำหนัก: heading **700** (`font-bold`), หัวข้อย่อย/ปุ่ม/label **500-600**, เนื้อหา **400**
- หัวข้อ section ในฟอร์ม: เล็ก + `uppercase tracking-wide text-slate-500` + แถบ accent ฟ้าซ้าย (ดู §6)

---

## 3. ระยะห่าง / มุม / เงา / เส้น

- **Spacing:** ยึดจังหวะ 4/8px (`gap-2/3/4/5/6`, `p-4/5/6`) — section spacing `gap-5/6`
- **Radius:** ปุ่ม/อินพุต `rounded-lg`(antd 10) · การ์ด `rounded-xl`/`rounded-2xl` · pill `rounded-full` · โลโก้ `rounded-xl`
- **Shadow:** เบามาก — การ์ดใช้ `border border-slate-200` เป็นหลัก + เงาจาง (`shadow-sm`/hover `shadow-md`) ไม่ใช้เงาหนา
- **Border:** เส้นคั่น `border-slate-100`, ขอบการ์ด `border-slate-200`
- **Density (data-dense vars):** card-padding ~12–16px, table row ~36px (`size="middle"/"small"`), gap 8px, header 56px, sidebar 240px, font 12–14px
- **Container:** หน้า list/dashboard ใช้ **เต็มความกว้าง** · ฟอร์มเดี่ยว `max-w-4xl` · หน้าแก้ไขใช้ 2 คอลัมน์ (form + rail) บน `xl`

---

## 4. Layout หลัก

- **Dashboard shell** (`features/navigation/DashboardLayout`): **sidebar navy เข้ม `brand.navy` 240px** (เดสก์ท็อป) / Drawer (มือถือ) + topbar ขาว 56px (h-14) + content `bg-slate-50 p-4`
- **Auth (login):** split-panel — แผงซ้าย gradient navy→blue (`.bg-brand-gradient`) / แผงขวา ฟอร์ม; มือถือซ่อนแผงซ้าย
- Responsive: ทำงานได้ ~375px → desktop, ใช้ breakpoint `sm md lg xl` สม่ำเสมอ, ไม่มี horizontal scroll

---

## 5. Signature elements (เอกลักษณ์)

1. **Brand mark:** โลโก้สี่เหลี่ยมมน `bg-brand` + ไอคอน `ReadOutlined` ขาว + wordmark "ชุมโค / ระบบบริหารโรงเรียน" (ขาวบน navy)
2. **Dark navy sidebar:** พื้น `brand.navy` + ไอคอนเมนูสีตามหมวด (sky/emerald/violet/amber-400) — active = บล็อกสีทึบ `bg-brand text-white`
3. **KPI cards:** แถบสี accent ด้านบน (`ACCENTS[x].strip`) + ไอคอน chip สี + ตัวเลขใหญ่ `.num` สี navy
4. **Login gradient:** แผง `.bg-brand-gradient` (navy→blue) + วงตกแต่ง `bg-white/10`, `bg-brand-bright/20`

---

## 6. แพทเทิร์น Component

**Page header** (`shared/ui/PageHeader`): ลิงก์ย้อนกลับ (← ...) + ไอคอนกรอบ `bg-brand/10 text-brand` + title `text-xl font-bold` + subtitle `text-slate-500` + actions ฝั่งขวา

**Section card** (`shared/ui/SectionCard`): `rounded-2xl border-slate-200 bg-white`; header `border-b border-slate-100 px-5 py-4` (ไอคอน + title + description + extra); body `p-5`

**Section header ในฟอร์ม:** `border-l-[3px] border-brand pl-2.5 text-sm font-semibold uppercase tracking-wide text-slate-500`

**Form field:** label `text-sm font-medium text-slate-700` (มองเห็นเสมอ ไม่ placeholder-only) → antd input ผ่าน `Controller` → error ไทยใต้ field `text-sm text-red-500`; required มี `*`; input `size="large"` ในฟอร์มหลัก/หน้า login

**Badge/Tag:** `Tag bordered={false}` — role: teacher=blue, executive=purple; สถานะ: ใช้งาน=success, ระงับ=default; "ปัจจุบัน"/"หัวหน้ากลุ่ม"=green/gold; pill "เร็ว ๆ นี้" = `bg-slate-100 text-slate-400 text-[10px] rounded-full`

**Row list** (sub-resource): `divide-y divide-slate-100`, แต่ละแถว `py-3 flex items-center justify-between`, ปุ่มลบเป็น icon button (`type="text" danger`) + `Popconfirm`

**Avatar:** วงกลม gradient `from-brand to-brand-deep` + ตัวอักษรแรกขาว (topbar) หรือ `bg-brand/10 text-brand` initial (ตาราง)

**Empty state:** ข้อความบอกทางออก + ปุ่มถ้าทำได้ (ไม่ปล่อยว่าง) เช่น "ยังไม่มีข้อมูล — กด …"

**Table:** ครอบในกรอบการ์ด (`rounded-xl border`), header `bg-slate-50`, `scroll={{x}}` เมื่อคอลัมน์เยอะ, action เป็น icon button ชิดขวา

**ปุ่ม:** primary = action หลัก 1 ปุ่ม/บริบท, ข้อความ active voice ("เพิ่มบุคลากร", "บันทึกการแก้ไข", "มอบหมาย"); ปุ่มกับ toast สอดคล้องกัน; มี loading state

---

## 7. UX / Accessibility / PDPA

- **Active voice ไทย** ทุกปุ่ม/ข้อความ; error บอกสาเหตุ+ทางแก้; empty/loading/error ครบทุก action
- feedback: loading (spinner/skeleton), success (toast), error (alert/toast ไทยจาก backend)
- a11y: label ผูก input, icon-only button มี `aria-label`, focus ring คงไว้, contrast ผ่าน, ไม่สื่อด้วยสีอย่างเดียว
- micro-interaction: transition 150–300ms (`transition-colors/all`), hover ยกการ์ด `-translate-y-0.5`
- **PDPA:** เลขบัตรประชาชนแสดง **mask** เสมอ (`1-2345-xxxxx-xx-1`), ตอนแก้ไขไม่ดึงเลขเต็ม, ไม่ log ข้อมูลส่วนบุคคล

---

## 8. Anti-patterns (อย่าทำ)

- ❌ ใช้ emoji เป็นไอคอน (ใช้ `@ant-design/icons`) · ❌ hardcode hex · ❌ เงาหนา/หลายระดับมั่ว
- ❌ ฟ้าทั้งหน้าจนล้า (ฟ้าเป็น accent, ขาว/slate เป็นพื้น) · ❌ ตัวอักษรเทาบนเทา
- ❌ placeholder แทน label · ❌ error รวมบนสุดอย่างเดียว (วางใต้ field) · ❌ หน้า empty ปล่อยว่าง
- ❌ เปลี่ยน stack เป็น shadcn/อื่น · ❌ business logic ใน component (อยู่ใน hook/service)

---

## อ้างอิงการใช้งานจริง

- tokens: `tailwind.config.ts`, `src/lib/theme.ts`, `src/app/globals.css`
- shell/nav: `src/features/navigation/DashboardLayout.tsx`
- shared ui: `src/shared/ui/`
- หน้าอ้างอิง: `src/app/login/page.tsx`, `src/app/(dashboard)/page.tsx`, `src/app/(dashboard)/personnel/*`
