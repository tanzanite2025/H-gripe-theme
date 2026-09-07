<template>
  <div class="company-page">
    <h1 class="sr-only">{{ t('terms.title') }}</h1>

    <div class="policies-content space-y-8 tz-text-secondary">
      <p class="text-sm tz-text-secondary">
        {{ t('terms.pageIntro') }}
      </p>

      <section>
        <h2 class="text-lg font-semibold tz-text-primary mb-3">
          {{ t('terms.sections.agreement.title') }}
        </h2>
        <p class="text-sm leading-relaxed">
          {{ t('terms.sections.agreement.body') }}
        </p>
      </section>

      <section>
        <h2 class="text-lg font-semibold tz-text-primary mb-3">
          {{ t('terms.sections.license.title') }}
        </h2>
        <p class="text-sm leading-relaxed mb-3">
          {{ t('terms.sections.license.body') }}
        </p>
        <ul class="list-disc list-inside text-sm space-y-2 tz-text-secondary">
          <li>{{ t('terms.sections.license.restrictions.0') }}</li>
          <li>{{ t('terms.sections.license.restrictions.1') }}</li>
          <li>{{ t('terms.sections.license.restrictions.2') }}</li>
        </ul>
      </section>

      <section>
        <h2 class="text-lg font-semibold tz-text-primary mb-3">
          {{ t('terms.sections.products.title') }}
        </h2>
        <p class="text-sm leading-relaxed">
          {{ t('terms.sections.products.body') }}
        </p>
      </section>

      <section>
        <h2 class="text-lg font-semibold tz-text-primary mb-3">
          {{ t('terms.sections.orders.title') }}
        </h2>
        <p class="text-sm leading-relaxed">
          {{ t('terms.sections.orders.body') }}
        </p>
      </section>

      <section>
        <h2 class="text-lg font-semibold tz-text-primary mb-3">
          {{ t('terms.sections.shipping.title') }}
        </h2>
        <p class="text-sm leading-relaxed">
          {{ t('terms.sections.shipping.body') }}
        </p>
      </section>

      <section>
        <h2 class="text-lg font-semibold tz-text-primary mb-3">
          {{ t('terms.sections.returns.title') }}
        </h2>
        <p class="text-sm leading-relaxed">
          {{ t('terms.sections.returns.prefix') }}
          <NuxtLink :to="localePath('/policies/refund-cancellation')" class="tz-text-accent underline">
            {{ t('terms.sections.returns.link') }}
          </NuxtLink>
          {{ t('terms.sections.returns.suffix') }}
        </p>
      </section>

      <section>
        <h2 class="text-lg font-semibold tz-text-primary mb-3">
          {{ t('terms.sections.warranty.title') }}
        </h2>
        <p class="text-sm leading-relaxed">
          {{ t('terms.sections.warranty.prefix') }}
          <NuxtLink :to="localePath('/support/after-sales')" class="tz-text-accent underline">
            {{ t('terms.sections.warranty.link') }}
          </NuxtLink>
          {{ t('terms.sections.warranty.suffix') }}
        </p>
      </section>

      <section>
        <h2 class="text-lg font-semibold tz-text-primary mb-3">
          {{ t('terms.sections.liability.title') }}
        </h2>
        <p class="text-sm leading-relaxed">
          {{ t('terms.sections.liability.body') }}
        </p>
      </section>

      <section>
        <h2 class="text-lg font-semibold tz-text-primary mb-3">
          {{ t('terms.sections.intellectualProperty.title') }}
        </h2>
        <p class="text-sm leading-relaxed">
          {{ t('terms.sections.intellectualProperty.body') }}
        </p>
      </section>

      <section>
        <h2 class="text-lg font-semibold tz-text-primary mb-3">
          {{ t('terms.sections.governingLaw.title') }}
        </h2>
        <p class="text-sm leading-relaxed">
          {{ t('terms.sections.governingLaw.body') }}
        </p>
      </section>

      <section>
        <h2 class="text-lg font-semibold tz-text-primary mb-3">
          {{ t('terms.sections.changes.title') }}
        </h2>
        <p class="text-sm leading-relaxed">
          {{ t('terms.sections.changes.body') }}
        </p>
      </section>

      <section>
        <h2 class="text-lg font-semibold tz-text-primary mb-3">
          {{ t('terms.sections.contact.title') }}
        </h2>
        <p class="text-sm leading-relaxed">
          {{ t('terms.sections.contact.prefix') }}
          <NuxtLink :to="localePath('/company/contact')" class="tz-text-accent underline">
            {{ t('terms.sections.contact.link') }}
          </NuxtLink>
          {{ t('terms.sections.contact.suffix') }}
        </p>
      </section>

      <div class="pt-6 border-t tz-border-subtle">
        <p class="text-xs tz-text-muted">{{ t('terms.lastUpdated') }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { definePageMeta, useHead, useI18n, useLocalePath } from '#imports'
import { usePageMessages } from '~/composables/usePageMessages'

definePageMeta({
  layout: 'products',
  footerLabelKey: 'policyTabs.terms',
  footerLabelFallback: 'Terms',
})

const { locale, t } = useI18n()
const localePath = useLocalePath()
const { loadPageMessages } = usePageMessages('terms')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

useHead(() => ({
  title: t('terms.title'),
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
