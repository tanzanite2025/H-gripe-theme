<template>
  <div class="company-page">

    <h1 class="company-page__title company-page__title--sr-only">About us</h1>

    <AboutFactory
      v-show="activeTab === 'factory'"
    />

    <AboutAppearance
      v-show="activeTab === 'appearance'"
    />

    <AboutHolePatterns
      v-show="activeTab === 'hole-patterns'"
    />

    <section
      v-show="activeTab === 'facility'"
      id="facility"
      class="company-section"
    >
      <h2 class="company-section__title">Facility</h2>
      <p class="company-section__body">
        Describe the key facilities that support the factory: warehousing, assembly
        areas, test labs, packing zones, and any dedicated spaces for training or
        demonstrations.
      </p>
      <p class="company-section__body">
        You can later replace this placeholder text with concrete details about
        climate control, storage systems, and how the layout helps keep builds
        organized and predictable.
      </p>
    </section>

    <section
      v-show="activeTab === 'manufacture'"
      id="manufacture"
      class="company-section"
    >
      <SmartAccordion default-id="rim-build">
        <AccordionItem id="rim-build" title="1. RIM BUILD">
           <div class="p-4 tz-text-secondary text-sm">
              Details about Rim Build...
           </div>
        </AccordionItem>

        <AccordionItem id="wheelset-build" title="2. WHEELSET BUILD">
           <div class="p-4 tz-text-secondary text-sm">
              Details about Wheelset Build...
           </div>
        </AccordionItem>

        <AccordionItem id="carbon-spoke-build" title="3. CARBON SPOKE BUILD">
           <div class="p-4 tz-text-secondary text-sm">
              Details about Carbon Spoke Build...
           </div>
        </AccordionItem>
      </SmartAccordion>
    </section>

    <section
      v-show="activeTab === 'qualitycontrol'"
      id="qualitycontrol"
      class="company-section"
    >
      <h2 class="company-section__title">Quality control</h2>
      <p class="company-section__body">
        Summarize how quality control works at Tanzanite: incoming material checks,
        in-process measurements, and final wheel verification before shipping.
      </p>
      <p class="company-section__body">
        This placeholder can later be replaced with your actual test procedures,
        measurement tolerances, and any certifications or standards that the
        factory follows.
      </p>
    </section>

    <div class="company-feedback">
      <UserFeedbackThread
        threadKey="company-ourstory"
        title="Share your feedback about Our Story and the factory"
      />
    </div>


  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useHead, definePageMeta, useRoute } from '#imports'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'
import AboutFactory from '~/components/company/AboutFactory.vue'
import AboutAppearance from '~/components/company/AboutAppearance.vue'
import AboutHolePatterns from '~/components/company/AboutHolePatterns.vue'
import SmartAccordion from '~/components/ui/SmartAccordion.vue'
import AccordionItem from '~/components/ui/AccordionItem.vue'
import {
  companyAboutTabs,
  isPageSubNavigationTabId,
  type CompanyAboutTabId,
} from '~/utils/pageSubNavigation'

const tabs = companyAboutTabs

const activeTab = ref<CompanyAboutTabId>('factory')
const route = useRoute()

const getTabFromHash = (hash: string): CompanyAboutTabId | null => {
  const raw = String(hash || '').replace(/^#/, '')
  return isPageSubNavigationTabId(tabs, raw) ? raw : null
}

watch(
  () => route.hash,
  (hash) => {
    const next = getTabFromHash(hash)
    if (next) activeTab.value = next
  },
  { immediate: true }
)

definePageMeta({
  layout: 'products',
})

useHead({
  title: 'About us',
})
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
  color: #f9fafb;
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
  border: 1px solid rgba(148, 163, 184, 0.6);
  background: rgba(15, 23, 42, 0.9);
  color: var(--tz-text-primary);
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
  background: rgba(15, 23, 42, 0.8);
  border: 1px solid rgba(148, 163, 184, 0.24);
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
