<template>
  <div class="w-full pb-6">
    <h1 class="sr-only">{{ t('companyGlobalPartners.title') }}</h1>

    <div class="w-full max-w-none">
      <section class="text-center">
        <h2 class="mb-1 text-lg font-semibold tz-text-primary">
          {{ t('companyGlobalPartners.title') }}
        </h2>
        <p class="mb-4 text-sm tz-text-secondary">
          {{ t('companyGlobalPartners.hero.tagline') }}
        </p>
        <div class="mb-4 flex flex-wrap items-center justify-center gap-3">
          <NuxtLink
            :to="localePath('/company/contact')"
            class="premium-button premium-button--active justify-center"
          >
            {{ t('companyGlobalPartners.hero.contactCta') }}
          </NuxtLink>
        </div>
        <WorldMapXiamen />

        <!-- Data highlights cards -->
        <section class="mt-6 grid gap-4 sm:grid-cols-3">
          <!-- 30+ countries -->
          <div class="premium-card rounded-2xl p-4">
            <p class="text-xs font-medium uppercase tracking-wide tz-text-muted">
              {{ t('companyGlobalPartners.stats.countries.label') }}
            </p>
            <p class="mt-2 text-3xl font-semibold tz-text-primary">
              30<span class="align-top text-lg">+</span>
            </p>
            <p class="mt-1 text-xs tz-text-secondary">
              {{ t('companyGlobalPartners.stats.countries.description') }}
            </p>
          </div>

          <!-- 100+ brand partners -->
          <div class="premium-card rounded-2xl p-4">
            <p class="text-xs font-medium uppercase tracking-wide tz-text-muted">
              {{ t('companyGlobalPartners.stats.brandPartners.label') }}
            </p>
            <p class="mt-2 text-3xl font-semibold tz-text-primary">
              100<span class="align-top text-lg">+</span>
            </p>
            <p class="mt-1 text-xs tz-text-secondary">
              {{ t('companyGlobalPartners.stats.brandPartners.description') }}
            </p>
          </div>

          <!-- 500,000+ products delivered -->
          <div class="premium-card rounded-2xl p-4">
            <p class="text-xs font-medium uppercase tracking-wide tz-text-muted">
              {{ t('companyGlobalPartners.stats.productsDelivered.label') }}
            </p>
            <p class="mt-2 text-3xl font-semibold tz-text-primary">
              500,000<span class="align-top text-lg">+</span>
            </p>
            <p class="mt-1 text-xs tz-text-secondary">
              {{ t('companyGlobalPartners.stats.productsDelivered.description') }}
            </p>
          </div>
        </section>

        <!-- Confidentiality notice for partner brands -->
        <div
          class="premium-card mt-4 rounded-2xl p-3 text-center text-sm tz-text-accent"
        >
          <p class="m-0">
            {{ t('companyGlobalPartners.confidentiality') }}
          </p>
        </div>

        <!-- Value proposition: Why partner with us -->
        <section class="mt-8">
          <h3 class="mb-3 text-sm font-semibold tz-text-primary text-center">
            {{ t('companyGlobalPartners.whyPartner.title') }}
          </h3>
          <div class="grid gap-4 sm:grid-cols-3">
            <!-- Long-term trust -->
            <div
              class="premium-card rounded-2xl p-4 text-center"
            >
              <div
                class="mb-3 inline-flex h-8 w-8 items-center justify-center rounded-full tz-surface-muted text-xs font-semibold tz-text-accent"
              >
                01
              </div>
              <p class="text-sm font-semibold tz-text-primary">
                {{ t('companyGlobalPartners.whyPartner.trust.title') }}
              </p>
              <p class="mt-1 text-xs tz-text-secondary">
                {{ t('companyGlobalPartners.whyPartner.trust.description') }}
              </p>
            </div>

            <!-- Technology leadership -->
            <div
              class="premium-card rounded-2xl p-4"
            >
              <div
                class="mb-3 inline-flex h-8 w-8 items-center justify-center rounded-full tz-surface-muted text-xs font-semibold tz-text-accent"
              >
                02
              </div>
              <p class="text-sm font-semibold tz-text-primary">
                {{ t('companyGlobalPartners.whyPartner.technology.title') }}
              </p>
              <p class="mt-1 text-xs tz-text-secondary">
                {{ t('companyGlobalPartners.whyPartner.technology.description') }}
              </p>
            </div>

            <!-- Global support -->
            <div
              class="premium-card rounded-2xl p-4"
            >
              <div
                class="mb-3 inline-flex h-8 w-8 items-center justify-center rounded-full tz-surface-muted text-xs font-semibold tz-text-accent"
              >
                03
              </div>
              <p class="text-sm font-semibold tz-text-primary">
                {{ t('companyGlobalPartners.whyPartner.support.title') }}
              </p>
              <p class="mt-1 text-xs tz-text-secondary">
                {{ t('companyGlobalPartners.whyPartner.support.description') }}
              </p>
            </div>
          </div>
        </section>
      </section>
    </div>

    <!-- Feedback Section -->
    <div class="w-full max-w-none px-0 mt-4">
      <UserFeedbackThread
        threadKey="company-global-partners"
        :title="t('companyGlobalPartners.feedbackTitle')"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { definePageMeta, useHead, useI18n, useLocalePath } from '#imports'
import WorldMapXiamen from '~/components/WorldMapXiamen.vue'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'
import { usePageMessages } from '~/composables/usePageMessages'

const { locale, t } = useI18n()
const localePath = useLocalePath()
const { loadPageMessages } = usePageMessages('companyGlobalPartners')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

definePageMeta({
  layout: 'products',
  footerLabelKey: 'company.nav.globalPartners',
  footerLabelFallback: 'Global Partners',
})

useHead(() => ({
  title: t('companyGlobalPartners.title'),
}))
</script>

