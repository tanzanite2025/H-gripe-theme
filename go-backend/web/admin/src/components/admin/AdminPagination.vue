<template>
  <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
    <span class="text-xs text-muted-foreground">
      共 {{ total }} 项<span v-if="total">，显示 {{ rangeStart }}-{{ rangeEnd }}</span>
    </span>

    <div class="flex flex-wrap items-center gap-2">
      <Select v-model="pageSizeModel">
        <SelectTrigger class="h-8 w-24">
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
            @click="goTo(pageNumber)"
          >
            {{ pageNumber }}
          </Button>
        </div>

        <span class="min-w-14 text-center text-xs text-muted-foreground sm:hidden">
          {{ page }}/{{ totalPages }}
        </span>

        <Button
          variant="outline"
          size="icon"
          :disabled="page >= totalPages"
          aria-label="下一页"
          @click="goTo(page + 1)"
        >
          <ChevronRight class="size-3.5" />
        </Button>
      </div>

      <label class="hidden items-center gap-1 text-xs text-muted-foreground lg:flex">
        前往
        <Input
          v-model="jumpValue"
          type="number"
          min="1"
          :max="totalPages"
          class="h-8 w-16 text-center"
          @keyup.enter="jumpToPage"
          @change="jumpToPage"
        />
        页
      </label>
    </div>
  </div>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { ChevronLeft, ChevronRight } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

const props = defineProps({
  page: {
    type: Number,
    required: true
  },
  pageSize: {
    type: Number,
    required: true
  },
  total: {
    type: Number,
    required: true
  },
  pageSizes: {
    type: Array,
    default: () => [10, 20, 50, 100]
  }
})

const emit = defineEmits(['update:page', 'update:pageSize'])
const jumpValue = ref(String(props.page))

const totalPages = computed(() => Math.max(1, Math.ceil(props.total / props.pageSize)))
const rangeStart = computed(() => (props.total === 0 ? 0 : (props.page - 1) * props.pageSize + 1))
const rangeEnd = computed(() => Math.min(props.page * props.pageSize, props.total))
const pageSizeModel = computed({
  get: () => String(props.pageSize),
  set: (value) => {
    emit('update:pageSize', Number(value))
  }
})
const visiblePages = computed(() => {
  const start = Math.max(1, Math.min(props.page - 2, totalPages.value - 4))
  const end = Math.min(totalPages.value, start + 4)
  return Array.from({ length: end - start + 1 }, (_, index) => start + index)
})

const goTo = (nextPage) => {
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
