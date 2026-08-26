<template>
  <div>
    <!-- SEO Title: Visually hidden but present for search engines and screen readers -->
    <h1 class="sr-only">Picture warehouse</h1>

    <!-- Riders Tab -->
    <section v-show="activeTab === 'riders'" class="mt-4 space-y-3">
      <div>
        <h2 class="mb-1 text-sm font-semibold tz-text-primary">Riders photos</h2>
        <p class="mb-3 text-xs tz-text-secondary">
          Real builds from riders around the world.
        </p>

        <div class="min-h-[60px]">
          <!-- Riders 加载状态 -->
          <div v-if="userLoading" class="grid gap-3 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
            <div
              v-for="n in 3"
              :key="n"
              class="flex flex-col overflow-hidden rounded-xl border tz-border-subtle bg-[var(--tz-card-surface)] shadow-md"
            >
              <div class="aspect-square w-full tz-surface-panel animate-pulse"></div>
              <div class="px-2.5 py-2 flex flex-col gap-1">
                <div class="h-2.5 w-3/4 rounded bg-slate-200 animate-pulse"></div>
                <div class="h-2 w-1/2 rounded tz-surface-subtle animate-pulse"></div>
              </div>
            </div>
          </div>

          <template v-else>
            <!-- Riders 无数据 -->
            <p v-if="!userPhotos.length" class="text-xs tz-text-secondary">
              No rider photos published yet.
            </p>

            <!-- Riders 数据（真实或占位） -->
            <div v-if="userPhotos.length" class="space-y-2">
              <div class="grid gap-3 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
                <button
                  v-for="(photo, index) in visibleUserPhotos"
                  :key="photo.id"
                  type="button"
                  class="group flex flex-col overflow-hidden rounded-xl border tz-border-subtle bg-[var(--tz-card-surface)] shadow-md transition-colors hover:tz-border-subtle hover:tz-surface-subtle"
                  @click="openLightbox('user', index)"
                >
                  <div class="aspect-square w-full overflow-hidden tz-surface-panel group-hover:tz-surface-muted transition-colors">
                    <StorefrontImage
                      v-if="photoCover(photo)"
                      :src="photoCover(photo)"
                      :alt="photo.title"
                      class="size-full object-cover transition duration-200 group-hover:scale-[1.03]"
                      preset="card"
                    />
                  </div>
                  <div class="px-2.5 py-2 flex flex-col gap-0.5">
                    <p class="tz-caption font-medium tz-text-primary truncate">
                      {{ photo.title }}
                    </p>
                    <p class="tz-micro-label tz-text-muted truncate">
                      {{ photo.region }}<span v-if="photo.nickname"> · {{ photo.nickname }}</span>
                    </p>
                  </div>
                </button>
              </div>

              <div v-if="hasMoreUserPhotos" class="pt-1">
                <button
                  type="button"
                  class="tz-micro-label text-emerald-600 hover:text-emerald-700 underline underline-offset-2"
                  @click="toggleUserPhotos"
                >
                  {{ showAllUserPhotos ? 'Show fewer photos' : 'Show more photos' }}
                </button>
              </div>
            </div>
          </template>
        </div>
      </div>
    </section>

    <!-- Brand Tab -->
    <section v-show="activeTab === 'brand'" class="mt-4 space-y-3">
      <div>
        <h2 class="mb-1 text-sm font-semibold tz-text-primary">Brand photos</h2>
        <p class="mb-3 text-xs tz-text-secondary">
          Official product and detail shots curated by our team.
        </p>

        <div class="min-h-[60px]">
          <!-- 加载状态 -->
          <div v-if="brandLoading" class="grid gap-3 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
            <div
              v-for="n in 3"
              :key="n"
              class="flex flex-col overflow-hidden rounded-xl border tz-border-subtle tz-surface-card shadow-[0_3px_9px_rgb(15_23_42_/_0.08)]"
            >
              <div class="aspect-square w-full bg-[var(--tz-image-loading-surface)] animate-pulse"></div>
              <div class="px-2.5 py-2 flex flex-col gap-1">
                <div class="h-2.5 w-3/4 rounded tz-surface-panel animate-pulse"></div>
                <div class="h-2 w-1/2 rounded tz-surface-panel animate-pulse"></div>
              </div>
            </div>
          </div>

          <template v-else>
            <!-- 无数据 -->
            <p v-if="!brandPhotos.length" class="text-xs tz-text-secondary">
              No brand photos published yet.
            </p>

            <!-- 数据（真实或占位） -->
            <div v-if="brandPhotos.length" class="space-y-2">
              <div class="grid gap-3 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
                <button
                  v-for="(photo, index) in visibleBrandPhotos"
                  :key="photo.id"
                  type="button"
                  class="group flex flex-col overflow-hidden rounded-xl border tz-border-subtle tz-surface-card shadow-[0_3px_9px_rgb(15_23_42_/_0.08)] hover:shadow-[0_4px_12px_rgb(15_23_42_/_0.12)] transition-all"
                  @click="openLightbox('brand', index)"
                >
                  <div class="aspect-square w-full overflow-hidden bg-[var(--tz-image-loading-surface)] group-hover:tz-surface-muted transition-colors">
                    <StorefrontImage
                      v-if="photoCover(photo)"
                      :src="photoCover(photo)"
                      :alt="photo.title"
                      class="size-full object-cover transition duration-200 group-hover:scale-[1.03]"
                      preset="card"
                    />
                  </div>
                  <div class="px-2.5 py-2 flex flex-col gap-0.5">
                    <p class="tz-caption font-medium tz-text-primary truncate">
                      {{ photo.title }}
                    </p>
                    <p class="tz-micro-label tz-text-muted truncate">
                      {{ photo.region }}<span v-if="photo.nickname"> · {{ photo.nickname }}</span>
                    </p>
                  </div>
                </button>
              </div>

              <div v-if="hasMoreBrandPhotos" class="pt-1">
                <button
                  type="button"
                    class="tz-micro-label text-emerald-600 hover:text-emerald-700 underline underline-offset-2"
                  @click="toggleBrandPhotos"
                >
                  {{ showAllBrandPhotos ? 'Show fewer photos' : 'Show more photos' }}
                </button>
              </div>
            </div>
          </template>
        </div>
      </div>
    </section>

    <!-- 上传表单（Phase 3：调用 /tanz-photo/v1/upload） - 移至底部通栏 -->
    <section class="mt-10 border-t tz-border-subtle pt-8">
      <div class="picture-upload-card mx-auto max-w-3xl rounded-2xl px-4 py-4 sm:px-6 sm:py-5">
        <div class="mb-4 text-center">
          <h4 class="text-sm font-semibold tz-text-primary">
            Share your build (login required)
          </h4>
          <p class="tz-caption tz-text-secondary">
            Join the gallery! {{ uploadSpecHint('user_showcase_image') }}. Uploaded photos will appear after review.
          </p>
        </div>
        
        <form class="space-y-3" @submit.prevent="submitUpload">
          <div class="flex flex-col gap-1">
            <label class="tz-caption tz-text-secondary">
              Order <span class="text-red-400">*</span>
            </label>
            <select
              v-model="uploadOrderID"
              class="picture-upload-control h-8 rounded-lg px-2.5 text-xs"
              :disabled="uploadOrdersLoading || uploading || !uploadOrders.length"
              required
            >
              <option value="" disabled>
                {{ uploadOrdersLoading ? 'Loading orders...' : 'Select a completed order' }}
              </option>
              <option
                v-for="order in uploadOrders"
                :key="order.id"
                :value="String(order.id)"
                :disabled="!order.eligible"
              >
                {{ order.order_number }} · {{ order.status }}
              </option>
            </select>
            <p v-if="uploadOrdersError" class="tz-caption text-red-400">
              {{ uploadOrdersError }}
            </p>
            <p v-else-if="!uploadOrdersLoading && !hasEligibleUploadOrder" class="tz-caption tz-text-muted">
              A completed order is required before you can upload photos.
            </p>
          </div>

          <div class="grid gap-3 sm:grid-cols-2">
            <div class="flex flex-col gap-1">
              <label class="tz-caption tz-text-secondary">
                Region <span class="text-red-400">*</span>
              </label>
              <input
                v-model="uploadRegion"
                type="text"
                class="picture-upload-control h-8 rounded-lg px-2.5 text-xs"
                placeholder="e.g. Germany"
                required
              />
            </div>
            <div class="flex flex-col gap-1">
              <label class="tz-caption tz-text-secondary">Location</label>
              <input
                v-model="uploadLocation"
                type="text"
                class="picture-upload-control h-8 rounded-lg px-2.5 text-xs"
                placeholder="e.g. Berlin"
              />
            </div>
          </div>

          <div class="flex flex-col gap-1">
            <label class="tz-caption tz-text-secondary">Nickname</label>
            <input
              v-model="uploadNickname"
              type="text"
              class="picture-upload-control h-8 rounded-lg px-2.5 text-xs"
              placeholder="Your name or handle"
            />
          </div>

          <div class="flex flex-col gap-1">
            <label class="tz-caption tz-text-secondary">Notes</label>
            <textarea
              v-model="uploadNotes"
              rows="2"
              class="picture-upload-control rounded-lg px-2.5 py-1.5 text-xs"
              placeholder="Tell us about your build..."
            ></textarea>
          </div>

          <div class="flex flex-col gap-1">
            <label class="tz-caption tz-text-secondary">Photos (WEBP, max 10)</label>
            <div class="picture-upload-picker">
              <div class="flex flex-wrap items-center justify-between gap-2">
                <button
                  type="button"
                  class="picture-upload-file-trigger inline-flex h-8 items-center gap-1.5 rounded-lg px-3 text-xs font-semibold disabled:cursor-not-allowed disabled:opacity-50"
                  :disabled="uploading || uploadFiles.length >= 10 || !selectedUploadOrderEligible"
                  @click="openUploadFilePicker"
                >
                  <Icon name="lucide:image-plus" class="h-3.5 w-3.5" />
                  <span>{{ uploadFiles.length >= 10 ? 'Maximum reached' : 'Choose files' }}</span>
                </button>
                <span class="tz-caption tz-text-muted">
                  {{ uploadFiles.length ? `${uploadFiles.length}/10 selected` : 'No photos selected' }}
                </span>
              </div>

              <input
                ref="uploadFileInput"
                type="file"
                :accept="uploadSpecAccept('user_showcase_image')"
                multiple
                :disabled="uploading || !selectedUploadOrderEligible"
                @change="onUploadFileChange"
                class="hidden"
              />
              <p class="mt-2 tz-caption tz-text-muted">{{ uploadSpecHint('user_showcase_image') }}</p>

              <div
                v-if="uploadPreviews.length"
                class="mt-3 grid grid-cols-2 gap-2 sm:grid-cols-4 md:grid-cols-5"
              >
                <div
                  v-for="(preview, index) in uploadPreviews"
                  :key="preview.key"
                  class="picture-upload-preview group"
                >
                  <div class="relative aspect-square overflow-hidden rounded-lg tz-surface-panel">
                    <StorefrontImage
                      :src="preview.url"
                      :alt="preview.file.name"
                      class="size-full object-cover"
                      preset="thumbnail"
                    />
                    <button
                      type="button"
                      class="picture-upload-preview-remove"
                      :aria-label="`Remove ${preview.file.name}`"
                      title="Remove photo"
                      @click="removeUploadFile(index)"
                    >
                      <Icon name="lucide:x" class="h-3.5 w-3.5" />
                    </button>
                  </div>
                  <p class="mt-1 truncate text-[10px] tz-text-muted" :title="preview.file.name">
                    {{ preview.file.name }}
                  </p>
                </div>
              </div>
            </div>
          </div>

          <div class="flex items-center justify-between gap-2 pt-2 border-t tz-border-subtle mt-2">
            <div class="flex-1">
               <p v-if="uploadSuccess" class="tz-caption text-emerald-600">
                {{ uploadSuccess }}
              </p>
              <p v-else-if="uploadError" class="tz-caption text-red-400">
                {{ uploadError }}
              </p>
            </div>
            <button
              type="submit"
              class="picture-upload-submit inline-flex items-center justify-center h-9 rounded-full px-5 text-xs font-semibold disabled:cursor-not-allowed disabled:opacity-50"
              :disabled="uploading || !selectedUploadOrderEligible"
            >
              <span v-if="uploading">Uploading…</span>
              <span v-else>Submit for review</span>
            </button>
          </div>
        </form>
      </div>
    </section>

    <!-- 单张图片详情弹窗（Phase 1/2：仅大图 + 标题 + 左右切换 + 关闭，评论/分享/推荐为 UI 占位） -->
    <teleport to="body">
      <transition
        enter-active-class="transition-opacity duration-200 ease-out"
        leave-active-class="transition-opacity duration-150 ease-in"
        enter-from-class="opacity-0"
        leave-to-class="opacity-0"
      >
        <div
          v-if="isLightboxOpen && activeKind === 'user'"
          class="tz-standard-modal-mask tz-standard-modal-mask--compact fixed inset-0 z-[1400] flex items-center justify-center px-3 tz-mobile-safe-modal-mask"
          @click.self="closeLightbox"
        >
          <div
            class="picture-lightbox-panel tz-standard-modal-surface relative w-full max-w-[960px] max-h-[90vh] tz-surface-panel flex flex-col overflow-hidden md:overflow-y-auto"
          >
            <!-- 顶部标题 + 关闭按钮 -->
            <header
              class="px-4 py-3 flex items-center justify-between border-b tz-border-subtle tz-surface-panel"
            >
              <h3 class="text-sm sm:text-base font-semibold tz-text-primary truncate">
                {{ activePhoto?.title || 'Picture' }}
              </h3>
              <button
                type="button"
                class="tz-global-close-btn ml-4"
                @click="closeLightbox"
                aria-label="Close"
              >
                ×
              </button>
            </header>

            <!-- 中部：大图区域 + 左右切换 -->
            <div class="relative flex-1 flex items-center justify-center bg-[var(--tz-image-loading-surface)] p-4 overflow-hidden">
              <button
                type="button"
                class="tz-directional-arrow tz-directional-arrow--large absolute left-2 z-20 sm:left-4"
                @click.stop="goPrev"
                aria-label="Previous photo"
              >
                <Icon name="lucide:chevron-left" aria-hidden="true" />
              </button>

              <div class="relative w-full h-full flex flex-col items-center justify-center">
                <!-- 大图 -->
                <div class="relative flex items-center justify-center w-full h-full">
                  <StorefrontImage
                    v-if="currentImageUrl"
                    :src="currentImageUrl"
                    alt="Photo"
                    class="picture-lightbox-image max-w-full max-h-[65vh] object-contain rounded shadow-lg"
                    preset="gallery"
                    loading="eager"
                  />
                  <div v-else class="w-full max-w-[500px] aspect-square tz-surface-panel rounded flex items-center justify-center tz-text-muted">
                    No image
                  </div>
                </div>

                <!-- 悬浮缩略图 (仅当有多张图时显示) -->
                <div
                  v-if="activePhoto?.galleryImages && activePhoto.galleryImages.length > 1"
                  class="mt-4 flex items-center gap-2 overflow-x-auto max-w-full pb-2 px-2 snap-x"
                >
                  <button
                    v-for="(img, idx) in activePhoto.galleryImages"
                    :key="idx"
                    type="button"
                    class="relative flex-shrink-0 w-12 h-12 rounded overflow-hidden border-2 transition-all snap-start"
                    :class="idx === currentGalleryIndex ? 'border-emerald-400 opacity-100' : 'border-transparent opacity-60 hover:opacity-90'"
                    @click.stop="currentGalleryIndex = idx"
                  >
                    <StorefrontImage :src="img" alt="" class="w-full h-full object-cover" preset="thumbnail" />
                  </button>
                </div>
              </div>

              <button
                type="button"
                class="tz-directional-arrow tz-directional-arrow--large absolute right-2 z-20 sm:right-4"
                @click.stop="goNext"
                aria-label="Next photo"
              >
                <Icon name="lucide:chevron-right" aria-hidden="true" />
              </button>
            </div>
            <div
              class="border-t tz-border-subtle grid grid-cols-1 md:grid-cols-[minmax(0,2fr)_minmax(0,1fr)] tz-surface-panel"
            >
              <!-- 左：操作按钮栏 + 评论占位 -->
              <div class="px-4 py-3">
                <div class="mb-2 flex flex-wrap gap-2 tz-caption">
                  <button
                    type="button"
                    class="px-3 py-1 rounded-full bg-[#1877F2] tz-text-primary border border-transparent opacity-80 cursor-default"
                    title="Coming soon"
                  >
                    Share to Facebook
                  </button>
                  <button
                    type="button"
                      class="px-3 py-1 rounded-full tz-surface-muted tz-text-secondary border tz-border-strong/25 opacity-80 cursor-default"
                    title="Coming soon"
                  >
                    Share to X
                  </button>
                  <button
                    type="button"
                    class="px-3 py-1 rounded-full bg-[#FF4500] tz-text-primary border border-transparent opacity-80 cursor-default"
                    title="Coming soon"
                  >
                    Share to Reddit
                  </button>
                  <button
                    type="button"
                    class="px-3 py-1 rounded-full tz-surface-subtle tz-text-primary border tz-border-subtle hover:tz-surface-subtle transition-colors disabled:opacity-60 disabled:cursor-not-allowed"
                    @click="copyShareLink"
                    :disabled="shareCopying || !activePhoto"
                  >
                    <span v-if="shareCopying">Copying...</span>
                    <span v-else>Copy link</span>
                  </button>
                </div>
                <p v-if="shareMessage" class="tz-caption mb-2 tz-text-muted">
                  {{ shareMessage }}
                </p>
                <div class="tz-caption rounded-lg border tz-border-subtle px-3 py-3 tz-text-secondary">
                  <!-- 评论列表 -->
                  <div class="mb-2 flex items-center justify-between">
                    <h4 class="font-semibold tz-text-primary">Comments</h4>
                    <span v-if="commentsLoading" class="tz-caption tz-text-muted">
                      Loading...
                    </span>
                  </div>

                  <div v-if="commentsError" class="tz-caption mb-2 text-red-400">
                    Unable to load comments. Please try again later.
                  </div>

                  <div v-else-if="!commentsLoading && !comments.length" class="tz-caption mb-2 tz-text-muted">
                    No comments yet.
                  </div>

                  <ul v-else class="mb-3 space-y-2 max-h-40 overflow-y-auto pr-1">
                    <li
                      v-for="comment in comments"
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
                      <p v-if="comment.location" class="tz-caption mt-0.5 tz-text-muted">
                        {{ comment.location }}
                      </p>
                    </li>
                  </ul>

                  <!-- 简单评论表单 -->
                  <form class="space-y-1.5" @submit.prevent="submitComment">
                    <textarea
                      v-model="commentContent"
                      rows="2"
                      class="w-full rounded-lg border tz-border-subtle px-2 py-1 text-sm tz-text-primary tz-surface-input shadow-none focus:outline-none focus:border-emerald-600 focus:ring-1 focus:ring-emerald-600/30"
                      placeholder="Write a comment (login required)"
                    ></textarea>
                    <div class="grid grid-cols-[minmax(0,1.7fr)_minmax(0,1.1fr)] gap-1.5 items-center">
                      <input
                        v-model="commentLocation"
                        type="text"
                        class="h-8 rounded-lg border tz-border-subtle px-2 text-xs tz-text-primary tz-surface-input shadow-none focus:outline-none focus:border-emerald-600 focus:ring-1 focus:ring-emerald-600/30"
                        placeholder="Location (optional)"
                      />
                      <div class="flex items-center justify-end gap-2">
                        <button
                          type="submit"
                          class="inline-flex items-center justify-center rounded-full tz-surface-subtle px-3 py-1 text-xs font-semibold tz-text-primary disabled:cursor-not-allowed disabled:opacity-50"
                          :disabled="commentSubmitting || !activePhoto"
                        >
                          <span v-if="commentSubmitting">Sending...</span>
                          <span v-else>Post comment</span>
                        </button>
                      </div>
                    </div>
                    <p v-if="commentSuccess" class="tz-caption text-emerald-600">
                      {{ commentSuccess }}
                    </p>
                    <p v-else-if="commentError" class="tz-caption text-red-400">
                      {{ commentError }}
                    </p>
                  </form>
                </div>
              </div>

              <!-- 右：推荐模块 -->
              <div
                class="px-4 py-3 border-t md:border-t-0 md:border-l tz-border-subtle tz-caption tz-text-secondary"
              >
                <div class="mb-2 font-semibold tz-text-primary">Like This? Get The Same Build.</div>
                <div class="space-y-2">
                  <NuxtLink
                    v-for="product in activePhotoProductLinks"
                    :key="product.product_id"
                    :to="productLinkPath(product)"
                        class="block rounded-lg tz-surface-subtle px-3 py-2 tz-text-secondary transition-colors hover:tz-surface-muted hover:text-emerald-600"
                  >
                    <span class="block truncate font-semibold">{{ product.name || product.slug }}</span>
                    <span class="mt-0.5 block truncate font-mono text-[10px] tz-text-primary0">{{ product.slug }}</span>
                  </NuxtLink>
                  <p v-if="!activePhotoProductLinks.length" class="tz-text-muted">
                    No related products configured.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </transition>
    </teleport>

    <GlobalBrandGalleryLightbox
      :open="isLightboxOpen && activeKind === 'brand'"
      :gallery="activeBrandPhoto"
      @close="closeLightbox"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { definePageMeta, useHead, useLocalePath, useRuntimeConfig } from '#imports'

