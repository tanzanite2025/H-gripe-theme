<template>
  <TabsContent value="expiring" class="space-y-3">
    <div class="rounded-[24px] border border-dashed bg-card p-4">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 class="text-sm font-black tracking-tighter italic uppercase">30 天内到期</h2>
          <p class="mt-1 text-xs text-muted-foreground">从后端 `/registrations/expiring` 读取，不在前端重新推算。</p>
        </div>
        <AdminStatusBadge tone="amber">{{ expiring.length }} 条</AdminStatusBadge>
      </div>
    </div>

    <AdminTablePanel :loading="loading">
      <Table class="min-w-[980px]">
        <TableHeader>
          <TableRow>
            <TableHead>商品</TableHead>
            <TableHead>序列号</TableHead>
            <TableHead>客户</TableHead>
            <TableHead>保修到期</TableHead>
            <TableHead>状态</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="expiring.length === 0" :colspan="5">
            <div class="flex flex-col items-center text-muted-foreground">
              <Clock3 class="mb-2 size-7 opacity-55" />
              <span class="text-xs">暂无即将到期记录</span>
            </div>
          </TableEmpty>
          <TableRow v-for="item in expiring" :key="item.id">
            <TableCell>{{ productName(item.product) }}</TableCell>
            <TableCell class="font-mono text-xs font-bold">{{ item.serial_number || '-' }}</TableCell>
            <TableCell>{{ userName(item.user) }}</TableCell>
            <TableCell>{{ formatDate(item.warranty_expires) }}</TableCell>
            <TableCell>
              <AdminStatusBadge :tone="registrationStatusTone(item.status)">
                {{ registrationStatusLabel(item.status) }}
              </AdminStatusBadge>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </AdminTablePanel>
  </TabsContent>
</template>

<script setup>
import { Clock3 } from '@lucide/vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { TabsContent } from '@/components/ui/tabs'
import {
  formatDate,
  productName,
  registrationStatusLabel,
  registrationStatusTone,
  userName
} from '@/lib/warrantyPresentation'

defineProps({
  expiring: { type: Array, required: true },
  loading: { type: Boolean, default: false }
})
</script>
