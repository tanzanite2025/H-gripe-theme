<template>
  <section class="bg-transparent text-white pt-[var(--site-header-offset,140px)] md:pt-[var(--site-header-offset,96px)]">
    <div class="mx-auto max-w-6xl px-1 sm:px-4 pb-6 pt-2 lg:pb-8 lg:pt-3">
      <div class="flex flex-col items-center text-center">
        <!-- Mobile: Media above CTAs -->
        <h1 class="mt-4 px-2 text-xl font-semibold leading-[1.05] tracking-tight sm:text-3xl lg:text-4xl">
          {{ t('home.hero.title') }}
        </h1>
        <p class="mt-2 max-w-2xl px-2 text-sm leading-relaxed text-white/80 sm:text-lg">
          {{ t('home.hero.subtitle') }}
        </p>

        <div class="mt-2 w-full">
          <div class="relative mx-auto w-full max-w-5xl">
            <div
              class="relative rounded-3xl pt-[72px] sm:pt-[92px]"
              :aria-label="t('home.hero.stackAriaLabel')"
              tabindex="0"
              @keydown.left.prevent="prev"
              @keydown.right.prevent="next"
            >
              <div class="absolute inset-x-0 top-[14px] z-40 flex justify-center gap-2 px-4">
                <button
                  v-for="(_, index) in cards"
                  :key="index"
                  type="button"
                  class="h-1.5 rounded-full transition-[width,background-color] duration-200"
                  :class="index === activeIndex ? 'w-7 bg-white/90' : 'w-1.5 bg-white/25 hover:bg-white/40'"
                  :aria-label="t('home.hero.dotAriaLabel', { index: index + 1 })"
                  @click="goTo(index)"
                ></button>
              </div>

              <div class="relative aspect-[16/10] overflow-visible sm:aspect-[16/9] md:aspect-[21/9]">
                <ul class="absolute inset-0 overflow-visible" aria-live="polite">
                  <li
                    v-for="(card, index) in cards"
                    :key="card.src"
                    class="absolute inset-0 origin-center will-change-transform transition-[transform,filter,opacity] duration-[260ms] ease-[cubic-bezier(0.2,0,0.2,1)]"
                    :class="cardClass(index)"
                  >
                    <img
                      :src="card.src"
                      :alt="t(card.altKey)"
                      class="h-full w-full object-cover"
                      loading="eager"
                      decoding="async"
                    />
                    <div
                      class="absolute inset-0 rounded-3xl bg-gradient-to-t from-black/55 via-black/10 to-transparent"
                      aria-hidden="true"
                    ></div>
                  </li>
                </ul>

                <button
                  type="button"
                  class="absolute left-[12px] top-1/2 z-40 inline-flex h-[46px] w-[46px] -translate-y-1/2 items-center justify-center rounded-full bg-black/70 text-white backdrop-blur transition-colors hover:bg-black/80"
                  :aria-label="t('common.previous')"
                  @click="prev"
                >
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
                    <path d="M15 19L8 12L15 5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
                  </svg>
                </button>
                <button
                  type="button"
                  class="absolute right-[12px] top-1/2 z-40 inline-flex h-[46px] w-[46px] -translate-y-1/2 items-center justify-center rounded-full bg-black/70 text-white backdrop-blur transition-colors hover:bg-black/80"
                  :aria-label="t('common.next')"
                  @click="next"
                >
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
                    <path d="M9 5L16 12L9 19" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
                  </svg>
                </button>
              </div>
            </div>
          </div>
        </div>

        <div class="mt-2 w-full">
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
              :to="localePath('/picture-warehouse')"
              class="inline-flex w-full items-center justify-center rounded-full bg-white/5 px-3 py-2 text-[13px] font-semibold text-white/90 shadow-[4px_4px_0_rgba(0,0,0,1)] transition-colors hover:bg-white/10 sm:w-auto sm:px-6 sm:py-3 sm:text-sm"
            >
              {{ t('home.hero.cta.pictureWarehouse') }}
            </NuxtLink>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, useI18n, useLocalePath } from '#imports'

const { t } = useI18n()
const localePath = useLocalePath()

const cards = computed(() => [
  {
    src: '/company/ourstory/ourstory/tanzanite-ourstory.webp',
    altKey: 'home.hero.cards.0.alt'
  },
  {
    src: '/company/ourstory/factory/tanzanite-factory-premoldlayupworkshop6.webp',
    altKey: 'home.hero.cards.1.alt'
  },
  {
    src: '/company/ourstory/factory/tanzanite-factory-inspectionpacking18.webp',
    altKey: 'home.hero.cards.2.alt'
  }
])

const activeIndex = ref(0)

const next = () => {
  activeIndex.value = (activeIndex.value + 1) % cards.value.length
}

const prev = () => {
  activeIndex.value = (activeIndex.value - 1 + cards.value.length) % cards.value.length
}

const goTo = (index: number) => {
  activeIndex.value = index
}

const relativeSlot = (index: number) => {
  return (index - activeIndex.value + cards.value.length) % cards.value.length
}

const cardClass = (index: number) => {
  const slot = relativeSlot(index)

  if (slot === 0) {
    return 'z-30 opacity-100 translate-y-0 scale-100 brightness-100 rounded-3xl overflow-hidden border-2 border-white/20 bg-slate-900/50 shadow-[0_25px_50px_rgba(0,0,0,0.7)]'
  }

  if (slot === 1) {
    return 'z-20 opacity-100 -translate-y-[10%] scale-[0.94] brightness-[0.85] rounded-3xl overflow-hidden border-2 border-white/20 bg-slate-900/50 shadow-[0_15px_30px_rgba(0,0,0,0.4)]'
  }

  return 'z-10 opacity-100 -translate-y-[20%] scale-[0.88] brightness-[0.7] rounded-3xl overflow-hidden border-2 border-white/20 bg-slate-900/50 shadow-[0_15px_30px_rgba(0,0,0,0.4)]'
}
</script>
