<template>
  <fieldset :disabled="locked" class="grid gap-3 rounded-lg bg-muted/20 p-3 md:grid-cols-3">
    <AdminFormField label="孔数" required><Input v-model="specification.spoke_hole_count" :disabled="persisted" type="number" min="1" max="128" step="1" /></AdminFormField>
    <AdminFormField label="实际钻孔形式" required>
      <select v-model="specification.drilling_pattern" :disabled="persisted" class="h-9 w-full rounded-md border bg-background px-3 text-sm">
        <option value="" disabled>请选择</option><option value="uniform">均匀分布</option><option value="paired">成对孔</option><option value="2-to-1">2:1 孔位</option><option value="g3">G3 孔位</option>
      </select>
    </AdminFormField>
    <AdminFormField label="涂装/表面标识" required description="例如 matte-black；保存后改变组合需新增 SKU。"><Input v-model="specification.finish_code" :disabled="persisted" maxlength="80" placeholder="matte-black" /></AdminFormField>
    <AdminFormField label="单只净重（g）" required><Input v-model="specification.net_weight_grams" type="number" min="0.01" step="any" /></AdminFormField>
    <AdminFormField label="ERD（mm，可选）"><Input v-model="specification.erd_mm" type="number" min="0.01" step="any" /></AdminFormField>
    <AdminFormField label="ERD 测量约定" description="填写 ERD 时必填；注明条帽、垫片和测量基准。"><Input v-model="specification.erd_measurement_notes" maxlength="1000" /></AdminFormField>
  </fieldset>
</template>

<script setup lang="ts">
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Input } from '@/components/ui/input'
import type { CarbonRimVariantSpecificationForm } from '@/modules/product/carbonRimDedicatedSpecificationFormValuesAndValidation'
defineProps<{ specification: CarbonRimVariantSpecificationForm; persisted?: boolean; locked?: boolean }>()
</script>
