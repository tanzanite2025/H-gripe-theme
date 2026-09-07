<template>
  <div class="company-page">
    <h1 class="sr-only">{{ t('cookie.title') }}</h1>

    <div class="policies-content">
      <p class="text-sm tz-text-secondary mb-6">
        {{ t('cookie.pageIntro') }}
      </p>

      <div class="mt-4 space-y-8 tz-text-secondary">
        <section>
          <h2 class="text-lg font-semibold tz-text-primary mb-3">
            {{ t('cookie.sections.whatAreCookies.title') }}
          </h2>
          <p class="text-sm leading-relaxed">
            {{ t('cookie.sections.whatAreCookies.body') }}
          </p>
        </section>

        <section>
          <h2 class="text-lg font-semibold tz-text-primary mb-3">
            {{ t('cookie.sections.types.title') }}
          </h2>
          <div class="space-y-4">
            <div class="premium-card rounded-xl p-4">
              <h3 class="text-base font-medium text-[#059669] mb-2">
                {{ t('cookie.sections.types.essential.title') }}
              </h3>
              <p class="text-sm leading-relaxed mb-2">
                {{ t('cookie.sections.types.essential.body') }}
              </p>
              <p class="text-xs tz-text-muted">
                {{ t('cookie.sections.types.essential.status') }}
              </p>
            </div>

            <div class="premium-card rounded-xl p-4">
              <h3 class="text-base font-medium text-[#059669] mb-2">
                {{ t('cookie.sections.types.performance.title') }}
              </h3>
              <p class="text-sm leading-relaxed mb-2">
                {{ t('cookie.sections.types.performance.body') }}
              </p>
              <p class="text-xs tz-text-muted">
                {{ t('cookie.sections.types.performance.status') }}
              </p>
            </div>

            <div class="premium-card rounded-xl p-4">
              <h3 class="text-base font-medium text-[#059669] mb-2">
                {{ t('cookie.sections.types.functional.title') }}
              </h3>
              <p class="text-sm leading-relaxed mb-2">
                {{ t('cookie.sections.types.functional.body') }}
              </p>
              <p class="text-xs tz-text-muted">
                {{ t('cookie.sections.types.functional.status') }}
              </p>
            </div>

            <div class="premium-card rounded-xl p-4">
              <h3 class="text-base font-medium text-[#059669] mb-2">
                {{ t('cookie.sections.types.advertising.title') }}
              </h3>
              <p class="text-sm leading-relaxed mb-2">
                {{ t('cookie.sections.types.advertising.body') }}
              </p>
              <p class="text-xs tz-text-muted">
                {{ t('cookie.sections.types.advertising.status') }}
              </p>
            </div>
          </div>
        </section>

        <section>
          <h2 class="text-lg font-semibold tz-text-primary mb-3">
            {{ t('cookie.sections.preferences.title') }}
          </h2>
          <p class="text-sm leading-relaxed mb-4">
            {{ t('cookie.sections.preferences.intro') }}
          </p>
          <ul class="list-disc list-inside text-sm space-y-2 tz-text-secondary">
            <li>{{ t('cookie.sections.preferences.options.0') }}</li>
            <li>{{ t('cookie.sections.preferences.options.1') }}</li>
            <li>{{ t('cookie.sections.preferences.options.2') }}</li>
            <li>{{ t('cookie.sections.preferences.options.3') }}</li>
          </ul>
          <p class="text-sm leading-relaxed mt-4 tz-text-secondary">
            {{ t('cookie.sections.preferences.note') }}
          </p>
        </section>

        <section>
          <h2 class="text-lg font-semibold tz-text-primary mb-3">
            {{ t('cookie.sections.thirdParty.title') }}
          </h2>
          <p class="text-sm leading-relaxed">
            {{ t('cookie.sections.thirdParty.body') }}
          </p>
        </section>

        <section>
          <h2 class="text-lg font-semibold tz-text-primary mb-3">
            {{ t('cookie.sections.updates.title') }}
          </h2>
          <p class="text-sm leading-relaxed">
            {{ t('cookie.sections.updates.body') }}
          </p>
        </section>

        <div class="pt-6 border-t tz-border-subtle">
          <p class="text-xs tz-text-muted">{{ t('cookie.lastUpdated') }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useHead, definePageMeta, useI18n } from '#imports'
import { usePageMessages } from '~/composables/usePageMessages'

definePageMeta({
  layout: 'products',
  footerLabelKey: 'policyTabs.cookie',
  footerLabelFallback: 'Cookie',
})

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('cookie')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

useHead(() => ({
  title: t('cookie.title'),
}))
</script>

<style scoped>
.company-page {
  width: 100%;
  max-width: none;
  margin: 0 auto;
  padding: 0 16px;
  display: flex;
  flex-direction: column;
  gap: 0;
}

.policies-content {
  margin-top: 0;
}

@media (max-width: 767px) {
  .company-page {
    max-width: 100%;
    padding-inline: 0;
  }
}
</style>
