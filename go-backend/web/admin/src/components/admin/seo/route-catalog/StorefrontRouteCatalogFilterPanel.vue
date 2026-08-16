<template>
  <section class="rounded-[24px] border border-dashed border-border/80 bg-card p-4 shadow-none">
    <div class="flex flex-col gap-3 xl:flex-row xl:items-start xl:justify-between">
      <div class="min-w-0">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">Route Catalog Control</p>
        <h2 class="mt-1 text-sm font-black">URL 来源、索引与可用性筛选</h2>
        <p class="mt-1 text-xs text-muted-foreground">
          当前显示 {{ paginationTotal }} 条。批量检查最多处理 200 条符合筛选条件的可检查 URL。
        </p>
      </div>
      <div class="shrink-0 text-left text-[10px] font-mono text-muted-foreground xl:text-right">
        <p>MANIFEST / {{ stats.manifest_version || '未同步' }}</p>
        <p class="mt-1">LAST SYNC / {{ formatRouteCatalogDate(stats.last_synced_at) }}</p>
      </div>
    </div>

    <form class="mt-4 grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-[minmax(220px,1.5fr)_repeat(5,minmax(120px,1fr))_auto]" @submit.prevent="emit('apply')">
      <label class="block space-y-1">
        <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">SEARCH / 搜索</span>
        <Input v-model="filters.search" placeholder="标题、路径、slug 或来源键" />
      </label>

      <label class="block space-y-1">
        <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">LOCALE / 语言</span>
        <Select v-model="filters.locale">
          <SelectTrigger><SelectValue placeholder="全部语言" /></SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部语言</SelectItem>
            <SelectItem v-for="option in localeFilterOptions" :key="option.value" :value="option.value">
              {{ option.label }}
            </SelectItem>
          </SelectContent>
        </Select>
      </label>

      <label class="block space-y-1">
        <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">SOURCE / 来源</span>
        <Select v-model="filters.source_type">
          <SelectTrigger><SelectValue placeholder="全部来源" /></SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部来源</SelectItem>
            <SelectItem value="static">静态页面</SelectItem>
            <SelectItem value="product">产品</SelectItem>
            <SelectItem value="blog">Blog</SelectItem>
            <SelectItem value="alias">兼容路由</SelectItem>
          </SelectContent>
        </Select>
      </label>

      <label class="block space-y-1">
        <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">ENTRY / 台账</span>
        <Select v-model="filters.entry_status">
          <SelectTrigger><SelectValue placeholder="全部台账状态" /></SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部台账状态</SelectItem>
            <SelectItem value="active">有效</SelectItem>
            <SelectItem value="alias">兼容</SelectItem>
            <SelectItem value="duplicate">重复</SelectItem>
            <SelectItem value="stale">失效</SelectItem>
          </SelectContent>
        </Select>
      </label>

      <label class="block space-y-1">
        <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">CHECK / 检查</span>
        <Select v-model="filters.check_status">
          <SelectTrigger><SelectValue placeholder="全部检查状态" /></SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部检查状态</SelectItem>
            <SelectItem value="ok">正常</SelectItem>
            <SelectItem value="redirect">发生跳转</SelectItem>
            <SelectItem value="not_found">404</SelectItem>
            <SelectItem value="server_error">5xx</SelectItem>
            <SelectItem value="canonical_mismatch">Canonical 不一致</SelectItem>
            <SelectItem value="error">检查失败</SelectItem>
          </SelectContent>
        </Select>
      </label>

      <label class="block space-y-1">
        <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">SEARCHABLE / 搜索</span>
        <Select v-model="filters.searchable">
          <SelectTrigger><SelectValue placeholder="全部搜索属性" /></SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部搜索属性</SelectItem>
            <SelectItem value="true">可被全局搜索</SelectItem>
            <SelectItem value="false">不进入全局搜索</SelectItem>
          </SelectContent>
        </Select>
      </label>

      <div class="flex items-end gap-2">
        <Button type="submit" class="h-9 px-3 text-xs font-black uppercase tracking-wider" :disabled="loading">
          <Search class="size-3.5" />
          查询
        </Button>
        <Button type="button" variant="outline" class="h-9 px-3 text-xs font-black uppercase tracking-wider" :disabled="loading" @click="emit('reset')">
          重置
        </Button>
      </div>
    </form>

    <div class="mt-4 flex flex-col gap-3 border-t border-dashed border-border/70 pt-3 text-xs sm:flex-row sm:items-center sm:justify-between">
      <label class="flex items-center gap-2">
        <Switch v-model="filters.includeAliases" size="sm" />
        <span>显示兼容路由</span>
        <span class="text-muted-foreground">如 /faq、旧产品路径</span>
      </label>
      <span class="font-mono text-[10px] text-muted-foreground">
        可搜索 {{ stats.searchable }} · 可检查 {{ stats.checkable }} · 已检查 {{ stats.checked }} · 未检查 {{ stats.unchecked }}
      </span>
    </div>
  </section>
</template>

<script setup lang="ts">
import { Search } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import type { StorefrontRouteCatalogFilters } from '@/composables/seo/useStorefrontRouteCatalog'
import type { StorefrontRouteCatalogStats } from '@/modules/seo/routeCatalogTypes'
import { formatRouteCatalogDate } from './routeCatalogPresentation'

defineProps<{
  filters: StorefrontRouteCatalogFilters
  stats: StorefrontRouteCatalogStats
  paginationTotal: number
  localeFilterOptions: Array<{ value: string; label: string }>
  loading?: boolean
}>()

const emit = defineEmits<{
  (event: 'apply'): void
  (event: 'reset'): void
}>()
</script>
