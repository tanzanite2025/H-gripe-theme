<template>
  <GuideCategoryChildRouteNavigationCards
    eyebrow="Spoke Guides"
    heading="Spoke Guides"
    description="Choose a spoke tool or reference by the task you need to complete."
    open-label="Open guide"
    :cards="spokeGuideCards"
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useHead, useI18n } from '#imports'
import GuideCategoryChildRouteNavigationCards from '~/components/GuideCategoryChildRouteNavigationCards.vue'
import { usePageMessages } from '~/composables/usePageMessages'
import { spokeGuideTabs } from '~/utils/pageSubNavigation'

definePageMeta({
  path: '/guides/spokeguides',
  layout: 'products',
  breadcrumbLabelKey: 'products.nav.spokeGuides',
  breadcrumbLabelFallback: 'Spoke Guides',
  footerLabelKey: 'products.nav.spokeGuides',
  footerLabelFallback: 'Spoke Guides',
})

const { locale, t } = useI18n()
const spokeDislocationMessages = usePageMessages('guidesSpokeDislocationMechanics')
const spokeCalculatorMessages = usePageMessages('resourcesSpokeCalculator')
const wheelsetSpokeSpecsMessages = usePageMessages('brandWheelsetSpokeSpecs')

await Promise.all([
  spokeDislocationMessages.loadPageMessages(locale.value),
  spokeCalculatorMessages.loadPageMessages(locale.value),
  wheelsetSpokeSpecsMessages.loadPageMessages(locale.value),
])

useHead(() => ({
  title: t('products.nav.spokeGuides', 'Spoke Guides'),
}))

const spokeGuideCards = computed(() => spokeGuideTabs.map(tab => ({
  id: tab.id,
  label: t(tab.labelKey, tab.fallback),
  description: tab.description,
  to: tab.to,
})))
</script>
