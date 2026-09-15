<template>
  <div class="workbench-page">
    <header class="workbench-page__header">
      <p class="workbench-page__eyebrow">WORKBENCH FEED</p>
      <h1 class="workbench-page__title">Workshop notes</h1>
      <p class="workbench-page__intro">
        Real assembly, repair, measurements, and fitment notes from the workshop floor.
      </p>
    </header>

    <div v-if="initialError" class="workbench-page__state workbench-page__state--error" role="alert">
      <Icon name="lucide:triangle-alert" aria-hidden="true" />
      <span>Workshop notes are temporarily unavailable.</span>
      <button type="button" class="workbench-page__state-action" @click="reloadInitialEntries">
        <Icon name="lucide:refresh-cw" aria-hidden="true" />
        Retry
      </button>
    </div>

    <div v-else-if="initialLoading" class="workbench-page__state">
      <Icon name="lucide:loader-circle" class="workbench-page__spinner" aria-hidden="true" />
      <span>Loading workshop notes...</span>
    </div>

    <template v-else>
      <section v-if="entries.length" class="workbench-page__feed" aria-label="Workshop notes">
        <WorkbenchFeedEntryCard
          v-for="entry in entries"
          :key="entry.id"
          :entry="entry"
          @open-media="openMedia"
          @open-product="openProduct"
        />
      </section>

      <div v-else class="workbench-page__empty">
        <Icon name="lucide:clipboard-list" aria-hidden="true" />
        <p>No workshop notes have been published yet.</p>
      </div>

      <div ref="loadMoreSentinel" class="workbench-page__sentinel" aria-hidden="true"></div>

      <div v-if="canLoadMore" class="workbench-page__load-more">
        <button
          type="button"
          class="workbench-page__load-more-button"
          :disabled="loadingMore"
          @click="loadMore"
        >
          <Icon
            v-if="loadingMore"
            name="lucide:loader-circle"
            class="workbench-page__spinner"
            aria-hidden="true"
          />
          <span>{{ loadingMore ? 'Loading...' : 'Load more notes' }}</span>
        </button>
      </div>

      <p v-else-if="entries.length" class="workbench-page__end">
        You are up to date.
      </p>
    </template>

    <WorkbenchFeedLightbox
      :open="isLightboxOpen"
      :media="activeEntry?.media || []"
      :active-media="activeMedia"
      @close="closeLightbox"
      @previous="moveMedia(-1)"
      @next="moveMedia(1)"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useAsyncData, useHead, useI18n } from '#imports'
import WorkbenchFeedEntryCard from '~/features/workbench-feed/WorkbenchFeedEntryCard.vue'
import WorkbenchFeedLightbox from '~/features/workbench-feed/WorkbenchFeedLightbox.vue'
import { useWorkbenchFeedApi } from '~/features/workbench-feed/api'
import type {
  WorkbenchFeedEntry,
  WorkbenchFeedListResult,
  WorkbenchTaggedProduct,
} from '~/features/workbench-feed/types'
import { useGlobalProductDetailBottomSheet } from '~/composables/useGlobalProductDetailBottomSheet'

definePageMeta({
  layout: 'products',
  footerLabelFallback: 'Workbench Feed',
})

const { locale } = useI18n()
const feedApi = useWorkbenchFeedApi()
const lang = computed(() => String(locale.value || 'en'))
const pageSize = 10

const {
  data: initialResponse,
  error: initialError,
  pending: initialLoading,
  refresh: refreshInitial,
  clear: clearInitial,
} = await useAsyncData<WorkbenchFeedListResult | null>(
  `workbench-feed-${lang.value}`,
  () => feedApi.list({ locale: lang.value, page: 1, pageSize }),
  { watch: [lang] },
)

const entries = ref<WorkbenchFeedEntry[]>(initialResponse.value?.entries || [])
const pagination = ref(initialResponse.value?.pagination || {
  page: 1,
  page_size: pageSize,
  total: 0,
  total_pages: 0,
})
const loadingMore = ref(false)
const activeEntry = ref<WorkbenchFeedEntry | null>(null)
const activeMediaIndex = ref(0)
const loadMoreSentinel = ref<HTMLElement | null>(null)
let intersectionObserver: IntersectionObserver | null = null

const isLightboxOpen = computed(() => Boolean(activeEntry.value))
const activeMedia = computed(() => activeEntry.value?.media[activeMediaIndex.value] || null)
const canLoadMore = computed(() => pagination.value.page < pagination.value.total_pages)

watch(initialResponse, (response) => {
  entries.value = response?.entries || []
  pagination.value = response?.pagination || {
    page: 1,
    page_size: pageSize,
    total: 0,
    total_pages: 0,
  }
})

