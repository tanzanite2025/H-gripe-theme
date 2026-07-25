<template>
  <TabsContent value="registrations" class="space-y-3">
    <div class="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h2 class="text-sm font-black tracking-tighter italic uppercase">产品注册记录</h2>
        <p class="mt-1 text-xs text-muted-foreground">按序列号、用户、商品和保修到期时间管理注册状态。</p>
      </div>
      <div class="w-full sm:w-48">
        <Select v-model="filters.status">
          <SelectTrigger class="h-9 w-full rounded-full">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部状态</SelectItem>
            <SelectItem v-for="option in statusOptions" :key="option.value" :value="option.value">
              {{ option.label }}
            </SelectItem>
          </SelectContent>
        </Select>
      </div>
    </div>

    <AdminTablePanel :loading="loading">
      <Table class="min-w-[1080px]">
        <TableHeader>
          <TableRow>
            <TableHead>注册商品</TableHead>
            <TableHead>序列号</TableHead>
            <TableHead>客户</TableHead>
            <TableHead>购买 / 到期</TableHead>
            <TableHead>凭证</TableHead>
            <TableHead class="w-40">状态</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="registrations.length === 0" :colspan="6">
            <div class="flex flex-col items-center text-muted-foreground">
              <ShieldCheck class="mb-2 size-7 opacity-55" />
              <span class="text-xs">暂无产品注册记录</span>
            </div>
          </TableEmpty>
          <TableRow v-for="registration in registrations" :key="registration.id">
            <TableCell>
              <span class="block font-bold text-xs">{{ productName(registration.product) }}</span>
              <span class="block font-mono text-[10px] text-muted-foreground/70">product_id={{ registration.product_id || '-' }}</span>
            </TableCell>
            <TableCell>
              <span class="font-mono text-xs font-bold">{{ registration.serial_number || '-' }}</span>
            </TableCell>
            <TableCell>
              <span class="block text-xs font-bold">{{ userName(registration.user) }}</span>
              <span class="block max-w-56 truncate text-[10px] text-muted-foreground/70">{{ registration.user?.email || '-' }}</span>
            </TableCell>
            <TableCell>
              <span class="block text-xs">购买：{{ formatDate(registration.purchase_date) }}</span>
              <span class="block text-[10px] text-muted-foreground/70">到期：{{ formatDate(registration.warranty_expires) }}</span>
            </TableCell>
            <TableCell>
              <Button
                v-if="registration.purchase_proof"
                variant="ghost"
                size="sm"
                class="h-7 rounded-full px-2 text-xs"
                as-child
              >
                <a :href="registration.purchase_proof" target="_blank" rel="noopener noreferrer">查看凭证</a>
              </Button>
              <span v-else class="text-xs text-muted-foreground">无凭证</span>
            </TableCell>
            <TableCell>
              <Select
                :model-value="registration.status || 'active'"
                :disabled="statusUpdating.registration === registration.id || !canEdit"
                @update:model-value="$emit('update-status', registration, $event)"
              >
                <SelectTrigger class="h-8 w-full rounded-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem v-for="option in statusOptions" :key="option.value" :value="option.value">
                    {{ option.label }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
      <template #footer>
        <AdminPagination
          :page="pagination.page"
          :page-size="pagination.pageSize"
          :total="pagination.total"
          @update:page="$emit('update-page', $event)"
          @update:page-size="$emit('update-page-size', $event)"
        />
      </template>
    </AdminTablePanel>
  </TabsContent>
</template>

<script setup>
import { ShieldCheck } from '@lucide/vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { TabsContent } from '@/components/ui/tabs'
import {
  formatDate,
  productName,
  userName
} from '@/lib/warrantyPresentation'

defineProps({
  registrations: { type: Array, required: true },
  loading: { type: Boolean, default: false },
  filters: { type: Object, required: true },
  pagination: { type: Object, required: true },
  statusUpdating: { type: Object, required: true },
  statusOptions: { type: Array, required: true },
  canEdit: { type: Boolean, default: false }
})

defineEmits(['update-status', 'update-page', 'update-page-size'])
</script>
