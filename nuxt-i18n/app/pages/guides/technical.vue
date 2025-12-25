<template>
  <div>
    <h2 class="products-page__title products-page__title--sr-only">Technical</h2>

    <div class="technical-page">
      <div class="technical-tabs" role="tablist">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          type="button"
          class="technical-tabs__item"
          :class="{ 'technical-tabs__item--active': activeTab === tab.id }"
          @click="setActiveTab(tab.id)"
        >
          {{ tab.label }}
        </button>
      </div>

      <!-- Rims -->
      <section
        v-show="activeTab === 'rims'"
        id="rims"
        class="technical-section sizecharts-section"
      >
        <h2 class="sizecharts-section__title">Rims</h2>
        <p class="sizecharts-section__intro">
          Overview of Tanzanite rims and how rim dimensions relate to tire
          compatibility and spoke length.
        </p>
        <ul class="sizecharts-section__list">
          <li>Rim diameter standards (ISO / ETRTO such as 622, 584, etc.).</li>
          <li>Internal / external width and their impact on tire choice.</li>
          <li>Spoke hole count (28H / 32H / 36H …) and recommended use cases.</li>
          <li>ERD (Effective Rim Diameter) definition and basic measurement notes.</li>
          <li>How ERD is used when entering data into the Spoke Calculator.</li>
        </ul>
      </section>

      <!-- Spokes -->
      <section
        v-show="activeTab === 'spokes'"
        id="spokes"
        class="technical-section sizecharts-section"
      >
        <h2 class="sizecharts-section__title">Spokes</h2>
        <TechnicalSpokesSection />
      </section>

      <!-- Spoke pattern -->
      <section
        v-show="activeTab === 'spoke-pattern'"
        id="spoke-pattern"
        class="sizecharts-section"
      >
        <h2 class="sizecharts-section__title">Spoke pattern</h2>
        <p class="sizecharts-section__intro">
          Content coming soon...
        </p>
      </section>

      <!-- Spoke length -->
      <section
        v-show="activeTab === 'spoke-length'"
        id="spoke-length"
        class="sizecharts-section"
      >
        <h2 class="sizecharts-section__title">Spoke length</h2>
      </section>

      <!-- Hubs -->
      <section
        v-show="activeTab === 'hubs'"
        id="hubs"
        class="sizecharts-section"
      >
        <h2 class="sizecharts-section__title">Hubs</h2>
        <TechnicalHubsSection />
      </section>

      <!-- Tension -->
      <section
        v-show="activeTab === 'tension'"
        id="tension"
        class="sizecharts-section"
      >
        <h2 class="sizecharts-section__title">Tension &amp; formulas</h2>
        <TechnicalTensionSection />
      </section>
      <div class="technical-feedback">
        <UserFeedbackThread
          threadKey="guides-technical"
          title="Share your feedback about this Technical guide"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useLocalePath, useRoute } from '#imports'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'
import TechnicalSpokesSection from '~/components/TechnicalSpokesSection.vue'
import TechnicalHubsSection from '~/components/TechnicalHubsSection.vue'
import TechnicalTensionSection from '~/components/TechnicalTensionSection.vue'

definePageMeta({
  layout: 'products',
})

useHead({
  title: 'Technical',
})

type TechnicalTabId = 'rims' | 'spokes' | 'spoke-pattern' | 'spoke-length' | 'hubs' | 'tension'

const localePath = useLocalePath()
const route = useRoute()

const tabs: { id: TechnicalTabId; label: string }[] = [
  { id: 'rims', label: 'Rims' },
  { id: 'spokes', label: 'Spokes' },
  { id: 'spoke-pattern', label: 'Spoke pattern' },
  { id: 'spoke-length', label: 'Spoke length' },
  { id: 'hubs', label: 'Hubs' },
  { id: 'tension', label: 'Tension' },
]

const activeTab = ref<TechnicalTabId>('rims')

