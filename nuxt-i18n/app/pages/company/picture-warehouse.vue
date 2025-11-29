<template>
  <div>
    <h2 class="company-page__title">Picture warehouse</h2>
    <p class="company-page__intro">
      Browse rider photos and official Tanzanite shots. Upload and review flows will be added
      later; this version focuses on layout and navigation.
    </p>

    <!-- 两栏照片墙：左侧用户照片，右侧品牌/官方照片（使用 mock 数据） -->
    <section
      class="mt-4 grid gap-6 lg:grid-cols-[minmax(0,1.4fr)_minmax(0,1.2fr)] lg:items-start"
    >
      <!-- 左栏：用户照片 -->
      <div>
        <h3 class="mb-1 text-sm font-semibold text-white">Riders photos</h3>
        <p class="mb-3 text-xs text-slate-400">
          Real builds from riders around the world. Uploaded photos will appear here after
          approval.
        </p>

        <!-- 上传表单（Phase 3：调用 /tanz-photo/v1/upload） -->
        <div class="mb-4 rounded-xl border border-white/10 bg-white/5 px-3 py-3">
          <h4 class="mb-1 text-[11px] font-semibold text-slate-100">
            Share your build (login required)
          </h4>
          <p class="mb-2 text-[10px] text-slate-400">
            WEBP only, longest side up to 800px. Uploaded photos will appear after review.
          </p>
          <form class="space-y-2" @submit.prevent="submitUpload">
            <div class="flex flex-col gap-1">
              <label class="text-[10px] text-slate-300">
                Region <span class="text-red-400">*</span>
              </label>
              <input
                v-model="uploadRegion"
                type="text"
                class="h-7 rounded border border-white/15 bg-slate-900/60 px-2 text-[11px] text-slate-100 focus:border-sky-400 focus:outline-none"
                required
              />
            </div>
            <div class="grid grid-cols-2 gap-2">
              <div class="flex flex-col gap-1">
                <label class="text-[10px] text-slate-300">Location</label>
                <input
                  v-model="uploadLocation"
                  type="text"
                  class="h-7 rounded border border-white/15 bg-slate-900/60 px-2 text-[11px] text-slate-100 focus:border-sky-400 focus:outline-none"
                />
              </div>
              <div class="flex flex-col gap-1">
                <label class="text-[10px] text-slate-300">Nickname</label>
                <input
                  v-model="uploadNickname"
                  type="text"
                  class="h-7 rounded border border-white/15 bg-slate-900/60 px-2 text-[11px] text-slate-100 focus:border-sky-400 focus:outline-none"
                />
              </div>
            </div>
            <div class="flex flex-col gap-1">
              <label class="text-[10px] text-slate-300">Bike / wheelset</label>
              <input
                v-model="uploadBikeModel"
                type="text"
                class="h-7 rounded border border-white/15 bg-slate-900/60 px-2 text-[11px] text-slate-100 focus:border-sky-400 focus:outline-none"
              />
            </div>
            <div class="flex flex-col gap-1">
              <label class="text-[10px] text-slate-300">Notes</label>
              <textarea
                v-model="uploadNotes"
                rows="2"
                class="rounded border border-white/15 bg-slate-900/60 px-2 py-1 text-[11px] text-slate-100 focus:border-sky-400 focus:outline-none"
              ></textarea>
            </div>
            <div class="flex flex-col gap-1">
              <label class="text-[10px] text-slate-300">Photo (WEBP, ≤ 5MB)</label>
              <input
                type="file"
                accept="image/webp"
                @change="onUploadFileChange"
                class="block w-full text-[10px] text-slate-200 file:mr-2 file:rounded file:border-0 file:bg-white/10 file:px-2 file:py-1 file:text-[10px] file:text-slate-100"
              />
            </div>
            <div class="flex items-center justify-between gap-2 pt-1">
              <button
                type="submit"
                class="inline-flex items-center justify-center rounded-full bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] px-3.5 py-1 text-[11px] font-semibold text-slate-950 disabled:cursor-not-allowed disabled:opacity-50"
                :disabled="uploading"
              >
                <span v-if="uploading">Uploading…</span>
                <span v-else>Submit for review</span>
              </button>
              <p v-if="uploadSuccess" class="text-[10px] text-emerald-400">
                {{ uploadSuccess }}
              </p>
              <p v-else-if="uploadError" class="text-[10px] text-red-400">
                {{ uploadError }}
              </p>
            </div>
          </form>
        </div>

        <div class="min-h-[60px]">
          <!-- Riders 加载状态 -->
          <div v-if="userLoading" class="grid gap-3 sm:grid-cols-2">
            <div
              v-for="n in 3"
              :key="n"
              class="flex flex-col overflow-hidden rounded-xl border border-white/10 bg-white/5"
            >
              <div class="aspect-[4/5] w-full bg-slate-800/70 animate-pulse"></div>
              <div class="px-2.5 py-2 flex flex-col gap-1">
                <div class="h-2.5 w-3/4 rounded bg-slate-700/70 animate-pulse"></div>
                <div class="h-2 w-1/2 rounded bg-slate-800/70 animate-pulse"></div>
              </div>
            </div>
          </div>

          <template v-else>
            <!-- Riders 加载失败提示，但仍可显示占位数据 -->
            <p v-if="userError" class="mb-1 text-xs text-slate-400">
              Unable to load rider photos from the server. Showing local sample data.
            </p>

            <!-- Riders 无数据 -->
            <p v-else-if="!userPhotos.length" class="text-xs text-slate-400">
              No rider photos published yet.
            </p>

            <!-- Riders 数据（真实或占位） -->
            <div v-if="userPhotos.length" class="space-y-2">
              <div class="grid gap-3 sm:grid-cols-2">
                <button
                  v-for="(photo, index) in visibleUserPhotos"
                  :key="photo.id"
                  type="button"
                  class="group flex flex-col overflow-hidden rounded-xl border border-white/10 bg-white/5 hover:border-white/25 hover:bg-white/10 transition-colors"
                  @click="openLightbox('user', index)"
                >
                  <div
                    class="aspect-[4/5] w-full bg-slate-800 group-hover:bg-slate-700 transition-colors"
                  ></div>
                  <div class="px-2.5 py-2 flex flex-col gap-0.5">
                    <p class="text-[11px] font-medium text-slate-100 truncate">
                      {{ photo.title }}
                    </p>
                    <p class="text-[10px] text-slate-400 truncate">
                      {{ photo.region }}<span v-if="photo.nickname"> · {{ photo.nickname }}</span>
                    </p>
                  </div>
                </button>
              </div>

              <div v-if="hasMoreUserPhotos" class="pt-1">
                <button
                  type="button"
                  class="text-[10px] text-sky-300 hover:text-sky-200 underline underline-offset-2"
                  @click="toggleUserPhotos"
                >
                  {{ showAllUserPhotos ? 'Show fewer photos' : 'Show more photos' }}
                </button>
              </div>
            </div>
          </template>
        </div>
      </div>

      <!-- 右栏：品牌/官方照片 -->
      <div>
        <h3 class="mb-1 text-sm font-semibold text-white">Tanzanite photos</h3>
        <p class="mb-3 text-xs text-slate-400">
          Official product and detail shots curated by the Tanzanite team.
        </p>

        <div class="min-h-[60px]">
          <!-- 加载状态 -->
          <div v-if="brandLoading" class="grid gap-3 sm:grid-cols-2">
            <div
              v-for="n in 3"
              :key="n"
              class="flex flex-col overflow-hidden rounded-xl border border-white/10 bg-white/5"
            >
              <div class="aspect-[4/5] w-full bg-slate-900/70 animate-pulse"></div>
              <div class="px-2.5 py-2 flex flex-col gap-1">
                <div class="h-2.5 w-3/4 rounded bg-slate-700/70 animate-pulse"></div>
                <div class="h-2 w-1/2 rounded bg-slate-800/70 animate-pulse"></div>
              </div>
            </div>
          </div>

          <template v-else>
            <!-- 加载失败提示，但仍可显示占位数据 -->
            <p v-if="brandError" class="mb-1 text-xs text-slate-400">
              Unable to load brand photos from the server. Showing local sample data.
            </p>

            <!-- 无数据 -->
            <p v-else-if="!brandPhotos.length" class="text-xs text-slate-400">
              No brand photos published yet.
            </p>

            <!-- 数据（真实或占位） -->
            <div v-if="brandPhotos.length" class="space-y-2">
              <div class="grid gap-3 sm:grid-cols-2">
                <button
                  v-for="(photo, index) in visibleBrandPhotos"
                  :key="photo.id"
                  type="button"
                  class="group flex flex-col overflow-hidden rounded-xl border border-white/10 bg-white/5 hover:border-white/25 hover:bg-white/10 transition-colors"
                  @click="openLightbox('brand', index)"
                >
                  <div
                    class="aspect-[4/5] w-full bg-slate-900 group-hover:bg-slate-800 transition-colors"
                  ></div>
                  <div class="px-2.5 py-2 flex flex-col gap-0.5">
                    <p class="text-[11px] font-medium text-slate-100 truncate">
                      {{ photo.title }}
                    </p>
                    <p class="text-[10px] text-slate-400 truncate">
                      {{ photo.region }}<span v-if="photo.nickname"> · {{ photo.nickname }}</span>
                    </p>
                  </div>
                </button>
              </div>

              <div v-if="hasMoreBrandPhotos" class="pt-1">
                <button
                  type="button"
                  class="text-[10px] text-sky-300 hover:text-sky-200 underline underline-offset-2"
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

    <!-- 单张图片详情弹窗（Phase 1/2：仅大图 + 标题 + 左右切换 + 关闭，评论/分享/推荐为 UI 占位） -->
    <teleport to="body">
      <transition
        enter-active-class="transition-opacity duration-200 ease-out"
        leave-active-class="transition-opacity duration-150 ease-in"
        enter-from-class="opacity-0"
        leave-to-class="opacity-0"
      >
        <div
          v-if="isLightboxOpen"
          class="fixed inset-0 z-[1400] flex items-center justify-center bg-black/75 px-3"
          @click.self="closeLightbox"
        >
          <div
            class="relative w-full max-w-[960px] max-h-[90vh] bg-slate-950 rounded-2xl flex flex-col overflow-hidden md:overflow-y-auto"
          >
            <!-- 顶部标题 + 关闭按钮 -->
            <header
              class="px-4 py-3 flex items-center justify-between border-b border-white/10 bg-slate-950/90"
            >
              <h3 class="text-sm sm:text-base font-semibold text-white truncate">
                {{ activePhoto?.title || 'Picture' }}
              </h3>
              <button
                type="button"
                class="ml-4 inline-flex h-7 w-7 items-center justify-center rounded-full bg-white/10 text-white text-sm hover:bg-white/20"
                @click="closeLightbox"
                aria-label="Close"
              >
                ×
              </button>
            </header>

            <!-- 中部：大图区域 + 左右切换 -->
            <div class="relative flex-1 flex items-center justify-center bg-slate-900">
              <button
                type="button"
                class="absolute left-3 sm:left-4 inline-flex h-8 w-8 items-center justify-center rounded-full bg-black/40 text-white text-lg hover:bg-black/70"
                @click.stop="goPrev"
                aria-label="Previous photo"
              >
                ‹
              </button>

              <div
                class="w-full max-w-[640px] aspect-[4/5] max-h-[45vh] md:max-h-[60vh] bg-slate-800 rounded-md"
              ></div>

              <button
                type="button"
                class="absolute right-3 sm:right-4 inline-flex h-8 w-8 items-center justify-center rounded-full bg-black/40 text-white text-lg hover:bg-black/70"
                @click.stop="goNext"
                aria-label="Next photo"
              >
                ›
              </button>
            </div>
            <div
              class="border-t border-white/10 grid grid-cols-1 md:grid-cols-[minmax(0,2fr)_minmax(0,1fr)] bg-slate-950/95"
            >
              <!-- 左：操作按钮栏 + 评论占位 -->
              <div class="px-4 py-3">
                <div class="mb-2 flex flex-wrap gap-2 text-[11px]">
                  <button
                    type="button"
                    class="px-3 py-1 rounded-full bg-[#1877F2] text-slate-50 border border-transparent opacity-80 cursor-default"
                    title="Coming soon"
                  >
                    Share to Facebook
                  </button>
                  <button
                    type="button"
                    class="px-3 py-1 rounded-full bg-black text-slate-50 border border-white/25 opacity-80 cursor-default"
                    title="Coming soon"
                  >
                    Share to X
                  </button>
                  <button
                    type="button"
                    class="px-3 py-1 rounded-full bg-[#FF4500] text-slate-50 border border-transparent opacity-80 cursor-default"
                    title="Coming soon"
                  >
                    Share to Reddit
                  </button>
                  <button
                    type="button"
                    class="px-3 py-1 rounded-full bg-white/10 text-slate-100 border border-white/20 hover:bg-white/20 transition-colors disabled:opacity-60 disabled:cursor-not-allowed"
                    @click="copyShareLink"
                    :disabled="shareCopying || !activePhoto"
                  >
                    <span v-if="shareCopying">Copying...</span>
                    <span v-else>Copy link</span>
                  </button>
                </div>
                <p v-if="shareMessage" class="mb-2 text-[10px] text-slate-400">
                  {{ shareMessage }}
                </p>
                <div class="rounded-lg border border-white/15 px-3 py-3 text-[11px] text-slate-200">
                  <!-- 评论列表 -->
                  <div class="mb-2 flex items-center justify-between">
                    <h4 class="font-semibold text-slate-100">Comments</h4>
                    <span v-if="commentsLoading" class="text-[10px] text-slate-400">
                      Loading...
                    </span>
                  </div>

                  <div v-if="commentsError" class="mb-2 text-[10px] text-red-400">
                    Unable to load comments. Please try again later.
                  </div>

                  <div v-else-if="!commentsLoading && !comments.length" class="mb-2 text-[10px] text-slate-400">
                    No comments yet.
                  </div>

                  <ul v-else class="mb-3 space-y-2 max-h-40 overflow-y-auto pr-1">
                    <li
                      v-for="comment in comments"
                      :key="comment.id"
                      class="rounded border border-white/10 bg-slate-900/80 px-2.5 py-1.5"
                    >
                      <div class="mb-0.5 flex items-center justify-between gap-2">
                        <span class="font-semibold text-[10px] text-slate-100">
                          {{ comment.author }}
                        </span>
                        <span class="text-[9px] text-slate-500">
                          {{ comment.dateGmtFormatted }}
                        </span>
                      </div>
                      <p class="text-[10px] text-slate-200">
                        {{ comment.content }}
                      </p>
                      <p v-if="comment.location" class="mt-0.5 text-[9px] text-slate-400">
                        {{ comment.location }}
                      </p>
                    </li>
                  </ul>

                  <!-- 简单评论表单 -->
                  <form class="space-y-1.5" @submit.prevent="submitComment">
                    <textarea
                      v-model="commentContent"
                      rows="2"
                      class="w-full rounded border border-white/15 bg-slate-950/80 px-2 py-1 text-[11px] text-slate-100 focus:border-sky-400 focus:outline-none"
                      placeholder="Write a comment (login required)"
                    ></textarea>
                    <div class="grid grid-cols-[minmax(0,1.7fr)_minmax(0,1.1fr)] gap-1.5 items-center">
                      <input
                        v-model="commentLocation"
                        type="text"
                        class="h-7 rounded border border-white/15 bg-slate-950/80 px-2 text-[10px] text-slate-100 focus:border-sky-400 focus:outline-none"
                        placeholder="Location (optional)"
                      />
                      <div class="flex items-center justify-end gap-2">
                        <button
                          type="submit"
                          class="inline-flex items-center justify-center rounded-full bg-white/15 px-3 py-1 text-[10px] font-semibold text-slate-50 disabled:cursor-not-allowed disabled:opacity-50"
                          :disabled="commentSubmitting || !activePhoto"
                        >
                          <span v-if="commentSubmitting">Sending...</span>
                          <span v-else>Post comment</span>
                        </button>
                      </div>
                    </div>
                    <p v-if="commentSuccess" class="text-[10px] text-emerald-400">
                      {{ commentSuccess }}
                    </p>
                    <p v-else-if="commentError" class="text-[10px] text-red-400">
                      {{ commentError }}
                    </p>
                  </form>
                </div>
              </div>

              <!-- 右：推荐模块 -->
              <div
                class="px-4 py-3 border-t md:border-t-0 md:border-l border-white/10 text-[11px] text-slate-300"
              >
                <div class="mb-2 font-semibold text-slate-100">Like This? Get The Same Build.</div>
                <div class="space-y-2">
                  <div>
                    <p class="mb-0.5 font-semibold text-slate-200">Rim</p>
                    <p v-if="activePhoto?.productRefs?.rim" class="text-slate-200 truncate">
                      <a
                        :href="activePhoto?.productRefs?.rim"
                        target="_blank"
                        rel="noopener noreferrer"
                        class="underline underline-offset-2 hover:text-sky-300"
                      >
                        {{ activePhoto?.productRefs?.rim }}
                      </a>
                    </p>
                  </div>

                  <div>
                    <p class="mb-0.5 font-semibold text-slate-200">Wheel(s)</p>
                    <p v-if="activePhoto?.productRefs?.wheel" class="text-slate-200 truncate">
                      <a
                        :href="activePhoto?.productRefs?.wheel"
                        target="_blank"
                        rel="noopener noreferrer"
                        class="underline underline-offset-2 hover:text-sky-300"
                      >
                        {{ activePhoto?.productRefs?.wheel }}
                      </a>
                    </p>
                  </div>

                  <div>
                    <p class="mb-0.5 font-semibold text-slate-200">Hub</p>
                    <p v-if="activePhoto?.productRefs?.hub" class="text-slate-200 truncate">
                      <a
                        :href="activePhoto?.productRefs?.hub"
                        target="_blank"
                        rel="noopener noreferrer"
                        class="underline underline-offset-2 hover:text-sky-300"
                      >
                        {{ activePhoto?.productRefs?.hub }}
                      </a>
                    </p>
                  </div>

                  <div>
                    <p class="mb-0.5 font-semibold text-slate-200">Tire</p>
                    <p v-if="activePhoto?.productRefs?.tire" class="text-slate-200 truncate">
                      <a
                        :href="activePhoto?.productRefs?.tire"
                        target="_blank"
                        rel="noopener noreferrer"
                        class="underline underline-offset-2 hover:text-sky-300"
                      >
                        {{ activePhoto?.productRefs?.tire }}
                      </a>
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </transition>
    </teleport>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

