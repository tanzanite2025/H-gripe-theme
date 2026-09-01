<template>
  <div class="flex h-full min-h-0 flex-col gap-4 overflow-hidden">
    <AdminPageHeader
      class="shrink-0"
      title="页面留言"
      description="管理前台页面底部留言，按页面查看、审核并回复"
    >
      <template #actions>
        <Button variant="outline" size="sm" :disabled="loading || riskLoading" @click="refreshAll">
          <RefreshCw :class="['size-3.5', (loading || riskLoading) && 'animate-spin']" />
          刷新
        </Button>
      </template>
    </AdminPageHeader>

    <PageFeedbackRiskOverview
      :overview="riskOverview"
      :loading="riskLoading"
      @filter-page="filterByRiskPage"
    />

    <PageFeedbackFilters
      :filters="filters"
      :loading="loading"
      @apply="applyFilters"
      @reset="resetFilters"
    />

    <section class="grid min-h-0 flex-1 grid-cols-1 gap-3 overflow-hidden 2xl:grid-cols-[minmax(0,1fr)_430px]">
      <PageFeedbackTablePanel
        :items="feedbackItems"
        :selected-id="selected?.id ?? null"
        :loading="loading"
        :pagination="pagination"
        @select="selectFeedback"
        @update:page="updatePage"
        @update:page-size="updatePageSize"
      />
      <PageFeedbackDetailPanel
        :selected="selected"
        :can-edit="canEdit"
        :submitting="submitting"
        @save="saveSelected"
      />
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { RefreshCw } from '@lucide/vue'
import { toast } from 'vue-sonner'
import pageFeedbackApi, {
  type PageFeedbackItem,
  type PageFeedbackStatus,
} from '@/api/pageFeedback'
import type { PageFeedbackRiskOverview as PageFeedbackRiskOverviewData } from '@/api/pageFeedback'
import PageFeedbackDetailPanel from '@/components/admin/content-feedback/PageFeedbackDetailPanel.vue'
import PageFeedbackFilters from '@/components/admin/content-feedback/PageFeedbackFilters.vue'
import PageFeedbackRiskOverview from '@/components/admin/content-feedback/PageFeedbackRiskOverview.vue'
import PageFeedbackTablePanel from '@/components/admin/content-feedback/PageFeedbackTablePanel.vue'
import {
  createDefaultPageFeedbackFilters,
  type PageFeedbackFiltersState,
  type PageFeedbackPagination,
} from '@/modules/content-feedback/pageFeedbackTypes'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import { Button } from '@/components/ui/button'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const feedbackItems = ref<PageFeedbackItem[]>([])
const selected = ref<PageFeedbackItem | null>(null)
const filters = reactive<PageFeedbackFiltersState>(createDefaultPageFeedbackFilters())
const loading = ref(false)
const riskLoading = ref(false)
const riskOverview = ref<PageFeedbackRiskOverviewData | null>(null)
const submitting = ref(false)
const pagination = reactive<PageFeedbackPagination>({
  page: 1,
  page_size: 20,
  total: 0,
  total_pages: 0,
})

const canEdit = computed(() => authStore.hasPermission('content:edit'))

const selectFeedback = (item: PageFeedbackItem): void => {
  selected.value = item
}

const fetchFeedback = async (): Promise<void> => {
  loading.value = true
  try {
    const result = await pageFeedbackApi.list({
      status: filters.status,
      page_path: filters.pagePath || undefined,
      thread_key: filters.threadKey || undefined,
      search: filters.search || undefined,
      page: pagination.page,
      page_size: pagination.page_size,
    })
    feedbackItems.value = result.data
    Object.assign(pagination, result.pagination)

    if (selected.value) {
      const refreshed = feedbackItems.value.find((item) => item.id === selected.value?.id)
      selected.value = refreshed || null
    }
  } catch (error) {
    console.error('Failed to load page feedback:', error)
    feedbackItems.value = []
    toast.error('页面留言加载失败')
  } finally {
    loading.value = false
  }
}

const fetchRiskOverview = async (): Promise<void> => {
  riskLoading.value = true
  try {
    riskOverview.value = await pageFeedbackApi.riskOverview(24)
  } catch (error) {
    console.error('Failed to load page feedback risk overview:', error)
    riskOverview.value = null
  } finally {
    riskLoading.value = false
  }
}

const refreshAll = async (): Promise<void> => {
  await Promise.all([fetchFeedback(), fetchRiskOverview()])
}

const applyFilters = (nextFilters: PageFeedbackFiltersState): void => {
  Object.assign(filters, nextFilters)
  pagination.page = 1
  void fetchFeedback()
}

const resetFilters = (): void => {
  Object.assign(filters, createDefaultPageFeedbackFilters())
  pagination.page = 1
  void fetchFeedback()
}

const filterByRiskPage = (filter: { kind: 'page_path' | 'thread_key'; value: string }): void => {
  const value = filter.value.trim()
  if (!value) return
  filters.pagePath = filter.kind === 'page_path' ? value : ''
  filters.threadKey = filter.kind === 'thread_key' ? value : ''
  pagination.page = 1
  void fetchFeedback()
}

const saveSelected = async (
  payload: { status: PageFeedbackStatus; reply_content: string },
): Promise<void> => {
  if (!selected.value || submitting.value) return

  submitting.value = true
  try {
    const updated = await pageFeedbackApi.update(selected.value.id, payload)
    selectFeedback(updated)
    toast.success('页面留言已更新')
    await Promise.all([fetchFeedback(), fetchRiskOverview()])
  } catch (error: any) {
    toast.error(error?.response?.data?.error || '页面留言更新失败')
  } finally {
    submitting.value = false
  }
}

const updatePage = (page: number): void => {
  pagination.page = page
}

const updatePageSize = (pageSize: number): void => {
  pagination.page = 1
  pagination.page_size = pageSize
}

watch([() => pagination.page, () => pagination.page_size], () => {
  void fetchFeedback()
})

onMounted(() => {
  void refreshAll()
})
</script>

