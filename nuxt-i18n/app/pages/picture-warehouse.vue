<template>
  <div>
    <!-- SEO Title: Visually hidden but present for search engines and screen readers -->
    <h1 class="sr-only">Picture warehouse</h1>

    <!-- Tabs -->
    <div class="company-tabs" role="tablist">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        type="button"
        class="company-tabs__item"
        :class="{ 'company-tabs__item--active': activeTab === tab.id }"
        @click="setActiveTab(tab.id)"
      >
        {{ tab.label }}
      </button>
    </div>

    <!-- Riders Tab -->
    <section v-show="activeTab === 'riders'" class="mt-4 space-y-3">
      <div>
        <h2 class="mb-1 text-sm font-semibold text-white">Riders photos</h2>
        <p class="mb-3 text-xs text-slate-400">
          Real builds from riders around the world.
        </p>

        <div class="min-h-[60px]">
          <!-- Riders 加载状态 -->
          <div v-if="userLoading" class="grid gap-3 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
            <div
              v-for="n in 3"
              :key="n"
              class="flex flex-col overflow-hidden rounded-xl bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.96),rgba(15,23,42,0.98))] shadow-[0_3px_9px_rgba(0,0,0,0.9)] backdrop-blur-md"
            >
              <div class="aspect-square w-full bg-slate-800/80 animate-pulse"></div>
              <div class="px-2.5 py-2 flex flex-col gap-1">
                <div class="h-2.5 w-3/4 rounded bg-slate-700/80 animate-pulse"></div>
                <div class="h-2 w-1/2 rounded bg-slate-800/80 animate-pulse"></div>
              </div>
            </div>
          </div>

          <template v-else>
            <!-- Riders 无数据 -->
            <p v-if="!userPhotos.length" class="text-xs text-slate-400">
              No rider photos published yet.
            </p>

            <!-- Riders 数据（真实或占位） -->
            <div v-if="userPhotos.length" class="space-y-2">
              <div class="grid gap-3 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
                <button
                  v-for="(photo, index) in visibleUserPhotos"
                  :key="photo.id"
                  type="button"
                  class="group flex flex-col overflow-hidden rounded-xl bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.96),rgba(15,23,42,0.98))] shadow-[0_3px_9px_rgba(0,0,0,0.9)] backdrop-blur-md hover:shadow-[0_4px_12px_rgba(0,0,0,0.9)] transition-all"
                  @click="openLightbox('user', index)"
                >
                  <div
                    class="aspect-square w-full bg-slate-800/90 group-hover:bg-slate-700/90 transition-colors"
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
    </section>

    <!-- Brand Tab -->
    <section v-show="activeTab === 'brand'" class="mt-4 space-y-3">
      <div>
        <h2 class="mb-1 text-sm font-semibold text-white">Tanzanite photos</h2>
        <p class="mb-3 text-xs text-slate-400">
          Official product and detail shots curated by the Tanzanite team.
        </p>

        <div class="min-h-[60px]">
          <!-- 加载状态 -->
          <div v-if="brandLoading" class="grid gap-3 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
            <div
              v-for="n in 3"
              :key="n"
              class="flex flex-col overflow-hidden rounded-xl bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.96),rgba(15,23,42,0.98))] shadow-[0_3px_9px_rgba(0,0,0,0.9)] backdrop-blur-md"
            >
              <div class="aspect-square w-full bg-slate-900/80 animate-pulse"></div>
              <div class="px-2.5 py-2 flex flex-col gap-1">
                <div class="h-2.5 w-3/4 rounded bg-slate-700/80 animate-pulse"></div>
                <div class="h-2 w-1/2 rounded bg-slate-800/80 animate-pulse"></div>
              </div>
            </div>
          </div>

          <template v-else>
            <!-- 无数据 -->
            <p v-if="!brandPhotos.length" class="text-xs text-slate-400">
              No brand photos published yet.
            </p>

            <!-- 数据（真实或占位） -->
            <div v-if="brandPhotos.length" class="space-y-2">
              <div class="grid gap-3 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
                <button
                  v-for="(photo, index) in visibleBrandPhotos"
                  :key="photo.id"
                  type="button"
                  class="group flex flex-col overflow-hidden rounded-xl bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.96),rgba(15,23,42,0.98))] shadow-[0_3px_9px_rgba(0,0,0,0.9)] backdrop-blur-md hover:shadow-[0_4px_12px_rgba(0,0,0,0.9)] transition-all"
                  @click="openLightbox('brand', index)"
                >
                  <div
                    class="aspect-square w-full bg-slate-900/90 group-hover:bg-slate-800/90 transition-colors"
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

    <!-- 上传表单（Phase 3：调用 /tanz-photo/v1/upload） - 移至底部通栏 -->
    <section class="mt-10 border-t border-white/10 pt-8">
      <div class="mx-auto max-w-3xl rounded-2xl px-4 py-4 sm:px-6 sm:py-5 bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.96),rgba(15,23,42,0.98))] backdrop-blur-xl shadow-[0_3px_9px_rgba(0,0,0,0.9)]">
        <div class="mb-4 text-center">
          <h4 class="text-sm font-semibold text-slate-100">
            Share your build (login required)
          </h4>
          <p class="text-xs text-slate-400">
            Join the gallery! WEBP only, longest side up to 800px. Uploaded photos will appear after review.
          </p>
        </div>
        
        <form class="space-y-3" @submit.prevent="submitUpload">
          <div class="grid gap-3 sm:grid-cols-2">
            <div class="flex flex-col gap-1">
              <label class="text-[11px] text-slate-300">
                Region <span class="text-red-400">*</span>
              </label>
              <input
                v-model="uploadRegion"
                type="text"
                class="h-8 rounded-lg px-2.5 text-xs text-slate-100 bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] border-none shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(0,0,0,0.7)] focus:outline-none focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.9)]"
                placeholder="e.g. Germany"
                required
              />
            </div>
            <div class="flex flex-col gap-1">
              <label class="text-[11px] text-slate-300">Location</label>
              <input
                v-model="uploadLocation"
                type="text"
                class="h-8 rounded-lg px-2.5 text-xs text-slate-100 bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] border-none shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(0,0,0,0.7)] focus:outline-none focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.9)]"
                placeholder="e.g. Berlin"
              />
            </div>
          </div>

          <div class="grid gap-3 sm:grid-cols-2">
            <div class="flex flex-col gap-1">
              <label class="text-[11px] text-slate-300">Nickname</label>
              <input
                v-model="uploadNickname"
                type="text"
                class="h-8 rounded-lg px-2.5 text-xs text-slate-100 bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] border-none shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(0,0,0,0.7)] focus:outline-none focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.9)]"
                placeholder="Your name or handle"
              />
            </div>
            <div class="flex flex-col gap-1">
              <label class="text-[11px] text-slate-300">Bike / wheelset</label>
              <input
                v-model="uploadBikeModel"
                type="text"
                class="h-8 rounded-lg px-2.5 text-xs text-slate-100 bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] border-none shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(0,0,0,0.7)] focus:outline-none focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.9)]"
                placeholder="Model info"
              />
            </div>
          </div>

          <div class="flex flex-col gap-1">
            <label class="text-[11px] text-slate-300">Notes</label>
            <textarea
              v-model="uploadNotes"
              rows="2"
              class="rounded-lg px-2.5 py-1.5 text-xs text-slate-100 bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] border-none shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(0,0,0,0.7)] focus:outline-none focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.9)]"
              placeholder="Tell us about your build..."
            ></textarea>
          </div>

          <div class="flex flex-col gap-1">
            <label class="text-[11px] text-slate-300">Photos (WEBP, Max 10)</label>
            <input
              type="file"
              accept="image/webp"
              multiple
              @change="onUploadFileChange"
              class="block w-full text-xs text-slate-200 file:mr-2 file:rounded file:border-0 file:bg-white/10 file:px-3 file:py-1.5 file:text-xs file:text-slate-100 hover:file:bg-white/20 transition-colors"
            />
          </div>

          <div class="flex items-center justify-between gap-2 pt-2 border-t border-white/10 mt-2">
            <div class="flex-1">
               <p v-if="uploadSuccess" class="text-[11px] text-emerald-400">
                {{ uploadSuccess }}
              </p>
              <p v-else-if="uploadError" class="text-[11px] text-red-400">
                {{ uploadError }}
              </p>
            </div>
            <button
              type="submit"
              class="inline-flex items-center justify-center h-9 rounded-full bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] px-5 text-xs font-semibold text-slate-900 disabled:cursor-not-allowed disabled:opacity-50 hover:brightness-110 transition-all shadow-[0_4px_12px_-4px_rgba(0,0,0,0.95)]"
              :disabled="uploading"
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
            <div class="relative flex-1 flex items-center justify-center bg-slate-900 p-4 overflow-hidden">
              <button
                type="button"
                class="absolute left-2 sm:left-4 z-20 inline-flex h-10 w-10 items-center justify-center rounded-full bg-black/40 text-white text-2xl hover:bg-black/70 transition-colors"
                @click.stop="goPrev"
                aria-label="Previous photo"
              >
                ‹
              </button>

              <div class="relative w-full h-full flex flex-col items-center justify-center">
                <!-- 大图 -->
                <div class="relative flex items-center justify-center w-full h-full">
                  <img
                    v-if="currentImageUrl"
                    :src="currentImageUrl"
                    alt="Photo"
                    class="max-w-full max-h-[65vh] object-contain rounded shadow-lg"
                  />
                  <div v-else class="w-full max-w-[500px] aspect-square bg-slate-800 rounded flex items-center justify-center text-slate-500">
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
                    :class="idx === currentGalleryIndex ? 'border-sky-400 opacity-100' : 'border-transparent opacity-60 hover:opacity-90'"
                    @click.stop="currentGalleryIndex = idx"
                  >
                    <img :src="img" class="w-full h-full object-cover" loading="lazy" />
                  </button>
                </div>
              </div>

              <button
                type="button"
                class="absolute right-2 sm:right-4 z-20 inline-flex h-10 w-10 items-center justify-center rounded-full bg-black/40 text-white text-2xl hover:bg-black/70 transition-colors"
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
                      class="rounded-lg px-2.5 py-1.5 bg-[radial-gradient(circle_at_top_left,rgba(31,41,55,0.96),rgba(15,23,42,0.98))] shadow-[0_3px_9px_rgba(0,0,0,0.9)] backdrop-blur-md"
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
                      class="w-full rounded-lg px-2 py-1 text-[11px] text-slate-100 bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] border-none shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(0,0,0,0.7)] focus:outline-none focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.9)]"
                      placeholder="Write a comment (login required)"
                    ></textarea>
                    <div class="grid grid-cols-[minmax(0,1.7fr)_minmax(0,1.1fr)] gap-1.5 items-center">
                      <input
                        v-model="commentLocation"
                        type="text"
                        class="h-7 rounded-lg px-2 text-[10px] text-slate-100 bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] border-none shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(0,0,0,0.7)] focus:outline-none focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.9)]"
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
import { computed, onMounted, ref, watch } from 'vue'
import { definePageMeta, useHead, useRoute } from '#imports'

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
  galleryImages?: string[]
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
const uploadFiles = ref<File[]>([])
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
    uploadFiles.value = []
    return
  }
  uploadFiles.value = Array.from(target.files)
}

