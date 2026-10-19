<template>
  <dl v-if="specification" class="grid grid-cols-2 gap-2 rounded-lg border p-3 text-sm" data-carbon-rim-selected-sku-facts>
    <dt>{{ locale === 'zh_cn' ? '单只车圈净重' : 'Net weight per rim' }}</dt>
    <dd>{{ specification.net_weight_grams }} g</dd>
    <template v-if="specification.erd_mm != null">
      <dt>ERD</dt><dd>{{ specification.erd_mm }} mm</dd>
      <dt>{{ locale === 'zh_cn' ? '测量约定' : 'Measurement convention' }}</dt>
      <dd class="break-words">{{ specification.erd_measurement_notes }}</dd>
    </template>
  </dl>
</template>

<script setup lang="ts">
import { useI18n } from '#imports'
import type { CarbonRimVariantSpecification } from '~/types/productDetail'
defineProps<{ specification?: CarbonRimVariantSpecification | null }>()
const { locale } = useI18n()
</script>