definePageMeta({
  layout: 'products',
})

useHead({
  title: 'Picture warehouse',
})

type PhotoKind = 'user' | 'brand'

interface PicturePhoto {
  id: string
  kind: PhotoKind
  title: string
  region: string
  nickname?: string
  productRefs?: ProductRefs
}

interface PhotoComment {
  id: number
  author: string
  content: string
  dateGmt: string
  dateGmtFormatted: string
  location?: string
}

interface ProductRefs {
  rim?: string
  wheel?: string
  hub?: string
  tire?: string
}

const userPhotos = ref<PicturePhoto[]>([
  {
    id: 'sample-user-1',
    kind: 'user',
    title: 'Sample rider photo (dev only)',
    region: 'Sample region',
    nickname: 'Sample rider',
  },
])
const userLoading = ref(true)
const userError = ref<string | null>(null)

const uploadRegion = ref('')
const uploadLocation = ref('')
const uploadNickname = ref('')
const uploadBikeModel = ref('')
const uploadNotes = ref('')
const uploadFile = ref<File | null>(null)
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

const onUploadFileChange = (event: Event) => {
  const target = event.target as HTMLInputElement | null
  if (!target || !target.files || target.files.length === 0) {
    uploadFile.value = null
    return
  }
  uploadFile.value = target.files[0] ?? null
}