const getTabFromHash = (hash: string): TechnicalTabId | null => {
  const raw = String(hash || '').replace(/^#/, '')
  const allowed: TechnicalTabId[] = ['rims', 'spokes', 'spoke-pattern', 'spoke-length', 'hubs', 'tension']
  return (allowed as string[]).includes(raw) ? (raw as TechnicalTabId) : null
}

watch(
  () => route.hash,
  (hash) => {
    const next = getTabFromHash(hash)
    if (next) activeTab.value = next
  },
  { immediate: true }
)

const setActiveTab = (id: TechnicalTabId) => {
  activeTab.value = id
  if (typeof window !== 'undefined') {
    const url = new URL(window.location.href)
    url.hash = `#${id}`
    window.history.replaceState(null, '', url.toString())
  }
}
</script>

<style scoped>
.products-page__title {
  margin: 0 0 0.75rem;
  font-size: 1.5rem;
  font-weight: 600;
  color: #f9fafb;
}

.products-page__title--sr-only {
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

.products-page__intro {
  margin: 0 0 1rem;
  font-size: 0.95rem;
  color: rgba(148, 163, 184, 0.9);
}

.technical-page {
  margin: 0.25rem auto 0;
  max-width: 900px;
}

.technical-tabs {
  display: flex;
  overflow-x: auto;
  gap: 12px;
  padding: 4px 16px;
  margin: 0 -16px 1rem;
  max-width: calc(100% + 32px);
  -webkit-overflow-scrolling: touch;
  scrollbar-width: none;
  touch-action: pan-x;
}

.technical-tabs::-webkit-scrollbar {
  display: none;
}

.technical-tabs__item {
  flex-shrink: 0;
  border: none;
  border-radius: 9999px;
  padding: 8px 18px;
  font-size: 0.85rem;
  font-weight: 500;
  color: #ffffff;
  background: rgba(31, 41, 55, 0.9);
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  white-space: nowrap;
  backdrop-filter: blur(4px);
  box-shadow:
    0 3px 9px -6px rgba(0, 0, 0, 0.9),
    0 0 9px rgba(0, 0, 0, 0.85);
}

.technical-tabs__item:active {
  transform: scale(0.96);
}

.technical-tabs__item:hover {
  background: rgba(51, 65, 85, 0.95);
  color: #ffffff;
}

.technical-tabs__item--active {
  background: linear-gradient(135deg, #2dd4bf 0%, #3b82f6 100%);
  color: #000000;
  border: none;
  font-weight: 600;
  box-shadow: 0 4px 12px rgba(45, 212, 191, 0.3);
}

@media (min-width: 768px) {
  .technical-tabs {
    flex-wrap: wrap;
    justify-content: center;
    margin: 0 0 1rem;
    padding: 4px 0;
    max-width: 100%;
  }
}

.technical-section {
  margin-top: 0.75rem;
}

.technical-materials {
  margin-top: 0.75rem;
}

.technical-materials > .sizecharts-section__list > li {
  padding: 0.75rem 0.85rem;
  border-radius: 0.75rem;
  background: rgba(15, 23, 42, 0.75);
  border: 1px solid rgba(148, 163, 184, 0.35);
}

.technical-materials > .sizecharts-section__list > li + li {
  margin-top: 0.6rem;
}

.technical-materials__heading {
  display: block;
  margin-bottom: 0.25rem;
  color: #fbbf24;
}

.technical-link-button {
  margin-left: 0.5rem;
  padding: 0.15rem 0.6rem;
  border-radius: 9999px;
  border: 1px solid rgba(56, 189, 248, 0.9);
  font-size: 0.8rem;
  font-weight: 500;
  color: #e0f2fe;
  text-decoration: none;
  background: rgba(15, 23, 42, 0.9);
  transition: background-color 0.15s ease, border-color 0.15s ease,
    color 0.15s ease;
}

.technical-link-button:hover {
  background: rgba(56, 189, 248, 0.2);
  border-color: rgba(56, 189, 248, 1);
  color: #ffffff;
}

.technical-brand-button {
  margin-left: 0.5rem;
  margin-top: 0.15rem;
  padding: 0.25rem 0.8rem;
  border-radius: 9999px;
  border: 1px solid rgba(56, 189, 248, 0.9);
  background-image: linear-gradient(
    135deg,
    rgba(56, 189, 248, 0.9),
    rgba(59, 130, 246, 0.95)
  );
  color: #0b1020;
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.technical-inline-image {
  margin-top: 0.4rem;
}

.technical-inline-image img {
  display: block;
  max-width: 260px;
  width: 100%;
  height: auto;
  border-radius: 0.5rem;
}

.technical-inline-image--spoke-length img {
  max-width: 380px;
}

.technical-inline-image--spoke-shapes {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
}

.technical-inline-image--spoke-length {
  /* AUTO-FIT GRID for GuideImage rows: 1 image = full row, 2 images = two equal columns */
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 0.75rem;
}

.technical-inline-image--erd {
  /* AUTO-FIT GRID for GuideImage rows: 1 image = full row, 2 images = two equal columns */
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 0.75rem;
}

.technical-spoke-length-heading {
  color: #38bdf8;
}

.technical-list-item--no-bullet {
  padding-left: 0;
}

.technical-list-item--no-bullet::before {
  content: none;
}

.technical-feedback {
  margin-top: 2.5rem;
}

/* Nested lists inside shared guide lists on this page */
.technical-page .sizecharts-section__list ul {
  margin-top: 0.15rem;
  padding-left: 1.1rem;
  list-style-type: none;
}

@media (min-width: 768px) {
  .technical-section {
    margin-top: 1rem;
  }
}

@media (max-width: 768px) {
  .technical-tabs {
    justify-content: flex-start;
  }
}
</style>
