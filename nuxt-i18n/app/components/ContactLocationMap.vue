<template>
  <div class="w-full">
    <div :class="containerClass">
      <div class="flex flex-col gap-4">
        <div class="flex flex-col">
          <component
            :is="titleTag"
            class="text-white/90 font-semibold"
            :class="variant === 'compact' ? 'text-base' : 'text-lg'"
          >
            {{ t('contactLocation.title') }}
          </component>
          <div class="mt-[6px] h-1 w-14 rounded-full bg-gradient-to-r from-[#2dd4bf] to-[#3b82f6] shadow-[0_0_18px_rgba(45,212,191,0.25)]"></div>
          <p class="mt-[3px] text-white/70 leading-relaxed break-words" :class="variant === 'compact' ? 'text-sm' : 'text-base'">
            {{ contactLocation.addressText }}
          </p>
        </div>

        <div class="flex flex-wrap gap-2">
          <a
            :href="contactLocation.openGoogleDirectionsUrl || contactLocation.openGoogleMapsUrl"
            target="_blank"
            rel="noopener"
            class="px-4 py-2 rounded-full text-sm font-semibold bg-white text-slate-950 hover:bg-white/90 transition-all"
          >
            {{ t('contactLocation.getDirections') }}
          </a>
          <a
            :href="contactLocation.openGoogleMapsUrl"
            target="_blank"
            rel="noopener"
            class="px-4 py-2 rounded-full text-sm font-semibold bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] text-white/90 shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(0,0,0,0.7)] hover:bg-[linear-gradient(135deg,rgba(31,41,55,0.98),rgba(15,23,42,0.98))] hover:shadow-[0_4px_12px_-4px_rgba(0,0,0,0.95),0_0_8px_rgba(0,0,0,0.9)] transition-all"
          >
            {{ t('contactLocation.openGoogle') }}
          </a>
          <a
            v-if="contactLocation.openAppleMapsUrl"
            :href="contactLocation.openAppleMapsUrl"
            target="_blank"
            rel="noopener"
            class="px-4 py-2 rounded-full text-sm font-semibold bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] text-white/90 shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(0,0,0,0.7)] hover:bg-[linear-gradient(135deg,rgba(31,41,55,0.98),rgba(15,23,42,0.98))] hover:shadow-[0_4px_12px_-4px_rgba(0,0,0,0.95),0_0_8px_rgba(0,0,0,0.9)] transition-all"
          >
            {{ t('contactLocation.openApple') }}
          </a>
        </div>
      </div>

      <div class="relative overflow-hidden rounded-2xl border border-white/10 bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.96),rgba(15,23,42,0.98))] shadow-[0_4px_14px_-10px_rgba(0,0,0,0.9),0_0_12px_rgba(15,23,42,0.85)]">
        <button
          v-if="!showInteractive"
          type="button"
          class="relative w-full text-left"
          @click="handleOpenInteractive"
        >
          <img
            v-if="contactLocation.previewImageSrc && !previewImageFailed"
            :src="contactLocation.previewImageSrc"
            :alt="t('contactLocation.previewAlt')"
            class="w-full object-cover"
            :class="mapHeightClass"
            loading="lazy"
            decoding="async"
            @error="previewImageFailed = true"
          />
          <div v-else class="w-full" :class="[mapHeightClass, 'bg-[radial-gradient(circle_at_top,rgba(56,189,248,0.18),transparent_60%)]']"></div>

          <div class="absolute inset-0 bg-black/25"></div>
          <div class="absolute inset-0 flex items-center justify-center px-4">
            <span class="px-5 py-2.5 rounded-full text-sm font-semibold bg-black/60 text-white backdrop-blur-sm border border-white/15">
              {{ t('contactLocation.loadMap') }}
            </span>
          </div>
        </button>

        <ClientOnly>
          <iframe
            v-if="showInteractive"
            :src="contactLocation.googleEmbedUrl"
            class="w-full border-0"
            :class="mapHeightClass"
            allowfullscreen
            loading="lazy"
            referrerpolicy="no-referrer-when-downgrade"
            :title="t('contactLocation.iframeTitle')"
          ></iframe>
        </ClientOnly>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from '#imports'
import { contactLocation } from '~/utils/contactLocation'

const props = withDefaults(defineProps<{ variant?: 'default' | 'compact'; titleTag?: 'h2' | 'h3'; layout?: 'stack' | 'split' }>(), {
  variant: 'default',
  titleTag: 'h2',
  layout: 'stack',
})

const { t } = useI18n()

const showInteractive = ref(false)
const previewImageFailed = ref(false)

const mapHeightClass = computed(() => {
  return props.variant === 'compact'
    ? 'h-[220px] sm:h-[260px]'
    : 'h-[280px] sm:h-[340px] lg:h-[380px]'
})

const containerClass = computed(() => {
  if (props.layout === 'split') {
    return 'grid grid-cols-1 gap-4 lg:grid-cols-[minmax(0,360px)_minmax(0,1fr)] lg:items-start'
  }
  return 'flex flex-col gap-4'
})

const handleOpenInteractive = () => {
  if (contactLocation.googleEmbedUrl) {
    showInteractive.value = true
    return
  }

  if (typeof window !== 'undefined') {
    window.open(String(contactLocation.openGoogleMapsUrl), '_blank', 'noopener')
  }
}
</script>
