<template>
  <AdminFormSection id="carbon-rim-specifications" title="碳纤维车圈规格" description="单只碟刹车圈。一个商品维护同一型号、同一截面；不同孔数和涂装在销售 SKU 中维护。">
    <fieldset :disabled="locked" class="grid gap-4 md:grid-cols-3">
      <AdminFormField label="车圈型号" required><Input v-model="specification.model" maxlength="120" /></AdminFormField>
      <AdminFormField label="标称轮径" required description="700C 与 29 英寸都使用 622 mm BSD，请按实际销售规格选择。">
        <select v-model="specification.wheel_size" aria-label="标称轮径" class="h-9 w-full rounded-md border bg-background px-3 text-sm">
          <option value="" disabled>请选择轮径</option>
          <option v-for="choice in carbonRimWheelSizeChoices" :key="choice.value" :value="choice.value">{{ choice.label }} · BSD {{ choice.beadSeatDiameterMM }} mm</option>
        </select>
      </AdminFormField>
      <AdminFormField label="材质" required description="此入口用于碳纤维车圈；具体材料信息在下方补充。">
        <select v-model="specification.material" aria-label="材质" class="h-9 w-full rounded-md border bg-background px-3 text-sm"><option value="carbon_fiber">碳纤维</option></select>
      </AdminFormField>
      <AdminFormField label="材质牌号 / 铺层说明" class="md:col-span-3" description="可选，按制造商资料填写，例如 T700 / T800；未确认时留空。"><Input v-model="specification.material_details" aria-label="材质牌号 / 铺层说明" maxlength="500" placeholder="填写已确认的纤维牌号或材料构成" /></AdminFormField>
      <AdminFormField label="适用骑行场景" required class="md:col-span-3" description="可多选，按制造商公布的用途选择；不会根据轮径自动勾选。">
        <div role="group" aria-label="适用骑行场景" class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          <label v-for="choice in carbonRimRidingUseChoices" :key="choice.value" class="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
            <input v-model="specification.riding_uses" type="checkbox" :value="choice.value" class="size-4 accent-primary" />{{ choice.label }}
          </label>
        </div>
      </AdminFormField>
      <AdminFormField v-for="field in geometryFields" :key="field.key" :label="field.label" required>
        <Input v-model="specification[field.key]" type="number" min="0.01" step="any" />
      </AdminFormField>
      <AdminFormField label="胎圈结构" required>
        <select v-model="specification.bead_structure" class="h-9 w-full rounded-md border bg-background px-3 text-sm">
          <option value="" disabled>请选择</option><option value="hooked">有钩</option><option value="hookless">无钩</option>
        </select>
      </AdminFormField>
      <AdminFormField label="支持真空胎" required>
        <select v-model="specification.tubeless_ready" class="h-9 w-full rounded-md border bg-background px-3 text-sm">
          <option :value="null" disabled>请明确选择</option><option :value="true">是</option><option :value="false">否</option>
        </select>
      </AdminFormField>
      <AdminFormField label="需要胎垫">
        <select v-model="specification.rim_tape_required" class="h-9 w-full rounded-md border bg-background px-3 text-sm">
          <option :value="null">未确认</option><option :value="true">是</option><option :value="false">否</option>
        </select>
      </AdminFormField>
      <AdminFormField label="气嘴孔直径（mm）"><Input v-model="specification.valve_hole_diameter_mm" type="number" min="0.01" step="any" /></AdminFormField>
      <AdminFormField label="偏心量绝对值（mm）" description="对称车圈填 0；偏心方向在技术说明中注明。"><Input v-model="specification.asymmetric_offset_mm" type="number" min="0" step="any" /></AdminFormField>
      <AdminFormField label="制造商技术说明" class="md:col-span-3" description="仅填写有依据的胎压、负载、辐条张力限制及适用条件；没有资料时留空。此处不会自动进行适配认证。">
        <Textarea v-model="specification.technical_notes" maxlength="4000" />
      </AdminFormField>
    </fieldset>
    <p v-if="locked" class="mt-3 text-sm text-muted-foreground">工程参数由原语言商品维护；此处只读。</p>
    <p v-if="error" role="alert" class="mt-3 text-sm text-destructive">{{ error }}</p>
  </AdminFormSection>
</template>

<script setup lang="ts">
import AdminFormSection from '@/components/admin/AdminFormSection.vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import type { CarbonRimSpecificationForm } from '@/modules/product/carbonRimDedicatedSpecificationFormValuesAndValidation'
import { carbonRimWheelSizeChoices, carbonRimRidingUseChoices } from '@/modules/product/carbonRimWheelSizeAndRidingUseFormChoices'
defineProps<{ specification: CarbonRimSpecificationForm; locked?: boolean; error?: string }>()
const geometryFields = [
  { key: 'bead_seat_diameter_mm', label: '胎圈座直径 BSD（mm）' },
  { key: 'depth_mm', label: '框高（mm）' },
  { key: 'internal_width_mm', label: '内宽（mm）' },
  { key: 'external_width_mm', label: '外宽（mm）' },
] as const
</script>
