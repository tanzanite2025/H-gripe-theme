<template>
  <div class="space-y-4">
    <AdminPageHeader title="上线前检查 / 图片尺寸" description="独立核对原图尺寸与标准尺寸派生图">
      <template #actions>
        <Button
          size="icon"
          variant="outline"
          title="刷新图片尺寸检测"
          :disabled="loading"
          @click="loadFindings"
        >
          <RefreshCw :class="['size-4', { 'animate-spin': loading }]" />
        </Button>
      </template>
    </AdminPageHeader>

    <section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-5">
      <div v-for="item in summaryItems" :key="item.label" class="border bg-card px-4 py-3">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">{{ item.label }}</p>
        <p class="mt-2 text-2xl font-black" :class="item.tone">{{ item.value }}</p>
      </div>
    </section>

    <section class="border bg-card">
      <div class="flex flex-wrap items-start justify-between gap-3 border-b px-4 py-3">
        <div>
          <h2 class="text-sm font-black">当前支持的尺寸转换</h2>
          <p class="mt-1 text-xs text-muted-foreground">按原图比例生成，小于目标宽度的图片不会被放大。</p>
        </div>
        <span class="font-mono text-xs font-bold text-muted-foreground">{{ presets.length }} 个预设</span>
      </div>
      <div class="grid divide-y sm:grid-cols-3 sm:divide-x sm:divide-y-0">
        <div v-for="preset in presets" :key="preset.name" class="px-4 py-3">
          <div class="flex items-center justify-between gap-2">
            <p class="font-bold">{{ preset.label }}</p>
            <span class="font-mono text-[10px] font-bold text-muted-foreground">v{{ preset.generation_version || 1 }}</span>
          </div>
          <p class="mt-1 font-mono text-xs text-muted-foreground">{{ preset.name }}</p>
          <p class="mt-2 text-xs font-bold">最大宽度 {{ preset.max_width }} px</p>
        </div>
        <p v-if="presets.length === 0" class="px-4 py-5 text-sm text-muted-foreground sm:col-span-3">暂未读取到尺寸转换定义</p>
      </div>
    </section>

    <section class="border bg-card p-4">
      <div class="grid gap-3 lg:grid-cols-[minmax(0,1fr)_12rem_auto] lg:items-end">
        <label class="grid gap-1.5">
          <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">搜索图片</span>
          <Input v-model="search" placeholder="文件名或 URL" @keyup.enter="applyFilters" />
        </label>
        <label class="grid gap-1.5">
          <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">检测结果</span>
          <Select v-model="state">
            <SelectTrigger><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="attention">待补齐</SelectItem>
              <SelectItem value="all">全部图片</SelectItem>
              <SelectItem value="missing_dimensions">缺少原图尺寸</SelectItem>
              <SelectItem value="missing_variants">缺少标准尺寸</SelectItem>
              <SelectItem value="ready">已就绪</SelectItem>
            </SelectContent>
          </Select>
        </label>
        <Button :disabled="loading" @click="applyFilters">
          <Search class="size-4" />
          应用筛选
        </Button>
      </div>
    </section>

    <section class="overflow-hidden border bg-card">
      <div class="overflow-x-auto">
        <table class="w-full min-w-[860px] text-sm">
          <thead class="border-b bg-muted/30 text-left text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">
            <tr>
              <th class="px-4 py-3">图片</th>
              <th class="px-4 py-3">原图尺寸</th>
              <th class="px-4 py-3">标准尺寸</th>
              <th class="px-4 py-3">状态</th>
              <th class="px-4 py-3 text-right">处理</th>
            </tr>
          </thead>
          <tbody class="divide-y">
            <tr v-if="loading">
              <td colspan="5" class="px-4 py-12 text-center text-sm text-muted-foreground">正在核对图片尺寸</td>
            </tr>
            <tr v-else-if="findings.length === 0">
              <td colspan="5" class="px-4 py-12 text-center text-sm text-muted-foreground">当前筛选没有图片</td>
            </tr>
            <tr v-for="item in findings" :key="String(item.asset.id)">
              <td class="px-4 py-3">
                <div class="flex min-w-[19rem] items-center gap-3">
                  <img
                    v-if="assetAccessURL(item.asset)"
                    :src="assetAccessURL(item.asset)"
                    :alt="assetTitle(item.asset)"
                    class="size-12 shrink-0 border bg-muted object-cover"
                  />
                  <div class="min-w-0">
                    <p class="truncate font-bold">{{ assetTitle(item.asset) }}</p>
                    <p class="mt-1 truncate font-mono text-[10px] text-muted-foreground">{{ item.asset.storage_key || item.asset.url }}</p>
                  </div>
                </div>
              </td>
              <td class="px-4 py-3 font-mono text-xs">{{ formatMediaDimensions(item.asset.width, item.asset.height) }}</td>
              <td class="px-4 py-3">
                <span v-if="item.missing_presets?.length" class="font-mono text-xs text-amber-700">
                  缺少 {{ formatPresetList(item.missing_presets) }}
                </span>
                <span v-else class="font-mono text-xs text-emerald-700">{{ supportedPresetList }}</span>
              </td>
              <td class="px-4 py-3">
                <AdminStatusBadge :tone="stateTone(item.state)">{{ stateLabel(item.state) }}</AdminStatusBadge>
              </td>
              <td class="px-4 py-3 text-right">
                <Button
                  v-if="item.state !== 'ready'"
                  size="sm"
                  :disabled="!canEdit || reconcilingID === item.asset.id"
                  @click="reconcile(item.asset.id)"
                >
                  <LoaderCircle v-if="reconcilingID === item.asset.id" class="size-3.5 animate-spin" />
                  <WandSparkles v-else class="size-3.5" />
                  立即补齐
                </Button>
                <span v-else class="text-xs font-bold text-emerald-700">已就绪</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div class="border-t px-4 py-3">
        <AdminPagination
          :page="pagination.page"
          :page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[20, 40, 80, 100]"
          @update:page="changePage"
          @update:page-size="changePageSize"
        />
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { LoaderCircle, RefreshCw, Search, WandSparkles } from '@lucide/vue'
import { toast } from 'vue-sonner'
import preflightApi, {
  type PreflightImageDimensionFinding,
  type PreflightImageDimensionPreset,
  type PreflightImageDimensionState,
  type PreflightImageDimensionSummary,
} from '@/api/preflight'
import type { MediaID } from '@/api/media'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminStatusBadge, { type AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { assetAccessURL, assetTitle, formatMediaDimensions } from '@/lib/mediaPresentation'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const loading = ref(false)
const search = ref('')
const state = ref<PreflightImageDimensionState>('attention')
const findings = ref<PreflightImageDimensionFinding[]>([])
const presets = ref<PreflightImageDimensionPreset[]>([])
const summary = ref<PreflightImageDimensionSummary>({
  total: 0,
  attention: 0,
  ready: 0,
  missing_dimensions: 0,
  missing_variants: 0,
})
const pagination = ref({ page: 1, pageSize: 40, total: 0 })
const reconcilingID = ref<MediaID | null>(null)
const canEdit = computed(() => authStore.hasPermission('media:edit'))
const presetsByName = computed(() => new Map(presets.value.map((preset) => [preset.name, preset])))
const supportedPresetList = computed(() => formatPresetList(presets.value.map((preset) => preset.name)) || '-')

const summaryItems = computed(() => [
  { label: '图片总数', value: summary.value.total, tone: '' },
  { label: '待补齐', value: summary.value.attention, tone: summary.value.attention ? 'text-amber-600' : 'text-emerald-600' },
  { label: '原图尺寸缺失', value: summary.value.missing_dimensions, tone: summary.value.missing_dimensions ? 'text-rose-600' : 'text-emerald-600' },
  { label: '标准尺寸缺失', value: summary.value.missing_variants, tone: summary.value.missing_variants ? 'text-amber-600' : 'text-emerald-600' },
  { label: '已就绪', value: summary.value.ready, tone: 'text-emerald-600' },
])

const loadFindings = async (): Promise<void> => {
  loading.value = true
  try {
    const result = await preflightApi.listImageDimensions({
      page: pagination.value.page,
      page_size: pagination.value.pageSize,
      search: search.value.trim() || undefined,
      state: state.value,
    })
    findings.value = result.items
    summary.value = result.summary
    presets.value = result.presets
    pagination.value.total = result.pagination.total
  } catch (error: any) {
    toast.error(error?.response?.data?.error || '图片尺寸检测加载失败')
  } finally {
    loading.value = false
  }
}

const applyFilters = (): void => {
  pagination.value.page = 1
  void loadFindings()
}

const changePage = (page: number): void => {
  pagination.value.page = page
  void loadFindings()
}

const changePageSize = (pageSize: number): void => {
  pagination.value.pageSize = pageSize
  pagination.value.page = 1
  void loadFindings()
}

const reconcile = async (id: MediaID): Promise<void> => {
  if (!canEdit.value || reconcilingID.value) return
  reconcilingID.value = id
  try {
    await preflightApi.reconcileImageDimensions(id)
    toast.success('已回填原图尺寸并补齐标准尺寸')
    await loadFindings()
  } catch (error: any) {
    toast.error(error?.response?.data?.error || '图片尺寸补齐失败')
  } finally {
    reconcilingID.value = null
  }
}

const formatPresetList = (names: string[]): string => names
  .map((name) => {
    const preset = presetsByName.value.get(name)
    return preset ? `${preset.label} ${preset.max_width}px v${preset.generation_version || 1}` : name
  })
  .join(' / ')

const stateTone = (value: PreflightImageDimensionFinding['state']): AdminStatusTone => {
  if (value === 'ready') return 'green'
  if (value === 'missing_dimensions' || value === 'missing_dimensions_and_variants') return 'coral'
  return 'amber'
}

const stateLabel = (value: PreflightImageDimensionFinding['state']): string => ({
  ready: '已就绪',
  missing_dimensions: '缺少原图尺寸',
  missing_variants: '缺少标准尺寸',
  missing_dimensions_and_variants: '需要完整补齐',
}[value])

onMounted(() => {
  void loadFindings()
})
</script>