import GlobalBrandGalleryLightbox from '~/components/global/gallery/GlobalBrandGalleryLightbox.vue'
import { useAuth } from '~/composables/useAuth'
import { useBrandGalleryPhotos } from '~/composables/useBrandGalleryPhotos'
import type {
  BrandGalleryPhoto,
  PictureWarehouseProductLink,
} from '~/types/brandGalleryPhotos'
import {
  pictureWarehouseTabs,
} from '~/utils/pageSubNavigation'
import { buildProductPath } from '~/utils/seo/urls'
import { usePageSubNavigationTab } from '~/composables/usePageSubNavigationTab'
import {
  createStorefrontMediaContext,
  normalizeStorefrontMediaUrl,
} from '~/utils/storefrontMedia'
import {
  uploadSpecAccept,
  uploadSpecHint,
  validateStorefrontUploadFiles,
} from '~/utils/uploadSpecs'

definePageMeta({
  layout: 'products',
  footerLabelKey: 'company.nav.pictureWarehouse',
  footerLabelFallback: 'Picture Warehouse',
})

useHead({
  title: 'Picture warehouse',
})

const auth = useAuth()
const mediaContext = createStorefrontMediaContext(useRuntimeConfig())
const localePath = useLocalePath()
const {
  brandPhotos,
  brandLoading,
  brandError,
  fetchBrandPhotos,
  loadBrandGalleryDetails,
} = useBrandGalleryPhotos()

