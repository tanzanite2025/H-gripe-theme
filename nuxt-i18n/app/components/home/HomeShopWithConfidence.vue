<template>
  <section class="bg-transparent text-white">
    <div class="mx-auto max-w-6xl px-1 sm:px-4 pt-[3px] pb-0">
      <div class="mb-[9px] text-center">
        <h2 class="text-lg font-semibold text-white/90 lg:text-xl">{{ t('home.trust.title') }}</h2>
        <div
          class="mt-[6px] mb-[9px] h-1 w-14 mx-auto rounded-full bg-gradient-to-r from-[#2dd4bf] to-[#3b82f6] shadow-[0_0_18px_rgba(45,212,191,0.25)]"
        ></div>
      </div>

      <div class="grid grid-cols-1 gap-[9px] sm:grid-cols-2 lg:grid-cols-4">
        <div
          v-for="card in cards"
          :key="card.titleKey"
          class="rounded-2xl bg-white/5 p-5 shadow-[6px_6px_0_rgba(0,0,0,1)]"
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
  return Array.from({ length: 4 }, (_, index) => ({
    titleKey: `home.shopWithConfidence.items.${index}.title`,
    descriptionKey: `home.shopWithConfidence.items.${index}.description`,
    bullets: [0, 1, 2].map((bulletIndex) => `home.shopWithConfidence.items.${index}.bullets.${bulletIndex}`),
  }))
})
</script>
