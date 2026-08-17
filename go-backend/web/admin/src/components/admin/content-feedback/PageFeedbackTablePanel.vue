<template>
  <AdminTablePanel :loading="loading" scroll-body>
    <Table class="min-w-[1020px]">
      <TableHeader>
        <TableRow>
          <TableHead class="w-20">ID</TableHead>
          <TableHead class="w-64">页面</TableHead>
          <TableHead class="w-40">留言人</TableHead>
          <TableHead>留言</TableHead>
          <TableHead class="w-24">回复</TableHead>
          <TableHead class="w-28">状态</TableHead>
          <TableHead class="w-40">提交时间</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow v-if="!loading && items.length === 0">
          <TableCell colspan="7" class="h-40 text-center text-sm text-muted-foreground">
            当前筛选下没有页面留言
          </TableCell>
        </TableRow>

        <TableRow
          v-for="item in items"
          :key="item.id"
          class="cursor-pointer"
          :class="selectedId === item.id ? 'bg-primary/5' : ''"
          @click="emit('select', item)"
        >
          <TableCell class="font-mono text-xs">#{{ item.id }}</TableCell>
          <TableCell>
            <p class="truncate text-xs font-semibold">{{ displayPageTitle(item) }}</p>
            <p class="truncate font-mono text-[10px] text-muted-foreground">{{ displayPagePath(item) }}</p>
            <p class="truncate font-mono text-[10px] text-muted-foreground/80">{{ item.thread_key }}</p>
          </TableCell>
          <TableCell>
            <p class="truncate text-xs font-semibold">{{ item.name || '未填写' }}</p>
            <p class="truncate text-[10px] text-muted-foreground">{{ item.email || `用户 #${item.user_id}` }}</p>
            <p class="truncate font-mono text-[10px] text-muted-foreground/80">
              来源 {{ displaySourceHashPreview(item.source_hash_preview) }}
            </p>
          </TableCell>
          <TableCell>
            <p class="max-w-[360px] truncate text-xs text-muted-foreground">{{ item.content }}</p>
          </TableCell>
          <TableCell>
            <Badge :variant="item.reply_content ? 'default' : 'outline'">
              {{ item.reply_content ? '已回复' : '未回复' }}
            </Badge>
          </TableCell>
          <TableCell>
            <Badge :variant="pageFeedbackStatusBadgeVariant(item.status)">
              {{ pageFeedbackStatusLabel(item.status) }}
            </Badge>
          </TableCell>
          <TableCell class="text-xs text-muted-foreground">
            {{ formatPageFeedbackDate(item.created_at) }}
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>

    <template #footer>
      <AdminPagination
        :page="pagination.page"
        :page-size="pagination.page_size"
        :total="pagination.total"
        :page-sizes="[10, 20, 50]"
        @update:page="emit('update:page', $event)"
        @update:page-size="emit('update:page-size', $event)"
      />
    </template>
  </AdminTablePanel>
</template>

<script setup lang="ts">
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  displayPagePath,
  displayPageTitle,
  displaySourceHashPreview,
  formatPageFeedbackDate,
  pageFeedbackStatusBadgeVariant,
  pageFeedbackStatusLabel,
  type PageFeedbackItem,
  type PageFeedbackPagination,
} from './pageFeedbackTypes'

defineProps<{
  items: PageFeedbackItem[]
  selectedId: number | null
  loading: boolean
  pagination: PageFeedbackPagination
}>()

const emit = defineEmits<{
  select: [item: PageFeedbackItem]
  'update:page': [page: number]
  'update:page-size': [pageSize: number]
}>()
</script>
