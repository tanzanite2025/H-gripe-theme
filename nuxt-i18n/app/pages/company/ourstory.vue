<template>
  <div class="company-page">

    <h1 class="company-page__title company-page__title--sr-only">Our Story</h1>

    <div class="company-tabs" role="tablist">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        type="button"
        class="company-tabs__item"
        :class="{ 'company-tabs__item--active': activeTab === tab.id }"
        @click="setActiveTab(tab.id)"
      >
        {{ tab.label }}
      </button>
    </div>

    <section
      v-show="activeTab === 'factory'"
      id="factory"
      class="company-section"
    >
      <h2 class="company-section__title">Factory</h2>
      <p class="company-section__body">
        后续将在此 tab 展示工厂介绍、产线流程和照片等内容。目前为占位文案，结构已保留，后续直接替换即可。
      </p>
      <p class="company-section__body">
        你可以在这里继续添加步骤卡片、图库或数据列表，TAB 结构已经固定，填充内容不会影响切换。
      </p>
    </section>

    <section
      v-show="activeTab === 'appearance'"
      id="appearance"
      class="company-section"
    >
      <h3 class="company-section__title">Appearance</h3>
      <p class="company-section__body">
        Use this tab to highlight the visual details of Tanzanite rims and wheelsets: finishes, decals, logo styles, and paint or paintless options.
      </p>
      <p class="company-section__body">
        Later you can replace this placeholder with your own photos and descriptions of different appearance packages so riders can quickly understand what their wheels will look like on the bike.
      </p>
    </section>

    <section
      v-show="activeTab === 'facility'"
      id="facility"
      class="company-section"
    >
      <h3 class="company-section__title">Facility</h3>
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
      <h3 class="company-section__title">Manufacture</h3>
      <p class="company-section__body">
        Use this tab to walk visitors through how a Tanzanite wheel is manufactured,
        from selecting rims and hubs to lacing, tensioning, truing, and final
        inspection.
      </p>
      <p class="company-section__body">
        Later you can add step-by-step photos, short videos, or diagrams so
        builders understand what level of craftsmanship and tooling goes into
        every wheel.
      </p>
    </section>

    <section
      v-show="activeTab === 'qualitycontrol'"
      id="qualitycontrol"
      class="company-section"
    >
      <h3 class="company-section__title">Quality control</h3>
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
import { ref } from 'vue'
import { useHead, definePageMeta } from '#imports'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'

type OurStoryTabId = 'factory' | 'appearance' | 'facility' | 'manufacture' | 'qualitycontrol'

const tabs: { id: OurStoryTabId; label: string }[] = [
  { id: 'factory', label: 'Factory' },
  { id: 'appearance', label: 'Appearance' },
  { id: 'facility', label: 'Facility' },
  { id: 'manufacture', label: 'Manufacture' },
  { id: 'qualitycontrol', label: 'Quality control' },
]

const activeTab = ref<OurStoryTabId>('factory')

const setActiveTab = (id: OurStoryTabId) => {
  activeTab.value = id
}

definePageMeta({
  layout: 'products',
})

useHead({
  title: 'Our Story',
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
  font-size: 1.5rem;
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
  color: rgba(148, 163, 184, 0.9);
}

.company-tabs {
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

.company-tabs::-webkit-scrollbar {
  display: none;
}

.company-tabs__item {
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

.company-tabs__item:active {
  transform: scale(0.96);
}

.company-tabs__item:hover {
  background: rgba(51, 65, 85, 0.95);
  color: #ffffff;
}

.company-tabs__item--active {
  background: linear-gradient(135deg, #2dd4bf 0%, #3b82f6 100%);
  color: #000000;
  border: none;
  font-weight: 600;
  box-shadow: 0 4px 12px rgba(45, 212, 191, 0.3);
}

.company-section {
  margin-top: 0;
}

.company-section__title {
  margin: 0 0 0.5rem;
  font-size: 1.1rem;
  font-weight: 600;
  color: #e5e7eb;
}

.company-section__body {
  margin: 0 0 0.5rem;
  font-size: 0.95rem;
  color: rgba(148, 163, 184, 0.95);
}

.company-section__list {
  margin: 0;
  padding-left: 1.1rem;
  list-style-type: disc;
  font-size: 0.95rem;
  color: rgba(148, 163, 184, 0.95);
}

.company-section__list li + li {
  margin-top: 0.25rem;
}

.factory-flow-grid {
  margin-top: 1rem;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.75rem;
}

.factory-flow-card {
  position: relative;
  aspect-ratio: 1 / 1;
  border-radius: 0.9rem;
  overflow: hidden;
  border: 1px solid rgba(148, 163, 184, 0.35);
  background: #020617;
  box-shadow: 0 0 25px rgba(15, 23, 42, 0.9);
}

.factory-flow-card__image {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  transform-origin: center;
  transition: transform 0.35s ease;
}

.factory-flow-card:hover .factory-flow-card__image {
  transform: scale(1.05);
}

.factory-flow-card__overlay {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  justify-content: flex-end;
  padding: 0.75rem 0.8rem 0.65rem;
  background: linear-gradient(
    to top,
    rgba(15, 23, 42, 0.96) 0%,
    rgba(15, 23, 42, 0.9) 35%,
    rgba(15, 23, 42, 0.6) 55%,
    transparent 100%
  );
  color: #f9fafb;
}

.factory-flow-card__step {
  margin: 0 0 0.25rem;
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  color: #22d3ee;
}

.factory-flow-card__step-dot {
  width: 6px;
  height: 6px;
  border-radius: 999px;
  background: radial-gradient(circle at 30% 30%, #a5f3fc, #22d3ee 55%, #0ea5e9 100%);
  box-shadow: 0 0 9px rgba(56, 189, 248, 0.9);
}

.factory-flow-card__title {
  margin: 0 0 0.1rem;
  font-size: 0.9rem;
  font-weight: 600;
}

.factory-flow-card__text {
  margin: 0;
  font-size: 0.78rem;
  color: rgba(226, 232, 240, 0.9);
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
  color: #f9fafb;
}

.company-values__body {
  margin: 0;
  font-size: 0.9rem;
  color: rgba(148, 163, 184, 0.95);
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
  color: #e5e7eb;
  min-width: 3.5rem;
}

.company-timeline__content {
  font-size: 0.9rem;
  color: rgba(148, 163, 184, 0.95);
}

.company-feedback {
  margin-top: 2.5rem;
}

@media (min-width: 768px) {
  .company-page__title {
    font-size: 1.75rem;
  }

  .company-values {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .company-tabs {
    flex-wrap: wrap;
    justify-content: center;
    margin: 0 0 1rem;
    padding: 4px 0;
    max-width: 100%;
  }
}

@media (max-width: 768px) {
  .company-tabs {
    justify-content: flex-start;
  }

  .factory-flow-grid {
    grid-template-columns: 1fr;
  }
}
</style>
