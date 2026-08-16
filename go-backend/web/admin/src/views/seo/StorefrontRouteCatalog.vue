<template>
  <div class="space-y-4">
    <AdminPageHeader
      title="SEO / URL 台账"
      description="统一查看前台静态页面、产品、Blog 路由及最近一次可用性检查"
    >
      <template #actions>
        <Button variant="outline" :disabled="loading || statsLoading || syncing || checking" @click="refreshAll">
          <RefreshCw :class="['size-4', loading || statsLoading ? 'animate-spin' : '']" />
          刷新
        </Button>
        <Button variant="outline" :disabled="loading || syncing || !canEdit" @click="syncCatalog">
          <RefreshCw :class="['size-4', syncing ? 'animate-spin' : '']" />
          同步 URL
        </Button>
        <Button :disabled="checking || !canEdit || pagination.total === 0" @click="checkCatalog">
          <CircleCheck :class="['size-4', checking ? 'animate-spin' : '']" />
          检查筛选结果
        </Button>
      </template>
    </AdminPageHeader>

    <AdminStatsGrid :items="statItems" />

    <StorefrontRouteCatalogFilterPanel
      :filters="filters"
      :stats="stats"
      :pagination-total="pagination.total"
      :locale-filter-options="localeFilterOptions"
      :loading="loading"
      @apply="applyFilters"
      @reset="resetFilters"
    />

    <StorefrontRouteCatalogTable
      :items="items"
      :pagination="pagination"
      :loading="loading"
      @open-detail="openDetail"
      @update-page="updatePage"
      @update-page-size="updatePageSize"
    />

    <StorefrontRouteCatalogDetailDialog
      v-model:open="detailOpen"
      :selected-entry="selectedEntry"
      :history-items="historyItems"
      :history-pagination="historyPagination"
      :latest-history-item="latestHistoryItem"
      :detail-loading="detailLoading"
      :history-loading="historyLoading"
      :checking-selected="checkingSelected"
      :can-edit="canEdit"
      @check-selected="checkSelected"
      @update-history-page="updateHistoryPage"
      @update-history-page-size="updateHistoryPageSize"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import {
  CircleCheck,
  Eye,
  RefreshCw,
  Search,
} from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import StorefrontRouteCatalogDetailDialog from '@/components/admin/seo/route-catalog/StorefrontRouteCatalogDetailDialog.vue'
import StorefrontRouteCatalogFilterPanel from '@/components/admin/seo/route-catalog/StorefrontRouteCatalogFilterPanel.vue'
import StorefrontRouteCatalogTable from '@/components/admin/seo/route-catalog/StorefrontRouteCatalogTable.vue'
import { Button } from '@/components/ui/button'
import { useSupportedLanguages } from '@/composables/useSupportedLanguages'
import { useStorefrontRouteCatalog } from '@/composables/seo/useStorefrontRouteCatalog'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const supportedLanguages = useSupportedLanguages()
const localeFilterOptions = supportedLanguages.localeFilterOptions
const canEdit = authStore.hasPermission('seo:edit')

const {
  stats,
  items,
  loading,
  statsLoading,
  syncing,
  checking,
  detailLoading,
  historyLoading,
  checkingSelected,
  detailOpen,
  selectedEntry,
  historyItems,
  filters,
  pagination,
  historyPagination,
  latestHistoryItem,
  refreshAll,
  applyFilters,
  resetFilters,
  updatePage,
  updatePageSize,
  syncCatalog,
  checkCatalog,
  openDetail,
  checkSelected,
  updateHistoryPage,
  updateHistoryPageSize,
} = useStorefrontRouteCatalog(canEdit)

const statItems = computed(() => [
  { key: 'total', label: 'URL 总量', value: stats.value.total, icon: Eye, tone: 'blue' },
  { key: 'healthy', label: '正常可用', value: stats.value.ok, icon: CircleCheck, tone: 'green' },
  { key: 'attention', label: '需要处理', value: stats.value.needs_attention, icon: RefreshCw, tone: stats.value.needs_attention ? 'coral' : 'gray' },
  { key: 'not-found', label: '404', value: stats.value.not_found, icon: Search, tone: stats.value.not_found ? 'coral' : 'gray' },
  { key: 'unchecked', label: '未检查', value: stats.value.unchecked, icon: RefreshCw, tone: stats.value.unchecked ? 'amber' : 'gray' },
  { key: 'duplicate', label: '路径重复', value: stats.value.duplicate, icon: Eye, tone: stats.value.duplicate ? 'amber' : 'gray' },
])

onMounted(async () => {
  await supportedLanguages.fetchLanguages()
  await refreshAll()
})
</script>
