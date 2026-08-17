<template>
  <AdminTablePanel :loading="loading">
 <Table class="min-w-[1240px]">
      <TableHeader>
        <TableRow>
 <TableHead class="w-[320px]">页面 / 路径</TableHead>
 <TableHead class="w-36">来源 / 语言</TableHead>
 <TableHead class="w-40">台账属性</TableHead>
 <TableHead class="w-44">最近检查</TableHead>
 <TableHead class="w-[280px]">Canonical / 最终地址</TableHead>
 <TableHead class="w-40">最近检查时间</TableHead>
 <TableHead class="w-20 text-right">详情</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow v-if="items.length === 0">
 <TableCell colspan="7" class="h-40 text-center text-sm text-muted-foreground">
            {{ loading ? '正在加载 URL 台账' : '当前筛选下没有 URL，请先同步或调整筛选条件' }}
          </TableCell>
        </TableRow>
        <TableRow
          v-for="item in items"
          :key="item.id"
 class="cursor-pointer"
          @click="emit('openDetail', item)"
        >
          <TableCell>
 <div class="min-w-0">
 <p class="truncate font-bold">{{ item.title || '(未命名页面)'}}</p>
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
 <p v-if="item.duplicate_group_key" class="mt-1 truncate text-[10px] text-rose-600">
                路径冲突 / {{ item.duplicate_group_key }}
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
 <div class="flex flex-wrap items-center gap-1">
              <AdminStatusBadge :tone="entryTone(item.entry_status)">
                {{ entryLabel(item.entry_status) }}
              </AdminStatusBadge>
              <AdminStatusBadge v-if="item.is_alias" tone="gray">Alias</AdminStatusBadge>
            </div>
 <div class="mt-2 flex flex-wrap gap-x-2 gap-y-1 text-[10px] text-muted-foreground">
 <span :class="item.is_searchable ? 'text-emerald-600': ''">搜 {{ item.is_searchable ? '是': '否'}}</span>
 <span :class="item.is_checkable ? 'text-emerald-600': ''">检 {{ item.is_checkable ? '是': '否'}}</span>
 <span :class="item.is_indexable ? 'text-emerald-600': ''">索引 {{ item.is_indexable ? '是': '否'}}</span>
            </div>
          </TableCell>
          <TableCell>
            <AdminStatusBadge :tone="checkTone(item.last_check_status)">
              {{ checkLabel(item.last_check_status, item.last_http_status) }}
            </AdminStatusBadge>
 <p class="mt-1 text-[10px] text-muted-foreground">
              {{ item.last_response_ms ? `${item.last_response_ms} ms` : '暂无响应耗时' }}
              <span v-if="item.last_redirect_count"> · {{ item.last_redirect_count }} 次跳转</span>
            </p>
 <p v-if="item.last_check_error" class="mt-1 max-w-40 truncate text-[10px] text-rose-600" :title="item.last_check_error">
              {{ item.last_check_error }}
            </p>
          </TableCell>
          <TableCell>
 <p class="truncate font-mono text-[10px] text-muted-foreground" :title="item.canonical_path || ''">
              预期 / {{ item.canonical_path || '-' }}
            </p>
 <p class="mt-1 truncate font-mono text-[10px]" :class="item.last_final_url ? 'text-foreground': 'text-muted-foreground'">
              实际 / {{ item.last_final_url || '尚未检查' }}
            </p>
          </TableCell>
 <TableCell class="text-xs text-muted-foreground">
            {{ formatRouteCatalogDate(item.last_checked_at) }}
          </TableCell>
 <TableCell class="text-right">
            <Button
              variant="ghost"
              size="icon"
              title="查看 URL 详情"
              aria-label="查看 URL 详情"
              @click.stop="emit('openDetail', item)"
            >
 <Eye class="size-4" />
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
        @update:page="(page) => emit('updatePage', page)"
        @update:page-size="(pageSize) => emit('updatePageSize', pageSize)"
      />
    </template>
  </AdminTablePanel>
</template>

<script setup lang="ts">
import { Eye } from '@lucide/vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { storefrontHref } from '@/modules/seo/routes'
import type { SEOResourcePagination } from '@/modules/seo/types'
import type { StorefrontRouteCatalogEntry } from '@/modules/url-management/routeCatalogTypes'
import {
  checkLabel,
  checkTone,
  entryLabel,
  entryTone,
  formatRouteCatalogDate,
  sourceLabel,
  sourceTone,
} from './routeCatalogPresentation'

defineProps<{
  items: StorefrontRouteCatalogEntry[]
  pagination: SEOResourcePagination
  loading?: boolean
}>()

const emit = defineEmits<{
  (event: 'openDetail', item: StorefrontRouteCatalogEntry): void
  (event: 'updatePage', page: number): void
  (event: 'updatePageSize', pageSize: number): void
}>()
</script>
