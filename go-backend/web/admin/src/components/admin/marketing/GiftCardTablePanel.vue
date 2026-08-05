<template>
  <div class="space-y-3">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div>
        <h3 class="text-sm font-black tracking-tighter italic uppercase text-foreground">已发行礼品卡</h3>
        <p class="mt-1 text-xs text-muted-foreground">这里仅查看已经生成的真实卡号和余额。</p>
      </div>
      <label class="w-48 space-y-1.5">
        <span class="text-xs font-medium text-muted-foreground">状态</span>
        <Select v-model="filters.status" @update:model-value="emit('filter-change')">
          <SelectTrigger class="h-9 w-full"><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部状态</SelectItem>
            <SelectItem value="active">活跃</SelectItem>
            <SelectItem value="used">已使用</SelectItem>
            <SelectItem value="expired">已过期</SelectItem>
            <SelectItem value="cancelled">已取消</SelectItem>
          </SelectContent>
        </Select>
      </label>
    </div>

    <AdminTablePanel :loading="loading">
      <Table class="min-w-[980px]">
        <TableHeader>
          <TableRow>
            <TableHead class="w-44">卡号</TableHead>
            <TableHead class="w-32 text-right">初始金额</TableHead>
            <TableHead class="w-32 text-right">余额</TableHead>
            <TableHead>收件人</TableHead>
            <TableHead class="w-24">状态</TableHead>
            <TableHead class="w-44">到期时间</TableHead>
            <TableHead class="w-44">创建时间</TableHead>
            <TableHead class="w-16 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="giftCards.length === 0" :colspan="8">
            <div class="flex flex-col items-center text-muted-foreground">
              <Gift class="mb-2 size-7 opacity-55" />
              <span class="text-xs">暂无礼品卡</span>
            </div>
          </TableEmpty>
          <TableRow v-for="giftCard in giftCards" :key="giftCard.id">
            <TableCell class="font-mono text-xs font-bold">{{ giftCard.code }}</TableCell>
            <TableCell class="text-right tabular-nums">{{ formatCurrency(giftCard.initial_value, giftCard.currency) }}</TableCell>
            <TableCell class="text-right font-bold tabular-nums">{{ formatCurrency(giftCard.balance, giftCard.currency) }}</TableCell>
            <TableCell>
              <span class="block font-bold text-xs">{{ giftCard.recipient_name || '-' }}</span>
              <span class="block text-xs text-muted-foreground">{{ giftCard.recipient_email || '-' }}</span>
            </TableCell>
            <TableCell>
              <AdminStatusBadge :tone="giftCardStatusTone(giftCard.status)">{{ giftCardStatusName(giftCard.status) }}</AdminStatusBadge>
            </TableCell>
            <TableCell class="text-xs text-muted-foreground">{{ formatDate(giftCard.expires_at) }}</TableCell>
            <TableCell class="text-xs text-muted-foreground">{{ formatDate(giftCard.created_at) }}</TableCell>
            <TableCell class="text-right">
              <Button variant="ghost" size="icon" :aria-label="`查看礼品卡 ${giftCard.code}`" @click="emit('view', giftCard)">
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
  </div>
</template>

<script setup>
import { Eye, Gift } from '@lucide/vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'

defineProps({
  loading: { type: Boolean, default: false },
  giftCards: { type: Array, default: () => [] },
  filters: { type: Object, required: true },
  pagination: { type: Object, required: true },
  formatCurrency: { type: Function, required: true },
  formatDate: { type: Function, required: true },
  giftCardStatusName: { type: Function, required: true },
  giftCardStatusTone: { type: Function, required: true },
})

const emit = defineEmits(['filter-change', 'view', 'update-page', 'update-page-size'])
</script>
