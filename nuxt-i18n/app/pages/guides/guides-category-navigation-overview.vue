<template>
  <GuideCategoryChildRouteNavigationCards
    eyebrow="Guides"
    heading="Guides"
    description="Browse tire, wheelset, and spoke engineering guides from one clear starting point."
    open-label="Open guide"
    :cards="guideCategoryCards"
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useHead, useI18n } from '#imports'
import GuideCategoryChildRouteNavigationCards from '~/components/GuideCategoryChildRouteNavigationCards.vue'

definePageMeta({
  path: '/guides',
  layout: 'products',
  breadcrumbLabelKey: 'breadcrumbs.guides',
  breadcrumbLabelFallback: 'Guides',
})

const { t } = useI18n()

useHead(() => ({
  title: t('breadcrumbs.guides', 'Guides'),
}))

const guideCategoryCards = computed(() => [
  {
    id: 'tire-guides',
    label: t('products.nav.tireSizeCharts', 'Tire Guides'),
    description: 'Tire sizing, pressure, clearance, tubeless setup, and installation references.',
    to: '/guides/tireguides',
  },
  {
    id: 'wheelset-buyers-guide',
    label: t('products.nav.wheelsetBuyersGuide', 'Wheelset Guide'),
    description: 'Wheel selection, safety, custom builds, appearance, and freehub guidance.',
    to: '/guides/wheelset-buyers',
  },
  {
    id: 'spoke-guides',
    label: t('products.nav.spokeGuides', 'Spoke Guides'),
    description: 'Spoke length, wheelset specifications, and stainless steel spoke engineering tools.',
    to: '/guides/spokeguides',
  },
])
</script>
