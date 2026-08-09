<template>
  <AdminTablePanel :loading="loading">
    <Table class="min-w-[1440px]">
      <TableHeader>
        <TableRow>
          <TableHead class="w-16">ID</TableHead>
          <TableHead class="w-36">用户</TableHead>
          <TableHead class="w-24">操作</TableHead>
          <TableHead class="w-28">资源</TableHead>
          <TableHead class="w-24">资源 ID</TableHead>
          <TableHead class="w-20">方法</TableHead>
          <TableHead>路径</TableHead>
          <TableHead class="w-36">IP 地址</TableHead>
          <TableHead class="w-24">状态</TableHead>
          <TableHead class="w-24 text-right">耗时</TableHead>
          <TableHead class="w-44">时间</TableHead>
          <TableHead class="w-16 text-right">详情</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableEmpty v-if="logs.length === 0" :colspan="12">
          <div class="flex flex-col items-center text-muted-foreground">
            <ScrollText class="mb-2 size-7 opacity-55" />
            <span class="text-xs">暂无审计日志</span>
          </div>
        </TableEmpty>
        <TableRow v-for="log in logs" :key="log.id">
          <TableCell class="font-mono text-xs text-muted-foreground">{{ log.id }}</TableCell>
          <TableCell>
            <span class="block truncate font-medium">{{ log.username || '-' }}</span>
            <span class="block font-mono text-[11px] text-muted-foreground">ID {{ log.user_id || '-' }}</span>
          </TableCell>
          <TableCell><AdminStatusBadge :tone="actionTone(log.action)">{{ actionName(log.action) }}</AdminStatusBadge></TableCell>
          <TableCell>{{ resourceName(log.resource) }}</TableCell>
          <TableCell class="font-mono text-xs">{{ log.resource_id || '-' }}</TableCell>
          <TableCell><AdminStatusBadge :tone="methodTone(log.method)">{{ log.method || '-' }}</AdminStatusBadge></TableCell>
          <TableCell class="max-w-96 truncate font-mono text-xs text-muted-foreground">{{ log.path || '-' }}</TableCell>
          <TableCell class="font-mono text-xs">{{ log.ip_address || '-' }}</TableCell>
          <TableCell><AdminStatusBadge :tone="log.status === 'success' ? 'green' : 'coral'">{{ log.status === 'success' ? '成功' : '失败' }}</AdminStatusBadge></TableCell>
          <TableCell class="text-right tabular-nums" :class="durationClass(log.duration)">{{ log.duration || 0 }} ms</TableCell>
          <TableCell class="text-xs text-muted-foreground">{{ formatDate(log.created_at) }}</TableCell>
          <TableCell class="text-right">
            <Button variant="ghost" size="icon" :aria-label="`查看日志 ${log.id}`" @click="emit('view-detail', log)">
              <Eye class="size-4" />
            </Button>
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>
    <template #footer>
      <AdminPagination
        :page="pagination.page"
        :page-size="pagination.pageSize"
        :total="pagination.total"
        @update:page="emit('update-page', $event)"
        @update:page-size="emit('update-page-size', $event)"
      />
    </template>
  </AdminTablePanel>
</template>

<script setup lang="ts">
import { Eye, ScrollText } from '@lucide/vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import type {
  AuditLogDateFormatter,
  AuditLogDurationClassResolver,
  AuditLogLabelResolver,
  AuditLogPagination,
  AuditLogRecord,
  AuditLogToneResolver
} from './auditLogTypes'

withDefaults(defineProps<{
  loading?: boolean
  logs?: AuditLogRecord[]
  pagination: AuditLogPagination
  actionName: AuditLogLabelResolver
  actionTone: AuditLogToneResolver
  resourceName: AuditLogLabelResolver
  methodTone: AuditLogToneResolver
  durationClass: AuditLogDurationClassResolver
  formatDate: AuditLogDateFormatter
}>(), {
  loading: false,
  logs: () => []
})

const emit = defineEmits<{
  (event: 'view-detail', log: AuditLogRecord): void
  (event: 'update-page', page: number): void
  (event: 'update-page-size', pageSize: number): void
}>()
</script>
