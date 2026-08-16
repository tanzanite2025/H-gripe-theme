<template>
  <div class="flex min-h-0 flex-col gap-4 overflow-auto pb-2">
    <AdminPageHeader
      class="shrink-0"
      title="客服分析"
      description="按统计日查看客户地区、回复效率与会员成交转化"
    >
      <template #actions>
        <div class="flex flex-wrap items-end gap-2">
          <label class="space-y-1">
            <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">DATE / 统计日期</span>
            <Input
              v-model="selectedDate"
              type="date"
              class="h-9 w-40"
              :disabled="loading"
              @change="loadAnalytics"
            />
          </label>
          <Button variant="outline" size="sm" class="h-9 rounded-full" :disabled="loading" @click="loadAnalytics">
            <RefreshCw :class="['size-3.5', { 'animate-spin': loading }]" />
            刷新
          </Button>
        </div>
      </template>
    </AdminPageHeader>

    <AdminStatsGrid class="shrink-0" :items="statItems" />

    <div class="grid gap-4 xl:grid-cols-[minmax(0,1.35fr)_minmax(340px,0.85fr)]">
      <CustomerServiceRegionAnalytics
        :analytics="analytics"
        :loading="loading"
      />

      <section class="rounded-3xl border bg-card p-4 shadow-sm">
        <div class="flex items-start justify-between gap-3">
          <div>
            <h2 class="flex items-center gap-2 text-sm font-black tracking-tight">
              <Clock3 class="size-4 text-primary" />
              回复效率
            </h2>
            <p class="mt-1 text-xs text-muted-foreground">
              客户发言到下一次客服回复的平均间隔
            </p>
          </div>
          <span class="rounded-full bg-primary/10 px-2 py-1 text-[11px] font-black text-primary">
            {{ analytics?.reply_interval_count || 0 }} 个已匹配轮次
          </span>
        </div>

        <div class="mt-6">
          <p class="text-4xl font-black tracking-tight text-foreground">
            {{ averageReplyLabel }}
          </p>
          <p class="mt-1 text-xs font-bold text-muted-foreground">平均回复间隔</p>
        </div>

        <dl class="mt-6 divide-y divide-border/70 border-y border-border/70">
          <div class="flex items-center justify-between gap-3 py-3 text-xs">
            <dt class="text-muted-foreground">已匹配回复轮次</dt>
            <dd class="font-black tabular-nums text-foreground">{{ analytics?.reply_interval_count || 0 }}</dd>
          </div>
          <div class="flex items-center justify-between gap-3 py-3 text-xs">
            <dt class="text-muted-foreground">仍未回复客户轮次</dt>
            <dd class="font-black tabular-nums text-foreground">{{ analytics?.unanswered_customer_turns || 0 }}</dd>
          </div>
          <div class="flex items-center justify-between gap-3 py-3 text-xs">
            <dt class="text-muted-foreground">统计日期</dt>
            <dd class="font-black tabular-nums text-foreground">{{ analytics?.date || selectedDate }}</dd>
          </div>
        </dl>
      </section>
    </div>

    <section class="border-y py-4">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 class="flex items-center gap-2 text-sm font-black tracking-tight">
            <ShoppingCart class="size-4 text-primary" />
            聊天与下单转化
          </h2>
          <p class="mt-1 text-xs text-muted-foreground">
            统计日内，完成支付的会员用户 / 当日聊过的会员用户
          </p>
        </div>
        <span class="rounded-full bg-emerald-500/10 px-2 py-1 text-[11px] font-black text-emerald-700">
          {{ conversionLabel }}
        </span>
      </div>

      <div class="mt-5 grid gap-4 sm:grid-cols-3">
        <div class="rounded-2xl border bg-muted/25 p-4">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">聊天会员</p>
          <p class="mt-2 text-2xl font-black tabular-nums text-foreground">{{ analytics?.member_customer_count || 0 }}</p>
          <p class="mt-1 text-xs text-muted-foreground">按用户去重</p>
        </div>
        <div class="rounded-2xl border bg-muted/25 p-4">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">完成支付</p>
          <p class="mt-2 text-2xl font-black tabular-nums text-foreground">{{ analytics?.converted_member_customer_count || 0 }}</p>
          <p class="mt-1 text-xs text-muted-foreground">统计日内产生已支付订单</p>
        </div>
        <div class="rounded-2xl border bg-primary/5 p-4">
          <p class="text-[10px] font-black uppercase tracking-widest text-primary/70">会员成交率</p>
          <p class="mt-2 text-2xl font-black tabular-nums text-primary">{{ conversionLabel }}</p>
          <p class="mt-1 text-xs text-muted-foreground">仅统计可归因的登录会员</p>
        </div>
      </div>

      <p class="mt-4 flex items-start gap-2 text-[11px] leading-5 text-muted-foreground">
        <Info class="mt-0.5 size-3.5 shrink-0" />
        匿名访客目前没有稳定的订单归因键，因此不会被强行计入成交率分母。
      </p>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import {
  Clock3,
  Info,
  MapPin,
  MessagesSquare,
  RefreshCw,
  ShoppingCart,
} from '@lucide/vue'
import { toast } from 'vue-sonner'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import CustomerServiceRegionAnalytics from '@/components/admin/customer-service/CustomerServiceRegionAnalytics.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import customerServiceApi from '@/api/customerService'
import type { CustomerServiceAnalytics } from '@/components/admin/customer-service/customerServiceTypes'

