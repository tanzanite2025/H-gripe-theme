<template>
  <section class="bg-transparent text-white">
    <div class="mx-auto max-w-6xl px-4 pt-[21px] pb-0">
      <div class="mb-[9px]">
        <h2 class="text-xl font-semibold text-white/90 sm:text-2xl">{{ t('home.testimonials.title') }}</h2>
        <div class="mt-[6px] h-[3px] w-48 rounded-full bg-[linear-gradient(90deg,transparent,#2dd4bf,#3b82f6,transparent)]"></div>
        <p class="mt-[3px] max-w-2xl text-sm text-white/70 sm:text-base">{{ t('home.testimonials.subtitle') }}</p>
      </div>

      <div class="grid grid-cols-1 gap-[9px] sm:grid-cols-2 lg:grid-cols-3">
        <div
          v-for="item in items"
          :key="item.quoteKey"
          class="rounded-2xl bg-white/5 p-6 shadow-[6px_6px_0_rgba(0,0,0,1)]"
        >
          <p class="text-sm leading-relaxed text-white/80">{{ t(item.quoteKey) }}</p>
          <div class="mt-5">
            <div class="text-sm font-semibold text-white/90">{{ t(item.nameKey) }}</div>
            <div class="mt-1 text-xs text-white/60">{{ t(item.metaKey) }}</div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '#imports'

const { t } = useI18n()

const items = computed(() => {
  return [
    {
      quoteKey: 'home.testimonials.items.0.quote',
      nameKey: 'home.testimonials.items.0.name',
      metaKey: 'home.testimonials.items.0.meta',
    },
    {
      quoteKey: 'home.testimonials.items.1.quote',
      nameKey: 'home.testimonials.items.1.name',
      metaKey: 'home.testimonials.items.1.meta',
    },
    {
      quoteKey: 'home.testimonials.items.2.quote',
      nameKey: 'home.testimonials.items.2.name',
      metaKey: 'home.testimonials.items.2.meta',
    },
  ]
})
</script>
