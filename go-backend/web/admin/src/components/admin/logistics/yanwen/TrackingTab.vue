<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { AlertTriangle, CheckCircle2, ExternalLink, RadioTower, RefreshCw, Search } from '@lucide/vue'
import { yanwenLogisticsAdminApi, type YanwenTrackingAlert, type YanwenTrackingCheckpoint, type YanwenTrackingResult } from '@/api/yanwenLogisticsAdminApi'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

const trackingQuery = ref('')
const trackingQueried = ref(false)
const trackingRefreshing = ref(false)
const trackingError = ref('')
const trackingResults = ref<YanwenTrackingResult[]>([])
const trackingAlerts = ref<YanwenTrackingAlert[]>([])
const trackingAlertsLoading = ref(false)
const trackingAlertsError = ref('')

const trackingNumbers = computed(() => Array.from(new Set(
  trackingQuery.value
    .split(/[\s,，]+/)
    .map((value) => value.trim())
    .filter(Boolean),
)))
const trackingCanQuery = computed(() => (
  trackingNumbers.value.length > 0
  && trackingNumbers.value.length <= 30
))

const trackingStatusMap = [
  { codes: ['OR10', 'PU10'], label: '集货交仓', detail: '燕文揽收 / 操作中心分拣入库', tone: 'text-blue-600 bg-blue-500/10' },
  { codes: ['LH20'], label: '干线干飞', detail: '航班起飞离港', tone: 'text-indigo-600 bg-indigo-500/10' },
  { codes: ['S303', 'IC50', 'IC60'], label: '目的国清关', detail: '进口清关中 / 开始清关 / 清关完成', tone: 'text-amber-700 bg-amber-500/10' },
  { codes: ['LM10', 'LM20', 'LM25'], label: '尾程派送', detail: '到达目的国 / 到达末端网点 / 出库派送', tone: 'text-purple-600 bg-purple-500/10' },
  { codes: ['LM40'], label: '末端妥投', detail: '妥投成功、签收成功或完成提货', tone: 'text-emerald-700 bg-emerald-500/10' },
  { codes: ['PU30', 'EC30', 'IC51', 'IC70', 'LM50', 'LM90'], label: '异常阻断', detail: '揽收失败、报关/清关失败、派送失败或退回', tone: 'text-rose-700 bg-rose-500/10' },
]

const trackingStages = [
  { code: 'COLLECTED', label: '集货交仓', rank: 1 },
  { code: 'IN_TRANSIT_AIR', label: '干线干飞', rank: 2 },
  { code: 'CUSTOMS', label: '目的国清关', rank: 3 },
  { code: 'LAST_MILE', label: '尾程派送', rank: 4 },
  { code: 'DELIVERED', label: '末端妥投', rank: 5 },
] as const

const trackingStageLabel = (stage: YanwenTrackingResult['tracking_stage']) => (
  trackingStages.find((candidate) => candidate.code === stage)?.label || '未知阶段'
)

const statusPresentation = (code: string) => trackingStatusMap.find((status) => status.codes.includes(code)) || {
  codes: [code],
  label: '燕文原始状态',
  detail: '该状态码尚未配置本地阶段名称，保留官方原值。',
  tone: 'text-slate-600 bg-slate-500/10',
}

const checkpointFlightNumber = (checkpoint: YanwenTrackingCheckpoint) => {
  const value = checkpoint.extra_properties?.FlightNumber ?? checkpoint.extra_properties?.flightNumber
  return typeof value === 'string' || typeof value === 'number' ? String(value) : ''
}

const trackingAlertTypeLabel = (alert: YanwenTrackingAlert) => (
  alert.alert_type === 'CUSTOMS_STAGNATION' ? '清关停滞' : '在途超时'
)

const trackingAlertSeverityTone = (severity: YanwenTrackingAlert['severity']) => (
  severity === 'CRITICAL'
    ? 'border-rose-500/30 bg-rose-500/10 text-rose-700 dark:text-rose-300'
    : 'border-amber-500/30 bg-amber-500/10 text-amber-700 dark:text-amber-300'
)

const refreshTrackingAlerts = async () => {
  trackingAlertsLoading.value = true
  trackingAlertsError.value = ''
  try {
    trackingAlerts.value = await yanwenLogisticsAdminApi.listYanwenTrackingAlerts()
  } catch (error: any) {
    trackingAlertsError.value = error?.message || '读取燕文停滞告警失败。'
    trackingAlerts.value = []
  } finally {
    trackingAlertsLoading.value = false
  }
}

