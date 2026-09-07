<template>
  <div class="support-wheelset-test-report rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-3 md:p-6">
    <h3 class="support-section__title text-center">{{ t('testReportWheelset.title') }}</h3>

    <p class="support-section__body mt-4 text-center">
      {{ t('testReportWheelset.intro') }}
    </p>

    <div class="mt-3 flex justify-center">
      <div class="support-wheelset-test-report__assembly-note inline-flex items-center rounded-full px-3 py-1 text-xs sm:text-sm">
        <span>{{ t('testReportWheelset.assemblyNote.prefix') }}</span>
        <button
          type="button"
          class="support-wheelset-test-report__assembly-note-action ml-1 underline underline-offset-2 transition-colors duration-150"
          @click="goToWheelsetAssembly"
        >
          {{ t('testReportWheelset.assemblyNote.link') }}
        </button>
      </div>
    </div>

    <div
      v-for="test in wheelsetTests"
      :key="test.id"
      class="rim-test-card mt-4"
      :class="{ 'mt-6': test.id === 'lateral' }"
    >
      <h4 class="sizecharts-section__subheading text-emerald-700 font-semibold">
        {{ t(wheelsetMessage(test.id, 'title')) }}
      </h4>
      <p class="support-section__body">
        <strong>{{ t('testReportWheelset.labels.purpose') }}</strong>
        {{ t(wheelsetMessage(test.id, 'purpose')) }}
      </p>
      <p class="support-section__body mt-2">
        <strong>{{ t('testReportWheelset.labels.testMethod') }}</strong>
        <span class="support-wheelset-test-report__method-badge inline-flex items-center gap-1.5 ml-2 px-2.5 py-0.5 rounded-md text-xs font-medium tz-text-secondary align-middle">
          <svg class="w-3.5 h-3.5 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
          </svg>
          {{ t(wheelsetMessage(test.id, 'standard')) }}
        </span>
      </p>
      <ul class="sizecharts-section__list support-section__body">
        <li v-for="index in test.methodCount" :key="`method-${index}`">
          {{ t(wheelsetMessage(test.id, `method.${index - 1}`)) }}
        </li>
      </ul>
      <p class="support-section__body mt-2">
        <strong>{{ t('testReportWheelset.labels.evaluationCriteria') }}</strong>
      </p>
      <ul class="sizecharts-section__list support-section__body">
        <li v-for="index in test.criteriaCount" :key="`criteria-${index}`">
          {{ t(wheelsetMessage(test.id, `criteria.${index - 1}`)) }}
        </li>
      </ul>
    </div>

    <div class="mt-4 flex justify-center">
      <div
        class="support-video-thumbnail"
        @click="openWheelsetVideo"
      >
        <img
          class="support-video-thumbnail__image"
          src="/testreport/wheelsettestreport/wheelssettestroport-video-firstpicture.webp"
          :alt="t('testReportWheelset.video.alt')"
          loading="lazy"
        />
        <div class="support-video-thumbnail__overlay">
          <span class="support-video-thumbnail__icon">▶</span>
          <span class="support-video-thumbnail__label">{{ t('testReportWheelset.video.label') }}</span>
        </div>
      </div>
    </div>

    <div class="support-wheelset-test-report__disclaimer mt-6 rounded-lg px-4 py-3 text-sm leading-relaxed">
      <h4 class="support-wheelset-test-report__disclaimer-label mb-3 font-semibold">
        {{ t('testReportWheelset.disclaimer.title') }}
      </h4>
      <p class="mb-2">{{ t('testReportWheelset.disclaimer.body.0') }}</p>
      <p>{{ t('testReportWheelset.disclaimer.body.1') }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useI18n } from '#imports'
import { usePageMessages } from '~/composables/usePageMessages'

const { openWheelsetVideo, goToWheelsetAssembly } = defineProps<{
  openWheelsetVideo: () => void | Promise<void>
  goToWheelsetAssembly: () => void | Promise<void>
}>()

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('testReportWheelset')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

