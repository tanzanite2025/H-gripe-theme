<template>
  <AdminFormSection title="税务与清关资料" :description="description">
    <div class="mb-4 grid gap-3 rounded-lg border bg-muted/20 p-3 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-end">
      <AdminFormField label="清关资料模板" description="选择后会填入 HS Code、CN Code、原产国代码和英文报关品名，仍可继续手动覆盖。">
        <Select :model-value="customsClassificationSelectValue" @update:model-value="emit('customs-classification-select', $event)">
          <SelectTrigger class="w-full"><SelectValue placeholder="请选择清关资料模板" /></SelectTrigger>
          <SelectContent>
            <SelectItem value="__none__">不套用模板</SelectItem>
            <SelectItem v-for="profile in customsClassifications" :key="profile.id" :value="String(profile.id)">
              {{ profile.name }} · {{ profile.hs_code }}{{ profile.cn_code ? ` / ${profile.cn_code}` : '' }}
            </SelectItem>
          </SelectContent>
        </Select>
      </AdminFormField>
      <Button type="button" variant="outline" size="sm" as-child>
        <RouterLink to="/catalog/customs-classifications">
          <Tags class="size-3.5" />
          清关资料中心
        </RouterLink>
      </Button>
      <div class="flex flex-wrap gap-2 text-[11px] lg:col-span-2">
        <span :class="form.hs_code ? 'bg-emerald-500/10 text-emerald-700' : 'bg-amber-500/10 text-amber-700'" class="rounded-full px-2 py-0.5 font-medium">HS</span>
        <span :class="form.cn_code ? 'bg-emerald-500/10 text-emerald-700' : 'bg-amber-500/10 text-amber-700'" class="rounded-full px-2 py-0.5 font-medium">CN</span>
        <span :class="form.country_of_origin ? 'bg-emerald-500/10 text-emerald-700' : 'bg-amber-500/10 text-amber-700'" class="rounded-full px-2 py-0.5 font-medium">原产国</span>
        <span :class="form.customs_description ? 'bg-emerald-500/10 text-emerald-700' : 'bg-amber-500/10 text-amber-700'" class="rounded-full px-2 py-0.5 font-medium">英文品名</span>
      </div>
    </div>
    <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      <AdminFormField label="HS Code" description="6 位数字" :error="errors.hs_code">
        <Input v-model="form.hs_code" inputmode="numeric" maxlength="6" placeholder="例如 871499" @input="handleProductCustomsFieldManualEdit('hs_code')" />
      </AdminFormField>
      <AdminFormField label="CN Code" description="欧盟 8 位编码，可选" :error="errors.cn_code">
        <Input v-model="form.cn_code" inputmode="numeric" maxlength="8" placeholder="例如 87149990" @input="handleProductCustomsFieldManualEdit('cn_code')" />
      </AdminFormField>
      <AdminFormField label="原产国代码" description="2 位国家代码" :error="errors.country_of_origin">
        <Input v-model="form.country_of_origin" class="font-mono uppercase" maxlength="2" placeholder="例如 CN" @input="handleProductCustomsFieldManualEdit('country_of_origin')" />
      </AdminFormField>
      <AdminFormField label="英文报关品名" :error="errors.customs_description">
        <Input v-model="form.customs_description" maxlength="255" :placeholder="customsDescriptionPlaceholder" @input="handleProductCustomsFieldManualEdit('customs_description')" />
      </AdminFormField>
    </div>
  </AdminFormSection>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import { Tags } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminFormSection from '@/components/admin/AdminFormSection.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

interface CustomsClassificationRecord {
  id: number | string
  name: string
  hs_code: string
  cn_code?: string
  product_specification_template_id?: number | null
}

const props = withDefaults(defineProps<{
  form: Record<string, any>
  errors?: Record<string, string>
  customsClassifications?: CustomsClassificationRecord[]
  description?: string
  customsDescriptionPlaceholder?: string
}>(), {
  errors: () => ({}),
  customsClassifications: () => [],
  description: '维护商品的基础清关属性；申报价值不在产品上固定，后续按订单确认。',
  customsDescriptionPlaceholder: '例如 Bicycle frame',
})

const emit = defineEmits<{
  (event: 'customs-classification-select', value: unknown): void
  (event: 'customs-classification-manual-edit', field: string): void
  (event: 'clear-error', field: string): void
}>()

const customsClassificationSelectValue = computed(() => (
  props.form.customs_classification_profile_id ? String(props.form.customs_classification_profile_id) : '__none__'
))

const handleProductCustomsFieldManualEdit = (field: string): void => {
  emit('customs-classification-manual-edit', field)
  emit('clear-error', field)
}
</script>
