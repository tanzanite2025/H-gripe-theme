<template>
  <AdminTablePanel :loading="loading">
    <Table class="min-w-[820px]">
      <TableHeader>
        <TableRow>
          <TableHead class="w-20">ID</TableHead>
          <TableHead>用户名</TableHead>
          <TableHead>邮箱</TableHead>
          <TableHead>姓名</TableHead>
          <TableHead class="w-24">状态</TableHead>
          <TableHead class="w-44">注册时间</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableEmpty v-if="customers.length === 0" :colspan="6">
          <div class="flex flex-col items-center text-muted-foreground">
            <UsersRound class="mb-2 size-7 opacity-55" />
            <span class="text-xs">暂无客户账户</span>
          </div>
        </TableEmpty>

        <TableRow v-for="customer in customers" :key="customer.id">
          <TableCell class="font-mono text-xs text-muted-foreground">{{ customer.id }}</TableCell>
          <TableCell class="font-bold text-xs">{{ customer.username }}</TableCell>
          <TableCell class="max-w-64 truncate text-muted-foreground">{{ customer.email }}</TableCell>
          <TableCell>{{ formatFullName(customer) }}</TableCell>
          <TableCell>
            <AdminStatusBadge :tone="statusTone(customer.status)">{{ getStatusName(customer.status) }}</AdminStatusBadge>
          </TableCell>
          <TableCell class="text-xs text-muted-foreground">{{ formatDate(customer.created_at) }}</TableCell>
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
import { UsersRound } from '@lucide/vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import type {
  CustomerAccount,
  CustomerDateFormatter,
  CustomerNameFormatter,
  CustomerPagination,
  CustomerStatusNameResolver,
  CustomerStatusToneResolver
} from './customerTypes'

withDefaults(defineProps<{
  loading?: boolean
  customers?: CustomerAccount[]
  pagination: CustomerPagination
  getStatusName: CustomerStatusNameResolver
  statusTone: CustomerStatusToneResolver
  formatDate: CustomerDateFormatter
  formatFullName: CustomerNameFormatter
}>(), {
  loading: false,
  customers: () => []
})

const emit = defineEmits<{
  (event: 'update-page', page: number): void
  (event: 'update-page-size', pageSize: number): void
}>()
</script>
