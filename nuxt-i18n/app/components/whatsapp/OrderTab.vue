<template>
  <div class="h-full overflow-y-auto px-1 md:p-6">
    <div v-if="isLoadingOrders" class="text-center tz-text-secondary py-10 md:py-12 text-sm">
      Loading orders...
    </div>
    <div v-else-if="ordersList.length > 0" class="space-y-2 md:space-y-0 md:grid md:grid-cols-2 md:gap-3">
      <div
        v-for="order in ordersList"
        :key="order.order_number"
        class="border border-white/15 md:border-white/10 rounded-2xl md:rounded-lg p-3 bg-black/35 transition-colors"
      >
        <div class="flex items-center justify-between mb-1 md:mb-2">
          <span class="text-white text-sm font-semibold md:font-medium">
            {{ order.order_number ? `Order #${order.order_number}` : 'Order' }}
          </span>
          <span class="tz-micro-label md:text-xs px-2 py-0.5 rounded-full bg-white/15 md:bg-white/10 tz-text-secondary">
            {{ order.status || 'Processing' }}
          </span>
        </div>
        <p class="tz-text-secondary text-xs">{{ order.total }} {{ order.currency || '' }}</p>
        <p v-if="order.item_count" class="tz-caption md:text-xs tz-text-secondary mt-1">
          {{ order.item_count }} item{{ Number(order.item_count) > 1 ? 's' : '' }}
        </p>
        <div v-if="order.items?.length" class="mt-2 space-y-1">
          <div
            v-for="item in order.items.slice(0, 2)"
            :key="item.id || `${item.product_id}-${item.sku}`"
            class="flex items-center justify-between gap-2 rounded-xl bg-white/[0.04] px-2 py-1 tz-caption tz-text-secondary"
          >
            <span class="truncate">{{ item.product_name || item.title || 'Product' }}</span>
            <span class="shrink-0">x{{ item.quantity || 1 }}</span>
          </div>
          <p v-if="order.items.length > 2" class="tz-micro-label tz-text-muted">
            +{{ order.items.length - 2 }} more
          </p>
        </div>
        <p class="tz-text-muted tz-caption md:text-xs mt-1">{{ order.date }}</p>
        <button
          type="button"
          class="mt-3 w-full rounded-full border border-white/40 bg-white px-3 py-2 tz-caption font-semibold text-slate-950 transition-colors hover:bg-white/90"
          @click="$emit('shareOrder', order)"
        >
          和客服确认订单
        </button>
      </div>
    </div>
    <div v-else class="text-center tz-text-secondary text-sm py-10 md:py-12">
      No orders yet
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  ordersList: any[]
  isLoadingOrders: boolean
}>()

defineEmits<{
  'shareOrder': [order: any]
}>()
</script>
