<template>
  <div class="w-full">
    <div :class="containerClass">
      <div class="flex flex-col gap-4">
        <div class="flex flex-col">
          <component
            :is="titleTag"
            class="font-bold tz-text-primary leading-tight"
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
            :href="googleDirectionsUrl || googleMapsUrl"
            v-if="googleDirectionsUrl || googleMapsUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="flex min-h-10 w-full items-center justify-center rounded-full bg-white px-4 py-2 text-sm font-semibold text-slate-950 transition-colors hover:bg-white/90"
          >
            {{ t('contactLocation.getDirections') }}
          </a>
          <a
            v-if="googleMapsUrl"
            :href="googleMapsUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="flex min-h-10 w-full items-center justify-center rounded-full border tz-border-subtle tz-surface-subtle px-4 py-2 text-sm font-semibold tz-text-primary/85 transition-colors hover:tz-surface-subtle hover:tz-text-primary"
          >
            {{ t('contactLocation.openGoogle') }}
          </a>
          <a
            v-if="appleMapsUrl"
            :href="appleMapsUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="flex min-h-10 w-full items-center justify-center rounded-full border tz-border-subtle tz-surface-subtle px-4 py-2 text-sm font-semibold tz-text-primary/85 transition-colors hover:tz-surface-subtle hover:tz-text-primary"
          >
            {{ t('contactLocation.openApple') }}
          </a>
        </div>
      </div>

      <div class="relative overflow-hidden rounded-2xl border tz-border-subtle tz-surface-subtle shadow-[0_4px_14px_-10px_rgba(20,32,43,0.12)]">
        <button
          v-if="!showInteractive"
          type="button"
          class="relative w-full text-left"
          @click="handleOpenInteractive"
        >
          <StorefrontImage
            v-if="contactLocation.previewImageSrc && !previewImageFailed"
            :src="contactLocation.previewImageSrc"
            :alt="t('contactLocation.previewAlt')"
            class="w-full object-cover"
            :class="mapHeightClass"
            preset="content"
            @error="previewImageFailed = true"
          />
          <div v-else class="w-full tz-surface-muted" :class="mapHeightClass"></div>

          <div class="absolute inset-0 bg-slate-900/10"></div>
          <div class="absolute inset-0 flex items-center justify-center px-4">
            <span class="px-5 py-2.5 rounded-full text-sm font-semibold tz-surface-card tz-text-primary backdrop-blur-sm border tz-border-subtle">
              {{ t('contactLocation.loadMap') }}
            </span>
          </div>
        </button>

        <ClientOnly>
          <iframe
            v-if="showInteractive && googleEmbedUrl"
            :src="googleEmbedUrl"
            class="w-full border-0"
            :class="mapHeightClass"
            allowfullscreen
            loading="lazy"
            referrerpolicy="strict-origin-when-cross-origin"
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
import {
  contactLocation,
  isAllowedMapEmbedUrl,
  isAllowedMapLinkUrl,
} from '~/utils/contactLocation'

const props = withDefaults(defineProps<{ variant?: 'default' | 'compact'; titleTag?: 'h2' | 'h3'; layout?: 'stack' | 'split' }>(), {
  variant: 'default',
  titleTag: 'h2',
  layout: 'stack',
})

const { t } = useI18n()

const showInteractive = ref(false)
const previewImageFailed = ref(false)

const googleEmbedUrl = computed(() => (
  isAllowedMapEmbedUrl(contactLocation.googleEmbedUrl) ? contactLocation.googleEmbedUrl : ''
))

const googleMapsUrl = computed(() => (
  isAllowedMapLinkUrl(contactLocation.openGoogleMapsUrl) ? contactLocation.openGoogleMapsUrl : ''
))

const googleDirectionsUrl = computed(() => (
  isAllowedMapLinkUrl(contactLocation.openGoogleDirectionsUrl) ? contactLocation.openGoogleDirectionsUrl : ''
))

const appleMapsUrl = computed(() => (
  isAllowedMapLinkUrl(contactLocation.openAppleMapsUrl) ? contactLocation.openAppleMapsUrl : ''
))

const mapHeightClass = computed(() => {
  return props.variant === 'compact'
    ? 'h-[220px] sm:h-[260px]'
    : 'h-[280px] sm:h-[340px] lg:h-[380px]'
})

const containerClass = computed(() => {
  if (props.layout === 'split') {
    return 'grid grid-cols-1 gap-5 lg:grid-cols-[minmax(15rem,3fr)_minmax(0,7fr)] lg:items-center lg:gap-8'
  }
  return 'flex flex-col gap-4'
})

const handleOpenInteractive = () => {
  if (googleEmbedUrl.value) {
    showInteractive.value = true
    return
  }

  if (typeof window !== 'undefined' && googleMapsUrl.value) {
    window.open(googleMapsUrl.value, '_blank', 'noopener,noreferrer')
  }
}
</script>
