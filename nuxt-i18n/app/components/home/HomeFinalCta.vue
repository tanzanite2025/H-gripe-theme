<template>
  <section class="bg-transparent text-white">
    <div class="mx-auto max-w-6xl px-1 sm:px-4 pt-[21px] pb-[15px]">
      <div class="rounded-3xl bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.9),rgba(15,23,42,0.95))] p-[9px] shadow-[6px_6px_0_rgba(0,0,0,1)]">
        <h2 class="text-lg font-semibold text-white/90 sm:text-xl text-center">{{ t('home.finalCta.title') }}</h2>
        <div class="mt-[6px] h-1 w-14 mx-auto rounded-full bg-gradient-to-r from-[#2dd4bf] to-[#3b82f6] shadow-[0_0_18px_rgba(45,212,191,0.25)]"></div>
        <p class="mt-[3px] mx-auto max-w-3xl text-sm text-white/70 sm:text-base text-center">{{ t('home.finalCta.subtitle') }}</p>

        <div class="mt-[9px] flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-center">
          <NuxtLink
            :to="localePath('/shop')"
            class="inline-flex items-center justify-center rounded-full bg-white px-6 py-3 text-sm font-semibold text-neutral-950 shadow-[4px_4px_0_rgba(0,0,0,1)] transition-colors hover:bg-white/90"
          >
            {{ t('home.finalCta.primary') }}
          </NuxtLink>

          <NuxtLink
            :to="localePath('/picture-warehouse')"
            class="inline-flex items-center justify-center rounded-full bg-white/5 px-6 py-3 text-sm font-semibold text-white/90 shadow-[4px_4px_0_rgba(0,0,0,1)] transition-colors hover:bg-white/10"
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
