<template>
  <AdminFormSection id="spokes-product-specification" title="辐条产品规格" description="只记录 SAPIM 对该型号公开的产品事实；没有可靠资料的字段留空。">
    <div class="grid gap-4 md:grid-cols-3">
      <AdminFormField label="制造商型号" required>
        <Input v-model="specification.model_name" :disabled="locked" placeholder="例如 CX-Ray" />
      </AdminFormField>
      <AdminFormField label="官方 MPN" description="仅填写 SAPIM 明确公布的零件编号。">
        <Input v-model="specification.manufacturer_part_number" :disabled="locked" />
      </AdminFormField>
      <AdminFormField label="制造商系列" description="例如 Aero spokes、Butted spokes 或 Basic spokes。">
        <Input v-model="specification.manufacturer_family" :disabled="locked" />
      </AdminFormField>
      <AdminFormField label="材质">
        <Input v-model="specification.material" :disabled="locked" />
      </AdminFormField>
      <AdminFormField label="Middle strength 最小值" description="单位 N/mm²，保留制造商原字段口径。">
        <Input v-model="specification.middle_strength_minimum_n_per_mm2" :disabled="locked" type="number" min="0" step="any" inputmode="decimal" />
      </AdminFormField>
      <AdminFormField label="Middle strength 最大值" description="单一值时最小值与最大值相同。">
        <Input v-model="specification.middle_strength_maximum_n_per_mm2" :disabled="locked" type="number" min="0" step="any" inputmode="decimal" />
      </AdminFormField>
      <AdminFormField label="制造商说明" class="md:col-span-3">
        <Textarea v-model="specification.manufacturer_description" :disabled="locked" class="min-h-20" />
      </AdminFormField>
    </div>

    <div class="mt-6 space-y-3">
      <div class="flex items-center justify-between gap-3">
        <h3 class="text-sm font-semibold">线规、尺寸与重量参考</h3>
        <Button type="button" variant="outline" size="sm" :disabled="locked" @click="specification.gauge_options.push(createEmptySpokeGaugeOptionForm())">
          <Plus class="size-4" />
          添加线规
        </Button>
      </div>

      <div v-for="(gauge, index) in specification.gauge_options" :key="index" class="rounded-md border p-3">
        <div class="mb-3 flex items-center justify-between gap-3">
          <span class="text-xs font-semibold text-muted-foreground">线规 {{ index + 1 }}</span>
          <Button v-if="specification.gauge_options.length > 1" type="button" variant="ghost" size="icon" :disabled="locked" :aria-label="`移除线规 ${index + 1}`" @click="specification.gauge_options.splice(index, 1)">
            <Trash2 class="size-4" />
          </Button>
        </div>
        <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
          <AdminFormField label="线规">
            <Input v-model="gauge.gauge" :disabled="locked" placeholder="例如 14G" />
          </AdminFormField>
          <AdminFormField label="等径直径 (mm)"><Input v-model="gauge.diameter_mm" :disabled="locked" type="number" min="0" step="any" /></AdminFormField>
          <AdminFormField label="花鼓端直径 (mm)"><Input v-model="gauge.hub_end_diameter_mm" :disabled="locked" type="number" min="0" step="any" /></AdminFormField>
          <AdminFormField label="中段直径 (mm)"><Input v-model="gauge.middle_diameter_mm" :disabled="locked" type="number" min="0" step="any" /></AdminFormField>
          <AdminFormField label="中段宽度 (mm)"><Input v-model="gauge.middle_width_mm" :disabled="locked" type="number" min="0" step="any" /></AdminFormField>
          <AdminFormField label="中段厚度 (mm)"><Input v-model="gauge.middle_thickness_mm" :disabled="locked" type="number" min="0" step="any" /></AdminFormField>
          <AdminFormField label="螺纹端直径 (mm)"><Input v-model="gauge.thread_end_diameter_mm" :disabled="locked" type="number" min="0" step="any" /></AdminFormField>
          <AdminFormField label="最短长度 (mm)"><Input v-model="gauge.length_min_mm" :disabled="locked" type="number" min="0" step="any" /></AdminFormField>
          <AdminFormField label="最长长度 (mm)"><Input v-model="gauge.length_max_mm" :disabled="locked" type="number" min="0" step="any" /></AdminFormField>
          <AdminFormField label="参考支数"><Input v-model="gauge.weight_reference.spoke_count" :disabled="locked" type="number" min="1" step="1" /></AdminFormField>
          <AdminFormField label="参考长度 (mm)"><Input v-model="gauge.weight_reference.length_mm" :disabled="locked" type="number" min="0" step="any" /></AdminFormField>
          <AdminFormField label="参考总重 (g)"><Input v-model="gauge.weight_reference.total_weight_g" :disabled="locked" type="number" min="0" step="any" /></AdminFormField>
        </div>
      </div>
    </div>

    <div class="mt-6 grid gap-4 md:grid-cols-2">
      <AdminFormField label="颜色选项" description="逗号分隔；只录入该型号已核实的颜色。">
        <Input v-model="specification.color_options" :disabled="locked" />
      </AdminFormField>
      <AdminFormField label="制造商选项">
        <Input v-model="specification.manufacturer_options" :disabled="locked" />
      </AdminFormField>
      <AdminFormField label="制造商推荐用途">
        <Input v-model="specification.recommended_uses" :disabled="locked" />
      </AdminFormField>
      <AdminFormField label="制造商宣称">
        <Input v-model="specification.manufacturer_claims" :disabled="locked" />
      </AdminFormField>
      <AdminFormField label="官方来源 URL" class="md:col-span-2" description="每行或以逗号分隔；只记录可公开核验的官方页面。">
        <Textarea v-model="specification.source_urls" :disabled="locked" class="min-h-20" />
      </AdminFormField>
    </div>
    <p v-if="error" role="alert" class="mt-3 text-sm text-destructive">{{ error }}</p>
  </AdminFormSection>
</template>

<script setup lang="ts">
import { Plus, Trash2 } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminFormSection from '@/components/admin/AdminFormSection.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { createEmptySpokeGaugeOptionForm, type SpokesProductSpecificationForm } from '@/modules/product/spokesDedicatedProductSpecificationFormValuesAndValidation'

withDefaults(defineProps<{
  specification: SpokesProductSpecificationForm
  locked?: boolean
  error?: string
}>(), {
  locked: false,
  error: '',
})
</script>
