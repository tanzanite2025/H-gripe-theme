<template>
  <div class="h-full overflow-y-auto px-1 md:p-6">
    <div v-if="isLoadingOrders" class="text-center text-white/50 py-10 md:py-12 text-sm">
      Loading orders...
    </div>
    <div v-else-if="ordersList.length > 0" class="space-y-2 md:space-y-0 md:grid md:grid-cols-2 md:gap-3">
      <div
        v-for="order in ordersList"
        :key="order.id"
        @click="$emit('shareOrder', order)"
        class="border border-white/15 md:border-white/10 rounded-2xl md:rounded-lg p-3 bg-black/35 md:hover:bg-white/[0.05] cursor-pointer transition-colors"
      >
        <div class="flex items-center justify-between mb-1 md:mb-2">
          <span class="text-white text-sm font-semibold md:font-medium">Order #{{ order.id }}</span>
          <span class="text-[10px] md:text-xs px-2 py-0.5 rounded-full bg-white/15 md:bg-white/10 text-white/70">
            {{ order.status || 'Processing' }}
          </span>
        </div>
        <p class="text-white/70 text-xs">{{ order.total }} {{ order.currency || '' }}</p>
        <p class="text-white/50 text-[11px] md:text-xs mt-1">{{ order.date }}</p>
      </div>
    </div>
    <div v-else class="text-center text-white/60 md:text-white/50 text-sm py-10 md:py-12">
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
