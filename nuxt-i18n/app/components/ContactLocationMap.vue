<template>
  <div class="w-full">
    <div :class="containerClass">
      <div class="flex flex-col gap-4">
        <div class="flex flex-col">
          <component
            :is="titleTag"
            class="font-bold text-white leading-tight"
            :class="variant === 'compact' ? 'text-2xl sm:text-3xl' : 'text-lg'"
          >
            {{ t('contactLocation.title') }}
          </component>
          <p
            v-if="variant !== 'compact'"
            class="mt-[3px] tz-text-secondary leading-relaxed break-words"
          >
            {{ contactLocation.addressText }}
          </p>
        </div>

        <div class="flex w-full max-w-[220px] flex-col gap-2">
          <a
            :href="contactLocation.openGoogleDirectionsUrl || contactLocation.openGoogleMapsUrl"
            target="_blank"
            rel="noopener"
            class="flex min-h-10 w-full items-center justify-center rounded-full bg-white px-4 py-2 text-sm font-semibold text-slate-950 transition-colors hover:bg-white/90"
          >
            {{ t('contactLocation.getDirections') }}
          </a>
          <a
            :href="contactLocation.openGoogleMapsUrl"
            target="_blank"
            rel="noopener"
            class="flex min-h-10 w-full items-center justify-center rounded-full border border-white/10 bg-white/5 px-4 py-2 text-sm font-semibold text-white/85 transition-colors hover:bg-white/10 hover:text-white"
          >
            {{ t('contactLocation.openGoogle') }}
          </a>
          <a
            v-if="contactLocation.openAppleMapsUrl"
            :href="contactLocation.openAppleMapsUrl"
            target="_blank"
            rel="noopener"
            class="flex min-h-10 w-full items-center justify-center rounded-full border border-white/10 bg-white/5 px-4 py-2 text-sm font-semibold text-white/85 transition-colors hover:bg-white/10 hover:text-white"
          >
            {{ t('contactLocation.openApple') }}
          </a>
        </div>
      </div>

      <div class="relative overflow-hidden rounded-2xl border border-white/10 bg-black shadow-[0_4px_14px_-10px_rgba(0,0,0,0.9)]">
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
          <div v-else class="w-full bg-black" :class="mapHeightClass"></div>

          <div class="absolute inset-0 bg-black/20"></div>
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
    return 'grid grid-cols-1 gap-5 lg:grid-cols-[minmax(0,240px)_minmax(0,720px)] lg:items-center lg:justify-center lg:gap-8'
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
