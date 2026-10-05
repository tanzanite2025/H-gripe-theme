<template>
  <div class="space-y-4">
    <AdminPageHeader
      title="URL 管理 / 搜索管理"
      description="先按语言、配置状态和关键词缩小范围，再维护每条 URL 的搜索词、权重和展示文案"
    >
      <template #actions>
        <Button variant="outline" :disabled="loading || profilesLoading" @click="reloadAll">
          <RefreshCw :class="['size-4', loading || profilesLoading ? 'animate-spin' : '']" />
          刷新
        </Button>
        <Button
          variant="outline"
          :disabled="loading || syncing || !canEdit"
          @click="syncCatalogAndReload"
        >
          <RefreshCw :class="['size-4', syncing ? 'animate-spin' : '']" />
          同步台账
        </Button>
      </template>
    </AdminPageHeader>

    <SearchManagementFilterPanel
      :filters="filters"
      :pagination-total="pagination.total"
      :locale-filter-options="localeFilterOptions"
      :loading="loading || profilesLoading"
      @apply="applyFilters"
      @reset="resetFilters"
    />

    <AdminStatsGrid :items="statItems" />

    <AdminTablePanel :loading="loading">
      <Table class="min-w-[1320px]">
        <TableHeader>
          <TableRow>
            <TableHead class="w-[280px]">URL</TableHead>
            <TableHead class="w-36">来源 / 语言</TableHead>
            <TableHead class="w-32">搜索状态</TableHead>
            <TableHead class="w-[260px]">关键词</TableHead>
            <TableHead>展示文案</TableHead>
            <TableHead class="w-36">更新时间</TableHead>
            <TableHead class="w-24 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-if="items.length === 0">
            <TableCell colspan="7" class="h-40 text-center text-sm text-muted-foreground">
              {{ loading ? '正在加载 URL 台账' : '当前筛选下没有 URL' }}
            </TableCell>
          </TableRow>

          <TableRow v-for="item in items" :key="item.id">
            <TableCell>
              <div class="min-w-0">
                <p class="truncate font-bold">{{ item.title || '(未命名页面)' }}</p>
                <a
                  :href="storefrontHref(item.path)"
                  target="_blank"
                  rel="noreferrer"
                  class="mt-1 block truncate font-mono text-[10px] text-primary hover:underline"
                  @click.stop
                >
                  {{ item.path }}
                </a>
                <p v-if="item.source_key" class="mt-1 truncate text-[10px] text-muted-foreground">
                  KEY / {{ item.source_key }}
                </p>
              </div>
            </TableCell>

            <TableCell>
              <AdminStatusBadge :tone="sourceTone(item.source_type)">
                {{ sourceLabel(item.source_type) }}
              </AdminStatusBadge>
              <p class="mt-1 font-mono text-[10px] text-muted-foreground">{{ item.locale }}</p>
            </TableCell>

            <TableCell>
              <AdminStatusBadge :tone="searchStatusTone(item)">
                {{ searchStatusLabel(item) }}
              </AdminStatusBadge>
              <p class="mt-1 text-[10px] text-muted-foreground">
                {{ searchConfigLabel(item) }}
              </p>
            </TableCell>

            <TableCell>
              <div class="flex flex-wrap gap-1">
                <span
                  v-for="keyword in profileFor(item)?.keywords || []"
                  :key="keyword"
                  class="rounded-full border border-border/70 bg-background px-2 py-0.5 font-mono text-[10px] text-muted-foreground"
                >
                  {{ keyword }}
                </span>
                <span v-if="!(profileFor(item)?.keywords || []).length" class="text-xs text-muted-foreground">
                  -
                </span>
              </div>
            </TableCell>

            <TableCell>
              <p class="truncate text-xs font-semibold">
                {{ displayTitle(item) }}
              </p>
              <p class="mt-1 line-clamp-2 text-[10px] text-muted-foreground">
                {{ displaySummary(item) }}
              </p>
            </TableCell>

            <TableCell class="text-xs text-muted-foreground">
              {{ formatRouteCatalogDate(profileFor(item)?.updated_at || profileFor(item)?.created_at || null) }}
            </TableCell>

            <TableCell class="text-right">
              <Button
                variant="ghost"
                size="icon"
                title="编辑搜索配置"
                aria-label="编辑搜索配置"
                :disabled="!canEdit || editingRouteID === item.id"
                @click="openEditor(item)"
              >
                <PencilLine class="size-4" />
              </Button>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>

      <template #footer>
        <AdminPagination
          :page="pagination.page"
          :page-size="pagination.page_size"
          :total="pagination.total"
          :page-sizes="[20, 50, 100]"
          @update:page="updatePage"
          @update:page-size="updatePageSize"
        />
      </template>
    </AdminTablePanel>

    <Dialog v-model:open="editorOpen">
      <DialogContent class="max-h-[calc(100dvh-1rem)] overflow-y-auto sm:max-w-3xl">
        <DialogHeader>
          <DialogTitle>URL 搜索配置</DialogTitle>
          <DialogDescription class="mt-1 break-all font-mono text-[10px]">
            {{ editingRoute?.path || '正在加载 URL 信息' }}
          </DialogDescription>
        </DialogHeader>

        <div v-if="editorLoading" class="py-12 text-center text-sm text-muted-foreground">
          正在读取搜索配置
        </div>

        <form v-else class="space-y-4" @submit.prevent="save">
          <div class="grid gap-3 md:grid-cols-2">
            <label class="block space-y-1.5">
              <span class="text-xs font-bold">启用搜索</span>
              <div class="flex items-center gap-2">
                <Switch v-model="form.enabled" />
                <span class="text-xs text-muted-foreground">
                  启用且路由有效、可搜索、可索引时进入前台搜索索引
                </span>
              </div>
            </label>

            <label class="block space-y-1.5">
              <span class="text-xs font-bold">权重</span>
              <Input v-model="form.search_weight" type="number" min="0" step="1" />
            </label>
          </div>

          <label class="block space-y-1.5">
            <span class="text-xs font-bold">展示标题</span>
            <Input v-model="form.display_title" autocomplete="off" placeholder="默认沿用 URL 标题" />
          </label>

          <label class="block space-y-1.5">
            <span class="text-xs font-bold">展示摘要</span>
            <Textarea
              v-model="form.display_summary"
              class="min-h-24"
              placeholder="默认沿用 URL 摘要"
            />
          </label>

          <label class="block space-y-1.5">
            <span class="text-xs font-bold">关键词</span>
            <Textarea
              v-model="keywordsModel"
              class="min-h-32 font-mono text-xs"
              placeholder="每行一个关键词，例如：\nfaq\nsupport\nshipping"
            />
            <p class="text-[10px] text-muted-foreground">
              支持换行、逗号、分号分隔。前台只读索引会按这些词匹配。
            </p>
          </label>

          <DialogFooter>
            <Button type="button" variant="outline" :disabled="saving" @click="editorOpen = false">
              取消
            </Button>
            <Button type="submit" :disabled="saving || !canEdit">
              <PencilLine class="size-4" />
              {{ saving ? '保存中' : '保存配置' }}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { PencilLine, RefreshCw } from '@lucide/vue'
