<template>
  <AdminFilterPanel>
    <form class="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-[minmax(200px,1.2fr)_repeat(3,minmax(130px,0.7fr))_repeat(2,minmax(130px,0.7fr))_auto]" @submit.prevent="emit('apply')">
      <label class="space-y-1 block">
        <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">SEARCH / 搜索</span>
        <div class="relative">
          <Search class="pointer-events-none absolute left-3 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground/60" />
          <Input v-model="filters.search" class="h-9 pl-9" placeholder="订单号、客户或 Email" />
        </div>
      </label>

      <AdminFilterSelect v-model="filters.status" label="订单状态" :options="orderStatusOptions" />
      <AdminFilterSelect v-model="filters.payment_status" label="支付状态" :options="paymentStatusOptions" />
      <AdminFilterSelect v-model="filters.shipping_status" label="物流状态" :options="shippingStatusOptions" />

      <label class="space-y-1 block">
        <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">START DATE / 开始日期</span>
        <Input v-model="filters.start_date" type="date" class="h-9" />
      </label>

      <label class="space-y-1 block">
        <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">END DATE / 结束日期</span>
        <Input v-model="filters.end_date" type="date" class="h-9" />
      </label>

      <label class="space-y-1 block">
        <span class="block text-[10px] font-black uppercase tracking-widest text-transparent select-none">ACTION / 操作</span>
        <div class="flex items-center gap-2">
          <Button type="submit" class="h-9 rounded-full px-4 font-black text-xs uppercase tracking-wider">
            <Search class="size-3.5" />
            搜索
          </Button>
          <Button type="button" variant="outline" class="h-9 rounded-full px-3 font-black text-xs uppercase tracking-wider" @click="emit('reset')">
            <RotateCcw class="size-3.5" />
            重置
          </Button>
        </div>
      </label>
    </form>
  </AdminFilterPanel>
</template>

<script setup lang="ts">
import { RotateCcw, Search } from '@lucide/vue'
import AdminFilterPanel from '@/components/admin/AdminFilterPanel.vue'
import AdminFilterSelect from '@/components/admin/AdminFilterSelect.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import type { OrderFilters, OrderStatusOption } from '@/modules/order/orderTypes'

withDefaults(defineProps<{
  filters: OrderFilters
  orderStatusOptions?: OrderStatusOption[]
  paymentStatusOptions?: OrderStatusOption[]
  shippingStatusOptions?: OrderStatusOption[]
}>(), {
  orderStatusOptions: () => [],
  paymentStatusOptions: () => [],
  shippingStatusOptions: () => []
})

const emit = defineEmits<{
  (event: 'apply'): void
  (event: 'reset'): void
}>()
</script>

