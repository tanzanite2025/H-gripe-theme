<template>
  <div class="tz-mobile-text-inset">
    <h2 class="support-page__title">{{ t('supportShipping.title') }}</h2>
    <p class="support-page__intro">
      {{ t('supportShipping.intro') }}
    </p>

    <section class="support-section">
      <h3 class="support-section__title">{{ t('supportShipping.sections.overview.title') }}</h3>
      <p class="support-section__body">
        {{ t('supportShipping.sections.overview.body') }}
      </p>
    </section>

    <section class="support-section">
      <h3 class="support-section__title">{{ t('supportShipping.sections.processing.title') }}</h3>
      <p class="support-section__body">
        {{ t('supportShipping.sections.processing.body') }}
      </p>
    </section>

    <section class="support-section">
      <h3 class="support-section__title">{{ t('supportShipping.sections.international.title') }}</h3>
      <p class="support-section__body">
        {{ t('supportShipping.sections.international.body') }}
      </p>
    </section>

    <section class="support-section">
      <h3 class="support-section__title">{{ t('supportShipping.sections.comingSoon.title') }}</h3>
      <p class="support-section__body">
        {{ t('supportShipping.sections.comingSoon.body') }}
      </p>
    </section>

  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { definePageMeta, useHead, useI18n } from '#imports'
import { usePageMessages } from '~/composables/usePageMessages'

definePageMeta({
  layout: 'support',
  footerLabelKey: 'support.nav.shipping',
  footerLabelFallback: 'Shipping',
})

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('supportShipping')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

useHead(() => ({
  title: t('supportShipping.title'),
}))
</script>

<style scoped>
.support-page__title {
  margin: 0 0 0.75rem;
  font-size: var(--tz-type-page-title);
  line-height: 1.18;
  font-weight: 600;
  color: var(--tz-text-primary);
}

.support-page__intro {
  margin: 0 0 1.25rem;
  font-size: 0.95rem;
  color: var(--tz-text-secondary);
}

.support-section {
  margin-top: 1.75rem;
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
</style>