const submitUpload = async () => {
  uploadError.value = null
  uploadSuccess.value = null

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

  uploading.value = true

  try {
    const formData = new FormData()
    
    // Append multiple files. Using 'file[]' ensures PHP treats it as an array.
    uploadFiles.value.forEach((file) => {
      formData.append('file[]', file)
    })
    
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

    uploadSuccess.value = 'Photos submitted for review.'
    uploadFiles.value = []
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
        galleryImages: Array.isArray(item.gallery_images) ? item.gallery_images : [],
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
const currentGalleryIndex = ref(0)

type PictureWarehouseTabId = 'riders' | 'brand'

const tabs: Array<{ id: PictureWarehouseTabId; label: string }> = [
  { id: 'riders', label: 'Riders photos' },
  { id: 'brand', label: 'Tanzanite photos' },
]
const activeTab = ref<PictureWarehouseTabId>('riders')

const route = useRoute()

const getTabFromHash = (hash: string | null | undefined): PictureWarehouseTabId | null => {
  if (!hash) return null
  const raw = hash.startsWith('#') ? hash.slice(1) : hash
  const allowed: PictureWarehouseTabId[] = ['riders', 'brand']
  return (allowed as string[]).includes(raw) ? (raw as PictureWarehouseTabId) : null
}

watch(
  () => route.hash,
  (hash) => {
    const next = getTabFromHash(hash)
    if (next) activeTab.value = next
  },
  { immediate: true }
)

const setActiveTab = (id: PictureWarehouseTabId) => {
  activeTab.value = id
  if (typeof window !== 'undefined') {
    const url = new URL(window.location.href)
    url.hash = `#${id}`
    window.history.replaceState(null, '', url.toString())
  }
}

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

const currentImageUrl = computed(() => {
  const photo = activePhoto.value
  if (!photo) return ''
  // Prefer gallery image at current index, fallback to first gallery image, fallback to placeholder
  if (photo.galleryImages && photo.galleryImages.length > 0) {
    return photo.galleryImages[currentGalleryIndex.value] ?? photo.galleryImages[0]
  }
  return '' // Should handle empty case or show a placeholder if needed
})

const openLightbox = (kind: PhotoKind, index: number) => {
  activeKind.value = kind
  activeIndex.value = index
  currentGalleryIndex.value = 0
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
  currentGalleryIndex.value = 0
  void loadCommentsForActivePhoto()
}

const goPrev = () => {
  const list = activeList.value
  if (!list || !list.length || activeIndex.value === null) return
  const prevIndex = (activeIndex.value - 1 + list.length) % list.length
  activeIndex.value = prevIndex
  currentGalleryIndex.value = 0
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

/* Tabs 样式（对齐 ourstory） */
.company-tabs {
  display: flex;
  overflow-x: auto;
  gap: 12px;
  padding: 4px 16px;
  margin: 0 -16px 1rem;
  max-width: calc(100% + 32px);
  -webkit-overflow-scrolling: touch;
  scrollbar-width: none;
  touch-action: pan-x;
}

.company-tabs::-webkit-scrollbar {
  display: none;
}

.company-tabs__item {
  flex-shrink: 0;
  border: none;
  border-radius: 9999px;
  padding: 8px 18px;
  font-size: 0.85rem;
  font-weight: 500;
  color: #ffffff;
  background: rgba(31, 41, 55, 0.9);
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  white-space: nowrap;
  backdrop-filter: blur(4px);
  box-shadow: 6px 8px 18px -12px rgba(0, 0, 0, 0.85);
}

.company-tabs__item:active {
  transform: scale(0.96);
}

.company-tabs__item:hover {
  background: rgba(51, 65, 85, 0.95);
  color: #ffffff;
}

.company-tabs__item--active {
  background: #ffffff;
  color: #0f172a;
  border: none;
  font-weight: 600;
  box-shadow: 8px 10px 22px -10px rgba(0, 0, 0, 0.9);
}

@media (min-width: 768px) {
  .company-tabs {
    flex-wrap: wrap;
    justify-content: center;
    margin: 0 0 1rem;
    padding: 4px 0;
    max-width: 100%;
  }
}

@media (max-width: 768px) {
  .company-tabs {
    justify-content: flex-start;
  }
}
</style>
