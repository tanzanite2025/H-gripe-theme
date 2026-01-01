<template>
  <section class="bg-transparent text-white">
    <div class="mx-auto max-w-6xl px-1 sm:px-4 pt-[21px] pb-0">
      <div class="mb-[9px] relative">
        <div class="mx-auto w-full max-w-3xl text-center">
          <h2 class="text-xl font-semibold text-white/90 sm:text-2xl">{{ t('home.featuredProducts.title') }}</h2>
          <div class="mt-[6px] h-1 w-14 mx-auto rounded-full bg-gradient-to-r from-[#2dd4bf] to-[#3b82f6] shadow-[0_0_18px_rgba(45,212,191,0.25)]"></div>
          <p class="mt-[3px] mx-auto max-w-2xl text-sm text-white/70 sm:text-base">{{ t('home.featuredProducts.subtitle') }}</p>

          <NuxtLink
            to="/shop"
            class="mt-[6px] inline-flex items-center text-sm font-semibold text-white/80 transition-colors hover:text-white"
            aria-label="Shop all products"
          >
            {{ t('home.featuredProducts.viewAll') }}
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" class="ml-2 h-5 w-5" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 6l6 6-6 6" />
            </svg>
          </NuxtLink>
        </div>
      </div>

      <div class="grid grid-cols-1 gap-[9px] sm:grid-cols-2 lg:grid-cols-4">
        <NuxtLink
          v-for="card in cards"
          :key="card.titleKey"
          to="/shop"
          class="group block overflow-hidden rounded-2xl premium-card"
        >
          <div class="relative aspect-[4/3]">
            <div class="absolute inset-0 bg-[radial-gradient(circle_at_top_left,rgba(255,255,255,0.08),transparent_60%)]"></div>
            <div class="absolute inset-0 bg-gradient-to-t from-black/70 via-black/30 to-transparent" aria-hidden="true"></div>
          </div>

          <div class="p-5">
            <h3 class="text-base font-semibold text-white/90">{{ t(card.titleKey) }}</h3>
            <p class="mt-2 text-sm leading-relaxed text-white/70">{{ t(card.descriptionKey) }}</p>
            <div class="mt-4 text-sm font-semibold text-white/90">{{ t(card.priceKey) }}</div>
          </div>
        </NuxtLink>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '#imports'

const { t } = useI18n()

const cards = computed(() => {
  return [
    {
      titleKey: 'home.featuredProducts.items.0.title',
      descriptionKey: 'home.featuredProducts.items.0.description',
      priceKey: 'home.featuredProducts.items.0.price',
    },
    {
      titleKey: 'home.featuredProducts.items.1.title',
      descriptionKey: 'home.featuredProducts.items.1.description',
      priceKey: 'home.featuredProducts.items.1.price',
    },
    {
      titleKey: 'home.featuredProducts.items.2.title',
      descriptionKey: 'home.featuredProducts.items.2.description',
      priceKey: 'home.featuredProducts.items.2.price',
    },
    {
      titleKey: 'home.featuredProducts.items.3.title',
      descriptionKey: 'home.featuredProducts.items.3.description',
      priceKey: 'home.featuredProducts.items.3.price',
    },
  ]
})
</script>