type PhotoKind = 'user' | 'brand'

interface RiderPhoto {
  id: string
  kind: 'user'
  title: string
  region: string
  nickname?: string
  galleryImages?: string[]
  productLinks?: PictureWarehouseProductLink[]
}

type PicturePhoto = RiderPhoto | BrandGalleryPhoto

interface PhotoComment {
  id: number
  author: string
  content: string
  dateGmt: string
  dateGmtFormatted: string
  location?: string
}

const userPhotos = ref<RiderPhoto[]>([])
const userLoading = ref(true)
const userError = ref<string | null>(null)

const uploadRegion = ref('')
const uploadLocation = ref('')
const uploadNickname = ref('')
const uploadNotes = ref('')
interface UploadOrderOption {
  id: number
  order_number: string
  status: string
  shipping_status: string
  completed_at?: string
  total_amount: number
  currency: string
  eligible: boolean
}

const uploadOrders = ref<UploadOrderOption[]>([])
const uploadOrdersLoading = ref(false)
const uploadOrdersError = ref<string | null>(null)
const uploadOrderID = ref('')
const selectedUploadOrderEligible = computed(() =>
  uploadOrders.value.some(
    (order) => String(order.id) === uploadOrderID.value && order.eligible,
  ),
)
const hasEligibleUploadOrder = computed(() => uploadOrders.value.some((order) => order.eligible))
interface UploadPreview {
  key: string
  file: File
  url: string
}