import { toast } from 'vue-sonner'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import AdminStatusBadge, { type AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import SearchManagementFilterPanel from '@/components/admin/url-management/search-management/SearchManagementFilterPanel.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Textarea } from '@/components/ui/textarea'
import { formatRouteCatalogDate } from '@/modules/url-management/routeCatalogPresentation'
import { useSupportedLanguages } from '@/composables/useSupportedLanguages'
import { useStorefrontRouteCatalog } from '@/composables/url-management/useStorefrontRouteCatalog'
import { storefrontHref } from '@/modules/seo/routes'
import type { StorefrontRouteCatalogEntry } from '@/modules/url-management/routeCatalogTypes'
import {
  storefrontURLSearchProfilesApi,
  type StorefrontURLSearchProfile,
  type StorefrontURLSearchProfileStats,
} from '@/modules/url-management/searchProfiles'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const canEdit = authStore.hasPermission('url:edit')
const supportedLanguages = useSupportedLanguages()
const {
  items,
  loading,
  syncing,
  pagination,
  refreshCatalogList,
  syncCatalog,
  updatePage,
  updatePageSize,
  filters,
} = useStorefrontRouteCatalog(canEdit)

const profiles = ref<StorefrontURLSearchProfile[]>([])
const profilesLoading = ref(false)
const searchStats = ref<StorefrontURLSearchProfileStats>(emptySearchStats())
const editorOpen = ref(false)
const editorLoading = ref(false)
const saving = ref(false)
const editingRoute = ref<StorefrontRouteCatalogEntry | null>(null)
const editingRouteID = ref<number | null>(null)
const localeFilterOptions = computed(() => supportedLanguages.localeFilterOptions.value)

const form = reactive({
  enabled: true,
  search_weight: 0,
  display_title: '',
  display_summary: '',
})
const keywordsModel = ref('')

const profileMap = computed(() => new Map(profiles.value.map((profile) => [profile.route_entry_id, profile])))

function emptySearchStats(): StorefrontURLSearchProfileStats {
  return {
    total_routes: 0,
    public_indexed: 0,
    explicit_profiles: 0,
    explicit_enabled: 0,
    keyword_count: 0,
    unconfigured: 0,
  }
}

const profileFor = (entry: StorefrontRouteCatalogEntry) => profileMap.value.get(entry.id)

const parseKeywords = (value: string): string[] => (
  value
    .split(/[\n,，;；]/)
    .map((item) => item.trim())
    .filter(Boolean)
)

const stringifyKeywords = (keywords: string[] | undefined): string => {
  if (!Array.isArray(keywords) || keywords.length === 0) return ''
  return keywords.join('\n')
}

const isPublicSearchEntry = (entry: StorefrontRouteCatalogEntry): boolean => (
  entry.entry_status === 'active' &&
  !entry.is_alias &&
  entry.is_searchable &&
  entry.is_indexable
)

const searchStatusLabel = (entry: StorefrontRouteCatalogEntry): string => {
  const profile = profileFor(entry)
  if (!isPublicSearchEntry(entry)) return '不收录'
  if (profile && !profile.enabled) return '已停用'
  return profile ? '已收录 / 显式' : '已收录 / 默认'
}

const searchStatusTone = (entry: StorefrontRouteCatalogEntry): AdminStatusTone => {
  const profile = profileFor(entry)
  if (!isPublicSearchEntry(entry)) return 'gray'
  if (profile && !profile.enabled) return 'amber'
  return 'green'
}

const searchConfigLabel = (entry: StorefrontRouteCatalogEntry): string => {
  const profile = profileFor(entry)
  if (!profile) return '默认规则 / 权重 0'
  return `显式配置 / ${profile.enabled ? '已启用' : '已停用'} / 权重 ${profile.search_weight}`
}

const sourceLabel = (value: string): string => ({
  static: '静态页面',
  product: '产品',
  blog: 'Blog',
  alias: '兼容路由',
}[value] || value || '-')

const sourceTone = (value: string): AdminStatusTone => {
  switch (value) {
    case 'product':
      return 'blue'
    case 'blog':
      return 'amber'
    case 'alias':
      return 'gray'
    default:
      return 'green'
  }
}

const displayTitle = (entry: StorefrontRouteCatalogEntry): string => {
  const profile = profileFor(entry)
  return profile?.display_title || entry.title || '(未命名页面)'
}

const displaySummary = (entry: StorefrontRouteCatalogEntry): string => {
  const profile = profileFor(entry)
  return profile?.display_summary || entry.summary || '暂无展示摘要'
}

let profilesRequestSequence = 0
let searchStatsRequestSequence = 0
let editorRequestSequence = 0

const loadProfiles = async (): Promise<void> => {
  const requestSequence = ++profilesRequestSequence
  const localeSnapshot = filters.locale === 'all' ? undefined : filters.locale
  profilesLoading.value = true
  try {
    const nextProfiles = await storefrontURLSearchProfilesApi.list(localeSnapshot)
    if (requestSequence === profilesRequestSequence) profiles.value = nextProfiles
  } catch (error) {
    console.error('Failed to load storefront URL search profiles:', error)
    if (requestSequence === profilesRequestSequence) toast.error('URL 搜索配置加载失败')
  } finally {
    if (requestSequence === profilesRequestSequence) profilesLoading.value = false
  }
}

const loadSearchStats = async (): Promise<void> => {
  const requestSequence = ++searchStatsRequestSequence
  const localeSnapshot = filters.locale === 'all' ? undefined : filters.locale
  try {
    const nextStats = await storefrontURLSearchProfilesApi.stats(localeSnapshot)
    if (requestSequence === searchStatsRequestSequence) {
      searchStats.value = { ...emptySearchStats(), ...nextStats }
    }
  } catch (error) {
    console.error('Failed to load storefront URL search profile stats:', error)
    if (requestSequence === searchStatsRequestSequence) toast.error('URL 搜索统计加载失败')
  }
}

const reloadAll = async (): Promise<void> => {
  await Promise.all([refreshCatalogList(), loadProfiles(), loadSearchStats()])
}

const syncCatalogAndReload = async (): Promise<void> => {
  await syncCatalog()
  await Promise.all([loadProfiles(), loadSearchStats()])
}

const applyFilters = async (): Promise<void> => {
  pagination.page = 1
  await refreshCatalogList()
}

const resetFilters = async (): Promise<void> => {
  const previousLocale = filters.locale
  const nextLocale = supportedLanguages.defaultLocale.value || 'all'
  filters.search = ''
  filters.locale = nextLocale
  filters.source_type = 'all'
  filters.searchable = 'all'
  filters.search_profile_status = 'all'
  filters.includeAliases = false
  pagination.page = 1
  await refreshCatalogList()
  if (previousLocale !== nextLocale) {
    await Promise.all([loadProfiles(), loadSearchStats()])
  }
}

const openEditor = async (entry: StorefrontRouteCatalogEntry): Promise<void> => {
  const requestSequence = ++editorRequestSequence
  editingRouteID.value = entry.id
  editingRoute.value = entry
  editorOpen.value = true
  editorLoading.value = true
  try {
    const profile = await storefrontURLSearchProfilesApi.get(entry.id)
    if (requestSequence !== editorRequestSequence) return
    editingRoute.value = profile.route_entry || entry
    form.enabled = profile.enabled ?? true
    form.search_weight = profile.search_weight ?? 0
    form.display_title = profile.display_title || ''
    form.display_summary = profile.display_summary || ''
    keywordsModel.value = stringifyKeywords(profile.keywords)
  } catch (error) {
    console.error('Failed to load URL search profile:', error)
    if (requestSequence === editorRequestSequence) {
      toast.error('搜索配置读取失败')
      editorOpen.value = false
    }
  } finally {
    if (requestSequence === editorRequestSequence) editorLoading.value = false
  }
}

const save = async (): Promise<void> => {
  if (!canEdit || saving.value || !editingRouteID.value) return
  saving.value = true
  try {
    await storefrontURLSearchProfilesApi.upsert(editingRouteID.value, {
      enabled: form.enabled,
      search_weight: Number(form.search_weight) || 0,
      display_title: form.display_title.trim(),
      display_summary: form.display_summary.trim(),
      keywords: parseKeywords(keywordsModel.value),
    })
    toast.success('搜索配置已保存')
    editorOpen.value = false
    await reloadAll()
  } catch (error) {
    console.error('Failed to save URL search profile:', error)
    toast.error('搜索配置保存失败')
  } finally {
    saving.value = false
  }
}

const statItems = computed(() => {
  return [
    { key: 'total', label: '可管理 URL', value: searchStats.value.total_routes, icon: RefreshCw, tone: 'blue' },
    { key: 'public-indexed', label: '前台实际收录', value: searchStats.value.public_indexed, icon: RefreshCw, tone: searchStats.value.public_indexed ? 'green' : 'gray' },
    { key: 'explicit-profiles', label: '显式配置', value: searchStats.value.explicit_profiles, icon: PencilLine, tone: searchStats.value.explicit_profiles ? 'green' : 'gray' },
    { key: 'explicit-enabled', label: '显式启用', value: searchStats.value.explicit_enabled, icon: RefreshCw, tone: searchStats.value.explicit_enabled ? 'green' : 'gray' },
    { key: 'keywords', label: '关键词', value: searchStats.value.keyword_count, icon: PencilLine, tone: searchStats.value.keyword_count ? 'amber' : 'gray' },
    { key: 'unconfigured', label: '未配置', value: searchStats.value.unconfigured, icon: RefreshCw, tone: searchStats.value.unconfigured ? 'amber' : 'gray' },
  ]
})

onMounted(() => {
  void (async () => {
    await supportedLanguages.fetchLanguages()
    if ((!filters.locale || filters.locale === 'all') && supportedLanguages.defaultLocale.value) {
      filters.locale = supportedLanguages.defaultLocale.value
    }
    await reloadAll()
  })()
})
</script>
