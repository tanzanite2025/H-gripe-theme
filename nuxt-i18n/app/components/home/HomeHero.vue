<template>
  <section class="bg-neutral-950 text-white pt-[var(--site-header-offset,140px)] md:pt-[var(--site-header-offset,96px)]">
    <div class="mx-auto grid max-w-6xl grid-cols-1 gap-5 px-4 pb-6 pt-2 lg:grid-cols-12 lg:gap-7 lg:pb-8 lg:pt-3">
      <div class="flex flex-col lg:col-span-5">
        <h1 class="text-2xl font-semibold leading-[1.05] tracking-tight sm:text-3xl lg:text-4xl">
          {{ t('home.hero.title') }}
        </h1>
        <p class="mt-3 w-full text-base leading-relaxed text-white/70 sm:text-lg lg:max-w-xl">
          {{ t('home.hero.subtitle') }}
        </p>

        <!-- Mobile: Media above CTAs -->
        <div class="mt-4 grid grid-cols-2 gap-2 lg:hidden">
          <div
            class="relative aspect-[4/3] overflow-hidden rounded-3xl bg-[radial-gradient(circle_at_top_left,rgba(255,255,255,0.08),transparent_60%)] shadow-[6px_6px_0_rgba(0,0,0,1)]"
            aria-label="Hero image placeholder 1"
          >
            <div
              class="absolute inset-0 bg-[linear-gradient(135deg,rgba(148,163,184,0.28),rgba(15,23,42,0.92))]"
              aria-hidden="true"
            ></div>
            <div class="absolute inset-0 bg-gradient-to-t from-black/35 via-transparent to-transparent" aria-hidden="true"></div>
          </div>

          <div
            class="relative aspect-[4/3] overflow-hidden rounded-3xl bg-[radial-gradient(circle_at_top_left,rgba(255,255,255,0.08),transparent_60%)] shadow-[6px_6px_0_rgba(0,0,0,1)]"
            aria-label="Hero image placeholder 2"
          >
            <div
              class="absolute inset-0 bg-[linear-gradient(135deg,rgba(56,189,248,0.18),rgba(15,23,42,0.92))]"
              aria-hidden="true"
            ></div>
            <div class="absolute inset-0 bg-gradient-to-t from-black/35 via-transparent to-transparent" aria-hidden="true"></div>
          </div>
        </div>

        <div class="mt-4 w-full">
          <div class="grid w-full grid-cols-3 gap-2 sm:flex sm:w-auto sm:flex-row sm:items-center">
            <NuxtLink
              to="/shop"
              class="inline-flex w-full items-center justify-center rounded-full bg-white px-4 py-2.5 text-sm font-semibold text-neutral-950 shadow-[4px_4px_0_rgba(0,0,0,1)] transition-colors hover:bg-white/90 sm:w-auto sm:px-6 sm:py-3"
            >
              {{ t('home.hero.cta.shop') }}
            </NuxtLink>

            <NuxtLink
              to="/company/about"
              class="inline-flex w-full items-center justify-center rounded-full bg-white/5 px-4 py-2.5 text-sm font-semibold text-white/90 shadow-[4px_4px_0_rgba(0,0,0,1)] transition-colors hover:bg-white/10 sm:w-auto sm:px-6 sm:py-3"
            >
              {{ t('home.hero.cta.about') }}
            </NuxtLink>

            <NuxtLink
              to="/picture-warehouse"
              class="inline-flex w-full items-center justify-center rounded-full bg-white/5 px-4 py-2.5 text-sm font-semibold text-white/90 shadow-[4px_4px_0_rgba(0,0,0,1)] transition-colors hover:bg-white/10 sm:hidden"
            >
              {{ t('home.hero.cta.pictureWarehouse') }}
            </NuxtLink>
          </div>

          <NuxtLink
            to="/picture-warehouse"
            class="mt-3 hidden items-center text-sm font-semibold text-white/80 transition-colors hover:text-white sm:inline-flex"
          >
            {{ t('home.hero.cta.pictureWarehouse') }}
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" class="ml-2 h-5 w-5" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 6l6 6-6 6" />
            </svg>
          </NuxtLink>
        </div>
      </div>

      <div class="lg:col-span-7">
        <div class="hidden grid-cols-2 gap-2 lg:grid lg:h-full">
          <div
            class="relative aspect-[2/1] overflow-hidden rounded-3xl bg-[radial-gradient(circle_at_top_left,rgba(255,255,255,0.08),transparent_60%)] shadow-[6px_6px_0_rgba(0,0,0,1)] lg:aspect-auto lg:h-full"
            aria-label="Hero image placeholder 1"
          >
            <div
              class="absolute inset-0 bg-[linear-gradient(135deg,rgba(148,163,184,0.28),rgba(15,23,42,0.92))]"
              aria-hidden="true"
            ></div>
            <div class="absolute inset-0 bg-gradient-to-t from-black/35 via-transparent to-transparent" aria-hidden="true"></div>
          </div>

          <div
            class="relative aspect-[2/1] overflow-hidden rounded-3xl bg-[radial-gradient(circle_at_top_left,rgba(255,255,255,0.08),transparent_60%)] shadow-[6px_6px_0_rgba(0,0,0,1)] lg:aspect-auto lg:h-full"
            aria-label="Hero image placeholder 2"
          >
            <div
              class="absolute inset-0 bg-[linear-gradient(135deg,rgba(56,189,248,0.18),rgba(15,23,42,0.92))]"
              aria-hidden="true"
            ></div>
            <div class="absolute inset-0 bg-gradient-to-t from-black/35 via-transparent to-transparent" aria-hidden="true"></div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from '#imports'

const { t } = useI18n()
</script>
