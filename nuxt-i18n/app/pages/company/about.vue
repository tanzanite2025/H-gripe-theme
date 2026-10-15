<template>
  <div class="company-page">
    <ReusablePageSectionUnderlineTabs
      :tabs="companyAboutSectionNavigationTabs"
      :active-id="activeTab"
      :tab-list-label="t('company.aboutTabs.ariaLabel')"
    />

    <component :is="companyAboutSectionComponents[activeTab]" :key="activeTab" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { definePageMeta, useI18n } from '#imports'
import ReusablePageSectionUnderlineTabs from '~/components/ReusablePageSectionUnderlineTabs.vue'
import AboutFactory from '~/components/company/AboutFactory.vue'
import AboutAppearance from '~/components/company/AboutAppearance.vue'
import AboutHolePatterns from '~/components/company/AboutHolePatterns.vue'
import { usePageSubNavigationTab } from '~/composables/usePageSubNavigationTab'
import { companyAboutTabs } from '~/utils/pageSubNavigation'

const { t } = useI18n()
const { activeTab, localizedTabPath } = usePageSubNavigationTab({
  tabs: companyAboutTabs,
  basePath: '/company/about',
  defaultValue: 'factory',
})

const companyAboutSectionComponents = {
  factory: AboutFactory,
  appearance: AboutAppearance,
  'hole-patterns': AboutHolePatterns,
} as const

const companyAboutSectionNavigationTabs = computed(() => companyAboutTabs.map(tab => ({
  id: tab.id,
  label: t(tab.labelKey),
  to: localizedTabPath(tab.id),
})))

definePageMeta({
  layout: 'products',
  footerLabelKey: 'company.nav.about',
  footerLabelFallback: 'About Us',
})
</script>

<style scoped>
.company-page {
  display: flex;
  flex-direction: column;
}
</style>
