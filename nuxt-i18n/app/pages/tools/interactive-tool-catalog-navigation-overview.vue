<template>
  <GuideCategoryChildRouteNavigationCards
    :eyebrow="t('toolsNavigation.eyebrow')"
    :heading="t('toolsNavigation.label')"
    :description="t('toolsNavigation.description')"
    :open-label="t('toolsNavigation.openTool')"
    :cards="availableInteractiveToolCards"
  >
    <template #header="{ headingId }">
      <InteractiveToolPageHeader
        :heading-id="headingId"
        :eyebrow="t('toolsNavigation.eyebrow')"
        :title="t('toolsNavigation.label')"
        :description="t('toolsNavigation.description')"
      />
    </template>
  </GuideCategoryChildRouteNavigationCards>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useHead, useI18n } from '#imports'
import GuideCategoryChildRouteNavigationCards from '~/components/GuideCategoryChildRouteNavigationCards.vue'
import InteractiveToolPageHeader from '~/components/InteractiveToolPageHeader.vue'

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
}, {
  id: 'brand-wheelset-spoke-specs',
  label: t('toolsNavigation.brandWheelsetSpokeSpecsLabel'),
  description: t('toolsNavigation.brandWheelsetSpokeSpecsDescription'),
  to: '/tools/brand-wheelset-spoke-specs',
}, {
  id: 'spoke-length-calculator',
  label: t('toolsNavigation.spokeLengthCalculatorLabel'),
  description: t('toolsNavigation.spokeLengthCalculatorDescription'),
  to: '/tools/spoke-length-calculator',
}, {
  id: 'spoke-lacing-topology-simulator',
  label: t('toolsNavigation.spokeLacingTopologySimulatorLabel'),
  description: t('toolsNavigation.spokeLacingTopologySimulatorDescription'),
  to: '/tools/spoke-lacing-topology-simulator',
}, {
  id: 'dynamic-tire-pressure-calculator',
  label: t('toolsNavigation.dynamicTirePressureCalculatorLabel'),
  description: t('toolsNavigation.dynamicTirePressureCalculatorDescription'),
  to: '/tools/dynamic-tire-pressure-calculator',
}, {
  id: 'tire-rim-width-compatibility-engine',
  label: t('toolsNavigation.tireRimWidthCompatibilityEngineLabel'),
  description: t('toolsNavigation.tireRimWidthCompatibilityEngineDescription'),
  to: '/tools/tire-rim-width-compatibility-engine',
}, {
  id: 'tire-size-marking-converter',
  label: t('toolsNavigation.tireSizeMarkingConverterLabel'),
  description: t('toolsNavigation.tireSizeMarkingConverterDescription'),
  to: '/tools/tire-size-marking-converter',
}, {
  id: 'valve-length-fitment-calculator',
  label: t('toolsNavigation.valveLengthFitmentCalculatorLabel'),
  description: t('toolsNavigation.valveLengthFitmentCalculatorDescription'),
  to: '/tools/valve-length-fitment-calculator',
}, {
  id: 'freehub-cassette-fitment-engine',
  label: t('toolsNavigation.freehubCassetteFitmentEngineLabel'),
  description: t('toolsNavigation.freehubCassetteFitmentEngineDescription'),
  to: '/tools/freehub-cassette-fitment-engine',
}, {
  id: 'spoke-stress-relief-mechanics',
  label: t('toolsNavigation.spokeStressReliefMechanicsLabel'),
  description: t('toolsNavigation.spokeStressReliefMechanicsDescription'),
  to: '/tools/spoke-stress-relief-mechanics',
}])
</script>
