<template>
  <Card class="gap-0 rounded-[24px] border-dashed border-border/80 py-0 shadow-none" :aria-busy="state.loading">
    <CardHeader class="flex flex-col gap-3 border-b border-dashed border-border/70 py-4 sm:flex-row sm:items-center sm:justify-between">
      <div class="space-y-1">
        <CardTitle class="flex items-center gap-2 text-base font-bold"><ShieldAlert class="size-4 text-amber-600" />业务跟进提醒</CardTitle>
        <CardDescription class="max-w-3xl text-sm leading-relaxed">业务发生后，系统还需发送客户通知、同步订单或登记物流。这里提醒执行出错的后续任务，帮助你核对业务进展并联系负责人。</CardDescription>
      </div>
      <div class="flex shrink-0 flex-wrap items-center gap-2">
        <Button variant="outline" size="sm" :disabled="state.loading" @click="emit('refresh')"><RefreshCw :class="['size-3.5', state.loading ? 'animate-spin' : '']" />刷新提醒</Button>
        <Button v-if="!listMode" variant="ghost" size="sm" @click="emit('open-follow-up-list')">打开跟进清单<ArrowRight class="size-3.5" /></Button>
      </div>
    </CardHeader>
    <CardContent class="space-y-4 p-4">
      <details class="rounded-xl bg-muted/40 p-3 text-sm leading-relaxed" :open="listMode">
        <summary class="cursor-pointer font-bold">第一次使用：看到提醒后，应该做什么？</summary>
        <ol class="mt-3 list-decimal space-y-2 pl-5">
          <li>先看“对业务有什么影响”，再打开相关订单或业务记录，确认客户目前的实际情况。例如付款确认邮件出错，先查付款记录，不能要求客户再付一次。</li>
          <li>“已安排自动重试”表示系统记录了再次尝试计划；客户正在等待时仍需按卡片指引跟进。“需要负责人处理”表示自动尝试已停止，应联系对应负责人。</li>
          <li>按“销售下一步怎么做”核对并沟通；可复制跟进说明交给负责人。负责人核实实际结果后，在技术页面决定如何处理。</li>
          <li>处理后点“刷新提醒”，并回到业务记录确认最终结果。提醒减少只表示任务离开出错清单，重新排队或忽略均不等于业务成功。</li>
        </ol>
        <p class="mt-3 text-xs text-muted-foreground">清单只覆盖已记录的后台任务执行错误，不覆盖所有退款结果、物流异常或客服未回复。执行结果尚不确定的任务需要负责人另行核查，本清单暂不展示；没有提醒也不能证明所有任务已完成或整个系统健康。</p>
      </details>
      <p v-if="state.loading" role="status" class="py-3 text-sm text-muted-foreground">正在查询业务跟进提醒…</p>
      <div v-else-if="state.loadError || !state.outboxFailures || state.outboxFailures.status === 'unavailable'" role="alert" class="rounded-xl border border-amber-500/40 bg-amber-500/10 p-4 text-sm text-amber-700 dark:text-amber-400">
        <p class="font-bold">业务跟进数据暂时不可用</p><p class="mt-1">当前无法确认有多少事项需要跟进，请刷新重试；客户正在等待时，请直接核查订单或联系对应业务负责人。</p>
      </div>
      <template v-else>
        <div class="flex flex-wrap items-center gap-3 text-sm">
          <span class="rounded-full bg-muted px-3 py-1 font-bold">{{ state.outboxFailures.total_count }} 件业务提醒</span>
          <span class="text-rose-700 dark:text-rose-400">需要负责人处理 {{ state.outboxFailures.manual_attention_count }} 件</span>
          <span class="text-amber-700 dark:text-amber-400">已安排自动重试 {{ state.outboxFailures.automatic_retry_count }} 件</span>
        </div>
        <p class="text-xs text-muted-foreground">只统计你有权查看的事项，数量按任务计算，同一订单可能有多项提醒。{{ state.lastSuccessfulTimestamp ? `查询时间：${formatBusinessFollowUpSnapshotTime(state.lastSuccessfulTimestamp)}` : '' }}</p>
        <p v-if="state.selectedEventNotice" role="status" class="rounded-xl bg-amber-500/10 p-3 text-sm text-amber-700 dark:text-amber-400">{{ state.selectedEventNotice }}</p>
        <p v-if="state.outboxFailures.total_count === 0 && !state.selectedEventOutsidePreview" class="py-3 text-sm text-muted-foreground">当前没有你有权查看的出错任务。客户的付款、退款、发货和通知结果仍应以相关业务记录为准。</p>
        <template v-else>
          <div v-if="listMode" class="flex flex-wrap gap-2" aria-label="筛选当前加载的提醒">
            <Button v-for="filter in processingStateFilters" :key="filter.value" size="sm" :variant="state.selectedProcessingState === filter.value ? 'default' : 'outline'" :aria-pressed="state.selectedProcessingState === filter.value" @click="emit('update-processing-state', filter.value)">{{ filter.label }}</Button>
          </div>
          <div class="space-y-4">
            <DashboardBusinessFollowUpImpactAndNextStepsCard v-for="item in state.visibleBusinessFollowUps" :key="item.id" :item="item" :highlighted="listMode && item.id === state.selectedEventId" :show-follow-up-detail-link="!listMode" />
          </div>
          <p v-if="state.visibleBusinessFollowUps.length === 0" class="text-sm text-muted-foreground">当前加载的事项中没有符合此筛选条件的提醒。</p>
          <p class="text-xs leading-relaxed text-muted-foreground">优先展示需要负责人处理的事项，其次按最近更新时间从早到晚展示。{{ listMode ? `本次加载前 ${state.outboxFailures.items.length} 件，最多 100 件；筛选只作用于已加载事项。` : '首页最多展示 5 件，更多事项请打开跟进清单。' }}{{ state.selectedEventOutsidePreview ? ' 已另外定位从首页打开的目标事项。' : '' }}</p>
        </template>
      </template>
    </CardContent>
  </Card>
</template>

<script setup lang="ts">
import { ArrowRight, RefreshCw, ShieldAlert } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import DashboardBusinessFollowUpImpactAndNextStepsCard from './DashboardBusinessFollowUpImpactAndNextStepsCard.vue'
import type { DashboardBusinessFollowUpPanelState } from '@/modules/dashboard/dashboardBusinessFollowUpPresentationTypes'

withDefaults(defineProps<{ listMode?: boolean; state: DashboardBusinessFollowUpPanelState }>(), { listMode: false })
const emit = defineEmits<{
  refresh: []
  'open-follow-up-list': []
  'update-processing-state': [value: string]
}>()
const processingStateFilters = [
  { value: 'all', label: '全部已加载事项' },
  { value: 'manual_attention', label: '需要负责人处理' },
  { value: 'automatic_retry_scheduled', label: '已安排自动重试' },
]
function formatBusinessFollowUpSnapshotTime(timestamp: string): string {
  return new Date(timestamp).toLocaleString('zh-CN', { hour12: false })
}
</script>