const uploadFiles = ref<File[]>([])
const uploadPreviews = ref<UploadPreview[]>([])
const uploadFileInput = ref<HTMLInputElement | null>(null)
const uploading = ref(false)
const uploadError = ref<string | null>(null)
const uploadSuccess = ref<string | null>(null)

const comments = ref<PhotoComment[]>([])
const commentsLoading = ref(false)
const commentsError = ref<string | null>(null)
const commentContent = ref('')
const commentLocation = ref('')
const commentSubmitting = ref(false)
const commentSuccess = ref<string | null>(null)
const commentError = ref<string | null>(null)
const shareCopying = ref(false)
const shareMessage = ref<string | null>(null)

const uploadFileKey = (file: File): string =>
  `${file.name}:${file.size}:${file.lastModified}`

const openUploadFilePicker = () => {
  if (!selectedUploadOrderEligible.value) return
  uploadFileInput.value?.click()
}

const onUploadFileChange = async (event: Event) => {
  if (!selectedUploadOrderEligible.value) {
    const target = event.target as HTMLInputElement | null
    if (target) target.value = ''
    return
  }

  const target = event.target as HTMLInputElement | null
  if (!target || !target.files || target.files.length === 0) {
    target && (target.value = '')
    return
  }

  uploadError.value = null
  uploadSuccess.value = null

  const selectedFiles = Array.from(target.files)
  const validation = await validateStorefrontUploadFiles([...uploadFiles.value, ...selectedFiles], 'user_showcase_image')
  if (!validation.ok) {
    uploadError.value = validation.error || 'The selected images do not meet the upload requirements.'
    target.value = ''
    return
  }

  const existingKeys = new Set(uploadFiles.value.map(uploadFileKey))
  const newFiles = selectedFiles.filter((file) => !existingKeys.has(uploadFileKey(file)))
  const availableSlots = Math.max(0, 10 - uploadFiles.value.length)
  const filesToAdd = newFiles.slice(0, availableSlots)

  if (newFiles.length > availableSlots) {
    uploadError.value = 'Maximum 10 files allowed.'
  }

  if (filesToAdd.length) {
    uploadFiles.value.push(...filesToAdd)
    uploadPreviews.value.push(
      ...filesToAdd.map((file) => ({
        key: uploadFileKey(file),
        file,
        url: URL.createObjectURL(file),
      })),
    )
  }

  // Reset the native input so selecting the same file again still emits change.
  target.value = ''
}

