<template>
  <div class="tension-section space-y-8">
    <div class="tension-panel tension-panel--intro p-5 md:p-6 text-center">
      <h3 class="tension-title mb-2">{{ t('testReportTension.title') }}</h3>
      <p class="tension-copy max-w-2xl mx-auto mb-6">
        {{ t('testReportTension.intro.before') }}
        <strong>{{ t('testReportTension.intro.riderWeight') }}</strong>{{ t('testReportTension.intro.betweenRiderAndUsage') }}
        <strong>{{ t('testReportTension.intro.usage') }}</strong>{{ t('testReportTension.intro.betweenUsageAndRim') }}
        <strong>{{ t('testReportTension.intro.rimMaterial') }}</strong>
        {{ t('testReportTension.intro.after') }}
      </p>
      <div class="flex justify-center flex-col items-center gap-4">
        <GuideImage
          src="/public/technical/tension/wheel-spoke-tension.webp"
          :alt="t('testReportTension.image.alt')"
          :zoomOnClick="true"
          :caption="t('testReportTension.image.caption')"
          class="tension-image rounded-xl max-w-lg w-full"
        />
        <NuxtLink
          :to="localePath('/support/test-report/wheelset-assembly')"
          class="premium-button tension-link"
        >
          {{ t('testReportTension.assemblyLink') }}
        </NuxtLink>
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <div class="tension-panel tension-panel--compact p-5 md:p-6">
        <h3 class="tension-title tension-title--section">
          {{ t('testReportTension.weight.title') }}
        </h3>
        <ul class="tension-list space-y-3 text-sm">
          <li
            v-for="index in 2"
            :key="`weight-${index}`"
            class="tension-list__item flex items-start gap-3"
          >
            <span class="tension-list__marker mt-1 shrink-0" />
            <span>
              <strong>{{ t(`testReportTension.weight.items.${index - 1}.label`) }}</strong>
              {{ t(`testReportTension.weight.items.${index - 1}.body`) }}
            </span>
          </li>
        </ul>
      </div>

      <div class="tension-panel tension-panel--compact p-5 md:p-6">
        <h3 class="tension-title tension-title--section">
          {{ t('testReportTension.usage.title') }}
        </h3>
        <div class="tension-options space-y-3 text-sm">
          <div
            v-for="index in 3"
            :key="`usage-${index}`"
            class="tension-option p-2 rounded-lg"
          >
            <strong class="tension-option__label block">
              {{ t(`testReportTension.usage.items.${index - 1}.label`) }}
            </strong>
            <span class="tension-option__copy text-xs">
              {{ t(`testReportTension.usage.items.${index - 1}.body`) }}
            </span>
          </div>
        </div>
      </div>
    </div>

    <div class="tension-panel tension-table-panel overflow-hidden">
      <div class="tension-table-header p-4 flex items-center justify-between">
        <h3 class="font-bold tz-text-primary">{{ t('testReportTension.table.title') }}</h3>
        <span class="tension-unit text-xs font-mono">{{ t('testReportTension.table.unit') }}</span>
      </div>
      <div class="overflow-x-auto">
        <table class="tension-table min-w-full text-left text-sm">
          <thead>
            <tr>
              <th>{{ t('testReportTension.table.headers.riderWeight') }}</th>
              <th>{{ t('testReportTension.table.headers.road') }}</th>
              <th>{{ t('testReportTension.table.headers.mtb') }}</th>
              <th>{{ t('testReportTension.table.headers.touring') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="index in 4" :key="`row-${index}`">
              <td class="tension-table__row-label">
                {{ t(`testReportTension.table.rows.${index - 1}.weight`) }}
              </td>
              <td>{{ t(`testReportTension.table.rows.${index - 1}.road`) }}</td>
              <td>{{ t(`testReportTension.table.rows.${index - 1}.mtb`) }}</td>
              <td>{{ t(`testReportTension.table.rows.${index - 1}.touring`) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <div class="tension-note text-xs">
        <strong>{{ t('testReportTension.table.noteLabel') }}</strong>
        {{ t('testReportTension.table.note') }}
      </div>
    </div>

    <div class="tension-panel tension-panel--process p-5 md:p-6">
      <h3 class="tension-title tension-title--section mb-6">
        {{ t('testReportTension.process.title') }}
      </h3>

      <div class="tension-timeline relative pl-6 space-y-8">
        <div
          v-for="index in 4"
          :key="`process-${index}`"
          class="tension-step relative"
          :class="`tension-step--${['one', 'two', 'three', 'four'][index - 1]}`"
        >
          <span class="tension-step__marker absolute">
            <span class="tension-step__dot block" />
          </span>
          <h4 class="text-sm font-bold tz-text-primary">
            {{ t(`testReportTension.process.steps.${index - 1}.title`) }}
          </h4>
          <p class="text-xs tz-text-secondary mt-1">
            {{ t(`testReportTension.process.steps.${index - 1}.body`) }}
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useI18n, useLocalePath } from '#imports'
import GuideImage from '~/components/GuideImage.vue'
import { usePageMessages } from '~/composables/usePageMessages'

const localePath = useLocalePath()
const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('testReportTension')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})
</script>

<style scoped>
.tension-section {
  --tension-accent: var(--tz-site-accent);
  --tension-accent-soft: rgba(5, 150, 105, 0.12);
  --tension-border: var(--tz-border-subtle);
  --tension-border-strong: var(--tz-border-strong);
  --tension-soft-surface: var(--tz-surface-subtle);
  color: var(--tz-text-secondary);
}

.tension-panel {
  border: 1px solid var(--tension-border);
  border-radius: 0.75rem;
  background: var(--tz-card-surface);
  box-shadow: 0 10px 24px rgba(0, 0, 0, 0.2);
}

.tension-panel--intro {
  border-top: 2px solid var(--tension-accent);
}

.tension-title {
  color: var(--tz-text-primary);
  font-size: var(--tz-type-card-title);
  font-weight: 700;
  line-height: 1.35;
}

.tension-title--section {
  display: flex;
  align-items: center;
  gap: 0.6rem;
}

.tension-panel--compact .tension-title--section {
  margin-bottom: 1rem;
}

.tension-panel--process .tension-title--section {
  margin-bottom: 1.5rem;
}

.tension-title--section::before {
  width: 0.4rem;
  height: 0.4rem;
  flex: 0 0 auto;
  border-radius: 9999px;
  background: var(--tension-accent);
  box-shadow: 0 0 0 4px var(--tension-accent-soft);
  content: '';
}

.tension-copy {
  color: var(--tz-text-secondary);
  font-size: var(--tz-type-description);
  line-height: 1.65;
}

.tension-copy strong,
.tension-list strong {
  color: var(--tz-text-primary);
  font-weight: 600;
}

.tension-image {
  border: 1px solid var(--tension-border-strong);
  box-shadow: 0 8px 22px rgba(0, 0, 0, 0.26);
}

.tension-link {
  border: 1px solid var(--tension-border) !important;
  background: var(--tension-soft-surface);
}

.tension-link:hover {
  border-color: var(--tension-accent-soft) !important;
  background: var(--tension-accent-soft);
  color: var(--tz-text-primary);
}

.tension-list__item {
  color: var(--tz-text-secondary);
}

.tension-list__marker {
  width: 0.4rem;
  height: 0.4rem;
  border-radius: 9999px;
  background: var(--tension-accent);
  box-shadow: 0 0 0 3px var(--tension-accent-soft);
}

.tension-option {
  border: 1px solid var(--tension-border);
  border-left: 2px solid var(--tension-accent);
  background: var(--tension-soft-surface);
}

.tension-option__label {
  color: var(--tz-text-primary);
  line-height: 1.35;
}

.tension-option__copy {
  color: var(--tz-text-muted);
  line-height: 1.55;
}

.tension-table-panel {
  overflow: hidden;
}

.tension-table-header {
  border-bottom: 1px solid var(--tension-border);
  background: var(--tension-soft-surface);
}

.tension-unit {
  border: 1px solid var(--tension-border);
  border-radius: 0.45rem;
  padding: 0.3rem 0.5rem;
  background: var(--tz-input-surface);
  color: var(--tz-text-muted);
}

.tension-table {
  border-collapse: collapse;
  color: var(--tz-text-secondary);
}

.tension-table th,
.tension-table td {
  padding: 0.8rem 1.25rem;
  border-bottom: 1px solid var(--tension-border);
}

.tension-table th {
  background: var(--tz-surface-subtle);
  color: var(--tz-text-secondary);
  font-weight: 600;
}

.tension-table tbody tr {
  transition: background-color 0.18s ease;
}

.tension-table tbody tr:hover {
  background: var(--tension-soft-surface);
}

.tension-table tbody tr:last-child td {
  border-bottom: 0;
}

.tension-table__row-label {
  color: var(--tz-text-primary);
  font-weight: 600;
}

.tension-note {
  border-top: 1px solid rgba(5, 150, 105, 0.16);
  background: rgba(5, 150, 105, 0.045);
  color: var(--tz-text-secondary);
  line-height: 1.6;
  padding: 1rem 1.25rem;
}

.tension-note strong {
  color: var(--tension-accent);
  font-weight: 600;
}

.tension-timeline {
  border-left: 1px solid rgba(5, 150, 105, 0.22);
}

.tension-step__marker {
  top: 0.05rem;
  left: -1.9rem;
  padding: 0.35rem;
  background: var(--tz-card-surface);
}

.tension-step__dot {
  width: 0.55rem;
  height: 0.55rem;
  border-radius: 9999px;
  background: var(--tension-accent);
  box-shadow: 0 0 0 3px var(--tension-accent-soft);
}

.tension-step--two .tension-step__dot {
  opacity: 0.82;
}

.tension-step--three .tension-step__dot {
  opacity: 0.64;
}

.tension-step--four .tension-step__dot {
  opacity: 0.46;
}

@media (max-width: 640px) {
  .tension-table th,
  .tension-table td {
    padding: 0.72rem 0.9rem;
  }

  .tension-table-header {
    align-items: flex-start;
    flex-direction: column;
    gap: 0.65rem;
  }
}
</style>