const todayLocalDate = () => {
  const now = new Date()
  const year = now.getFullYear()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  const day = String(now.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const adminTimezoneOffsetMinutes = () => -new Date().getTimezoneOffset()

const selectedDate = ref(todayLocalDate())
const analytics = ref<CustomerServiceAnalytics | null>(null)
const loading = ref(false)

const averageReplyLabel = computed(() => formatDuration(Number(analytics.value?.average_reply_interval_seconds || 0)))

const conversionLabel = computed(() => {
  const members = Number(analytics.value?.member_customer_count || 0)
  if (!members) return '暂无'
  return `${Number(analytics.value?.member_conversion_rate || 0).toFixed(1)}%`
})

const statItems = computed(() => [
  {
    key: 'conversations',
    label: '统计日会话',
    value: Number(analytics.value?.total_conversations || 0),
    icon: MessagesSquare,
    tone: 'blue',
  },
  {
    key: 'regions',
    label: '已知地区会话',
    value: Number(analytics.value?.known_region_count || 0),
    icon: MapPin,
    tone: 'green',
  },
  {
    key: 'reply-interval',
    label: '平均回复间隔',
    value: averageReplyLabel.value,
    icon: Clock3,
    tone: 'amber',
  },
  {
    key: 'conversion',
    label: '会员成交率',
    value: conversionLabel.value,
    icon: ShoppingCart,
    tone: 'coral',
  },
])

const formatDuration = (seconds: number) => {
  if (!Number.isFinite(seconds) || seconds <= 0) return '暂无'
  if (seconds < 60) return `${Math.round(seconds)} 秒`

  const minutes = Math.floor(seconds / 60)
  const remainingSeconds = Math.round(seconds % 60)
  if (minutes < 60) return `${minutes} 分 ${remainingSeconds} 秒`

  const hours = Math.floor(minutes / 60)
  const remainingMinutes = minutes % 60
  return `${hours} 小时 ${remainingMinutes} 分`
}

const loadAnalytics = async () => {
  loading.value = true
  try {
    analytics.value = await customerServiceApi.getAnalytics({
      date: selectedDate.value,
      tz_offset_minutes: adminTimezoneOffsetMinutes(),
    })
  } catch (error) {
    console.error('Failed to fetch customer-service analytics:', error)
    analytics.value = null
    toast.error('客服分析加载失败')
  } finally {
    loading.value = false
  }
}

onMounted(loadAnalytics)
</script>
