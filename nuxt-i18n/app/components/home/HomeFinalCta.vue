<template>
  <section class="bg-transparent text-white">
    <div class="page-content-shell px-0 md:px-4 pt-[21px] pb-[15px]">
      <div class="rounded-3xl bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.9),rgba(15,23,42,0.95))] p-[9px] border-none shadow-[0_4px_16px_rgba(0,0,0,0.5)]">
        <h2 class="text-lg font-semibold text-white/90 sm:text-xl text-center">{{ t('home.finalCta.title') }}</h2>
        <div class="mt-[6px] h-1 w-14 mx-auto rounded-full bg-gradient-to-r from-[#2dd4bf] to-[#3b82f6] shadow-[0_0_18px_rgba(45,212,191,0.25)]"></div>
        <p class="mt-[3px] mx-auto max-w-3xl text-sm text-white/70 sm:text-base text-center">{{ t('home.finalCta.subtitle') }}</p>

        <div class="mt-[9px] flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-center">
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