watch(lang, () => {
  activeEntry.value = null
  activeMediaIndex.value = 0
})

const reloadInitialEntries = async () => {
  clearInitial()
  await refreshInitial()
}

const loadMore = async () => {
  if (!canLoadMore.value || loadingMore.value) return

  loadingMore.value = true
  try {
    const nextPage = pagination.value.page + 1
    const result = await feedApi.list({
      locale: lang.value,
      page: nextPage,
      pageSize,
    })
    entries.value = [...entries.value, ...result.entries]
    pagination.value = result.pagination
  } finally {
    loadingMore.value = false
  }
}

const openMedia = (payload: { entry: WorkbenchFeedEntry; index: number }) => {
  activeEntry.value = payload.entry
  activeMediaIndex.value = Math.max(0, Math.min(payload.index, payload.entry.media.length - 1))
}

const closeLightbox = () => {
  activeEntry.value = null
  activeMediaIndex.value = 0
}

const moveMedia = (direction: 1 | -1) => {
  const mediaCount = activeEntry.value?.media.length || 0
  if (mediaCount < 2) return
  activeMediaIndex.value = (activeMediaIndex.value + direction + mediaCount) % mediaCount
}

const openProduct = (product: WorkbenchTaggedProduct) => {
  if (!product.product_slug) return
  const { openGlobalProductDetailBottomSheet } = useGlobalProductDetailBottomSheet()
  openGlobalProductDetailBottomSheet({
    id: product.product_id,
    variantId: product.variant_id,
    slug: product.product_slug,
    title: product.display_title,
  })
}

onMounted(() => {
  intersectionObserver = new IntersectionObserver((observations) => {
    if (observations.some(observation => observation.isIntersecting)) {
      void loadMore()
    }
  }, { rootMargin: '360px 0px' })

  if (loadMoreSentinel.value) {
    intersectionObserver.observe(loadMoreSentinel.value)
  }
})

onBeforeUnmount(() => {
  intersectionObserver?.disconnect()
  intersectionObserver = null
})

useHead(() => ({
  title: 'Workbench Feed | Workshop notes',
  meta: [
    {
      name: 'description',
      content: 'Real workshop assembly, repair, measurement, and fitment notes.',
    },
  ],
}))
</script>

<style scoped>
.workbench-page {
  width: min(100%, 48rem);
  margin: 0 auto;
  padding: 1rem 1rem 4rem;
}

.workbench-page__header {
  display: grid;
  gap: 0.45rem;
  padding: 0.25rem 0 1.35rem;
}

.workbench-page__eyebrow {
  margin: 0;
  color: var(--tz-text-muted);
  font: 800 0.68rem/1.2 var(--tz-font-ui);
  letter-spacing: 0.12em;
}

.workbench-page__title {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: clamp(1.65rem, 5vw, 2.35rem);
  font-weight: 850;
  line-height: 1.05;
}

.workbench-page__intro {
  max-width: 38rem;
  margin: 0;
  color: var(--tz-text-secondary);
  font-size: 0.9rem;
  line-height: 1.6;
}

.workbench-page__feed {
  border-top: 1px solid var(--tz-border-subtle);
}

.workbench-page__state,
.workbench-page__empty {
  display: grid;
  min-height: 18rem;
  place-items: center;
  align-content: center;
  gap: 0.7rem;
  color: var(--tz-text-muted);
  font-size: 0.82rem;
  text-align: center;
}

.workbench-page__state--error {
  color: var(--tz-status-danger-text);
}

.workbench-page__state-action,
.workbench-page__load-more-button {
  display: inline-flex;
  min-height: 2.35rem;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  padding: 0.45rem 0.75rem;
  border: 1px solid var(--tz-border-strong);
  border-radius: 999px;
  color: var(--tz-text-primary);
  background: var(--tz-surface-subtle);
  font-size: 0.75rem;
  font-weight: 800;
}

.workbench-page__load-more-button:hover:not(:disabled),
.workbench-page__state-action:hover {
  border-color: #059669;
  color: #047857;
}

.workbench-page__load-more-button:disabled {
  cursor: wait;
  opacity: 0.65;
}

.workbench-page__load-more {
  display: flex;
  justify-content: center;
  padding: 1.5rem 0 0.5rem;
}

.workbench-page__end {
  margin: 1.5rem 0 0;
  color: var(--tz-text-muted);
  font-size: 0.72rem;
  text-align: center;
}

.workbench-page__sentinel {
  height: 1px;
}

.workbench-page__spinner {
  animation: workbench-spin 0.85s linear infinite;
}

@keyframes workbench-spin {
  to { transform: rotate(360deg); }
}

@media (min-width: 768px) {
  .workbench-page {
    padding-top: 1.75rem;
  }
}
</style>
