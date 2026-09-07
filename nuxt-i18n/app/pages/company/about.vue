<template>
  <div class="company-page">

    <h1 class="company-page__title company-page__title--sr-only">{{ t('company.nav.about') }}</h1>

    <AboutFactory
      v-if="activeTab === 'factory'"
    />

    <AboutAppearance
      v-else-if="activeTab === 'appearance'"
    />

    <AboutHolePatterns
      v-else-if="activeTab === 'hole-patterns'"
    />

    <AboutFacility v-else-if="activeTab === 'facility'" />

    <AboutManufacture v-else-if="activeTab === 'manufacture'" />

    <AboutQualityControl v-else-if="activeTab === 'qualitycontrol'" />

    <div class="company-feedback">
      <UserFeedbackThread
        threadKey="company-ourstory"
      />
    </div>


  </div>
</template>

<script setup lang="ts">
import { useHead, definePageMeta, useI18n } from '#imports'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'
import AboutFactory from '~/components/company/AboutFactory.vue'
import AboutAppearance from '~/components/company/AboutAppearance.vue'
import AboutHolePatterns from '~/components/company/AboutHolePatterns.vue'
import AboutFacility from '~/components/company/AboutFacility.vue'
import AboutManufacture from '~/components/company/AboutManufacture.vue'
import AboutQualityControl from '~/components/company/AboutQualityControl.vue'
import { usePageSubNavigationTab } from '~/composables/usePageSubNavigationTab'
import { companyAboutTabs } from '~/utils/pageSubNavigation'

const { t } = useI18n()
const tabs = companyAboutTabs
const { activeTab } = usePageSubNavigationTab({
  tabs,
  basePath: '/company/about',
  defaultValue: 'factory',
})

definePageMeta({
  layout: 'products',
  footerLabelKey: 'company.nav.about',
  footerLabelFallback: 'About Us',
})

useHead(() => ({
  title: t('company.nav.about'),
}))
</script>

<style scoped>
.company-page {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.company-page__title {
  margin: 0 0 0.75rem;
  font-size: var(--tz-type-page-title);
  line-height: 1.18;
  font-weight: 600;
  color: var(--tz-text-primary);
}

.company-page__title--sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

.company-page__intro {
  margin: 0 0 0.75rem;
  font-size: 0.95rem;
  color: var(--tz-text-secondary);
}

/* Page-level tab entry points are rendered by the header/mobile mega menu. */

.company-section {
  margin-top: 0;
}

.company-section__title {
  margin: 0 0 0.5rem;
  font-size: var(--tz-type-section-title);
  line-height: 1.35;
  font-weight: 600;
  color: var(--tz-text-primary);
  text-align: center;
}

.company-section__body {
  margin: 0 0 0.5rem;
  font-size: 0.95rem;
  color: var(--tz-text-secondary);
  text-align: center;
}

.company-section__list {
  margin: 0.25rem auto 0;
  padding-left: 0;
  list-style-type: none;
  font-size: 0.95rem;
  color: var(--tz-text-secondary);
  text-align: center;
}

.company-section__list li + li {
  margin-top: 0.25rem;
}

.company-video-button {
  margin-top: 0.5rem;
  padding: 0.35rem 0.85rem;
  border-radius: 9999px;
  border: 1px solid var(--tz-action-primary);
  background: var(--tz-action-primary);
  color: var(--tz-action-primary-foreground);
  font-size: 0.85rem;
  font-weight: 500;
  cursor: pointer;
}





.company-section--values {
  border-top: 1px solid rgba(148, 163, 184, 0.25);
  padding-top: 1.25rem;
}

.company-section--timeline {
  border-top: 1px solid rgba(148, 163, 184, 0.25);
  padding-top: 1.25rem;
}

.company-section--cta {
  border-top: 1px solid rgba(148, 163, 184, 0.25);
  padding-top: 1.25rem;
}

.company-values {
  display: grid;
  grid-template-columns: 1fr;
  gap: 0.75rem;
}

.company-values__item {
  padding: 0.75rem 0.9rem;
  border-radius: 0.75rem;
  background: var(--tz-card-surface);
  border: 1px solid var(--tz-border-subtle);
}

.company-values__title {
  margin: 0 0 0.35rem;
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--tz-text-primary);
}

.company-values__body {
  margin: 0;
  font-size: 0.9rem;
  color: var(--tz-text-secondary);
}

.company-timeline {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.company-timeline__item {
  display: grid;
  grid-template-columns: auto 1fr;
  column-gap: 0.75rem;
  row-gap: 0.25rem;
  align-items: flex-start;
}

.company-timeline__year {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--tz-text-primary);
  min-width: 3.5rem;
}

.company-timeline__content {
  font-size: 0.9rem;
  color: var(--tz-text-secondary);
}

.company-feedback {
  margin-top: 2.5rem;
}

@media (min-width: 768px) {
  .company-values {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

}



</style>
