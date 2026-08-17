<template>
  <Dialog v-model:open="openModel">
 <DialogContent size="xl" class="max-h-[calc(100dvh-1rem)]">
      <DialogHeader>
 <div class="flex min-w-0 items-start justify-between gap-3 pr-8">
 <div class="min-w-0">
 <DialogTitle class="truncate">{{ selectedEntry?.title || 'URL 详情'}}</DialogTitle>
 <DialogDescription class="mt-1 truncate font-mono text-[10px]">
              {{ selectedEntry?.path || '正在读取 URL 详情' }}
            </DialogDescription>
          </div>
          <AdminStatusBadge v-if="selectedEntry" :tone="entryTone(selectedEntry.entry_status)">
            {{ entryLabel(selectedEntry.entry_status) }}
          </AdminStatusBadge>
        </div>
      </DialogHeader>

 <div v-if="detailLoading && !selectedEntry" class="flex min-h-56 items-center justify-center text-sm text-muted-foreground">
        正在加载 URL 详情
      </div>

 <div v-else-if="selectedEntry" class="min-h-0 space-y-4">
 <div class="flex flex-col gap-2 rounded-xl border border-dashed border-border/80 bg-muted/20 p-3 sm:flex-row sm:items-center sm:justify-between">
          <a
            :href="storefrontHref(selectedEntry.path)"
            target="_blank"
            rel="noreferrer"
 class="min-w-0 truncate font-mono text-xs text-primary hover:underline"
          >
            {{ storefrontHref(selectedEntry.path) }}
          </a>
 <div class="flex shrink-0 gap-2">
            <Button variant="outline" size="sm" as-child>
              <a :href="storefrontHref(selectedEntry.path)" target="_blank" rel="noreferrer">
 <ExternalLink class="size-3.5" />
                打开前台
              </a>
            </Button>
            <Button
              v-if="canCreateMigrationRedirect"
              variant="outline"
              size="sm"
              as-child
            >
              <RouterLink :to="{ name: 'URLRedirects', query: { source_path: selectedEntry.path } }">
 <GitBranch class="size-3.5" />
                迁移重定向
              </RouterLink>
            </Button>
            <Button size="sm" :disabled="checkingSelected || !canEdit || !selectedEntry.is_checkable" @click="emit('checkSelected')">
 <CircleCheck :class="['size-3.5', checkingSelected ? 'animate-spin': '']" />
              {{ checkingSelected ? '检查中' : '检查此 URL' }}
            </Button>
          </div>
        </div>

 <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
 <div class="rounded-xl bg-muted/30 p-3">
 <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">来源 / 语言</p>
 <p class="mt-1 text-sm font-bold">{{ sourceLabel(selectedEntry.source_type) }} · {{ selectedEntry.locale }}</p>
 <p class="mt-1 truncate font-mono text-[10px] text-muted-foreground">{{ selectedEntry.source_key || '无来源键'}}</p>
          </div>
 <div class="rounded-xl bg-muted/30 p-3">
 <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">全局搜索</p>
 <p class="mt-1 text-sm font-bold">{{ selectedEntry.is_searchable ? '允许进入搜索': '不进入搜索'}}</p>
 <p class="mt-1 text-[10px] text-muted-foreground">{{ selectedEntry.is_indexable ? '允许索引': '不允许索引'}}</p>
          </div>
 <div class="rounded-xl bg-muted/30 p-3">
 <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">最近 HTTP</p>
 <p class="mt-1 text-sm font-bold">{{ checkLabel(selectedEntry.last_check_status, selectedEntry.last_http_status) }}</p>
 <p class="mt-1 text-[10px] text-muted-foreground">
              {{ selectedEntry.last_response_ms ? `${selectedEntry.last_response_ms} ms` : '暂无响应耗时' }}
              <span v-if="selectedEntry.last_redirect_count"> · {{ selectedEntry.last_redirect_count }} 次跳转</span>
            </p>
          </div>
 <div class="rounded-xl bg-muted/30 p-3">
 <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">最近检查</p>
 <p class="mt-1 text-sm font-bold">{{ formatRouteCatalogDate(selectedEntry.last_checked_at) }}</p>
 <p class="mt-1 text-[10px] text-muted-foreground">{{ selectedEntry.last_check_error || '无错误信息'}}</p>
          </div>
        </div>

 <div class="grid gap-3 xl:grid-cols-2">
 <div class="rounded-xl border border-dashed border-border/80 p-3">
 <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">CANONICAL / 规范地址</p>
 <p class="mt-2 break-all font-mono text-xs">{{ selectedEntry.canonical_path || '-'}}</p>
 <p class="mt-3 text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">页面声明的 Canonical</p>
 <p class="mt-2 break-all font-mono text-xs text-muted-foreground">{{ selectedEntry.last_canonical_url || '尚未读取'}}</p>
          </div>
 <div class="rounded-xl border border-dashed border-border/80 p-3">
 <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">FINAL URL / 最终地址</p>
 <p class="mt-2 break-all font-mono text-xs">{{ selectedEntry.last_final_url || '尚未检查'}}</p>
 <p class="mt-3 text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">CONTENT HASH / 内容指纹</p>
 <p class="mt-2 break-all font-mono text-[10px] text-muted-foreground">{{ latestHistoryItem?.content_hash || '尚未读取'}}</p>
          </div>
        </div>

 <section class="min-h-0 rounded-xl border border-dashed border-border/80">
 <div class="flex flex-col gap-2 border-b border-dashed border-border/70 px-3 py-3 sm:flex-row sm:items-center sm:justify-between">
            <div>
 <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">CHECK HISTORY</p>
 <h3 class="mt-1 text-sm font-black">检查历史</h3>
            </div>
 <span class="font-mono text-[10px] text-muted-foreground">共 {{ historyPagination.total }} 次</span>
          </div>
 <div class="max-h-64 overflow-auto">
 <Table class="min-w-[760px]">
              <TableHeader>
                <TableRow>
                  <TableHead>检查时间</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>HTTP</TableHead>
                  <TableHead>耗时</TableHead>
                  <TableHead>最终地址</TableHead>
                  <TableHead>错误</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-if="historyItems.length === 0">
 <TableCell colspan="6" class="h-24 text-center text-xs text-muted-foreground">
                    {{ historyLoading ? '正在加载检查历史' : '暂无检查记录' }}
                  </TableCell>
                </TableRow>
                <TableRow v-for="history in historyItems" :key="history.id">
 <TableCell class="whitespace-nowrap text-xs text-muted-foreground">{{ formatRouteCatalogDate(history.checked_at) }}</TableCell>
                  <TableCell>
                    <AdminStatusBadge :tone="checkTone(history.status)">
                      {{ checkLabel(history.status, history.http_status) }}
                    </AdminStatusBadge>
                  </TableCell>
 <TableCell class="font-mono text-xs">{{ history.http_status || '-'}}</TableCell>
 <TableCell class="font-mono text-xs">{{ history.response_ms || '-'}} ms</TableCell>
 <TableCell class="max-w-64 truncate font-mono text-[10px]" :title="history.final_url || ''">
                    {{ history.final_url || '-' }}
                  </TableCell>
 <TableCell class="max-w-52 truncate text-[10px] text-rose-600" :title="history.error_message || ''">
                    {{ history.error_message || '-' }}
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </div>
 <div v-if="historyPagination.total_pages > 1" class="border-t border-dashed border-border/70 px-3 py-3">
            <AdminPagination
              :page="historyPagination.page"
              :page-size="historyPagination.page_size"
              :total="historyPagination.total"
              :page-sizes="[10, 20, 50]"
              @update:page="(page) => emit('updateHistoryPage', page)"
              @update:page-size="(pageSize) => emit('updateHistoryPageSize', pageSize)"
            />
          </div>
        </section>
      </div>

      <DialogFooter>
        <Button type="button" variant="outline" @click="openModel = false">关闭</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { CircleCheck, ExternalLink, GitBranch } from '@lucide/vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { storefrontHref } from '@/modules/seo/routes'
import type { SEOResourcePagination } from '@/modules/seo/types'
import type { StorefrontRouteCatalogEntry, StorefrontRouteCheckResult } from '@/modules/url-management/routeCatalogTypes'
import {
  checkLabel,
  checkTone,
  entryLabel,
  entryTone,
  formatRouteCatalogDate,
  sourceLabel,
} from './routeCatalogPresentation'

const props = defineProps<{
  open: boolean
  selectedEntry: StorefrontRouteCatalogEntry | null
  historyItems: StorefrontRouteCheckResult[]
  historyPagination: SEOResourcePagination
  latestHistoryItem?: StorefrontRouteCheckResult | null
  detailLoading?: boolean
  historyLoading?: boolean
  checkingSelected?: boolean
  canEdit?: boolean
}>()

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'checkSelected'): void
  (event: 'updateHistoryPage', page: number): void
  (event: 'updateHistoryPageSize', pageSize: number): void
}>()

const openModel = computed({
  get: () => props.open,
  set: (value: boolean) => emit('update:open', value),
})

const canCreateMigrationRedirect = computed(() => (
  Boolean(props.canEdit && props.selectedEntry)
  && props.selectedEntry?.entry_status === 'stale'
))
</script>
