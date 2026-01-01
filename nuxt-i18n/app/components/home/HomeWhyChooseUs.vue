<template>
  <section class="bg-transparent text-white">
    <div class="mx-auto max-w-6xl px-1 sm:px-4 pt-[21px] pb-0">
      <div class="mb-[9px]">
        <div class="mx-auto w-full max-w-3xl text-center">
          <h2 class="text-xl font-semibold text-white/90 sm:text-2xl">{{ t('home.whyChooseUs.title') }}</h2>
          <div class="mt-[6px] h-1 w-14 mx-auto rounded-full bg-gradient-to-r from-[#2dd4bf] to-[#3b82f6] shadow-[0_0_18px_rgba(45,212,191,0.25)]"></div>
          <p class="mt-[3px] mx-auto max-w-2xl text-sm text-white/70 sm:text-base">{{ t('home.whyChooseUs.subtitle') }}</p>
        </div>
      </div>

      <div class="grid grid-cols-1 gap-[9px] sm:grid-cols-2 lg:grid-cols-3">
        <div
          v-for="card in cards"
          :key="card.titleKey"
          class="rounded-2xl premium-card p-5"
        >
          <h3 class="text-base font-semibold text-white/95">{{ t(card.titleKey) }}</h3>
          <p class="mt-2 text-sm leading-relaxed text-white/70">{{ t(card.descriptionKey) }}</p>

          <ul class="mt-4 space-y-2 text-sm text-white/80">
            <li v-for="bulletKey in card.bullets" :key="bulletKey" class="flex gap-2">
              <span class="text-teal-300">•</span>
              <span>{{ t(bulletKey) }}</span>
            </li>
          </ul>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '#imports'

const { t } = useI18n()

const cards = computed(() => {
  return Array.from({ length: 6 }, (_, index) => ({
    titleKey: `home.whyChooseUs.items.${index}.title`,
    descriptionKey: `home.whyChooseUs.items.${index}.description`,
    bullets: [0, 1, 2].map((bulletIndex) => `home.whyChooseUs.items.${index}.bullets.${bulletIndex}`),
  }))
})
</script>
