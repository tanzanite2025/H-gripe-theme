<template>
 <div class="space-y-4">
    <AdminPageHeader
      :title="pageMeta.title"
      :description="pageMeta.description"
    >
      <template #actions>
        <Button variant="outline" :disabled="loading || statsLoading || syncing || checking" @click="refreshAll">
 <RefreshCw :class="['size-4', loading || statsLoading ? 'animate-spin': '']" />
          刷新
        </Button>
        <Button variant="outline" :disabled="loading || syncing || !canEdit" @click="syncCatalog">
 <RefreshCw :class="['size-4', syncing ? 'animate-spin': '']" />
          同步 URL
        </Button>
        <Button :disabled="checking || !canEdit || pagination.total === 0" @click="checkCatalog">
 <CircleCheck :class="['size-4', checking ? 'animate-spin': '']" />
          检查当前语言
        </Button>
      </template>
    </AdminPageHeader>

    <div class="overflow-x-auto pb-1">
      <Tabs
        :model-value="filters.locale"
        class="min-w-max"
        @update:model-value="selectLocale"
      >
        <TabsList class="w-max min-w-full flex-nowrap gap-1 rounded-xl border border-border/70 bg-card p-1">
          <TabsTrigger
            v-for="language in enabledLanguages"
            :key="language.code"
            :value="language.code"
            class="min-w-20 flex-none px-3 py-1.5 normal-case tracking-normal"
          >
            <span class="flex flex-col items-center leading-tight">
              <span>{{ language.native_name || language.name || language.code }}</span>
              <span class="font-mono text-[9px] opacity-60">{{ language.code }}</span>
            </span>
          </TabsTrigger>
        </TabsList>
      </Tabs>
    </div>

    <AdminStatsGrid :items="statItems" />

    <StorefrontRouteCatalogFilterPanel
      :filters="filters"
      :stats="stats"
      :pagination-total="pagination.total"
      :locale-label="selectedLocaleLabel"
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
import { computed, onMounted, watch } from 'vue'
import {
  CircleCheck,
  Eye,
  RefreshCw,
  Search,
} from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import StorefrontRouteCatalogDetailDialog from '@/components/admin/url-management/route-catalog/StorefrontRouteCatalogDetailDialog.vue'
import StorefrontRouteCatalogFilterPanel from '@/components/admin/url-management/route-catalog/StorefrontRouteCatalogFilterPanel.vue'
import StorefrontRouteCatalogTable from '@/components/admin/url-management/route-catalog/StorefrontRouteCatalogTable.vue'
import { Button } from '@/components/ui/button'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useSupportedLanguages } from '@/composables/useSupportedLanguages'
import { type RouteCatalogMode, useStorefrontRouteCatalog } from '@/composables/url-management/useStorefrontRouteCatalog'
import { useAuthStore } from '@/stores/auth'

const props = withDefaults(defineProps<{
  mode?: RouteCatalogMode
}>(), {
  mode: 'catalog',
})

const authStore = useAuthStore()
const supportedLanguages = useSupportedLanguages()
const enabledLanguages = supportedLanguages.enabledLanguages
const canEdit = authStore.hasPermission('url:edit')

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
  applyPreset,
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

const pageMeta = computed(() => ({
  catalog: {
    title: 'URL 管理 / 路由台账',
    description: '统一查看前台静态页面、产品、Blog 路由及最近一次可用性检查',
  },
  canonical: {
    title: 'URL 管理 / Canonical 与冲突',
    description: '检查 Canonical 不一致与同路径来源冲突',
  },
}[props.mode]))

const statItems = computed(() => [
  { key: 'total', label: 'URL 总量', value: stats.value.total, icon: Eye, tone: 'blue' },
  { key: 'healthy', label: '正常可用', value: stats.value.ok, icon: CircleCheck, tone: 'green' },
  { key: 'attention', label: '需要处理', value: stats.value.needs_attention, icon: RefreshCw, tone: stats.value.needs_attention ? 'coral' : 'gray' },
  { key: 'not-found', label: '404', value: stats.value.not_found, icon: Search, tone: stats.value.not_found ? 'coral' : 'gray' },
  { key: 'unchecked', label: '未检查', value: stats.value.unchecked, icon: RefreshCw, tone: stats.value.unchecked ? 'amber' : 'gray' },
  { key: 'duplicate', label: '路径重复', value: stats.value.duplicate, icon: Eye, tone: stats.value.duplicate ? 'amber' : 'gray' },
])
const selectedLocaleLabel = computed(() => supportedLanguages.localeName(filters.locale))

const selectLocale = (locale: string | number): void => {
  const nextLocale = String(locale)
  if (!nextLocale || nextLocale === filters.locale) return
  filters.locale = nextLocale
  pagination.page = 1
  void refreshAll()
}

watch(
  () => props.mode,
  (mode, previousMode) => {
    applyPreset(mode, Boolean(previousMode))
  },
  { immediate: true },
)

onMounted(async () => {
  await supportedLanguages.fetchLanguages()
  if (!filters.locale && supportedLanguages.defaultLocale.value) {
    filters.locale = supportedLanguages.defaultLocale.value
  }
  await refreshAll()
})
</script>
