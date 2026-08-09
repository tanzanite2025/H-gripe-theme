<template>
  <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
    <span class="text-[10px] font-mono font-bold uppercase tracking-wider text-muted-foreground/70">
      TOTAL: {{ total }}<span v-if="total"> ({{ rangeStart }}-{{ rangeEnd }})</span>
    </span>

    <div class="flex flex-wrap items-center gap-2">
      <Select v-model="pageSizeModel">
        <SelectTrigger class="h-8 w-24 rounded-full">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem v-for="size in pageSizes" :key="size" :value="String(size)">
            {{ size }} 条/页
          </SelectItem>
        </SelectContent>
      </Select>

      <div class="flex items-center gap-1">
        <Button
          variant="outline"
          size="icon"
          class="rounded-full size-8"
          :disabled="page <= 1"
          aria-label="上一页"
          @click="goTo(page - 1)"
        >
          <ChevronLeft class="size-3.5" />
        </Button>

        <div class="hidden items-center gap-1 sm:flex">
          <Button
            v-for="pageNumber in visiblePages"
            :key="pageNumber"
            :variant="pageNumber === page ? 'default' : 'outline'"
            size="icon"
            class="rounded-full size-8 font-mono font-bold text-xs"
            @click="goTo(pageNumber)"
          >
            {{ pageNumber }}
          </Button>
        </div>

        <span class="min-w-14 text-center text-[10px] font-mono font-bold text-muted-foreground sm:hidden">
          {{ page }}/{{ totalPages }}
        </span>

        <Button
          variant="outline"
          size="icon"
          class="rounded-full size-8"
          :disabled="page >= totalPages"
          aria-label="下一页"
          @click="goTo(page + 1)"
        >
          <ChevronRight class="size-3.5" />
        </Button>
      </div>

      <label class="hidden items-center gap-1.5 text-[10px] font-mono font-bold uppercase tracking-wider text-muted-foreground/70 lg:flex">
        GOTO / 前往
        <Input
          v-model="jumpValue"
          type="number"
          min="1"
          :max="totalPages"
          class="h-8 w-16 text-center font-mono font-bold text-xs rounded-xl"
          @keyup.enter="jumpToPage"
          @change="jumpToPage"
        />
        PAGE
      </label>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ChevronLeft, ChevronRight } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

const props = withDefaults(defineProps<{
  page: number
  pageSize: number
  total: number
  pageSizes?: number[]
}>(), {
  pageSizes: () => [10, 20, 50, 100]
})

const emit = defineEmits<{
  (event: 'update:page', value: number): void
  (event: 'update:pageSize', value: number): void
}>()
const jumpValue = ref(String(props.page))

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))
const rangeStart = computed(() => (props.total === 0 ? 0 : (props.page - 1) * props.pageSize + 1))
const rangeEnd = computed(() => Math.min(props.page * props.pageSize, props.total))
const pageSizeModel = computed<string>({
  get: () => String(props.pageSize),
  set: (value: string) => {
    emit('update:pageSize', Number(value))
  }
})
const visiblePages = computed(() => {
  const start = Math.max(1, Math.min(props.page - 2, totalPages.value - 4))
  const end = Math.min(totalPages.value, start + 4)
  return Array.from({ length: end - start + 1 }, (_, index) => start + index)
})

const goTo = (nextPage: number) => {
  const normalized = Math.min(totalPages.value, Math.max(1, nextPage))
  if (normalized !== props.page) emit('update:page', normalized)
}

const jumpToPage = () => {
  const nextPage = Number(jumpValue.value)
  if (Number.isFinite(nextPage)) goTo(nextPage)
  jumpValue.value = String(props.page)
}

watch(
  () => props.page,
  (page) => {
    jumpValue.value = String(page)
  }
)
</script>
