<template>
  <section class="bg-transparent text-white">
    <div class="mx-auto max-w-6xl px-4 pt-[21px] pb-0">
      <div class="mb-[9px]">
        <h2 class="text-xl font-semibold text-white/90 sm:text-2xl">{{ t('home.features.title') }}</h2>
        <div class="mt-[6px] h-[3px] w-48 rounded-full bg-[linear-gradient(90deg,transparent,#2dd4bf,#3b82f6,transparent)]"></div>
        <p class="mt-[3px] max-w-3xl text-sm text-white/70 sm:text-base">{{ t('home.features.subtitle') }}</p>
      </div>

      <div class="grid grid-cols-1 gap-[9px] sm:grid-cols-2 lg:grid-cols-4">
        <div
          v-for="item in items"
          :key="item.titleKey"
          class="rounded-2xl bg-white/5 p-6 shadow-[6px_6px_0_rgba(0,0,0,1)]"
        >
          <h3 class="text-base font-semibold text-white/90">{{ t(item.titleKey) }}</h3>
          <p class="mt-3 text-sm leading-relaxed text-white/70">{{ t(item.descriptionKey) }}</p>
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
      titleKey: 'home.features.items.qc.title',
      descriptionKey: 'home.features.items.qc.description',
    },
    {
      titleKey: 'home.features.items.options.title',
      descriptionKey: 'home.features.items.options.description',
    },
    {
      titleKey: 'home.features.items.oem.title',
      descriptionKey: 'home.features.items.oem.description',
    },
    {
      titleKey: 'home.features.items.lightweight.title',
      descriptionKey: 'home.features.items.lightweight.description',
    },
  ]
})
</script>
