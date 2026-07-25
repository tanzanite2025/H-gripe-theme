<template>
  <div class="space-y-3">
    <div class="flex flex-wrap items-end justify-between gap-3">
      <label class="w-48 space-y-1 block">
        <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">STATUS / 状态</span>
        <Select v-model="filters.status" @update:model-value="emit('filter-change')">
          <SelectTrigger class="h-9 w-full"><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部状态</SelectItem>
            <SelectItem value="active">生效中</SelectItem>
            <SelectItem value="expired">已过期</SelectItem>
            <SelectItem value="disabled">已停用</SelectItem>
          </SelectContent>
        </Select>
      </label>
      <Button v-if="canCreate" size="sm" @click="emit('create')">
        <Plus class="size-3.5" />
        创建优惠券
      </Button>
    </div>

    <AdminTablePanel :loading="loading">
      <Table class="min-w-[1080px]">
        <TableHeader>
          <TableRow>
            <TableHead class="w-36">优惠码</TableHead>
            <TableHead class="w-24">类型</TableHead>
            <TableHead class="w-28 text-right">折扣值</TableHead>
            <TableHead>描述</TableHead>
            <TableHead class="w-32 text-right">最低消费</TableHead>
            <TableHead class="w-28">使用情况</TableHead>
            <TableHead class="w-60">有效期</TableHead>
            <TableHead class="w-24">状态</TableHead>
            <TableHead class="w-16 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="coupons.length === 0" :colspan="9">
            <div class="flex flex-col items-center text-muted-foreground">
              <BadgePercent class="mb-2 size-7 opacity-55" />
              <span class="text-xs">暂无优惠券</span>
            </div>
          </TableEmpty>
          <TableRow v-for="coupon in coupons" :key="coupon.id">
            <TableCell class="font-mono text-xs font-bold">{{ coupon.code }}</TableCell>
            <TableCell>{{ coupon.type === 'fixed' ? '固定金额' : '百分比' }}</TableCell>
            <TableCell class="text-right font-medium tabular-nums">{{ couponValue(coupon) }}</TableCell>
            <TableCell class="max-w-64 truncate text-muted-foreground">{{ coupon.description || '-' }}</TableCell>
            <TableCell class="text-right tabular-nums">¥{{ formatMoney(coupon.min_amount) }}</TableCell>
            <TableCell class="tabular-nums">{{ coupon.used_count || 0 }} / {{ coupon.usage_limit || '不限' }}</TableCell>
            <TableCell class="text-xs text-muted-foreground">
              {{ formatDate(coupon.start_date) }}<br />{{ formatDate(coupon.end_date) }}
            </TableCell>
            <TableCell>
              <AdminStatusBadge :tone="couponStatus(coupon).tone">{{ couponStatus(coupon).label }}</AdminStatusBadge>
            </TableCell>
            <TableCell class="text-right">
              <DropdownMenu>
                <DropdownMenuTrigger as-child>
                  <Button variant="ghost" size="icon" :aria-label="`管理优惠券 ${coupon.code}`">
                    <MoreHorizontal class="size-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" class="w-36">
                  <DropdownMenuItem v-if="canEdit" @select="emit('edit', coupon)">
                    <Pencil class="size-4" />
                    编辑
                  </DropdownMenuItem>
                  <DropdownMenuSeparator v-if="canDelete" />
                  <DropdownMenuItem
                    v-if="canDelete"
                    class="text-destructive focus:text-destructive"
                    @select="emit('delete', coupon)"
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
  </div>
</template>

<script setup>
import { BadgePercent, MoreHorizontal, Pencil, Plus, Trash2 } from '@lucide/vue'
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
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'

defineProps({
  loading: { type: Boolean, default: false },
  coupons: { type: Array, default: () => [] },
  filters: { type: Object, required: true },
  pagination: { type: Object, required: true },
  canCreate: { type: Boolean, default: false },
  canEdit: { type: Boolean, default: false },
  canDelete: { type: Boolean, default: false },
  couponValue: { type: Function, required: true },
  couponStatus: { type: Function, required: true },
  formatMoney: { type: Function, required: true },
  formatDate: { type: Function, required: true },
})

const emit = defineEmits(['filter-change', 'create', 'edit', 'delete', 'update-page', 'update-page-size'])
</script>
