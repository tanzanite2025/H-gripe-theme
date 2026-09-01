<template>
  <AdminFilterPanel class="shrink-0">
    <form
      class="grid grid-cols-1 gap-3 xl:grid-cols-[150px_220px_220px_minmax(0,1fr)_auto]"
      @submit.prevent="apply"
    >
      <label class="block space-y-1">
        <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">STATUS / 状态</span>
        <select v-model="draft.status" class="h-9 w-full rounded-md border border-dashed border-border bg-background px-3 text-sm">
          <option value="pending">待处理</option>
          <option value="approved">已发布</option>
          <option value="rejected">已拒绝</option>
          <option value="hidden">已隐藏</option>
          <option value="all">全部</option>
        </select>
      </label>

      <label class="block space-y-1">
        <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">PAGE / 页面</span>
        <Input v-model="draft.pagePath" placeholder="/membershipandpoints/exchange" />
      </label>

      <label class="block space-y-1">
        <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">THREAD / 线程</span>
        <Input v-model="draft.threadKey" placeholder="membership-exchange" />
      </label>

      <label class="block space-y-1">
        <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">SEARCH / 搜索</span>
        <Input v-model="draft.search" placeholder="留言、姓名、邮箱、页面标题" />
      </label>

      <div class="flex items-end gap-2">
        <Button type="submit" class="h-9 px-3 text-xs font-black uppercase tracking-wider" :disabled="loading">
          <Search class="size-3.5" />
          查询
        </Button>
        <Button type="button" variant="outline" class="h-9 px-3 text-xs font-black uppercase tracking-wider" @click="reset">
          重置
        </Button>
      </div>
    </form>
  </AdminFilterPanel>
</template>

<script setup lang="ts">
import { reactive, watch } from 'vue'
import { Search } from '@lucide/vue'
import AdminFilterPanel from '@/components/admin/AdminFilterPanel.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import type { PageFeedbackFiltersState } from '@/modules/content-feedback/pageFeedbackTypes'

const props = defineProps<{
  filters: PageFeedbackFiltersState
  loading: boolean
}>()

const emit = defineEmits<{
  apply: [filters: PageFeedbackFiltersState]
  reset: []
}>()

const draft = reactive<PageFeedbackFiltersState>({ ...props.filters })

watch(
  () => props.filters,
  (value) => Object.assign(draft, value),
  { deep: true },
)

const apply = (): void => {
  emit('apply', {
    status: draft.status,
    pagePath: draft.pagePath.trim(),
    threadKey: draft.threadKey.trim(),
    search: draft.search.trim(),
  })
}

const reset = (): void => {
  emit('reset')
}
</script>

