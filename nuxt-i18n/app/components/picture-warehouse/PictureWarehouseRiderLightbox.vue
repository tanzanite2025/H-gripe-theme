<template>
  <teleport to="body">
    <transition
      enter-active-class="transition-opacity duration-200 ease-out"
      leave-active-class="transition-opacity duration-150 ease-in"
      enter-from-class="opacity-0"
      leave-to-class="opacity-0"
    >
      <div
        v-if="props.open"
        class="tz-standard-modal-mask tz-standard-modal-mask--compact fixed inset-0 z-[1400] flex items-center justify-center px-3 tz-mobile-safe-modal-mask"
        @click.self="emit('close')"
      >
        <div
          class="picture-lightbox-panel tz-standard-modal-surface relative flex max-h-[90vh] w-full max-w-[960px] flex-col overflow-hidden tz-surface-panel md:overflow-y-auto"
        >
          <header class="flex items-center justify-between border-b tz-border-subtle px-4 py-3 tz-surface-panel">
            <h2 class="truncate text-sm font-semibold tz-text-primary sm:text-base">
              {{ props.activePhoto?.title || t('resourcesPictureWarehouseShared.lightbox.picture') }}
            </h2>
            <button
              type="button"
              class="tz-global-close-btn ml-4"
              :aria-label="t('resourcesPictureWarehouseShared.lightbox.close')"
              @click="emit('close')"
            >
              ×
            </button>
          </header>

          <div class="relative flex flex-1 items-center justify-center overflow-hidden bg-[var(--tz-image-loading-surface)] p-4">
            <button
              type="button"
              class="tz-directional-arrow tz-directional-arrow--large absolute left-2 z-20 sm:left-4"
              :aria-label="t('resourcesPictureWarehouseShared.lightbox.previousPhoto')"
              @click.stop="emit('previous')"
            >
              <Icon name="lucide:chevron-left" aria-hidden="true" />
            </button>

            <div class="relative flex h-full w-full flex-col items-center justify-center">
              <div class="relative flex h-full w-full items-center justify-center">
                <StorefrontImage
                  v-if="props.currentImageUrl"
                  :src="props.currentImageUrl"
                  :alt="t('resourcesPictureWarehouseShared.lightbox.photoAlt')"
                  class="picture-lightbox-image max-h-[65vh] max-w-full rounded object-contain shadow-lg"
                  preset="gallery"
                  loading="eager"
                />
                <div
                  v-else
                  class="flex aspect-square w-full max-w-[500px] items-center justify-center rounded tz-surface-panel tz-text-muted"
                >
                  {{ t('resourcesPictureWarehouseShared.lightbox.noImage') }}
                </div>
              </div>

              <div
                v-if="props.activePhoto?.galleryImages && props.activePhoto.galleryImages.length > 1"
                class="mt-4 flex max-w-full snap-x items-center gap-2 overflow-x-auto px-2 pb-2"
              >
                <button
                  v-for="(image, index) in props.activePhoto.galleryImages"
                  :key="`${image}-${index}`"
                  type="button"
                  class="relative h-12 w-12 flex-shrink-0 snap-start overflow-hidden rounded border-2 transition-all"
                  :class="index === props.currentGalleryIndex
                    ? 'border-emerald-400 opacity-100'
                    : 'border-transparent opacity-60 hover:opacity-90'"
                  :aria-label="t('resourcesPictureWarehouseShared.lightbox.thumbnail', {
                    index: index + 1,
                  })"
                  @click.stop="emit('select-image', index)"
                >
                  <StorefrontImage
                    :src="image"
                    alt=""
                    class="size-full object-cover"
                    preset="thumbnail"
                  />
                </button>
              </div>
            </div>

            <button
              type="button"
              class="tz-directional-arrow tz-directional-arrow--large absolute right-2 z-20 sm:right-4"
              :aria-label="t('resourcesPictureWarehouseShared.lightbox.nextPhoto')"
              @click.stop="emit('next')"
            >
              <Icon name="lucide:chevron-right" aria-hidden="true" />
            </button>
          </div>

          <div class="grid grid-cols-1 border-t tz-border-subtle tz-surface-panel md:grid-cols-[minmax(0,2fr)_minmax(0,1fr)]">
            <div class="px-4 py-3">
              <div class="mb-2 flex flex-wrap gap-2 tz-caption">
                <button
                  type="button"
                  class="cursor-default rounded-full border border-transparent bg-[#1877F2] px-3 py-1 tz-text-primary opacity-80"
                  :title="t('resourcesPictureWarehouseShared.lightbox.comingSoon')"
                >
                  {{ t('resourcesPictureWarehouseShared.lightbox.shareFacebook') }}
                </button>
                <button
                  type="button"
                  class="cursor-default rounded-full border tz-border-strong/25 px-3 py-1 tz-surface-muted tz-text-secondary opacity-80"
                  :title="t('resourcesPictureWarehouseShared.lightbox.comingSoon')"
                >
                  {{ t('resourcesPictureWarehouseShared.lightbox.shareX') }}
                </button>
                <button
                  type="button"
                  class="cursor-default rounded-full border border-transparent bg-[#FF4500] px-3 py-1 tz-text-primary opacity-80"
                  :title="t('resourcesPictureWarehouseShared.lightbox.comingSoon')"
                >
                  {{ t('resourcesPictureWarehouseShared.lightbox.shareReddit') }}
                </button>
                <button
                  type="button"
                  class="rounded-full border tz-border-subtle px-3 py-1 tz-surface-subtle tz-text-primary transition-colors hover:tz-surface-subtle disabled:cursor-not-allowed disabled:opacity-60"
                  :disabled="props.shareCopying || !props.activePhoto"
                  @click="emit('copy-share-link')"
                >
                  {{ props.shareCopying
                    ? t('resourcesPictureWarehouseShared.lightbox.copying')
                    : t('resourcesPictureWarehouseShared.lightbox.copyLink') }}
                </button>
              </div>
              <p v-if="props.shareMessage" class="mb-2 tz-caption tz-text-muted">
                {{ props.shareMessage }}
              </p>

              <div class="rounded-lg border tz-border-subtle px-3 py-3 tz-caption tz-text-secondary">
                <div class="mb-2 flex items-center justify-between">
                  <h3 class="font-semibold tz-text-primary">
                    {{ t('resourcesPictureWarehouseShared.lightbox.comments') }}
                  </h3>
                  <span v-if="props.commentsLoading" class="tz-caption tz-text-muted">
                    {{ t('resourcesPictureWarehouseShared.lightbox.loading') }}
                  </span>
                </div>

                <div v-if="props.commentsError" class="mb-2 tz-caption text-red-400">
                  {{ t('resourcesPictureWarehouseShared.lightbox.commentsError') }}
                </div>
                <div
                  v-else-if="!props.commentsLoading && !props.comments.length"
                  class="mb-2 tz-caption tz-text-muted"
                >
                  {{ t('resourcesPictureWarehouseShared.lightbox.noComments') }}
                </div>

                <ul v-else class="mb-3 max-h-40 space-y-2 overflow-y-auto pr-1">
                  <li
                    v-for="comment in props.comments"
                    :key="comment.id"
                    class="rounded-lg border tz-border-subtle px-2.5 py-1.5 tz-surface-subtle"
                  >
                    <div class="mb-0.5 flex items-center justify-between gap-2">
                      <span class="tz-caption font-semibold tz-text-primary">
                        {{ comment.author }}
                      </span>
                      <span class="tz-caption tz-text-muted">
                        {{ comment.dateGmtFormatted }}
                      </span>
                    </div>
                    <p class="tz-description tz-text-secondary">
                      {{ comment.content }}
                    </p>
                    <p v-if="comment.location" class="mt-0.5 tz-caption tz-text-muted">
                      {{ comment.location }}
                    </p>
                  </li>
                </ul>

                <form class="space-y-1.5" @submit.prevent="emit('submit-comment')">
                  <textarea
                    :value="props.commentContent"
                    rows="2"
                    class="w-full rounded-lg border tz-border-subtle px-2 py-1 text-sm tz-surface-input tz-text-primary shadow-none focus:border-emerald-600 focus:outline-none focus:ring-1 focus:ring-emerald-600/30"
                    :placeholder="t('resourcesPictureWarehouseShared.lightbox.commentPlaceholder')"
                    @input="emit('update:commentContent', ($event.target as HTMLTextAreaElement).value)"
                  ></textarea>
                  <div class="grid grid-cols-[minmax(0,1.7fr)_minmax(0,1.1fr)] items-center gap-1.5">
                    <input
                      :value="props.commentLocation"
                      type="text"
                      class="h-8 rounded-lg border tz-border-subtle px-2 text-xs tz-surface-input tz-text-primary shadow-none focus:border-emerald-600 focus:outline-none focus:ring-1 focus:ring-emerald-600/30"
                      :placeholder="t('resourcesPictureWarehouseShared.lightbox.locationPlaceholder')"
                      @input="emit('update:commentLocation', ($event.target as HTMLInputElement).value)"
                    />
                    <div class="flex items-center justify-end gap-2">
                      <button
                        type="submit"
                        class="inline-flex items-center justify-center rounded-full px-3 py-1 text-xs font-semibold tz-surface-subtle tz-text-primary disabled:cursor-not-allowed disabled:opacity-50"
                        :disabled="props.commentSubmitting || !props.activePhoto"
                      >
                        {{ props.commentSubmitting
                          ? t('resourcesPictureWarehouseShared.lightbox.sending')
                          : t('resourcesPictureWarehouseShared.lightbox.postComment') }}
                      </button>
                    </div>
                  </div>
                  <p v-if="props.commentSuccess" class="tz-caption text-emerald-600">
                    {{ props.commentSuccess }}
                  </p>
                  <p v-else-if="props.commentError" class="tz-caption text-red-400">
                    {{ props.commentError }}
                  </p>
                </form>
              </div>
            </div>

            <div class="border-t px-4 py-3 tz-border-subtle tz-caption tz-text-secondary md:border-l md:border-t-0">
              <div class="mb-2 font-semibold tz-text-primary">
                {{ t('resourcesPictureWarehouseShared.lightbox.likeThis') }}
              </div>
              <div class="space-y-2">
                <NuxtLink
                  v-for="product in props.activePhotoProductLinks"
                  :key="product.product_id"
                  :to="props.productLinkPath(product)"
                  class="block rounded-lg px-3 py-2 tz-surface-subtle tz-text-secondary transition-colors hover:tz-surface-muted hover:text-emerald-600"
                >
                  <span class="block truncate font-semibold">{{ product.name || product.slug }}</span>
                  <span class="mt-0.5 block truncate font-mono text-[10px] tz-text-primary0">
                    {{ product.slug }}
                  </span>
                </NuxtLink>
                <p v-if="!props.activePhotoProductLinks.length" class="tz-text-muted">
                  {{ t('resourcesPictureWarehouseShared.lightbox.noRelatedProducts') }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </transition>
  </teleport>
</template>

<script setup lang="ts">
import type {
  PhotoComment,
  RiderPhoto,
} from '~/types/pictureWarehouse'
import type { PictureWarehouseProductLink } from '~/types/brandGalleryPhotos'
import { useI18n } from '#imports'

const props = defineProps<{
  open: boolean
  activePhoto: RiderPhoto | null
  currentImageUrl: string
  currentGalleryIndex: number
  comments: PhotoComment[]
  commentsLoading: boolean
  commentsError: string | null
  commentContent: string
  commentLocation: string
  commentSubmitting: boolean
  commentSuccess: string | null
  commentError: string | null
  shareCopying: boolean
  shareMessage: string | null
  activePhotoProductLinks: PictureWarehouseProductLink[]
  productLinkPath: (product: PictureWarehouseProductLink) => string
}>()

const emit = defineEmits<{
  close: []
  previous: []
  next: []
  'select-image': [index: number]
  'copy-share-link': []
  'update:commentContent': [value: string]
  'update:commentLocation': [value: string]
  'submit-comment': []
}>()

const { t } = useI18n()
</script>

