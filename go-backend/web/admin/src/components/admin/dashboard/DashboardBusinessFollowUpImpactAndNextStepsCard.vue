<template>
  <article :id="`business-follow-up-${item.id}`" class="min-w-0 space-y-4 rounded-2xl border border-border/80 bg-background p-4" :class="highlighted ? 'ring-2 ring-primary' : ''">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div class="min-w-0 space-y-1">
        <p class="text-xs text-muted-foreground">{{ item.category }} · 事项 #{{ item.id }}</p>
        <h3 class="break-words text-sm font-bold">{{ item.title }}</h3>
        <p v-if="item.related_record" class="break-words text-sm text-muted-foreground">{{ item.related_record.label }}</p>
      </div>
      <span class="rounded-full px-3 py-1 text-xs font-bold" :class="isManualAttention ? 'bg-rose-500/10 text-rose-700 dark:text-rose-400' : 'bg-amber-500/10 text-amber-700 dark:text-amber-400'">
        {{ isManualAttention ? '需要负责人处理' : '已安排自动重试' }}
      </span>
    </div>
    <div class="space-y-1 text-sm leading-relaxed">
      <p class="font-bold">对业务有什么影响</p>
      <p>{{ item.business_impact }}</p>
    </div>
    <div class="rounded-xl bg-muted/50 p-3 text-xs leading-relaxed text-muted-foreground">
      <p><span class="font-bold text-foreground">当前情况：</span>{{ item.reason_summary }}</p>
      <p class="mt-1">已尝试 {{ item.attempts }} 次，最多自动尝试 {{ item.max_attempts }} 次。
        {{ isManualAttention ? '自动尝试已停止，请联系下方负责人核查实际结果。' : '系统已记录下一次尝试计划；是否已恢复需刷新确认。' }}</p>
      <p v-if="item.next_retry_at" class="mt-1">计划再次尝试：{{ formatBusinessFollowUpTimestamp(item.next_retry_at) }}{{ new Date(item.next_retry_at).getTime() <= Date.now() ? '（计划时间已到，请刷新核实执行进展）' : '' }}</p>
      <p class="mt-1">任务创建：{{ formatBusinessFollowUpTimestamp(item.created_at) }} · 最近尝试：{{ item.last_attempt_at ? formatBusinessFollowUpTimestamp(item.last_attempt_at) : '尚无尝试时间记录' }}</p>
      <p class="mt-1">任务创建时间不等于首次出错时间。</p>
    </div>
    <div class="space-y-2 text-sm leading-relaxed">
      <p class="font-bold">销售下一步怎么做</p>
      <ol class="list-decimal space-y-1 pl-5">
        <li v-for="step in item.sales_next_steps" :key="step">{{ step }}</li>
      </ol>
      <p><span class="font-bold">建议联系：</span>{{ item.responsible_team }}</p>
      <p v-if="item.related_record_notice" class="text-xs text-amber-700 dark:text-amber-400">{{ item.related_record_notice }}</p>
    </div>
    <div class="flex flex-wrap gap-2">
      <Button v-if="item.related_record?.path && authStore.hasPermission(item.related_record.required_permission)" variant="outline" size="sm" @click="router.push(item.related_record.path)">{{ relatedRecordActionLabel }}</Button>
      <Button variant="outline" size="sm" @click="copyBusinessFollowUpExplanation">复制跟进说明</Button>
      <Button v-if="showFollowUpDetailLink" variant="ghost" size="sm" @click="router.push({ path: '/dashboard/business-follow-ups', query: { event_id: String(item.id) } })">查看跟进详情</Button>
      <Button v-if="authStore.hasPermission('ops:view')" variant="ghost" size="sm" @click="router.push({ path: '/ops/outbox-failures', query: { event_id: String(item.id) } })">负责人查看技术详情</Button>
    </div>
    <p class="text-xs text-muted-foreground">复制后请通过现有工作渠道交给负责人；复制操作不会通知任何人或改变任务状态。</p>
  </article>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import { Button } from '@/components/ui/button'
import { useAuthStore } from '@/stores/auth'
import type { DashboardBusinessFollowUpItem } from '@/api/dashboardCommandCenterOutboxFailures'

const props = withDefaults(defineProps<{ item: DashboardBusinessFollowUpItem; highlighted?: boolean; showFollowUpDetailLink?: boolean }>(), { highlighted: false, showFollowUpDetailLink: false })
const authStore = useAuthStore()
const router = useRouter()
const isManualAttention = computed(() => props.item.processing_state === 'manual_attention')
const relatedRecordActionLabel = computed(() => {
  const path = props.item.related_record?.path || ''
  if (path.startsWith('/orders/')) return '打开相关订单'
  if (path.startsWith('/support/')) return '打开客服会话'
  if (path.startsWith('/google-merchant')) return '打开商品渠道管理'
  return '打开商品管理'
})

function formatBusinessFollowUpTimestamp(timestamp: string): string {
  return new Date(timestamp).toLocaleString('zh-CN', { hour12: false })
}

async function copyBusinessFollowUpExplanation(): Promise<void> {
  const item = props.item
  const explanation = [
    `业务跟进事项 #${item.id}：${item.title}`,
    `相关业务：${item.related_record?.label || item.related_record_notice}`,
    `业务影响：${item.business_impact}`,
    `当前情况：${item.reason_summary}`,
    `处理状态：${isManualAttention.value ? '自动尝试已停止，需要负责人处理' : '已安排自动重试，请核实实际执行结果'}`,
    `已尝试：${item.attempts} / ${item.max_attempts} 次`,
    ...(item.next_retry_at ? [`计划再次尝试：${formatBusinessFollowUpTimestamp(item.next_retry_at)}`] : []),
    `任务创建：${formatBusinessFollowUpTimestamp(item.created_at)}`,
    `最近尝试：${item.last_attempt_at ? formatBusinessFollowUpTimestamp(item.last_attempt_at) : '无时间记录'}`,
    `建议联系：${item.responsible_team}`,
    '跟进步骤：', ...item.sales_next_steps.map((step, index) => `${index + 1}. ${step}`),
    '请核查业务记录与外部实际结果后处理，任务离开提醒清单不代表业务已成功。',
  ].join('\n')
  try {
    await navigator.clipboard.writeText(explanation)
    toast.success('跟进说明已复制，请交给对应负责人')
  } catch {
    toast.error('未能复制，请手动选取页面中的跟进说明')
  }
}
</script>
