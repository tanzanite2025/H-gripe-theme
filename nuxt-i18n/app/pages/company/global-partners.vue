<template>
  <div class="w-full pb-6">
    <h1 class="sr-only">{{ t('company.nav.globalPartners') }}</h1>

    <div class="w-full max-w-none">
      <section class="text-center">
        <h2 class="mb-1 text-lg font-semibold tz-text-primary">
          {{ t('company.nav.globalPartners') }}
        </h2>
        <p class="mb-4 text-sm tz-text-secondary">
          Trusted by leading brands worldwide
        </p>
        <div class="mb-4 flex flex-wrap items-center justify-center gap-3">
          <NuxtLink
            to="/company/contact"
            class="premium-button premium-button--active justify-center"
          >
            Contact us about partnerships
          </NuxtLink>
        </div>
        <WorldMapXiamen />

        <!-- Data highlights cards -->
        <section class="mt-6 grid gap-4 sm:grid-cols-3">
          <!-- 30+ countries -->
          <div class="premium-card rounded-2xl p-4">
            <p class="text-xs font-medium uppercase tracking-wide tz-text-muted">
              Countries / regions
            </p>
            <p class="mt-2 text-3xl font-semibold tz-text-primary">
              30<span class="align-top text-lg">+</span>
            </p>
            <p class="mt-1 text-xs tz-text-secondary">
              Served worldwide from Xiamen.
            </p>
          </div>

          <!-- 100+ brand partners -->
          <div class="premium-card rounded-2xl p-4">
            <p class="text-xs font-medium uppercase tracking-wide tz-text-muted">
              Brand partners
            </p>
            <p class="mt-2 text-3xl font-semibold tz-text-primary">
              100<span class="align-top text-lg">+</span>
            </p>
            <p class="mt-1 text-xs tz-text-secondary">
              OEM / ODM collaborations globally.
            </p>
          </div>

          <!-- 500,000+ products delivered -->
          <div class="premium-card rounded-2xl p-4">
            <p class="text-xs font-medium uppercase tracking-wide tz-text-muted">
              Products delivered
            </p>
            <p class="mt-2 text-3xl font-semibold tz-text-primary">
              500,000<span class="align-top text-lg">+</span>
            </p>
            <p class="mt-1 text-xs tz-text-secondary">
              Wheel components shipped since launch.
            </p>
          </div>
        </section>

        <!-- Confidentiality notice for partner brands -->
        <div
          class="premium-card mt-4 rounded-2xl p-3 text-center text-sm tz-text-accent"
        >
          <p class="m-0">
            Out of respect for confidentiality agreements, we do not list partner brands publicly.
            All production projects are treated as strictly confidential to protect our customers'
            commercial interests and privacy.
          </p>
        </div>

        <!-- Value proposition: Why partner with us -->
        <section class="mt-8">
          <h3 class="mb-3 text-sm font-semibold tz-text-primary text-center">
            Why Partner With Us
          </h3>
          <div class="grid gap-4 sm:grid-cols-3">
            <!-- Long-term trust -->
            <div
              class="premium-card rounded-2xl p-4 text-center"
            >
              <div
                class="mb-3 inline-flex h-8 w-8 items-center justify-center rounded-full bg-black/30 text-xs font-semibold tz-text-accent"
              >
                01
              </div>
              <p class="text-sm font-semibold tz-text-primary">
                Long-term trust
              </p>
              <p class="mt-1 text-xs tz-text-secondary">
                Reliable delivery and consistent quality for your catalog.
              </p>
            </div>

            <!-- Technology leadership -->
            <div
              class="premium-card rounded-2xl p-4"
            >
              <div
                class="mb-3 inline-flex h-8 w-8 items-center justify-center rounded-full bg-black/30 text-xs font-semibold tz-text-accent"
              >
                02
              </div>
              <p class="text-sm font-semibold tz-text-primary">
                Technology leadership
              </p>
              <p class="mt-1 text-xs tz-text-secondary">
                Continuous R&amp;D in carbon rims, wheel builds and silent hubs.
              </p>
            </div>

            <!-- Global support -->
            <div
              class="premium-card rounded-2xl p-4"
            >
              <div
                class="mb-3 inline-flex h-8 w-8 items-center justify-center rounded-full bg-black/30 text-xs font-semibold tz-text-accent"
              >
                03
              </div>
              <p class="text-sm font-semibold tz-text-primary">
                Global support
              </p>
              <p class="mt-1 text-xs tz-text-secondary">
                Cross-regional service from Xiamen to your distribution markets.
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
        title="Share your feedback about our Partnership Program"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { definePageMeta, useHead, useI18n } from '#imports'
import WorldMapXiamen from '~/components/WorldMapXiamen.vue'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'

const { t } = useI18n()

definePageMeta({
  layout: 'products',
})

useHead({
  title: t('company.nav.globalPartners'),
})
</script>

