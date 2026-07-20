<template>
  <div class="w-full pt-4 md:pt-8 pb-6">
    <h1 class="sr-only">{{ t('company.nav.globalPartners') }}</h1>

    <div class="w-full max-w-none">
      <section class="rounded-2xl bg-slate-950/60 p-4 md:p-6 text-center shadow-[3px_3px_10px_rgba(0,0,0,0.9)]">
        <h2 class="mb-1 text-lg font-semibold text-slate-100">
          {{ t('company.nav.globalPartners') }}
        </h2>
        <p class="mb-4 text-sm text-slate-400">
          Trusted by leading brands worldwide
        </p>
        <div class="mb-4 flex flex-wrap items-center justify-center gap-3">
          <NuxtLink
            to="/company/contact"
            class="inline-flex items-center justify-center rounded-full bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] px-4 py-2 text-sm font-semibold text-slate-950 shadow-[0_4px_12px_rgba(0,0,0,0.8)] transition hover:brightness-110"
          >
            Contact us about partnerships
          </NuxtLink>
          <span class="flex items-center gap-2 text-xs text-slate-400">
            <span>or email</span>
            <a
              href="mailto:support@tanzanite.site"
              class="inline-flex items-center justify-center rounded-full bg-slate-800/90 px-3 py-1.5 text-[11px] font-medium text-slate-100 shadow-[2px_2px_6px_rgba(0,0,0,0.85)] transition hover:bg-slate-700/95 hover:text-slate-50"
            >
              support@tanzanite.site
            </a>
          </span>
        </div>
        <WorldMapXiamen />

        <!-- Data highlights cards -->
        <section class="mt-6 grid gap-4 sm:grid-cols-3">
          <!-- 30+ countries -->
          <div
            class="rounded-2xl bg-slate-900/70 p-4 shadow-[3px_3px_10px_rgba(0,0,0,0.9)]"
          >
            <p class="text-xs font-medium uppercase tracking-wide text-slate-400">
              Countries / regions
            </p>
            <p class="mt-2 text-3xl font-semibold text-slate-50">
              30<span class="align-top text-lg">+</span>
            </p>
            <p class="mt-1 text-xs text-slate-400">
              Served worldwide from Xiamen.
            </p>
          </div>

          <!-- 100+ brand partners -->
          <div
            class="rounded-2xl bg-slate-900/70 p-4 shadow-[3px_3px_10px_rgba(0,0,0,0.9)]"
          >
            <p class="text-xs font-medium uppercase tracking-wide text-slate-400">
              Brand partners
            </p>
            <p class="mt-2 text-3xl font-semibold text-slate-50">
              100<span class="align-top text-lg">+</span>
            </p>
            <p class="mt-1 text-xs text-slate-400">
              OEM / ODM collaborations globally.
            </p>
          </div>

          <!-- 500,000+ products delivered -->
          <div
            class="rounded-2xl bg-slate-900/70 p-4 shadow-[3px_3px_10px_rgba(0,0,0,0.9)]"
          >
            <p class="text-xs font-medium uppercase tracking-wide text-slate-400">
              Products delivered
            </p>
            <p class="mt-2 text-3xl font-semibold text-slate-50">
              500,000<span class="align-top text-lg">+</span>
            </p>
            <p class="mt-1 text-xs text-slate-400">
              Wheel components shipped since launch.
            </p>
          </div>
        </section>

        <!-- Confidentiality notice for partner brands -->
        <div
          class="mt-4 rounded-2xl bg-slate-900/80 p-3 text-sm text-sky-300 text-center shadow-[3px_3px_10px_rgba(0,0,0,0.9)]"
        >
          <p class="m-0">
            Out of respect for confidentiality agreements, we do not list partner brands publicly.
            All production projects are treated as strictly confidential to protect our customers'
            commercial interests and privacy.
          </p>
        </div>

        <!-- Value proposition: Why partner with us -->
        <section class="mt-8">
          <h3 class="mb-3 text-sm font-semibold text-slate-100 text-center">
            Why Partner With Us
          </h3>
          <div class="grid gap-4 sm:grid-cols-3">
            <!-- Long-term trust -->
            <div
              class="rounded-2xl bg-slate-900/70 p-4 text-center shadow-[3px_3px_10px_rgba(0,0,0,0.9)]"
            >
              <div
                class="mb-3 inline-flex h-8 w-8 items-center justify-center rounded-full bg-slate-800 text-xs font-semibold text-sky-300"
              >
                01
              </div>
              <p class="text-sm font-semibold text-slate-100">
                Long-term trust
              </p>
              <p class="mt-1 text-xs text-slate-400">
                Reliable delivery and consistent quality for your catalog.
              </p>
            </div>

            <!-- Technology leadership -->
            <div
              class="rounded-2xl bg-slate-900/70 p-4 shadow-[3px_3px_10px_rgba(0,0,0,0.9)]"
            >
              <div
                class="mb-3 inline-flex h-8 w-8 items-center justify-center rounded-full bg-slate-800 text-xs font-semibold text-sky-300"
              >
                02
              </div>
              <p class="text-sm font-semibold text-slate-100">
                Technology leadership
              </p>
              <p class="mt-1 text-xs text-slate-400">
                Continuous R&amp;D in carbon rims, wheel builds and silent hubs.
              </p>
            </div>

            <!-- Global support -->
            <div
              class="rounded-2xl bg-slate-900/70 p-4 shadow-[3px_3px_10px_rgba(0,0,0,0.9)]"
            >
              <div
                class="mb-3 inline-flex h-8 w-8 items-center justify-center rounded-full bg-slate-800 text-xs font-semibold text-sky-300"
              >
                03
              </div>
              <p class="text-sm font-semibold text-slate-100">
                Global support
              </p>
              <p class="mt-1 text-xs text-slate-400">
                Cross-regional service from Xiamen to your distribution markets.
              </p>
            </div>
          </div>
        </section>
      </section>
    </div>

    <!-- FAQ Section -->
    <div class="w-full max-w-none mt-4 px-0">
      <PageFaq pageId="company-global-partners" />
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
import PageFaq from '~/components/PageFaq.vue'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'

const { t } = useI18n()

definePageMeta({
  layout: 'products',
})

useHead({
  title: t('company.nav.globalPartners'),
})
</script>