const submitUpload = async () => {
  uploadError.value = null
  uploadSuccess.value = null

  if (!uploadRegion.value.trim()) {
    uploadError.value = 'Please enter a region.'
    return
  }

  if (!uploadFile.value) {
    uploadError.value = 'Please choose a WEBP image to upload.'
    return
  }

  uploading.value = true

  try {
    const formData = new FormData()
    formData.append('file', uploadFile.value)
    formData.append('region', uploadRegion.value.trim())

    if (uploadLocation.value.trim()) formData.append('location', uploadLocation.value.trim())
    if (uploadNickname.value.trim()) formData.append('nickname', uploadNickname.value.trim())
    if (uploadBikeModel.value.trim())
      formData.append('bike_model', uploadBikeModel.value.trim())
    if (uploadNotes.value.trim()) formData.append('notes', uploadNotes.value.trim())

    const response = await fetch('/wp-json/tanz-photo/v1/upload', {
      method: 'POST',
      body: formData,
    })

    let payload: any = null
    try {
      payload = await response.json()
    } catch {
      payload = null
    }

    if (!response.ok) {
      const code = payload && (payload.code || payload?.data?.code)

      if (response.status === 401 || response.status === 403) {
        uploadError.value = 'Please log in before uploading.'
      } else if (response.status === 429) {
        uploadError.value = 'Too many uploads. Please try again later.'
      } else if (code === 'tpg_invalid_type') {
        uploadError.value = 'Only WEBP images are allowed.'
      } else if (code === 'tpg_file_too_large') {
        uploadError.value = 'File is too large. Please keep it under 5MB.'
      } else if (code === 'tpg_missing_region') {
        uploadError.value = 'Region is required.'
      } else {
        uploadError.value = 'Upload failed. Please try again later.'
      }

      return
    }

    uploadSuccess.value = 'Photo submitted for review.'
    uploadFile.value = null
    uploadBikeModel.value = ''
    uploadNotes.value = ''
  } catch {
    uploadError.value = 'Upload failed. Please try again later.'
  } finally {
    uploading.value = false
  }
}

