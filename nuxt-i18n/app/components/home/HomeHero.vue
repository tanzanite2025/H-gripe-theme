<template>
  <section class="bg-transparent text-white pt-[var(--site-header-offset,140px)] md:pt-[var(--site-header-offset,96px)]">
    <div class="mx-auto max-w-6xl px-1 sm:px-4 pb-6 pt-2 lg:pb-8 lg:pt-3">
      <div class="flex flex-col">
        <!-- Mobile: Media above CTAs -->
        <div
          class="mt-4 relative aspect-[3/2] overflow-hidden rounded-3xl bg-[radial-gradient(circle_at_top_left,rgba(255,255,255,0.08),transparent_60%)] shadow-[6px_6px_0_rgba(0,0,0,1)] sm:aspect-[16/9] md:aspect-[21/9]"
          aria-label="Hero collage placeholder"
        >
          <div
            class="absolute inset-0 bg-[linear-gradient(135deg,rgba(148,163,184,0.22),rgba(15,23,42,0.92))]"
            aria-hidden="true"
          ></div>
          <div class="absolute inset-0 bg-[radial-gradient(circle_at_top,rgba(56,189,248,0.18),transparent_60%)]" aria-hidden="true"></div>
          <div class="absolute inset-0 bg-gradient-to-t from-black/55 via-black/10 to-transparent" aria-hidden="true"></div>

          <div class="relative flex h-full flex-col items-center justify-center px-4 py-6 text-center transform -translate-y-2 sm:translate-y-0 sm:px-6 sm:py-8">
            <h1 class="text-xl font-semibold leading-[1.05] tracking-tight sm:text-3xl lg:text-4xl">
              {{ t('home.hero.title') }}
            </h1>
            <div class="mt-[6px] h-1 w-14 rounded-full bg-gradient-to-r from-[#2dd4bf] to-[#3b82f6] shadow-[0_0_18px_rgba(45,212,191,0.25)]"></div>
            <p class="mt-[3px] w-full max-w-xl text-base leading-relaxed text-white/80 sm:text-lg">
              {{ t('home.hero.subtitle') }}
            </p>

            <div class="mt-4 w-full">
              <div class="grid w-full grid-cols-3 gap-2 sm:flex sm:w-auto sm:flex-row sm:items-center sm:justify-center">
                <NuxtLink
                  :to="localePath('/shop')"
                  class="inline-flex w-full items-center justify-center rounded-full bg-white px-3 py-2 text-[13px] font-semibold text-neutral-950 shadow-[4px_4px_0_rgba(0,0,0,1)] transition-colors hover:bg-white/90 sm:w-auto sm:px-6 sm:py-3 sm:text-sm"
                >
                  {{ t('home.hero.cta.shop') }}
                </NuxtLink>

                <NuxtLink
                  :to="localePath('/company/about')"
                  class="inline-flex w-full items-center justify-center rounded-full bg-white/5 px-3 py-2 text-[13px] font-semibold text-white/90 shadow-[4px_4px_0_rgba(0,0,0,1)] transition-colors hover:bg-white/10 sm:w-auto sm:px-6 sm:py-3 sm:text-sm"
                >
                  {{ t('home.hero.cta.about') }}
                </NuxtLink>

                <NuxtLink
                  :to="localePath('/company/picture-warehouse')"
                  class="inline-flex w-full items-center justify-center rounded-full bg-white/5 px-3 py-2 text-[13px] font-semibold text-white/90 shadow-[4px_4px_0_rgba(0,0,0,1)] transition-colors hover:bg-white/10 sm:w-auto sm:px-6 sm:py-3 sm:text-sm"
                >
                  {{ t('home.hero.cta.pictureWarehouse') }}
                </NuxtLink>
              </div>
            </div>
          </div>
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
