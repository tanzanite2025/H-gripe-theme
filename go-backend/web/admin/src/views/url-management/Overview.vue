<template>
 <div class="space-y-4">
    <AdminPageHeader
      title="URL 管理 / 概览"
      description="查看前台路由健康度与待处理问题"
    >
      <template #actions>
        <Button variant="outline" :disabled="loading" @click="load">
 <RefreshCw :class="['size-4', loading ? 'animate-spin': '']" />
          刷新
        </Button>
        <Button variant="outline" as-child>
          <a :href="sitemapHref" target="_blank" rel="noreferrer">
 <ExternalLink class="size-4" />
            打开 Sitemap
          </a>
        </Button>
        <Button variant="outline" :disabled="loading || sitemapSyncing || !canEdit" @click="syncSitemap">
 <RefreshCw :class="['size-4', sitemapSyncing ? 'animate-spin': '']" />
          更新 Sitemap
        </Button>
        <Button as-child>
          <RouterLink :to="{ name: 'URLIssues' }">
 <ListChecks class="size-4" />
            问题队列
          </RouterLink>
        </Button>
      </template>
    </AdminPageHeader>

    <AdminStatsGrid :items="statItems" />

    <section class="rounded-2xl border bg-muted/20 p-4">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">Sitemap Mapping</p>
          <h2 class="mt-1 text-sm font-black">Sitemap 映射</h2>
          <p class="mt-1 text-xs text-muted-foreground">
            sitemap.xml 由后台路由台账驱动，公开入口指向 {{ sitemapOverview.public_path || '/sitemap.xml' }}。
          </p>
        </div>
        <div class="text-right text-[10px] font-mono text-muted-foreground">
          <p>VERSION / {{ sitemapOverview.manifest_version || stats.manifest_version || '未同步' }}</p>
          <p class="mt-1">LAST SYNC / {{ formatDate(sitemapOverview.last_synced_at || stats.last_synced_at) }}</p>
        </div>
      </div>

      <div class="mt-4 grid gap-3 md:grid-cols-2 xl:grid-cols-4">
        <div class="rounded-xl border bg-background/70 p-3">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">公开入口</p>
          <p class="mt-1 break-all font-mono text-xs text-foreground">{{ sitemapOverview.sitemap_url || sitemapHref }}</p>
        </div>
        <div class="rounded-xl border bg-background/70 p-3">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">数据源</p>
          <p class="mt-1 text-sm font-black text-foreground">{{ sitemapOverview.source || 'storefront route catalog' }}</p>
          <p class="mt-1 text-[10px] text-muted-foreground">{{ sitemapOverview.dynamic_source_path || '/api/v1/storefront/sitemap-routes' }}</p>
        </div>
        <div class="rounded-xl border bg-background/70 p-3">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">可索引</p>
          <p class="mt-1 text-sm font-black text-foreground">{{ sitemapOverview.entries ?? stats.sitemap_eligible }}</p>
        </div>
        <div class="rounded-xl border bg-background/70 p-3">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">总映射</p>
          <p class="mt-1 text-sm font-black text-foreground">{{ sitemapOverview.indexable ?? stats.indexable }}</p>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ExternalLink, ListChecks, RefreshCw, Route, TriangleAlert } from '@lucide/vue'
import { toast } from 'vue-sonner'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import { Button } from '@/components/ui/button'
import { storefrontRouteCatalogApi } from '@/modules/url-management/routeCatalog'
import type { StorefrontRouteCatalogStats, StorefrontSitemapOverview } from '@/modules/url-management/routeCatalogTypes'
import { storefrontURLIssuesApi, type StorefrontURLIssueStats } from '@/modules/url-management/urlIssues'
import { defaultStorefrontRouteCatalogStats } from '@/composables/url-management/useStorefrontRouteCatalog'
import { useAuthStore } from '@/stores/auth'

const loading = ref(false)
const sitemapSyncing = ref(false)
const sitemapOverview = ref<StorefrontSitemapOverview>({
  public_path: '/sitemap.xml',
  sitemap_url: '/sitemap.xml',
  source: 'storefront route catalog',
  dynamic_source_path: '/api/v1/storefront/sitemap-routes',
  entries: 0,
  indexable: 0,
  last_synced_at: null,
  manifest_version: '',
})
const stats = ref<StorefrontRouteCatalogStats>(defaultStorefrontRouteCatalogStats())
const issueStats = ref<StorefrontURLIssueStats>({
  active: 0,
  open: 0,
  acknowledged: 0,
  resolved: 0,
  verified: 0,
  suppressed: 0,
  critical: 0,
  high: 0,
})
const authStore = useAuthStore()
const canEdit = authStore.hasPermission('url:edit')
const sitemapHref = computed(() => sitemapOverview.value.sitemap_url || '/sitemap.xml')

const formatDate = (value?: string | null): string => {
  if (!value) return '未同步'
  const parsed = new Date(value)
  return Number.isNaN(parsed.getTime()) ? value : parsed.toLocaleString()
}

const load = async (): Promise<void> => {
  loading.value = true
  try {
    const [catalogStats, currentIssueStats, currentSitemapOverview] = await Promise.all([
      storefrontRouteCatalogApi.stats(),
      storefrontURLIssuesApi.summary(),
      storefrontRouteCatalogApi.sitemap(),
    ])
    stats.value = { ...defaultStorefrontRouteCatalogStats(), ...catalogStats }
    issueStats.value = { ...issueStats.value, ...currentIssueStats }
    sitemapOverview.value = { ...sitemapOverview.value, ...currentSitemapOverview }
  } catch (error) {
    console.error('Failed to load URL management overview:', error)
    toast.error('URL 管理概览加载失败')
  } finally {
    loading.value = false
  }
}

const syncSitemap = async (): Promise<void> => {
  if (!canEdit || sitemapSyncing.value) return
  sitemapSyncing.value = true
  try {
    const response = await storefrontRouteCatalogApi.syncSitemap()
    sitemapOverview.value = { ...sitemapOverview.value, ...response.sitemap }
    toast.success(`Sitemap 已更新：${response.sync.entries || 0} 条映射`)
    await load()
  } catch (error) {
    console.error('Failed to sync sitemap mapping:', error)
    toast.error('Sitemap 更新失败')
  } finally {
    sitemapSyncing.value = false
  }
}

const statItems = computed(() => [
  { key: 'total', label: 'URL 总量', value: stats.value.total, icon: Route, tone: 'blue' },
  { key: 'attention', label: '待处理', value: issueStats.value.active, icon: TriangleAlert, tone: issueStats.value.active ? 'coral' : 'gray' },
  { key: 'redirects', label: '兼容跳转', value: stats.value.redirects, icon: Route, tone: stats.value.redirects ? 'amber' : 'gray' },
  { key: 'canonical', label: 'Canonical', value: stats.value.canonical_mismatch, icon: TriangleAlert, tone: stats.value.canonical_mismatch ? 'amber' : 'gray' },
  { key: 'not-found', label: '404', value: stats.value.not_found, icon: TriangleAlert, tone: stats.value.not_found ? 'coral' : 'gray' },
  { key: 'unchecked', label: '未检查', value: stats.value.unchecked, icon: RefreshCw, tone: stats.value.unchecked ? 'amber' : 'gray' },
])

onMounted(() => {
  void load()
})
</script>
