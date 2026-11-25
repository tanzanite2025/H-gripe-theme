<template>
  <AdminFormSection id="bicycle-tire-specifications" title="自行车外胎规格" description="制造商没有提供的规格留空。">
    <div class="mb-4 rounded-lg border border-primary/20 bg-primary/5 p-3">
      <AdminFormField label="Schwalbe 型号 / Article No.">
        <Select
          :model-value="selectedArticleNo || '__none__'"
          :disabled="locked || catalogLoading || catalogItems.length === 0"
          @update:model-value="emit('catalog-model-select', String($event))"
        >
          <SelectTrigger class="w-full"><SelectValue :placeholder="catalogLoading ? '正在读取型号目录…' : '选择目录型号或手工填写'" /></SelectTrigger>
          <SelectContent>
            <SelectItem value="__none__">不使用目录</SelectItem>
            <SelectItem v-for="item in catalogItems" :key="item.article_no" :value="item.article_no">
              {{ item.model_name }} · {{ item.etrto }} · {{ item.article_no }}{{ item.product_exists ? '（已有商品）' : '' }}
            </SelectItem>
          </SelectContent>
        </Select>
      </AdminFormField>
      <p v-if="catalogLoading" class="mt-2 text-xs text-muted-foreground">正在读取 Schwalbe 型号目录。</p>
      <p v-else-if="catalogError" class="mt-2 flex items-center gap-2 text-xs text-destructive">
        {{ catalogError }}
        <Button type="button" variant="link" size="sm" class="h-auto p-0" @click="emit('retry-catalog')">重试</Button>
      </p>
      <p v-else-if="catalogItems.length === 0" class="mt-2 text-xs text-muted-foreground">目录不可用时可以直接手工填写外胎规格。</p>
      <p v-else class="mt-2 text-xs text-muted-foreground">选择目录型号会填入外胎规格和 EAN；各项仍可修改，缺失字段保持空白。</p>
    </div>
    <fieldset :disabled="locked" class="grid gap-4 md:grid-cols-3">
      <AdminFormField label="制造商型号"><Input v-model="specification.model_name" maxlength="180" /></AdminFormField>
      <AdminFormField label="Article No."><Input v-model="specification.article_no" maxlength="70" /></AdminFormField>
      <AdminFormField label="ETRTO"><Input v-model="specification.etrto" maxlength="32" placeholder="例如 40-622" /></AdminFormField>
      <AdminFormField label="英寸规格"><Input v-model="specification.inch_designation" maxlength="32" /></AdminFormField>
      <AdminFormField label="外胎净重（g）"><Input v-model="specification.weight_g" type="number" min="0.01" step="any" /></AdminFormField>
      <AdminFormField label="版本 / 胎体"><Input v-model="specification.version_label" maxlength="120" /></AdminFormField>
      <AdminFormField label="胶料"><Input v-model="specification.compound" maxlength="120" /></AdminFormField>
      <AdminFormField label="颜色"><Input v-model="specification.color" maxlength="120" /></AdminFormField>
      <AdminFormField label="胎圈"><Input v-model="specification.bead" maxlength="64" placeholder="例如 WIRED、Folding" /></AdminFormField>
      <AdminFormField label="E-bike 评级"><Input v-model="specification.e_bike_rating" maxlength="64" /></AdminFormField>
      <AdminFormField label="EPI"><Input v-model="specification.epi" type="number" min="1" step="1" /></AdminFormField>
      <AdminFormField label="载重（kg）"><Input v-model="specification.load_kg" type="number" min="0.01" step="any" /></AdminFormField>
      <AdminFormField label="密封类型"><Input v-model="specification.seal" maxlength="120" placeholder="例如 Tube、TLE" /></AdminFormField>
      <AdminFormField label="胎面 / 花纹"><Input v-model="specification.tread" maxlength="120" /></AdminFormField>
      <AdminFormField label="最小胎压（bar）"><Input v-model="specification.min_pressure_bar" type="number" min="0.01" step="any" /></AdminFormField>
      <AdminFormField label="最大胎压（bar）"><Input v-model="specification.max_pressure_bar" type="number" min="0.01" step="any" /></AdminFormField>
      <AdminFormField label="最小胎压（psi）"><Input v-model="specification.min_pressure_psi" type="number" min="0.01" step="any" /></AdminFormField>
      <AdminFormField label="最大胎压（psi）"><Input v-model="specification.max_pressure_psi" type="number" min="0.01" step="any" /></AdminFormField>
    </fieldset>
    <p v-if="locked" class="mt-3 text-sm text-muted-foreground">规格由源语言商品维护。</p>
    <p v-if="error" role="alert" class="mt-3 text-sm text-destructive">{{ error }}</p>
  </AdminFormSection>
</template>

<script setup lang="ts">
import AdminFormSection from '@/components/admin/AdminFormSection.vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import type { SchwalbeTireCatalogItem } from '@/api/schwalbeTireCatalog'
import type { BicycleTireSpecificationForm } from '@/modules/product/bicycleTireDedicatedSpecificationFormValuesAndValidation'

withDefaults(defineProps<{
  specification: BicycleTireSpecificationForm
  locked?: boolean
  error?: string
  catalogItems?: SchwalbeTireCatalogItem[]
  catalogLoading?: boolean
  catalogError?: string
  selectedArticleNo?: string
}>(), {
  locked: false,
  catalogItems: () => [],
  catalogLoading: false,
  catalogError: '',
  selectedArticleNo: '',
})
const emit = defineEmits<{ 'catalog-model-select': [articleNo: string]; 'retry-catalog': [] }>()
</script>
