<template>
  <NuxtLink
    v-if="isRimOrWheelsetSpecificationTemplate"
    :to="localePath('/company/about/appearance')"
    class="product-rim-appearance-guide-link"
  >
    {{ t('products.detail.rimAppearanceGuide') }}
  </NuxtLink>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n, useLocalePath } from '#imports'

const props = defineProps<{ productSpecificationTemplateSlug: string }>()
const { t } = useI18n()
const localePath = useLocalePath()
const isRimOrWheelsetSpecificationTemplate = computed(() => (
  ['rim', 'carbon_rim', 'wheelset'].includes(props.productSpecificationTemplateSlug)
))
</script>

<style scoped>
.product-rim-appearance-guide-link {
  align-self: flex-start;
  color: var(--tz-text-accent);
  font-size: 0.85rem;
  line-height: 1.5;
  text-decoration: underline;
  text-underline-offset: 0.2em;
}

.product-rim-appearance-guide-link:focus-visible {
  outline: 2px solid var(--tz-site-accent);
  outline-offset: 3px;
}
</style>
