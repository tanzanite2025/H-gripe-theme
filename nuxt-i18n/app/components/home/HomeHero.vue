<template>
  <section class="bg-transparent text-white pt-[140px] md:pt-[96px]">
    <div class="page-content-shell px-0 md:px-4 pb-6 pt-2 lg:pb-8 lg:pt-3">
      <div class="flex flex-col items-center text-center">
        <!-- Mobile: Media above CTAs -->
        <div class="mt-3 flex items-center justify-center gap-2.5 text-base font-medium leading-tight text-white sm:gap-3 sm:text-lg">
          <span class="welcome-hand-wave inline-flex shrink-0" aria-hidden="true">
            <Icon
              name="lucide:hand"
              class="h-7 w-7 text-[#B5FF6D] sm:h-8 sm:w-8"
            />
          </span>
          <span>{{ welcomeText }}</span>
        </div>
        <h1 class="mt-6 px-2 text-xl font-semibold leading-[1.05] tracking-tight sm:mt-7 sm:text-3xl lg:text-4xl">
          <template v-if="heroTitle.accent">
            {{ heroTitle.before }}<span class="text-[#B5FF6D]">{{ heroTitle.accent }}</span>{{ heroTitle.after }}
          </template>
          <template v-else>
            {{ heroTitle.full }}
          </template>
        </h1>
        <p class="mt-2 max-w-2xl px-2 text-sm leading-relaxed tz-text-secondary sm:text-lg">
          {{ t('home.hero.subtitle') }}
        </p>

        <div class="mt-2 w-full">
          <div class="relative mx-auto w-full">
            <div
              class="relative rounded-3xl pt-[72px] sm:pt-[92px]"
              :aria-label="t('home.hero.stackAriaLabel')"
              tabindex="0"
              @keydown.left.prevent="prev"
              @keydown.right.prevent="next"
            >
              <div class="tz-carousel-pagination home-hero-stack-pagination absolute inset-x-0 top-[14px] z-40 px-4">
                <button
                  v-for="(_, index) in cards"
                  :key="index"
                  type="button"
                  class="tz-carousel-pagination__dot"
                  :class="{ 'is-active': index === activeIndex }"
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
                    <NuxtImg
                      :src="card.src"
                      :alt="t(card.altKey)"
                      :width="card.width"
                      :height="card.height"
                      class="h-full w-full object-cover"
                      sizes="xs:100vw sm:100vw md:100vw lg:100vw xl:1280px"
                      densities="1x"
                      format="webp"
                      quality="84"
                      :loading="index === 0 ? 'eager' : 'lazy'"
                      :fetchpriority="index === 0 ? 'high' : 'low'"
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
                  class="tz-directional-arrow tz-directional-arrow--large absolute left-[12px] top-1/2 z-40 -translate-y-1/2"
                  :aria-label="t('common.previous')"
                  @click="prev"
                >
                  <Icon name="lucide:chevron-left" aria-hidden="true" />
                </button>
                <button
                  type="button"
                  class="tz-directional-arrow tz-directional-arrow--large absolute right-[12px] top-1/2 z-40 -translate-y-1/2"
                  :aria-label="t('common.next')"
                  @click="next"
                >
                  <Icon name="lucide:chevron-right" aria-hidden="true" />
                </button>
              </div>
            </div>
          </div>
        </div>

      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, useI18n } from '#imports'

const { t } = useI18n()

const welcomeText = computed(() => {
  const translated = t('home.hero.welcome')
  return translated === 'home.hero.welcome' ? 'WELCOME!' : translated
})

const heroTitle = computed(() => {
  const full = t('home.hero.title')
  const accent = 'Factory-Direct'
  const accentIndex = full.indexOf(accent)

  if (accentIndex === -1) {
    return { full, before: '', accent: '', after: '' }
  }

  return {
    full,
    before: full.slice(0, accentIndex),
    accent,
    after: full.slice(accentIndex + accent.length),
  }
})

const cards = computed(() => [
  {
    src: '/company/ourstory/ourstory/ourstory.webp',
    altKey: 'home.hero.cards.0.alt',
    width: 800,
    height: 600,
  },
  {
    src: '/company/ourstory/factory/factory-premoldlayupworkshop6.webp',
    altKey: 'home.hero.cards.1.alt',
    width: 800,
    height: 600,
  },
  {
    src: '/company/ourstory/factory/factory-inspectionpacking18.webp',
    altKey: 'home.hero.cards.2.alt',
    width: 800,
    height: 500,
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

<style scoped>
@keyframes welcome-hand-wave {
  0%,
  72%,
  100% {
    transform: rotate(0deg);
  }

  76% {
    transform: rotate(12deg);
  }

  80% {
    transform: rotate(-8deg);
  }

  84% {
    transform: rotate(10deg);
  }

  88% {
    transform: rotate(-4deg);
  }

  92% {
    transform: rotate(0deg);
  }
}

.welcome-hand-wave {
  transform-origin: 70% 90%;
  animation: welcome-hand-wave 4.5s ease-in-out infinite;
}

.home-hero-stack-pagination {
  --tz-carousel-pagination-dot-bg: rgba(255, 255, 255, 0.25);
  --tz-carousel-pagination-dot-hover-bg: rgba(255, 255, 255, 0.4);
  --tz-carousel-pagination-dot-active-bg: rgba(255, 255, 255, 0.9);
}

@media (prefers-reduced-motion: reduce) {
  .welcome-hand-wave {
    animation: none;
  }
}
</style>
