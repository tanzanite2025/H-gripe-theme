<template>
  <div class="flex flex-wrap items-center justify-between gap-2 text-xs text-muted-foreground">
    <span>共 {{ pagination.total || 0 }} 条，第 {{ pagination.page || 1 }} / {{ pagination.total_pages || 1 }} 页</span>
    <div class="flex items-center gap-2">
      <button
        type="button"
        class="rounded-full border border-dashed px-3 py-1 font-bold disabled:opacity-40"
        :disabled="(pagination.page || 1) <= 1"
        @click="emit('page', Math.max(1, (pagination.page || 1) - 1))"
      >
        上一页
      </button>
      <button
        type="button"
        class="rounded-full border border-dashed px-3 py-1 font-bold disabled:opacity-40"
        :disabled="(pagination.page || 1) >= (pagination.total_pages || 1)"
        @click="emit('page', (pagination.page || 1) + 1)"
      >
        下一页
      </button>
      <select
        class="h-7 rounded-full border border-dashed bg-background px-2"
        :value="pagination.page_size || 20"
        @change="emitPageSize"
      >
        <option v-for="size in pageSizes" :key="size" :value="size">{{ size }}/页</option>
      </select>
    </div>
  </div>
</template>

<script setup lang="ts">
interface RiskStrategyPagination {
  page?: number
  page_size?: number
  total?: number
  total_pages?: number
}

withDefaults(defineProps<{
  pagination: RiskStrategyPagination
  pageSizes?: number[]
}>(), {
  pageSizes: () => [10, 20, 50, 100],
})

const emit = defineEmits<{
  (event: 'page', page: number): void
  (event: 'page-size', pageSize: number): void
}>()

const emitPageSize = (event: Event) => {
  const target = event.target
  if (target instanceof HTMLSelectElement) {
    emit('page-size', Number(target.value))
  }
}
</script>
