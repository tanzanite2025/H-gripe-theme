<template>
  <section class="bg-transparent text-white">
    <div class="page-content-shell px-0 md:px-4 pt-6 pb-6 md:pt-8 md:pb-8">
      <div class="flex min-h-[240px] flex-col items-center justify-center rounded-[32px] bg-[var(--tz-card-surface)] px-6 py-12 text-center shadow-[0_18px_50px_rgba(0,0,0,0.2)] md:min-h-[280px] md:px-16 md:py-14">
        <div class="mb-5 inline-flex items-center gap-3 rounded-full bg-[#1b241a] px-5 py-3 text-sm font-medium text-white">
          <span class="h-2 w-2 rounded-full bg-[#B5FF6D] shadow-[0_0_8px_rgba(181,255,109,0.55)]"></span>
          <span>Available for work</span>
        </div>
        <h2 class="max-w-5xl text-2xl font-semibold leading-[1.08] text-white sm:text-3xl">{{ t('home.finalCta.title') }}</h2>
        <p class="mt-3 mx-auto max-w-3xl text-sm tz-text-secondary sm:text-base">{{ t('home.finalCta.subtitle') }}</p>

        <div class="mt-6 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-center">
          <NuxtLink
            :to="localePath('/shop')"
            class="premium-button premium-button--active w-full sm:w-auto justify-center"
          >
            {{ t('home.finalCta.primary') }}
          </NuxtLink>

          <NuxtLink
            :to="localePath('/picture-warehouse')"
            class="premium-button w-full sm:w-auto justify-center"
          >
            {{ t('home.finalCta.secondary') }}
          </NuxtLink>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n, useLocalePath } from '#imports'

const { t } = useI18n()
const localePath = useLocalePath()
</script>