const removeUploadFile = (index: number) => {
  const preview = uploadPreviews.value[index]
  if (!preview) return

  URL.revokeObjectURL(preview.url)
  uploadPreviews.value.splice(index, 1)
  uploadFiles.value.splice(index, 1)
}

const clearUploadFiles = () => {
  uploadPreviews.value.forEach((preview) => URL.revokeObjectURL(preview.url))
  uploadPreviews.value = []
  uploadFiles.value = []
}

onBeforeUnmount(clearUploadFiles)

const submitUpload = async () => {
  uploadError.value = null
  uploadSuccess.value = null

  if (!selectedUploadOrderEligible.value) {
    uploadError.value = 'Please select a completed order before uploading.'
    return
  }

  if (!uploadRegion.value.trim()) {
    uploadError.value = 'Please enter a region.'
    return
  }

  if (uploadFiles.value.length === 0) {
    uploadError.value = 'Please choose at least one WEBP image to upload.'
    return
  }
  
  if (uploadFiles.value.length > 10) {
    uploadError.value = 'Maximum 10 files allowed.'
    return
  }
  const validation = await validateStorefrontUploadFiles(uploadFiles.value, 'user_showcase_image')
  if (!validation.ok) {
    uploadError.value = validation.error || 'The selected images do not meet the upload requirements.'
    return
  }

  uploading.value = true

  try {
    const formData = new FormData()
    
    // Append multiple files. Using 'file[]' ensures PHP treats it as an array.
    uploadFiles.value.forEach((file) => {
      formData.append('file[]', file)
    })
    
    formData.append('order_id', uploadOrderID.value)
    formData.append('region', uploadRegion.value.trim())

    if (uploadLocation.value.trim()) formData.append('location', uploadLocation.value.trim())
    if (uploadNickname.value.trim()) formData.append('nickname', uploadNickname.value.trim())
    if (uploadNotes.value.trim()) formData.append('notes', uploadNotes.value.trim())

    try {
      await auth.request('/showcase/upload', {
        method: 'POST',
        // FormData will automatically set boundary headers, so don't set Content-Type here
        headers: { accept: 'application/json' },
        body: formData,
      }, 'Upload failed. Please try again later.')
    } catch (err: any) {
      const msg = err?.message || 'Upload failed. Please try again later.'
      if (err?.code === 'showcase_upload_order_not_eligible' || msg.includes('showcase_upload_order_not_eligible')) {
        uploadError.value = 'Only a completed order from your account can be used.'
      } else if (err?.code === 'showcase_upload_order_required' || msg.includes('showcase_upload_order_required')) {
        uploadError.value = 'Please select a completed order before uploading.'
      } else if (err?.status === 401 || msg.includes('401') || msg.toLowerCase().includes('login')) {
        uploadError.value = 'Please log in before uploading.'
      } else if (msg.includes('429')) {
        uploadError.value = 'Too many uploads. Please try again later.'
      } else if (msg.includes('invalid_type') || msg.includes('Only WEBP')) {
        uploadError.value = 'Only WEBP images are allowed.'
      } else if (msg.includes('file_too_large') || msg.includes('too large')) {
        uploadError.value = 'File is too large. Please keep it under 5MB.'
      } else if (msg.includes('missing_region')) {
        uploadError.value = 'Region is required.'
      } else {
        uploadError.value = msg
      }
      return
    }

    uploadSuccess.value = 'Photos submitted for review.'
    clearUploadFiles()
    if (uploadFileInput.value) uploadFileInput.value.value = ''
    uploadNotes.value = ''
  } catch {
    uploadError.value = 'Upload failed. Please try again later.'
  } finally {
    uploading.value = false
  }
}

