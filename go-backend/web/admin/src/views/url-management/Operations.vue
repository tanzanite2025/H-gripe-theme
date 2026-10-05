<template>
 <div class="space-y-4">
    <AdminPageHeader
      title="URL 管理 / 同步与检查"
      description="更新路由快照并执行前台可用性检查"
    >
      <template #actions>
        <Button variant="outline" :disabled="loading || syncing || checking" @click="refreshAll">
 <RefreshCw :class="['size-4', loading ? 'animate-spin': '']" />
          刷新
        </Button>
        <Button variant="outline" :disabled="syncing || !canEdit" @click="syncCatalog">
 <RefreshCw :class="['size-4', syncing ? 'animate-spin': '']" />
          同步 URL
        </Button>
        <Button :disabled="checking || !canEdit || pagination.total === 0" @click="checkCatalog">
 <CircleCheck :class="['size-4', checking ? 'animate-spin': '']" />
          检查当前语言
        </Button>
      </template>
    </AdminPageHeader>

    <AdminStorefrontLanguageDisplayCard
      :model-value="filters.locale"
      :language-options="languageOptions"
      :disabled="loading || syncing || checking"
      :loading="loading"
      aria-label="URL 同步与检查语言"
      @update:model-value="selectLocale"
    />

    <p class="text-xs text-muted-foreground">
      当前检查范围：{{ selectedLocaleLabel }}（{{ filters.locale }}）。同步 URL 会更新全量语言；检查会自动分批处理当前筛选范围内的全部可检查 URL；待处理卡片为全站工单汇总。
    </p>

    <AdminStatsGrid :items="statItems" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { CircleCheck, RefreshCw, TriangleAlert } from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import AdminStorefrontLanguageDisplayCard from '@/components/admin/AdminStorefrontLanguageDisplayCard.vue'
import { Button } from '@/components/ui/button'
import { useSupportedLanguages } from '@/composables/useSupportedLanguages'
import { useStorefrontRouteCatalog } from '@/composables/url-management/useStorefrontRouteCatalog'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const canEdit = authStore.hasPermission('url:edit')
const supportedLanguages = useSupportedLanguages()
const languageOptions = supportedLanguages.languageOptions
const {
  stats,
  issueStats,
  loading,
  syncing,
  checking,
  pagination,
  filters,
  refreshAll,
  syncCatalog,
  checkCatalog,
} = useStorefrontRouteCatalog(canEdit)

const selectedLocaleLabel = computed(() => supportedLanguages.localeName(filters.locale))

const selectLocale = (locale: string | number): void => {
  const nextLocale = String(locale)
  if (!nextLocale || nextLocale === filters.locale) return
  filters.locale = nextLocale
  pagination.page = 1
  void refreshAll()
}

const statItems = computed(() => [
  { key: 'checked', label: '已检查', value: stats.value.checked, icon: CircleCheck, tone: 'green' },
  { key: 'unchecked', label: '未检查', value: stats.value.unchecked, icon: RefreshCw, tone: stats.value.unchecked ? 'amber' : 'gray' },
  { key: 'attention', label: '全站未关闭工单', value: issueStats.value.active, icon: TriangleAlert, tone: issueStats.value.active ? 'coral' : 'gray' },
  { key: 'stale', label: '失效快照', value: stats.value.stale, icon: TriangleAlert, tone: stats.value.stale ? 'amber' : 'gray' },
])

onMounted(() => {
  void (async () => {
    await supportedLanguages.fetchLanguages()
    if (!filters.locale && supportedLanguages.defaultLocale.value) {
      filters.locale = supportedLanguages.defaultLocale.value
    }
    await refreshAll()
  })()
})
</script>
