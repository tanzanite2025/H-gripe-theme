<script setup lang="ts">
import { Activity, AlertTriangle, CheckCircle2, Globe2, PackageCheck, RefreshCw, Send, Settings2, Truck } from '@lucide/vue'
import { computed, onMounted, ref } from 'vue'
import yanwenLogisticsAdminApi, { type YanwenOperationsOverview } from '@/api/yanwenLogisticsAdminApi'
import { Button } from '@/components/ui/button'

type YanwenEnvironment = 'fat' | 'production'

const environment = ref<YanwenEnvironment>('production')
const overview = ref<YanwenOperationsOverview | null>(null)
const loading = ref(false)
const loadError = ref('')

const metrics = computed(() => {
  const current = overview.value
  return [
    {
      label: '今日创建运单',
      value: current ? String(current.today_created_waybill_count) : '—',
      detail: current ? `${current.environment === 'production' ? 'PRD' : 'FAT'} 本地真实建单成功记录` : '正在读取燕文运单台账',
      icon: Send,
    },
    {
      label: '官方待打印面单',
      value: current ? String(current.official_pending_print_waybill_count) : '—',
      detail: '仅统计已同步官方 isPrint=0 的运单',
      icon: PackageCheck,
    },
    {
      label: '有效在途包裹',
      value: current ? String(current.active_tracking_snapshot_count) : '—',
      detail: '燕文官方轨迹快照中未终止的记录',
      icon: Truck,
    },
    {
      label: '清关异常快照',
      value: current ? String(current.customs_exception_tracking_snapshot_count) : '—',
      detail: 'S303 / IC50 / IC51 / IC70 官方状态',
      icon: AlertTriangle,
    },
    {
      label: '本月预估运费',
      value: '—',
      detail: '燕文台账未保存官方账单金额',
      icon: Activity,
    },
  ]
})

const gatewayStatusLabel = computed(() => {
  if (!overview.value) return '等待读取'
  if (!overview.value.gateway.credentials_configured) return '凭据未完整配置'
  if (!overview.value.gateway.enabled) return '已停用'
  return '凭据已配置'
})

const gatewayStatusClass = computed(() => {
  if (!overview.value?.gateway.credentials_configured || overview.value.gateway.enabled === false) {
    return 'border-amber-500/30 bg-amber-500/10 text-amber-700 dark:text-amber-300'
  }
  return 'border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300'
})

const loadOverview = async () => {
  loading.value = true
  loadError.value = ''
  try {
    overview.value = await yanwenLogisticsAdminApi.getYanwenOperationsOverview(environment.value)
  } catch (error) {
    overview.value = null
    loadError.value = error instanceof Error ? error.message : '无法读取燕文运营概览'
  } finally {
    loading.value = false
  }
}

const selectEnvironment = async () => {
  await loadOverview()
}

onMounted(loadOverview)
</script>