const showAllUserPhotos = ref(false)
const showAllBrandPhotos = ref(false)

const visibleUserPhotos = computed<PicturePhoto[]>(() => {
  if (!userPhotos.value.length) return []
  return showAllUserPhotos.value ? userPhotos.value : userPhotos.value.slice(0, 6)
})

const visibleBrandPhotos = computed<PicturePhoto[]>(() => {
  if (!brandPhotos.value.length) return []
  return showAllBrandPhotos.value ? brandPhotos.value : brandPhotos.value.slice(0, 6)
})

const hasMoreUserPhotos = computed(() => userPhotos.value.length > 6)
const hasMoreBrandPhotos = computed(() => brandPhotos.value.length > 6)

const photoCover = (photo: PicturePhoto): string =>
  photo.kind === 'brand'
    ? photo.coverImage || photo.galleryImages?.[0] || ''
    : photo.galleryImages?.[0] || ''
const productLinkPath = (product: PictureWarehouseProductLink): string => {
  const slug = String(product.slug || '').trim()
  return slug ? localePath(buildProductPath(slug)) : localePath('/shop')
}

const toggleUserPhotos = () => {
  if (!hasMoreUserPhotos.value) return
  showAllUserPhotos.value = !showAllUserPhotos.value
}

const toggleBrandPhotos = () => {
  if (!hasMoreBrandPhotos.value) return
  showAllBrandPhotos.value = !showAllBrandPhotos.value
}

const mapPayloadToUserPhotos = (payload: any[]): RiderPhoto[] => {
  return payload
    .map((item: any): RiderPhoto | null => {
      if (!item) return null
      const id = item.id ?? item.ID ?? null
      if (!id) return null

      const galleryImages = Array.isArray(item.gallery_images)
        ? item.gallery_images
          .map((image: unknown) => normalizeStorefrontMediaUrl(image, mediaContext))
          .filter((image: string): image is string => Boolean(image))
        : []

      return {
        id: String(id),
        kind: 'user',
        title: String(item.title ?? item.post_title ?? 'Rider photo'),
        region: String(item.region ?? 'Unknown'),
        nickname: typeof item.nickname === 'string' ? item.nickname : undefined,
        galleryImages,
      }
    })
    .filter((p): p is RiderPhoto => p !== null)
}

const fetchUserPhotos = async () => {
  try {
    userLoading.value = true
    userError.value = null

    const payload = await auth.request<any[]>('/showcase/gallery?type=user&status=approved')
    const mapped = Array.isArray(payload) ? mapPayloadToUserPhotos(payload) : []

    userPhotos.value = mapped
  } catch (error) {
    userError.value = 'load_failed'
    userPhotos.value = []
  } finally {
    userLoading.value = false
  }
}

const fetchUploadOrders = async () => {
  uploadOrdersLoading.value = true
  uploadOrdersError.value = null
  uploadOrders.value = []
  uploadOrderID.value = ''

  try {
    const payload = await auth.request<UploadOrderOption[]>(
      '/showcase/upload-orders',
      {},
      'Unable to load your orders.',
    )
    uploadOrders.value = Array.isArray(payload) ? payload : []
  } catch (error: any) {
    uploadOrdersError.value =
      error?.status === 401 || String(error?.message || '').toLowerCase().includes('login')
        ? 'Please log in before uploading.'
        : 'Unable to load your orders.'
  } finally {
    uploadOrdersLoading.value = false
  }
}

watch(
  () => auth.isAuthenticated.value,
  (authenticated, wasAuthenticated) => {
    if (authenticated && !wasAuthenticated) {
      void fetchUploadOrders()
      return
    }
    if (!authenticated) {
      uploadOrders.value = []
      uploadOrderID.value = ''
      clearUploadFiles()
    }
  },
)