const brandPhotos = ref<PicturePhoto[]>([
  {
    id: 'sample-brand-1',
    kind: 'brand',
    title: 'Sample brand photo (dev only)',
    region: 'Studio',
  },
])
const brandLoading = ref(true)
const brandError = ref<string | null>(null)

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

const toggleUserPhotos = () => {
  if (!hasMoreUserPhotos.value) return
  showAllUserPhotos.value = !showAllUserPhotos.value
}

const toggleBrandPhotos = () => {
  if (!hasMoreBrandPhotos.value) return
  showAllBrandPhotos.value = !showAllBrandPhotos.value
}

const mapPayloadToPhotos = (payload: any[], kind: PhotoKind): PicturePhoto[] => {
  return payload
    .map((item: any): PicturePhoto | null => {
      if (!item) return null
      const id = item.id ?? item.ID ?? null
      if (!id) return null

      let productRefs: ProductRefs | undefined
      const rawRefs = item.product_refs
      if (rawRefs && typeof rawRefs === 'object') {
        productRefs = {}
        if (typeof rawRefs.rim === 'string' && rawRefs.rim) productRefs.rim = rawRefs.rim
        if (typeof rawRefs.wheel === 'string' && rawRefs.wheel) productRefs.wheel = rawRefs.wheel
        if (typeof rawRefs.hub === 'string' && rawRefs.hub) productRefs.hub = rawRefs.hub
        if (typeof rawRefs.tire === 'string' && rawRefs.tire) productRefs.tire = rawRefs.tire

        if (!productRefs.rim && !productRefs.wheel && !productRefs.hub && !productRefs.tire) {
          productRefs = undefined
        }
      }

      return {
        id: String(id),
        kind,
        title: String(item.title ?? item.post_title ?? (kind === 'user' ? 'Rider photo' : 'Brand photo')),
        region: String(item.region ?? (kind === 'user' ? 'Unknown' : 'Studio')),
        nickname: typeof item.nickname === 'string' ? item.nickname : undefined,
        productRefs,
      }
    })
    .filter((p: PicturePhoto | null): p is PicturePhoto => p !== null)
}

