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

    <p class="text-xs text-muted-foreground">
      当前检查范围：{{ selectedLocaleLabel }}（{{ filters.locale }}）。同步 URL 会更新全量语言；检查只处理当前语言，最多 200 条可检查 URL；待处理卡片为全站工单汇总。
    </p>

    <AdminStatsGrid :items="statItems" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { CircleCheck, RefreshCw, TriangleAlert } from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import { Button } from '@/components/ui/button'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useSupportedLanguages } from '@/composables/useSupportedLanguages'
import { useStorefrontRouteCatalog } from '@/composables/url-management/useStorefrontRouteCatalog'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const canEdit = authStore.hasPermission('url:edit')
const supportedLanguages = useSupportedLanguages()
const enabledLanguages = supportedLanguages.enabledLanguages
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
  { key: 'attention', label: '全站待处理', value: issueStats.value.active, icon: TriangleAlert, tone: issueStats.value.active ? 'coral' : 'gray' },
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
