<template>
  <section class="mt-4 space-y-3">
    <div>
      <h2 class="mb-1 text-sm font-semibold tz-text-primary">
        {{ t('resourcesPictureWarehouseRiders.title') }}
      </h2>
      <p class="mb-3 text-xs tz-text-secondary">
        {{ t('resourcesPictureWarehouseRiders.intro') }}
      </p>

      <div class="min-h-[60px]">
        <div
          v-if="props.loading"
          class="grid gap-3 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4"
        >
          <div
            v-for="n in 3"
            :key="n"
            class="flex flex-col overflow-hidden rounded-xl border tz-border-subtle bg-[var(--tz-card-surface)] shadow-md"
          >
            <div class="aspect-square w-full tz-surface-panel animate-pulse"></div>
            <div class="flex flex-col gap-1 px-2.5 py-2">
              <div class="h-2.5 w-3/4 rounded bg-slate-200 animate-pulse"></div>
              <div class="h-2 w-1/2 rounded tz-surface-subtle animate-pulse"></div>
            </div>
          </div>
        </div>

        <template v-else>
          <p v-if="!props.photos.length" class="text-xs tz-text-secondary">
            {{ t('resourcesPictureWarehouseRiders.empty') }}
          </p>

          <div v-if="props.photos.length" class="space-y-2">
            <div class="grid gap-3 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
              <button
                v-for="(photo, index) in props.visiblePhotos"
                :key="photo.id"
                type="button"
                class="group flex flex-col overflow-hidden rounded-xl border tz-border-subtle bg-[var(--tz-card-surface)] shadow-md transition-colors hover:tz-border-subtle hover:tz-surface-subtle"
                @click="emit('open', index)"
              >
                <div class="aspect-square w-full overflow-hidden tz-surface-panel transition-colors group-hover:tz-surface-muted">
                  <StorefrontImage
                    v-if="photo.galleryImages?.[0]"
                    :src="photo.galleryImages[0]"
                    :alt="photo.title"
                    class="size-full object-cover transition duration-200 group-hover:scale-[1.03]"
                    preset="card"
                  />
                </div>
                <div class="flex flex-col gap-0.5 px-2.5 py-2">
                  <p class="tz-caption truncate font-medium tz-text-primary">
                    {{ photo.title }}
                  </p>
                  <p class="tz-micro-label truncate tz-text-muted">
                    {{ photo.region }}<span v-if="photo.nickname"> · {{ photo.nickname }}</span>
                  </p>
                </div>
              </button>
            </div>

            <div v-if="props.hasMore" class="pt-1">
              <button
                type="button"
                class="tz-micro-label text-emerald-600 underline underline-offset-2 hover:text-emerald-700"
                @click="emit('toggle')"
              >
                {{ props.showAll
                  ? t('resourcesPictureWarehouseRiders.showFewer')
                  : t('resourcesPictureWarehouseRiders.showMore') }}
              </button>
            </div>
          </div>
        </template>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useI18n } from '#imports'
import { usePageMessages } from '~/composables/usePageMessages'
import type { RiderPhoto } from '~/types/pictureWarehouse'

const props = defineProps<{
  photos: RiderPhoto[]
  visiblePhotos: RiderPhoto[]
  loading: boolean
  showAll: boolean
  hasMore: boolean
}>()

const emit = defineEmits<{
  open: [index: number]
  toggle: []
}>()

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('resourcesPictureWarehouseRiders')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})
</script>