const wheelsetTests = [
  { id: 'lateral', methodCount: 3, criteriaCount: 3 },
  { id: 'torsional', methodCount: 4, criteriaCount: 3 },
  { id: 'environmental', methodCount: 5, criteriaCount: 4 },
  { id: 'balance', methodCount: 4, criteriaCount: 4 },
  { id: 'fatigue', methodCount: 4, criteriaCount: 3 },
  { id: 'braking', methodCount: 5, criteriaCount: 4 },
] as const

const wheelsetMessage = (testId: string, key: string) => `testReportWheelset.tests.${testId}.${key}`
</script>

<style src="~/assets/css/guide-sections.css"></style>

<style scoped>
.support-wheelset-test-report {
  --wheelset-report-accent: var(--tz-site-accent);
  --wheelset-report-border: rgba(255, 255, 255, 0.08);
  --wheelset-report-soft-surface: rgba(255, 255, 255, 0.035);
}

.support-section__title {
  margin: 0 0 0.5rem;
  font-size: var(--tz-type-section-title);
  line-height: 1.35;
  font-weight: 600;
  color: var(--tz-text-primary);
}

.support-section__body {
  margin: 0;
  font-size: 0.9rem;
  line-height: 1.6;
  color: var(--tz-text-secondary);
}

.support-wheelset-test-report .text-emerald-700 {
  color: var(--wheelset-report-accent) !important;
}

.support-wheelset-test-report .text-emerald-600 {
  color: var(--wheelset-report-accent) !important;
}

.support-wheelset-test-report__assembly-note {
  border: 1px solid rgba(5, 150, 105, 0.28);
  background: rgba(5, 150, 105, 0.04);
  color: var(--tz-text-secondary);
}

.support-wheelset-test-report__assembly-note-action {
  color: var(--wheelset-report-accent);
  text-underline-offset: 0.2em;
}

.support-wheelset-test-report__assembly-note-action:hover {
  color: var(--tz-site-accent-hover);
}

.rim-test-card {
  padding: 1.25rem;
  border-radius: 0.75rem;
  background: var(--tz-card-surface);
  border: 1px solid var(--wheelset-report-border);
  box-shadow: 0 8px 22px rgba(0, 0, 0, 0.22);
}

.support-wheelset-test-report__method-badge {
  border: 1px solid rgba(5, 150, 105, 0.18);
  background: var(--wheelset-report-soft-surface);
}

.support-wheelset-test-report__disclaimer {
  border: 1px solid rgba(5, 150, 105, 0.22);
  background: rgba(5, 150, 105, 0.04);
  color: var(--tz-text-secondary);
}

.support-wheelset-test-report__disclaimer-label {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  border: 1px solid rgba(5, 150, 105, 0.34);
  border-radius: 9999px;
  padding: 0.22rem 0.7rem;
  background: rgba(0, 0, 0, 0.22);
  color: var(--wheelset-report-accent);
  font-size: var(--tz-type-micro-label);
  letter-spacing: 0.08em;
  line-height: 1.2;
  text-transform: uppercase;
}

.support-wheelset-test-report__disclaimer-label::before {
  width: 0.38rem;
  height: 0.38rem;
  border-radius: 9999px;
  background: var(--wheelset-report-accent);
  box-shadow: 0 0 10px rgba(5, 150, 105, 0.7);
  content: '';
}

.support-video-thumbnail {
  margin-top: 0.75rem;
  position: relative;
  border-radius: 0.75rem;
  overflow: hidden;
  cursor: pointer;
  box-shadow: 0 16px 32px rgba(0, 0, 0, 0.85);
}

.support-video-thumbnail__image {
  display: block;
  width: 100%;
  height: auto;
}

.support-video-thumbnail__overlay {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.25rem;
  background: linear-gradient(
    to top,
    rgba(0, 0, 0, 0.82),
    rgba(0, 0, 0, 0.38)
  );
  color: var(--tz-text-primary);
}

.support-video-thumbnail__icon {
  font-size: 1.8rem;
}

.support-video-thumbnail__label {
  font-size: 0.9rem;
  font-weight: 500;
}
</style>