onMounted(() => {
  fetchUserPhotos()
  fetchBrandPhotos()
  fetchUploadOrders()
})

const activeKind = ref<PhotoKind | null>(null)
const activeIndex = ref<number | null>(null)
const currentGalleryIndex = ref(0)

const { activeTab } = usePageSubNavigationTab({
  tabs: pictureWarehouseTabs,
  basePath: '/resources/picture-warehouse',
  defaultValue: 'riders',
})

const isLightboxOpen = computed(() => activeKind.value !== null && activeIndex.value !== null)

const activeList = computed<PicturePhoto[] | null>(() => {
  if (!activeKind.value) return null
  return activeKind.value === 'user' ? userPhotos.value : brandPhotos.value
})

const activePhoto = computed<PicturePhoto | null>(() => {
  const list = activeList.value
  if (!list || !list.length || activeIndex.value === null) return null
  return list[activeIndex.value] ?? null
})

const activeBrandPhoto = computed<BrandGalleryPhoto | null>(() => {
  if (activeKind.value !== 'brand' || activeIndex.value === null) return null
  return brandPhotos.value[activeIndex.value] ?? null
})

const activePhotoProductLinks = computed<PictureWarehouseProductLink[]>(() =>
  (activePhoto.value?.productLinks || []).filter((product) => Boolean(product.slug || product.name))
)

const currentImageUrl = computed(() => {
  const photo = activePhoto.value
  if (!photo) return ''
  // Prefer gallery image at current index, fallback to first gallery image, fallback to placeholder
  if (photo.galleryImages && photo.galleryImages.length > 0) {
    return photo.galleryImages[currentGalleryIndex.value] ?? photo.galleryImages[0]
  }
  return '' // Should handle empty case or show a placeholder if needed
})

const loadActiveBrandGalleryDetails = () => {
  if (activeKind.value !== 'brand' || activeIndex.value === null) return
  void loadBrandGalleryDetails(activeIndex.value)
}

const openLightbox = (kind: PhotoKind, index: number) => {
  activeKind.value = kind
  activeIndex.value = index
  currentGalleryIndex.value = 0
  if (kind === 'user') {
    void loadCommentsForActivePhoto()
  } else {
    loadActiveBrandGalleryDetails()
  }
}

const closeLightbox = () => {
  activeKind.value = null
  activeIndex.value = null
  comments.value = []
  commentsError.value = null
  commentsLoading.value = false
  commentContent.value = ''
  commentLocation.value = ''
  commentSuccess.value = null
  commentError.value = null
}

const goNext = () => {
  const list = activeList.value
  if (!list || !list.length || activeIndex.value === null) return
  const nextIndex = (activeIndex.value + 1) % list.length
  activeIndex.value = nextIndex
  currentGalleryIndex.value = 0
  if (activeKind.value === 'user') {
    void loadCommentsForActivePhoto()
  } else {
    loadActiveBrandGalleryDetails()
  }
}

const goPrev = () => {
  const list = activeList.value
  if (!list || !list.length || activeIndex.value === null) return
  const prevIndex = (activeIndex.value - 1 + list.length) % list.length
  activeIndex.value = prevIndex
  currentGalleryIndex.value = 0
  if (activeKind.value === 'user') {
    void loadCommentsForActivePhoto()
  } else {
    loadActiveBrandGalleryDetails()
  }
}

const fetchCommentsForPhoto = async (photoId: string) => {
  try {
    commentsLoading.value = true
    commentsError.value = null

    const payload = await auth.request<any[]>(
      `/showcase/comments?photo_id=${encodeURIComponent(photoId)}&per_page=20`
    )
    if (!Array.isArray(payload)) {
      comments.value = []
      return
    }

    comments.value = payload.map((item: any): PhotoComment => {
      const rawDate = String(item.date_gmt ?? '')
      let formatted = rawDate
      const date = rawDate ? new Date(`${rawDate}Z`) : null
      if (date && !Number.isNaN(date.getTime())) {
        formatted = date.toLocaleDateString(undefined, {
          year: 'numeric',
          month: 'short',
          day: 'numeric',
        })
      }

      return {
        id: Number(item.id ?? 0),
        author: String(item.author ?? 'Anonymous'),
        content: String(item.content ?? ''),
        dateGmt: rawDate,
        dateGmtFormatted: formatted,
        location: item.location ? String(item.location) : '',
      }
    })
  } catch (error) {
    commentsError.value = 'load_failed'
    comments.value = []
  } finally {
    commentsLoading.value = false
  }
}

const loadCommentsForActivePhoto = async () => {
  const current = activePhoto.value
  if (!current) {
    comments.value = []
    commentsError.value = null
    commentsLoading.value = false
    return
  }
  if (current.kind !== 'user') {
    comments.value = []
    commentsError.value = null
    commentsLoading.value = false
    return
  }

  await fetchCommentsForPhoto(current.id)
}

const submitComment = async () => {
  commentError.value = null
  commentSuccess.value = null

  const current = activePhoto.value
  if (!current) {
    commentError.value = 'No photo selected.'
    return
  }
  if (current.kind !== 'user') {
    commentError.value = 'Comments are only available for rider photos.'
    return
  }

  if (!commentContent.value.trim()) {
    commentError.value = 'Please enter a comment.'
    return
  }

  commentSubmitting.value = true

  try {
    const body = {
      photo_id: Number(current.id),
      content: commentContent.value.trim(),
    } as any

    if (commentLocation.value.trim()) {
      body.location = commentLocation.value.trim()
    }

    try {
      await auth.request('/showcase/comments', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'accept': 'application/json',
        },
        body: JSON.stringify(body),
      }, 'Failed to submit comment. Please try again later.')
    } catch (err: any) {
      const msg = err?.message || 'Failed to submit comment. Please try again later.'
      if (msg.includes('401') || msg.includes('403') || msg.toLowerCase().includes('login')) {
        commentError.value = 'Please log in before commenting.'
      } else if (msg.includes('empty_comment') || msg.includes('cannot be empty')) {
        commentError.value = 'Comment content cannot be empty.'
      } else {
        commentError.value = msg
      }
      return
    }

    commentSuccess.value = 'Comment submitted for review.'
    commentContent.value = ''
    commentLocation.value = ''
  } catch {
    commentError.value = 'Failed to submit comment. Please try again later.'
  } finally {
    commentSubmitting.value = false
  }
}