const refreshTracking = async () => {
  if (!trackingCanQuery.value) {
    trackingError.value = trackingNumbers.value.length > 30
      ? '燕文轨迹接口一次最多查询 30 个单号，请减少后重试。'
      : '请输入燕文单号或尾程单号，可用逗号分隔多个单号。'
    trackingQueried.value = false
    trackingResults.value = []
    return
  }

  trackingRefreshing.value = true
  trackingError.value = ''
  trackingResults.value = []
  try {
    trackingResults.value = await yanwenLogisticsAdminApi.queryYanwenTracking(trackingNumbers.value)
    trackingQueried.value = true
    void refreshTrackingAlerts()
  } catch (error: any) {
    trackingError.value = error?.message || '读取燕文官方轨迹失败。'
    trackingQueried.value = false
  } finally {
    trackingRefreshing.value = false
  }
}

const clearTracking = () => {
  trackingQuery.value = ''
  trackingQueried.value = false
  trackingError.value = ''
  trackingResults.value = []
}

onMounted(() => {
  void refreshTrackingAlerts()
})
</script>

<template>
  <section class="space-y-4 rounded-[28px] border border-dashed border-border/80 bg-card p-5 shadow-sm sm:p-6">
    <div>
      <p class="text-[10px] font-black uppercase tracking-[0.18em] text-emerald-600">Milestones &amp; In-Transit Tracking</p>
      <h2 class="mt-1 text-lg font-black tracking-tight">全链路轨迹与时效看板</h2>
      <p class="mt-1 max-w-3xl text-xs leading-5 text-muted-foreground">这里直接读取燕文正式轨迹接口，展示官方返回的单号、节点、状态码和时区。后台轮询结果只保存在燕文专属快照中，不进入通用物流追踪链。</p>
    </div>

    <section class="space-y-4 rounded-[24px] border border-dashed border-border/80 bg-muted/20 p-4">
      <div class="flex flex-col gap-3 sm:flex-row sm:items-end">
        <label class="min-w-0 flex-1 space-y-1.5">
          <span class="flex items-center gap-1.5 text-[10px] font-black uppercase tracking-wider text-muted-foreground/80"><Search class="size-3" />查询燕文运单</span>
          <Input v-model="trackingQuery" placeholder="输入单号，可用逗号分隔，最多 30 个" @keyup.enter="refreshTracking" />
        </label>
        <div class="flex gap-2">
          <Button :disabled="trackingRefreshing" @click="refreshTracking"><RefreshCw :class="['size-3.5', { 'animate-spin': trackingRefreshing }]" />查询轨迹</Button>
          <Button variant="ghost" @click="clearTracking">清空</Button>
        </div>
      </div>
      <p class="rounded-2xl border border-amber-500/30 bg-amber-500/10 px-4 py-3 text-xs leading-5 text-amber-700 dark:text-amber-300">正式接口为 `GET http://api.track.yw56.com.cn/api/tracking`，后端通过 `Authorization` 传已保存的商户号/制单账号。燕文没有轨迹测试环境，浏览器不会直连官方接口。</p>
      <p v-if="trackingError" class="rounded-2xl border border-rose-500/30 bg-rose-500/10 px-4 py-3 text-xs leading-5 text-rose-700 dark:text-rose-300">{{ trackingError }}</p>

      <div v-if="!trackingQueried && !trackingError" class="flex min-h-40 flex-col items-center justify-center gap-3 rounded-2xl border border-dashed border-border/80 bg-card px-6 py-10 text-center">
        <span class="flex size-11 items-center justify-center rounded-2xl bg-muted text-muted-foreground"><RadioTower class="size-5" /></span>
        <div><p class="text-sm font-black">输入真实单号后查询</p><p class="mt-1 max-w-lg text-xs leading-5 text-muted-foreground">查询结果来自燕文正式轨迹接口；接口没有返回的节点不会在页面中补造。</p></div>
      </div>
      <div v-else-if="trackingQueried && trackingResults.length === 0" class="flex min-h-40 flex-col items-center justify-center gap-3 rounded-2xl border border-dashed border-border/80 bg-card px-6 py-10 text-center">
        <span class="flex size-11 items-center justify-center rounded-2xl bg-muted text-muted-foreground"><RadioTower class="size-5" /></span>
        <div><p class="text-sm font-black">官方接口未返回轨迹</p><p class="mt-1 max-w-lg text-xs leading-5 text-muted-foreground">燕文返回了成功响应，但当前单号没有可展示的结果。</p></div>
      </div>
    </section>

    <section v-if="trackingResults.length > 0" class="space-y-4">
      <article v-for="result in trackingResults" :key="result.tracking_number" class="space-y-4 rounded-[24px] border border-dashed border-border/80 bg-card p-4">
        <div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
          <div class="min-w-0">
            <div class="flex flex-wrap items-center gap-2">
              <h3 class="font-mono text-sm font-black">{{ result.tracking_number }}</h3>
              <span :class="statusPresentation(result.tracking_status).tone" class="rounded-full px-2.5 py-1 font-mono text-[10px] font-bold">{{ result.tracking_status }}</span>
              <span class="rounded-full bg-muted px-2.5 py-1 text-[10px] font-bold">{{ statusPresentation(result.tracking_status).label }}</span>
            </div>
            <p class="mt-1 text-xs text-muted-foreground">燕文单号：{{ result.waybill_number }}<span v-if="result.exchange_number"> · 尾程单号：{{ result.exchange_number }}</span></p>
          </div>
          <a v-if="result.last_mile_carrier_website" :href="result.last_mile_carrier_website" target="_blank" rel="noreferrer" class="inline-flex items-center gap-1 text-xs font-bold text-emerald-700 hover:underline dark:text-emerald-300"><ExternalLink class="size-3" />{{ result.last_mile_carrier || '尾程承运商' }}</a>
        </div>

        <div class="space-y-2 rounded-2xl border border-dashed border-border/80 bg-muted/20 p-3">
          <div class="flex flex-wrap items-center justify-between gap-2"><p class="text-[10px] font-black uppercase tracking-wider text-muted-foreground/80">燕文五阶段推进（本地归一化）</p><span class="rounded-full bg-muted px-2.5 py-1 text-[10px] font-bold">{{ trackingStageLabel(result.tracking_stage) }}</span></div>
          <div class="flex gap-1" role="progressbar" :aria-valuemin="0" :aria-valuemax="5" :aria-valuenow="result.tracking_stage_rank" :aria-label="`燕文轨迹阶段：${trackingStageLabel(result.tracking_stage)}`">
            <span v-for="stage in trackingStages" :key="stage.code" :class="stage.rank <= result.tracking_stage_rank ? 'bg-emerald-500' : 'bg-border'" class="h-2 flex-1 rounded-full" />
          </div>
          <div class="flex justify-between gap-2 text-[10px] text-muted-foreground"><span v-for="stage in trackingStages" :key="stage.code" class="min-w-0 flex-1 truncate">{{ stage.label }}</span></div>
          <p v-if="result.tracking_has_exception" class="rounded-xl border border-rose-500/30 bg-rose-500/10 px-3 py-2 text-[11px] leading-5 text-rose-700 dark:text-rose-300">当前官方状态属于异常阻断；阶段条只使用官方节点中已经确认的主阶段，没有可确认阶段时保持“未知”，原始状态码和官方消息不被覆盖。</p>
        </div>

        <div class="grid gap-2 text-xs sm:grid-cols-2 xl:grid-cols-4">
          <div class="rounded-2xl bg-muted/50 px-3 py-2"><span class="text-muted-foreground">起运 / 目的</span><p class="mt-1 font-mono font-bold">{{ result.origin_country || '—' }} / {{ result.destination_country || '—' }}</p></div>
          <div class="rounded-2xl bg-muted/50 px-3 py-2"><span class="text-muted-foreground">官方状态层级</span><p class="mt-1 font-mono font-bold">{{ result.tracking_status_waybill.level1 }} · {{ result.tracking_status_waybill.level2 }} · {{ result.tracking_status_waybill.level3 }}</p></div>
          <div class="rounded-2xl bg-muted/50 px-3 py-2"><span class="text-muted-foreground">尾程追踪</span><p class="mt-1 font-bold">{{ result.last_mile_tracking_expected ? '预计有尾程节点' : '官方未标记尾程节点' }}</p></div>
          <div class="rounded-2xl bg-muted/50 px-3 py-2"><span class="text-muted-foreground">尾程联系方式</span><p class="mt-1 font-bold">{{ result.last_mile_carrier_contact_number || '—' }}</p></div>
        </div>

        <div class="rounded-2xl border border-dashed border-border/80 p-3">
          <div class="mb-3 flex items-center gap-2"><CheckCircle2 class="size-4 text-emerald-600" /><h4 class="text-sm font-black">官方轨迹节点</h4><span class="text-[11px] text-muted-foreground">{{ result.checkpoints.length }} 个节点</span></div>
          <ol class="space-y-3">
            <li v-for="checkpoint in result.checkpoints" :key="`${checkpoint.time_stamp}-${checkpoint.tracking_status}`" class="flex gap-3">
              <span class="mt-0.5 flex size-7 shrink-0 items-center justify-center rounded-full bg-emerald-500/10 text-emerald-600"><CheckCircle2 class="size-3.5" /></span>
              <div class="min-w-0 flex-1">
                <div class="flex flex-wrap items-center gap-2"><span :class="statusPresentation(checkpoint.tracking_status).tone" class="rounded-full px-2 py-0.5 font-mono text-[10px] font-bold">{{ checkpoint.tracking_status }}</span><span class="text-xs font-bold">{{ statusPresentation(checkpoint.tracking_status).label }}</span><span v-if="checkpoint.is_last_mile_checkpoint" class="rounded-full bg-purple-500/10 px-2 py-0.5 text-[10px] font-bold text-purple-700 dark:text-purple-300">尾程节点</span></div>
                <p class="mt-1 text-xs leading-5 text-foreground">{{ checkpoint.message }}</p>
                <p v-if="checkpoint.location" class="mt-0.5 text-[11px] text-muted-foreground">地点：{{ checkpoint.location }}</p>
                <p v-if="checkpointFlightNumber(checkpoint)" class="mt-0.5 text-[11px] text-muted-foreground">航班：{{ checkpointFlightNumber(checkpoint) }}</p>
                <p class="mt-1 font-mono text-[10px] text-muted-foreground">{{ checkpoint.time_stamp }} {{ checkpoint.time_zone }}</p>
              </div>
            </li>
          </ol>
        </div>
      </article>
    </section>

    <div class="grid gap-4 xl:grid-cols-[minmax(0,1.15fr)_minmax(320px,0.85fr)]">
      <section class="overflow-hidden rounded-[24px] border border-dashed border-border/80 bg-card">
        <div class="border-b border-dashed border-border/80 px-5 py-4"><h3 class="text-sm font-black">燕文状态码标准化</h3><p class="mt-1 text-[11px] text-muted-foreground">页面只为官方节点提供显示分组，原始状态码和官方消息始终保留。</p></div>
        <div class="divide-y divide-dashed divide-border/70"><div v-for="status in trackingStatusMap" :key="status.codes.join('-')" class="flex items-center gap-3 px-5 py-3"><span :class="status.tone" class="rounded-full px-2.5 py-1 font-mono text-[10px] font-bold">{{ status.codes.join(' / ') }}</span><div class="min-w-0 flex-1"><p class="text-xs font-bold">{{ status.label }}</p><p class="mt-0.5 text-[11px] text-muted-foreground">{{ status.detail }}</p></div></div></div>
      </section>
      <section class="space-y-3 rounded-[24px] border border-dashed border-border/80 bg-card p-4">
        <div class="flex items-start justify-between gap-3">
          <div class="flex items-center gap-3"><span class="flex size-10 items-center justify-center rounded-2xl bg-amber-500/10 text-amber-700"><AlertTriangle class="size-4" /></span><div><h3 class="text-sm font-black">停滞告警</h3><p class="mt-1 text-[11px] text-muted-foreground">只读取燕文生产轨迹快照，按本地运营规则识别停滞，不生成客服工单。</p></div></div>
          <Button variant="ghost" size="sm" :disabled="trackingAlertsLoading" @click="refreshTrackingAlerts"><RefreshCw :class="['size-3.5', { 'animate-spin': trackingAlertsLoading }]" />刷新</Button>
        </div>
        <p v-if="trackingAlertsError" class="rounded-xl border border-rose-500/30 bg-rose-500/10 px-3 py-2 text-[11px] leading-5 text-rose-700 dark:text-rose-300">{{ trackingAlertsError }}</p>
        <div v-else-if="trackingAlertsLoading" class="flex min-h-24 items-center justify-center rounded-2xl border border-dashed border-border/80 bg-muted/20 text-xs text-muted-foreground">正在读取燕文生产快照…</div>
        <div v-else-if="trackingAlerts.length === 0" class="flex min-h-24 items-center justify-center rounded-2xl border border-dashed border-border/80 bg-muted/20 px-4 text-center text-xs leading-5 text-muted-foreground">当前没有达到清关停滞或在途超时阈值的燕文快照。</div>
        <div v-else class="space-y-2">
          <article v-for="alert in trackingAlerts" :key="`${alert.alert_type}-${alert.tracking_number}`" :class="trackingAlertSeverityTone(alert.severity)" class="rounded-2xl border px-3 py-3">
            <div class="flex flex-wrap items-center gap-2"><span class="rounded-full border px-2 py-0.5 font-mono text-[10px] font-black">{{ alert.severity }}</span><span class="text-xs font-black">{{ trackingAlertTypeLabel(alert) }}</span><span class="font-mono text-[11px]">{{ alert.tracking_number }}</span><span class="font-mono text-[10px]">{{ alert.tracking_status }}</span></div>
            <p class="mt-2 text-xs leading-5">{{ alert.message }}</p>
            <p class="mt-1 text-[10px] leading-4 opacity-80">节点时间：{{ alert.checkpoint_time_stamp }} {{ alert.checkpoint_time_zone }} · 快照同步：{{ alert.snapshot_last_synced_at }}</p>
          </article>
        </div>
      </section>
    </div>
  </section>
</template>