const fetchUserPhotos = async () => {
  try {
    userLoading.value = true
    userError.value = null

    const response = await fetch('/wp-json/tanz-photo/v1/gallery?type=user&status=approved')
    if (!response.ok) {
      throw new Error(`Request failed with status ${response.status}`)
    }

    const payload = await response.json()
    const mapped = Array.isArray(payload) ? mapPayloadToPhotos(payload, 'user') : []

    if (mapped.length) {
      userPhotos.value = mapped
    }
  } catch (error) {
    userError.value = 'load_failed'
  } finally {
    userLoading.value = false
  }
}

const fetchBrandPhotos = async () => {
  try {
    brandLoading.value = true
    brandError.value = null

    const response = await fetch('/wp-json/tanz-photo/v1/gallery?type=brand&status=approved')

    if (!response.ok) {
      throw new Error(`Request failed with status ${response.status}`)
    }

    const payload = await response.json()
    const mapped = Array.isArray(payload) ? mapPayloadToPhotos(payload, 'brand') : []

    if (mapped.length) {
      brandPhotos.value = mapped
    }
  } catch (error) {
    brandError.value = 'load_failed'
  } finally {
    brandLoading.value = false
  }
}

onMounted(() => {
  fetchUserPhotos()
  fetchBrandPhotos()
})

