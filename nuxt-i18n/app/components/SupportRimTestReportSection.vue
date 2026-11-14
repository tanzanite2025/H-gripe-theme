<template>
  <div class="support-rim-test-report rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-3 md:p-6">
    <component :is="headingTag" class="support-section__title text-center">{{ t('testReportRim.title') }}</component>

    <p class="support-section__body mt-4 text-center">
      {{ t('testReportRim.intro') }}
    </p>

    <component
      :is="testReportBlockRegistrationsById[testDefinition.sectionId].component"
      v-for="testDefinition in rimTestReportTestDefinitions"
      :key="testDefinition.id"
      :use-page-anchor="true"
      :heading-tag="headingTag === 'h1' ? 'h2' : 'h4'"
    >
      <template #header-actions>
        <TestReportDialogTriggerButton :test-id="testDefinition.sectionId" />
      </template>
    </component>

    <div class="support-rim-test-report__disclaimer mt-6 rounded-lg px-4 py-3 text-sm leading-relaxed">
      <component :is="headingTag === 'h1' ? 'h2' : 'h4'" class="support-rim-test-report__disclaimer-label mb-3 font-semibold">
        {{ t('testReportRim.disclaimer.title') }}
      </component>
      <p class="mb-2">{{ t('testReportRim.disclaimer.body.0') }}</p>
      <p>{{ t('testReportRim.disclaimer.body.1') }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useI18n } from '#imports'
import { usePageMessages } from '~/composables/usePageMessages'
import TestReportDialogTriggerButton from '~/components/technology/test-reports/TestReportDialogTriggerButton.vue'
import { testReportBlockRegistrationsById } from '~/utils/testReportBlockComponentRegistry'
import { rimTestReportTestDefinitions } from '~/data/rimTestReportTestDefinitionsAndMediaConfiguration'

const { headingTag = 'h3' } = defineProps<{
  headingTag?: 'h1' | 'h2' | 'h3'
}>()

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('testReportRim')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})
</script>

<style scoped>
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

.support-rim-test-report__disclaimer {
  border: 1px solid rgba(5, 150, 105, 0.22);
  background: rgba(5, 150, 105, 0.04);
  color: var(--tz-text-secondary);
}

.support-rim-test-report__disclaimer-label {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  border: 1px solid rgba(5, 150, 105, 0.34);
  border-radius: 9999px;
  padding: 0.22rem 0.7rem;
  background: rgba(0, 0, 0, 0.22);
  color: var(--tz-site-accent);
  font-size: var(--tz-type-micro-label);
  letter-spacing: 0.08em;
  line-height: 1.2;
  text-transform: uppercase;
}

.support-rim-test-report__disclaimer-label::before {
  width: 0.38rem;
  height: 0.38rem;
  border-radius: 9999px;
  background: var(--tz-site-accent);
  box-shadow: 0 0 10px rgba(5, 150, 105, 0.7);
  content: '';
}
</style>
