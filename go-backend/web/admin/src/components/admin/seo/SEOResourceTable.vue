<template>
  <AdminTablePanel :loading="loading">
    <Table class="min-w-[860px]">
      <TableHeader>
        <TableRow>
          <TableHead>页面资源</TableHead>
          <TableHead>实际路由</TableHead>
          <TableHead class="w-28">语言</TableHead>
          <TableHead class="w-28">SEO 状态</TableHead>
          <TableHead class="w-28 text-right">操作</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableEmpty v-if="items.length === 0" :colspan="5">
          <div class="flex flex-col items-center text-muted-foreground">
            <FileSearch class="mb-2 size-7 opacity-55" />
            <span class="text-xs">暂无{{ resourceLabel }}资源</span>
          </div>
        </TableEmpty>

        <TableRow v-for="item in items" :key="item.id">
          <TableCell>
            <div class="min-w-0">
              <p class="max-w-80 truncate text-xs font-bold">{{ item.title }}</p>
              <p class="mt-1 font-mono text-[10px] text-muted-foreground/70">ID {{ item.id }}</p>
            </div>
          </TableCell>
          <TableCell>
            <a
              v-if="item.href"
              :href="item.href"
              target="_blank"
              rel="noreferrer"
              class="group inline-flex max-w-[30rem] items-center gap-1.5 font-mono text-[11px] text-primary hover:underline"
              :title="item.routePath"
            >
              <span class="truncate">{{ item.routePath }}</span>
              <ExternalLink class="size-3.5 shrink-0 opacity-70 transition group-hover:opacity-100" />
            </a>
            <span v-else class="text-xs text-muted-foreground">路由不可用</span>
          </TableCell>
          <TableCell class="text-xs">{{ item.localeLabel }}</TableCell>
          <TableCell>
            <AdminStatusBadge :tone="statusTone(item)">
              {{ statusLabel(item) }}
            </AdminStatusBadge>
          </TableCell>
          <TableCell class="text-right">
            <Button v-if="canEdit" variant="outline" size="sm" @click="emit('edit', item)">
              <Pencil class="size-3.5" />
              SEO
            </Button>
            <span v-else class="text-xs text-muted-foreground">只读</span>
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>

    <template #footer>
      <AdminPagination
        :page="pagination.page"
        :page-size="pagination.page_size"
        :total="pagination.total"
        @update:page="emit('update-page', $event)"
        @update:page-size="emit('update-page-size', $event)"
      />
    </template>
  </AdminTablePanel>
</template>

<script setup lang="ts">
import { ExternalLink, FileSearch, Pencil } from '@lucide/vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import type { SEOResourceItem, SEOResourcePagination } from '@/modules/seo/types'

const props = withDefaults(defineProps<{
  items?: SEOResourceItem[]
  pagination: SEOResourcePagination
  resourceLabel: string
  loading?: boolean
  canEdit?: boolean
}>(), {
  items: () => [],
  loading: false,
  canEdit: false,
})

const emit = defineEmits<{
  (event: 'edit', item: SEOResourceItem): void
  (event: 'update-page', page: number): void
  (event: 'update-page-size', pageSize: number): void
}>()

const hasSeo = (item: SEOResourceItem): boolean => Boolean(
  item.metaTitle.trim() || item.metaDescription.trim() || item.canonicalUrl?.trim()
)

const statusTone = (item: SEOResourceItem): 'green' | 'amber' | 'gray' => {
  if (item.productDiagnostics) return item.productDiagnostics.ready ? 'green' : 'amber'
  return hasSeo(item) ? 'green' : 'gray'
}

const statusLabel = (item: SEOResourceItem): string => {
  if (item.productDiagnostics) return item.productDiagnostics.ready ? 'JSON-LD 可用' : '需补数据'
  return hasSeo(item) ? '已设置' : '未设置'
}
</script>