const activeKind = ref<PhotoKind | null>(null)
const activeIndex = ref<number | null>(null)

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

const openLightbox = (kind: PhotoKind, index: number) => {
  activeKind.value = kind
  activeIndex.value = index
  void loadCommentsForActivePhoto()
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
  void loadCommentsForActivePhoto()
}

const goPrev = () => {
  const list = activeList.value
  if (!list || !list.length || activeIndex.value === null) return
  const prevIndex = (activeIndex.value - 1 + list.length) % list.length
  activeIndex.value = prevIndex
  void loadCommentsForActivePhoto()
}

const fetchCommentsForPhoto = async (photoId: string) => {
  try {
    commentsLoading.value = true
    commentsError.value = null

    const response = await fetch(
      `/wp-json/tanz-photo/v1/comments?photo_id=${encodeURIComponent(photoId)}&per_page=20`
    )

    if (!response.ok) {
      throw new Error(`Request failed with status ${response.status}`)
    }

    const payload = await response.json()
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

    const response = await fetch('/wp-json/tanz-photo/v1/comments', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(body),
    })

    let payload: any = null
    try {
      payload = await response.json()
    } catch {
      payload = null
    }

    if (!response.ok) {
      const code = payload && (payload.code || payload?.data?.code)

      if (response.status === 401 || response.status === 403) {
        commentError.value = 'Please log in before commenting.'
      } else if (code === 'tpg_empty_comment') {
        commentError.value = 'Comment content cannot be empty.'
      } else {
        commentError.value = 'Failed to submit comment. Please try again later.'
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
  font-size: 1.5rem;
  font-weight: 600;
  color: #f9fafb;
}

.company-page__intro {
  margin: 0 0 0.75rem;
  font-size: 0.95rem;
  color: rgba(148, 163, 184, 0.9);
}
</style>
