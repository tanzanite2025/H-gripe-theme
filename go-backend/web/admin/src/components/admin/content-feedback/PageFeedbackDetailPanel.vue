<template>
  <aside class="min-h-0 overflow-auto rounded-[24px] border border-dashed border-border/80 bg-card p-4">
    <div v-if="selected" class="space-y-5">
      <div class="flex items-start justify-between gap-3">
        <div class="min-w-0">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">
            FEEDBACK #{{ selected.id }}
          </p>
          <h2 class="mt-1 truncate text-base font-black">{{ displayPageTitle(selected) }}</h2>
          <p class="mt-1 truncate font-mono text-[11px] text-muted-foreground">{{ displayPagePath(selected) }}</p>
        </div>
        <Badge :variant="pageFeedbackStatusBadgeVariant(selected.status)">
          {{ pageFeedbackStatusLabel(selected.status) }}
        </Badge>
      </div>

      <div class="grid grid-cols-2 gap-2 text-xs">
        <div class="rounded-xl bg-muted/40 p-3">
          <p class="font-black uppercase text-muted-foreground">线程</p>
          <p class="mt-1 truncate font-mono">{{ selected.thread_key }}</p>
          <p class="mt-1 text-[10px] text-muted-foreground">{{ selected.locale || '-' }}</p>
        </div>
        <div class="rounded-xl bg-muted/40 p-3">
          <p class="font-black uppercase text-muted-foreground">用户</p>
          <p class="mt-1 truncate">{{ selected.name || '未填写姓名' }}</p>
          <p class="truncate text-[10px] text-muted-foreground">{{ selected.email || `用户 #${selected.user_id}` }}</p>
          <p class="truncate font-mono text-[10px] text-muted-foreground/80">
            来源 {{ displaySourceHashPreview(selected.source_hash_preview) }}
          </p>
        </div>
      </div>

      <div class="space-y-2">
        <p class="text-xs font-semibold">用户留言</p>
        <div class="rounded-xl border border-dashed border-border/80 bg-muted/20 p-3 text-sm leading-6 whitespace-pre-wrap">
          {{ selected.content }}
        </div>
      </div>

      <div class="border-y border-dashed py-3 text-xs text-muted-foreground">
        <p>提交：{{ formatPageFeedbackDate(selected.created_at) }}</p>
        <p v-if="selected.reviewed_at">处理：{{ formatPageFeedbackDate(selected.reviewed_at) }}</p>
        <p v-if="selected.replied_at">回复：{{ formatPageFeedbackDate(selected.replied_at) }}</p>
      </div>

      <div class="space-y-2">
        <label for="feedback-reply-content" class="text-xs font-semibold">公开回复</label>
        <Textarea
          id="feedback-reply-content"
          v-model="replyContent"
          class="min-h-32 resize-y"
          maxlength="3000"
          placeholder="这里的回复会随已发布留言显示在前台。留空则不显示商家回复。"
          :disabled="!canEdit"
        />
      </div>

      <div class="space-y-2">
        <label class="text-xs font-semibold">处理状态</label>
        <select v-model="nextStatus" class="h-9 w-full rounded-md border border-dashed border-border bg-background px-3 text-sm" :disabled="!canEdit">
          <option value="pending">待处理</option>
          <option value="approved">已发布</option>
          <option value="rejected">已拒绝</option>
          <option value="hidden">已隐藏</option>
        </select>
      </div>

      <div v-if="canEdit" class="flex flex-wrap gap-2">
        <Button :disabled="submitting" @click="save">
          <Send class="size-4" />
          保存处理
        </Button>
        <Button
          v-if="selected.status !== 'approved'"
          variant="outline"
          :disabled="submitting"
          @click="quickUpdate('approved')"
        >
          <CircleCheck class="size-4" />
          通过
        </Button>
        <Button
          v-if="selected.status !== 'hidden'"
          variant="outline"
          :disabled="submitting"
          @click="quickUpdate('hidden')"
        >
          <EyeOff class="size-4" />
          隐藏
        </Button>
        <Button
          v-if="selected.status !== 'pending'"
          variant="outline"
          :disabled="submitting"
          @click="quickUpdate('pending')"
        >
          <RotateCcw class="size-4" />
          待处理
        </Button>
      </div>

      <div v-else class="rounded-xl border border-dashed border-border/80 bg-muted/20 p-3 text-xs text-muted-foreground">
        当前账号只有查看权限，需要 content:edit 才能回复或处理留言。
      </div>
    </div>

    <div v-else class="flex h-full min-h-64 items-center justify-center text-center text-sm text-muted-foreground">
      选择一条页面留言查看详情
    </div>
  </aside>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { CircleCheck, EyeOff, RotateCcw, Send } from '@lucide/vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import {
  displayPagePath,
  displayPageTitle,
  displaySourceHashPreview,
  formatPageFeedbackDate,
  pageFeedbackStatusBadgeVariant,
  pageFeedbackStatusLabel,
  type PageFeedbackItem,
  type PageFeedbackStatus,
} from './pageFeedbackTypes'

const props = defineProps<{
  selected: PageFeedbackItem | null
  canEdit: boolean
  submitting: boolean
}>()

const emit = defineEmits<{
  save: [payload: { status: PageFeedbackStatus; reply_content: string }]
}>()

const replyContent = ref('')
const nextStatus = ref<PageFeedbackStatus>('pending')

watch(
  () => props.selected,
  (item) => {
    replyContent.value = item?.reply_content || ''
    nextStatus.value = item?.status || 'pending'
  },
  { immediate: true },
)

const save = (): void => {
  emit('save', {
    status: nextStatus.value,
    reply_content: replyContent.value.trim(),
  })
}

const quickUpdate = (status: PageFeedbackStatus): void => {
  nextStatus.value = status
  save()
}
</script>
