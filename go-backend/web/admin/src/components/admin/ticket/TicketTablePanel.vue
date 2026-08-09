<template>
  <AdminTablePanel :loading="loading">
    <Table class="min-w-[1080px]">
      <TableHeader>
        <TableRow>
          <TableHead class="w-44">工单号</TableHead>
          <TableHead>标题</TableHead>
          <TableHead class="w-28">分类</TableHead>
          <TableHead class="w-24">状态</TableHead>
          <TableHead class="w-24">优先级</TableHead>
          <TableHead class="w-40">用户</TableHead>
          <TableHead class="w-32">负责人</TableHead>
          <TableHead class="w-44">创建时间</TableHead>
          <TableHead class="w-16 text-right">操作</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableEmpty v-if="tickets.length === 0" :colspan="9">
          <div class="flex flex-col items-center text-muted-foreground">
            <MessagesSquare class="mb-2 size-7 opacity-55" />
            <span class="text-xs">暂无工单</span>
          </div>
        </TableEmpty>

        <TableRow v-for="ticket in tickets" :key="ticket.id">
          <TableCell class="font-mono text-xs font-bold">{{ ticket.ticket_number }}</TableCell>
          <TableCell class="max-w-80 truncate font-bold text-xs">{{ ticket.subject }}</TableCell>
          <TableCell>{{ categoryName(ticket.category) }}</TableCell>
          <TableCell>
            <AdminStatusBadge :tone="statusTone(ticket.status)">{{ statusName(ticket.status) }}</AdminStatusBadge>
          </TableCell>
          <TableCell>
            <AdminStatusBadge :tone="priorityTone(ticket.priority)">{{ priorityName(ticket.priority) }}</AdminStatusBadge>
          </TableCell>
          <TableCell class="max-w-40 truncate">{{ customerName(ticket) }}</TableCell>
          <TableCell>{{ assigneeName(ticket.assigned_to) }}</TableCell>
          <TableCell class="text-xs text-muted-foreground">{{ formatDate(ticket.created_at) }}</TableCell>
          <TableCell class="text-right">
            <DropdownMenu>
              <DropdownMenuTrigger as-child>
                <Button variant="ghost" size="icon" :aria-label="`管理工单 ${ticket.ticket_number}`">
                  <MoreHorizontal class="size-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" class="w-40">
                <DropdownMenuItem @select="emit('view', ticket)">
                  <Eye class="size-4" />
                  查看详情
                </DropdownMenuItem>
                <DropdownMenuItem v-if="canEdit" @select="emit('assign', ticket)">
                  <UserRoundCog class="size-4" />
                  分配工单
                </DropdownMenuItem>
                <DropdownMenuSeparator v-if="canDelete" />
                <DropdownMenuItem
                  v-if="canDelete"
                  class="text-destructive focus:text-destructive"
                  @select="emit('delete', ticket)"
                >
                  <Trash2 class="size-4" />
                  删除
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
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
import { Eye, MessagesSquare, MoreHorizontal, Trash2, UserRoundCog } from '@lucide/vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import type {
  TicketAssigneeResolver,
  TicketCustomerResolver,
  TicketDateFormatter,
  TicketLabelResolver,
  TicketPagination,
  TicketRecord,
  TicketToneResolver
} from './ticketTypes'

withDefaults(defineProps<{
  loading?: boolean
  tickets?: TicketRecord[]
  pagination: TicketPagination
  canEdit?: boolean
  canDelete?: boolean
  categoryName: TicketLabelResolver
  statusName: TicketLabelResolver
  statusTone: TicketToneResolver
  priorityName: TicketLabelResolver
  priorityTone: TicketToneResolver
  customerName: TicketCustomerResolver
  assigneeName: TicketAssigneeResolver
  formatDate: TicketDateFormatter
}>(), {
  loading: false,
  tickets: () => [],
  canEdit: false,
  canDelete: false
})

const emit = defineEmits<{
  (event: 'view', ticket: TicketRecord): void
  (event: 'assign', ticket: TicketRecord): void
  (event: 'delete', ticket: TicketRecord): void
  (event: 'update-page', page: number): void
  (event: 'update-page-size', pageSize: number): void
}>()
</script>