<template>
  <div class="space-y-5">
    <section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-5" aria-label="燕文运营指标">
      <article v-for="metric in metrics" :key="metric.label" class="group relative min-h-36 overflow-hidden rounded-[24px] border border-dashed border-border/80 bg-card p-4">
        <div class="pointer-events-none absolute inset-0 bg-gradient-to-br from-emerald-500/5 via-transparent to-transparent" />
        <div class="relative flex h-full flex-col justify-between gap-4">
          <div class="flex items-center justify-between gap-2"><p class="text-[10px] font-black uppercase tracking-wider text-muted-foreground/80">{{ metric.label }}</p><component :is="metric.icon" class="size-4 text-muted-foreground/70" aria-hidden="true" /></div>
          <div><p class="font-mono text-3xl font-black tracking-tight text-foreground">{{ metric.value }}</p><p class="mt-1 text-[11px] leading-4 text-muted-foreground">{{ metric.detail }}</p></div>
        </div>
      </article>
    </section>

    <section class="space-y-4 rounded-[28px] border border-dashed border-border/80 bg-muted/5 p-5 sm:p-6">
      <div class="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p class="text-[10px] font-black uppercase tracking-[0.18em] text-muted-foreground/70">Yanwen Logistics Operations</p>
          <h2 class="mt-1 text-lg font-black tracking-tight">燕文跨境专线运营汇总</h2>
          <p class="mt-1 text-xs text-muted-foreground">数据只来自燕文运单台账和燕文专属轨迹快照；大盘不读取通用物流或其他承运商数据。</p>
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <select v-model="environment" class="h-9 rounded-md border border-dashed border-border bg-background px-3 text-sm" @change="selectEnvironment">
            <option value="production">PRD 生产环境</option>
            <option value="fat">FAT 测试环境</option>
          </select>
          <Button variant="outline" size="sm" :disabled="loading" @click="loadOverview"><RefreshCw :class="['size-3.5', { 'animate-spin': loading }]" />刷新</Button>
          <span :class="gatewayStatusClass" class="inline-flex items-center gap-2 rounded-full border px-3 py-1.5 text-[11px] font-bold"><span class="size-1.5 rounded-full bg-current" />{{ gatewayStatusLabel }}</span>
        </div>
      </div>

      <div v-if="loading" class="flex items-start gap-2 rounded-2xl border border-border/70 bg-card/70 px-4 py-3 text-xs leading-5 text-muted-foreground"><Activity class="mt-0.5 size-4 shrink-0" aria-hidden="true" /><p>正在读取燕文运营概览。</p></div>
      <div v-else-if="loadError" class="flex items-start gap-2 rounded-2xl border border-rose-500/30 bg-rose-500/10 px-4 py-3 text-xs leading-5 text-rose-700 dark:text-rose-300"><AlertTriangle class="mt-0.5 size-4 shrink-0" aria-hidden="true" /><p>概览读取失败：{{ loadError }}</p></div>

      <div v-else class="grid gap-4 xl:grid-cols-2">
        <article class="rounded-[24px] border border-dashed border-border/80 bg-card p-4">
          <div class="flex items-center justify-between gap-3">
            <div><p class="text-[10px] font-black uppercase tracking-[0.16em] text-muted-foreground/70">Destination Mix</p><h3 class="mt-1 text-sm font-black">专线国家流向 Top 5</h3></div>
            <Globe2 class="size-4 text-muted-foreground" aria-hidden="true" />
          </div>
          <div v-if="overview && overview.destination_flows.length > 0" class="mt-4 divide-y divide-dashed divide-border/70">
            <div v-for="destination in overview.destination_flows" :key="destination.country_code" class="flex items-center gap-3 py-3 first:pt-0 last:pb-0">
              <span class="flex size-9 shrink-0 items-center justify-center rounded-xl bg-muted font-mono text-[10px] font-black">{{ destination.country_code }}</span>
              <div class="min-w-0 flex-1"><p class="truncate text-sm font-bold">{{ destination.channel_name || '燕文渠道' }}</p><p class="mt-0.5 text-[11px] text-muted-foreground">{{ destination.waybill_count }} 条本地真实运单</p></div>
              <span class="rounded-full bg-muted px-2.5 py-1 text-[10px] font-bold text-muted-foreground">{{ destination.waybill_count }} 单</span>
            </div>
          </div>
          <div v-else class="mt-4 rounded-2xl border border-dashed border-border/80 bg-muted/30 px-4 py-8 text-center text-xs leading-5 text-muted-foreground">当前环境没有已成功持久化的燕文运单流向。</div>
        </article>

        <article class="rounded-[24px] border border-dashed border-border/80 bg-card p-4">
          <div class="flex items-center justify-between gap-3">
            <div><p class="text-[10px] font-black uppercase tracking-[0.16em] text-muted-foreground/70">API Gateway State</p><h3 class="mt-1 text-sm font-black">燕文网关配置状态</h3></div>
            <Settings2 class="size-4 text-muted-foreground" aria-hidden="true" />
          </div>
          <div class="mt-4 space-y-3">
            <div class="flex items-center gap-3 rounded-2xl border border-dashed border-border/80 bg-muted/30 p-3"><span :class="overview?.gateway.credentials_configured && overview.gateway.enabled ? 'bg-emerald-500/10 text-emerald-600' : 'bg-amber-500/10 text-amber-600'" class="flex size-10 items-center justify-center rounded-2xl"><component :is="overview?.gateway.credentials_configured && overview.gateway.enabled ? CheckCircle2 : AlertTriangle" class="size-4" /></span><div><p class="text-sm font-black">{{ overview?.gateway.environment === 'production' ? 'PRD 生产环境' : 'FAT 测试环境' }}</p><p class="mt-1 break-all text-[11px] leading-5 text-muted-foreground">{{ overview?.gateway.endpoint || '未读取网关端点' }}</p></div></div>
            <div class="grid gap-2 sm:grid-cols-2"><div class="rounded-2xl bg-muted/50 p-3"><p class="text-[10px] font-bold text-muted-foreground">凭据状态</p><p class="mt-1 text-xs font-black">{{ overview?.gateway.credentials_configured ? 'user_id / apitoken 已配置' : '凭据未完整配置' }}</p></div><div class="rounded-2xl bg-muted/50 p-3"><p class="text-[10px] font-bold text-muted-foreground">服务开关</p><p class="mt-1 text-xs font-black">{{ overview?.gateway.enabled ? '已启用' : '未启用' }}</p></div></div>
            <p class="text-[11px] leading-5 text-muted-foreground">大盘不重复发起 Ping；需要验证签名和延迟时，请到“网关配置”TAB 执行 `common.country.getlist`。</p>
          </div>
        </article>
      </div>

      <div class="flex items-start gap-2 rounded-2xl border border-border/70 bg-card/70 px-4 py-3 text-xs leading-5 text-muted-foreground"><AlertTriangle class="mt-0.5 size-4 shrink-0" aria-hidden="true" /><p>“本月预估运费”暂不展示金额，因为燕文建单台账没有官方账单字段；其余指标只统计已经获得并持久化的燕文事实，不根据缺失状态推断结果。</p></div>
    </section>
  </div>
</template>