const buildShareUrl = (photo: PicturePhoto): string => {
  if (typeof window === 'undefined') return ''
  try {
    const url = new URL(window.location.href)
    url.searchParams.set('photo', photo.id)
    url.searchParams.set('kind', photo.kind)
    return url.toString()
  } catch {
    return ''
  }
}

const copyShareLink = async () => {
  shareMessage.value = null

  const current = activePhoto.value
  if (!current) {
    shareMessage.value = 'No photo selected.'
    return
  }

  const shareUrl = buildShareUrl(current)
  if (!shareUrl) {
    shareMessage.value = 'Unable to build link.'
    return
  }

  shareCopying.value = true

  try {
    if (typeof navigator !== 'undefined' && navigator.clipboard && navigator.clipboard.writeText) {
      await navigator.clipboard.writeText(shareUrl)
    } else if (typeof document !== 'undefined') {
      const textarea = document.createElement('textarea')
      textarea.value = shareUrl
      textarea.style.position = 'fixed'
      textarea.style.opacity = '0'
      document.body.appendChild(textarea)
      textarea.select()
      document.execCommand('copy')
      document.body.removeChild(textarea)
    }

    shareMessage.value = 'Link copied.'

    setTimeout(() => {
      if (shareMessage.value === 'Link copied.') {
        shareMessage.value = null
      }
    }, 2500)
  } catch {
    shareMessage.value = 'Failed to copy link.'
  } finally {
    shareCopying.value = false
  }
}
</script>

<style scoped>
.company-page__title {
  margin: 0 0 0.75rem;
  font-size: var(--tz-type-page-title);
  line-height: 1.18;
  font-weight: 600;
  color: var(--tz-text-primary);
}

.company-page__intro {
  margin: 0 0 0.75rem;
  font-size: 0.95rem;
  color: var(--tz-text-secondary);
}

.picture-upload-card {
  border: 1px solid var(--tz-border-subtle);
  background: var(--tz-card-surface);
  box-shadow: 0 3px 9px rgb(15 23 42 / 0.08);
}

.picture-upload-control {
  box-sizing: border-box;
  border: 1px solid var(--tz-form-control-border);
  color: var(--tz-text-primary);
  background: var(--tz-form-control-surface);
  box-shadow: none;
}

.picture-upload-control::placeholder {
  color: var(--tz-text-muted);
}

.picture-upload-control:focus {
  outline: none;
  border-color: var(--tz-form-control-focus-border);
  box-shadow: 0 0 0 1px var(--tz-form-control-focus-ring);
}

.picture-upload-file {
  color: var(--tz-text-secondary);
}

.picture-upload-file::file-selector-button {
  margin-right: 0.5rem;
  border: 1px solid var(--tz-border-strong);
  border-radius: 0.375rem;
  background: var(--tz-surface-subtle);
  padding: 0.375rem 0.75rem;
  color: var(--tz-text-primary);
  font-size: 0.75rem;
  transition: background-color 160ms ease, border-color 160ms ease;
}

.picture-upload-file::file-selector-button:hover {
  border-color: var(--tz-site-accent);
  background: var(--tz-site-accent-soft-surface);
}

.picture-upload-picker {
  border: 1px solid var(--tz-form-control-border);
  border-radius: 0.75rem;
  background: var(--tz-form-control-surface);
  padding: 0.625rem;
}

.picture-upload-file-trigger {
  border: 1px solid var(--tz-border-strong);
  background: var(--tz-surface-subtle);
  color: var(--tz-text-primary);
  transition: background-color 160ms ease, border-color 160ms ease;
}

.picture-upload-file-trigger:hover:not(:disabled) {
  border-color: var(--tz-site-accent);
  background: var(--tz-site-accent-soft-surface);
}

.picture-upload-preview {
  min-width: 0;
}

.picture-upload-preview-remove {
  position: absolute;
  top: 0.35rem;
  right: 0.35rem;
  display: inline-flex;
  height: 1.5rem;
  width: 1.5rem;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 9999px;
  background: rgba(0, 0, 0, 0.68);
  color: white;
  opacity: 0;
  transition: opacity 160ms ease, background-color 160ms ease;
}

.picture-upload-preview:hover .picture-upload-preview-remove,
.picture-upload-preview-remove:focus-visible {
  opacity: 1;
}

.picture-upload-preview-remove:hover {
  background: rgba(220, 38, 38, 0.9);
}

.picture-upload-submit {
  color: #ffffff;
  background: var(--tz-action-primary);
  border-color: var(--tz-action-primary);
  box-shadow: 0 4px 12px -4px rgb(15 23 42 / 0.14);
  transition: background-color 160ms ease, opacity 160ms ease;
}

.picture-upload-submit:hover:not(:disabled) {
  background: var(--tz-action-primary-hover);
  border-color: var(--tz-action-primary-hover);
}

@media (max-width: 767px) {
  .picture-lightbox-panel {
    max-height: min(90vh, var(--tz-mobile-safe-viewport-height, 90vh));
  }

  .picture-lightbox-image {
    max-height: min(65vh, calc(var(--tz-mobile-safe-viewport-height, 100vh) - 12rem));
  }
}
</style>

