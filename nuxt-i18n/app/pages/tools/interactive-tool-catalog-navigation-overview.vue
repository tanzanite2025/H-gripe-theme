<template>
  <GuideCategoryChildRouteNavigationCards
    :eyebrow="t('toolsNavigation.eyebrow')"
    :heading="t('toolsNavigation.label')"
    :description="t('toolsNavigation.description')"
    :open-label="t('toolsNavigation.openTool')"
    :cards="availableInteractiveToolCards"
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useHead, useI18n } from '#imports'
import GuideCategoryChildRouteNavigationCards from '~/components/GuideCategoryChildRouteNavigationCards.vue'

definePageMeta({
  path: '/tools',
  layout: 'products',
  breadcrumbLabelKey: 'toolsNavigation.label',
  breadcrumbLabelFallback: 'Tools',
  footerLabelKey: 'toolsNavigation.label',
  footerLabelFallback: 'Tools',
})

const { t } = useI18n()

useHead(() => ({
  title: `${t('toolsNavigation.label')} | Tanzanite`,
  meta: [{ name: 'description', content: t('toolsNavigation.description') }],
}))

// List complete, available tools only. Add later migrations after acceptance.
const availableInteractiveToolCards = computed(() => [{
  id: 'schwalbe-tire-selector',
  label: t('toolsNavigation.schwalbeTireSelectorLabel'),
  description: t('toolsNavigation.schwalbeTireSelectorDescription'),
  to: '/tools/schwalbe-tire-selector',
}])
</script>
