<script setup lang="ts">
import { computed, ref } from 'vue'
import { Activity, AlertTriangle, Calculator, CheckCircle2, ClipboardCheck, Download, Globe2, PackageCheck, Plus, Power, Printer, RadioTower, RefreshCw, Search, Send, Settings2, Trash2, Truck } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'


const trackingQuery = ref('')
const trackingQueried = ref(false)
const trackingRefreshing = ref(false)
const trackingError = ref('')
const trackingEvents = ref<Array<{ code: string; label: string; detail: string; occurredAt: string }>>([])
const trackingNumbers = computed(() => trackingQuery.value.split(/[\s,，]+/).map((value) => value.trim()).filter(Boolean))
const trackingCanQuery = computed(() => trackingNumbers.value.length > 0 && trackingNumbers.value.length <= 30 && trackingNumbers.value.every((value) => value.length >= 4))
const refreshTracking = async () => {
  if (!trackingCanQuery.value) {
    trackingError.value = trackingNumbers.value.length > 30
      ? '燕文轨迹接口一次最多查询 30 个单号，请减少后重试。'
      : '请输入至少 4 位燕文单号或尾程单号，可用逗号分隔多个单号。'
    trackingQueried.value = false
    return
  }
  trackingRefreshing.value = true
  trackingError.value = ''
  trackingQueried.value = true
  trackingEvents.value = []
  await Promise.resolve()
  trackingRefreshing.value = false
}
const clearTracking = () => {
  trackingQuery.value = ''
  trackingQueried.value = false
  trackingError.value = ''
  trackingEvents.value = []
}
const trackingStatusMap = [
  { code: 'OR10 / PU10', label: '集货交仓', detail: '燕文揽收 / 操作中心分拣入库', tone: 'text-blue-600 bg-blue-500/10' },
  { code: 'LH20', label: '干线干飞', detail: '航班起飞离港，提取 FlightNumber', tone: 'text-indigo-600 bg-indigo-500/10' },
  { code: 'S303 / IC50 / IC60', label: '目的国清关', detail: '进口清关中 / 开始清关 / 清关完成', tone: 'text-amber-700 bg-amber-500/10' },
  { code: 'LM10 / LM20 / LM25', label: '尾程派送', detail: '到达目的国 / 到达末端网点 / 出库派送', tone: 'text-purple-600 bg-purple-500/10' },
  { code: 'LM40', label: '末端妥投', detail: '妥投成功、签收成功或完成提货', tone: 'text-emerald-700 bg-emerald-500/10' },
  { code: 'PU30 / EC30 / IC51 / IC70 / LM50 / LM90', label: '异常阻断', detail: '揽收失败、报关/清关失败、派送失败或退回', tone: 'text-rose-700 bg-rose-500/10' },
]
</script>

<template>
      <section class="space-y-4 rounded-[28px] border border-dashed border-border/80 bg-card p-5 shadow-sm sm:p-6">
        <div>
          <p class="text-[10px] font-black uppercase tracking-[0.18em] text-emerald-600">Milestones &amp; In-Transit Tracking</p>
          <h2 class="mt-1 text-lg font-black tracking-tight">全链路轨迹与时效看板</h2>
          <p class="mt-1 max-w-3xl text-xs leading-5 text-muted-foreground">燕文轨迹只在本域内轮询，统一转换为集货、干线、清关、尾程和妥投事件。接口异常保持可见，不用空数组掩盖错误。</p>
        </div>

        <section class="space-y-4 rounded-[24px] border border-dashed border-border/80 bg-muted/20 p-4">
          <div class="flex flex-col gap-3 sm:flex-row sm:items-end"><label class="min-w-0 flex-1 space-y-1.5"><span class="flex items-center gap-1.5 text-[10px] font-black uppercase tracking-wider text-muted-foreground/80"><Search class="size-3" />查询燕文运单</span><Input v-model="trackingQuery" placeholder="输入单号，可用逗号分隔，最多 30 个" @keyup.enter="refreshTracking" /></label><div class="flex gap-2"><Button :disabled="trackingRefreshing" @click="refreshTracking"><RefreshCw :class="['size-3.5', { 'animate-spin': trackingRefreshing }]" />查询轨迹</Button><Button variant="ghost" @click="clearTracking">清空</Button></div></div>
          <p class="rounded-2xl border border-amber-500/30 bg-amber-500/10 px-4 py-3 text-xs leading-5 text-amber-700 dark:text-amber-300">燕文轨迹使用独立 GET 接口 `http://api.track.yw56.com.cn/api/tracking`，请求头需带 `Authorization`，一次最多 30 个单号，官方未提供测试环境。后端接入前不会直接从浏览器请求。</p>
          <p v-if="trackingError" class="rounded-2xl border border-rose-500/30 bg-rose-500/10 px-4 py-3 text-xs leading-5 text-rose-700 dark:text-rose-300">{{ trackingError }}</p>
          <div v-if="trackingQueried && trackingEvents.length === 0" class="flex min-h-48 flex-col items-center justify-center gap-3 rounded-2xl border border-dashed border-border/80 bg-card px-6 py-10 text-center"><span class="flex size-11 items-center justify-center rounded-2xl bg-muted text-muted-foreground"><RadioTower class="size-5" /></span><div><p class="text-sm font-black">暂无真实轨迹节点</p><p class="mt-1 max-w-lg text-xs leading-5 text-muted-foreground">燕文独立轨迹接口尚未接入或没有返回节点，不展示模拟运输状态；停滞告警也将在真实节点接入后计算。</p></div></div>
          <ol v-else-if="trackingEvents.length > 0" class="space-y-4 rounded-2xl border border-dashed border-border/80 bg-card p-4"><li v-for="event in trackingEvents" :key="`${event.code}-${event.occurredAt}`" class="flex gap-3"><span class="mt-0.5 flex size-8 shrink-0 items-center justify-center rounded-full bg-emerald-500/10 text-emerald-600"><CheckCircle2 class="size-4" /></span><div><p class="text-sm font-bold">{{ event.label }}</p><p class="mt-1 text-xs text-muted-foreground">{{ event.detail }}</p><p class="mt-1 font-mono text-[10px] text-muted-foreground">{{ event.occurredAt }}</p></div></li></ol>
        </section>

        <div class="grid gap-4 xl:grid-cols-[minmax(0,1.15fr)_minmax(320px,0.85fr)]">
          <section class="overflow-hidden rounded-[24px] border border-dashed border-border/80 bg-card">
            <div class="border-b border-dashed border-border/80 px-5 py-4"><h3 class="text-sm font-black">燕文状态码标准化</h3><p class="mt-1 text-[11px] text-muted-foreground">原生节点映射为统一领域事件，供燕文域内部使用。</p></div>
            <div class="divide-y divide-dashed divide-border/70"> <div v-for="status in trackingStatusMap" :key="status.code" class="flex items-center gap-3 px-5 py-3"><span :class="status.tone" class="rounded-full px-2.5 py-1 font-mono text-[10px] font-bold">{{ status.code }}</span><div class="min-w-0 flex-1"><p class="text-xs font-bold">{{ status.label }}</p><p class="mt-0.5 text-[11px] text-muted-foreground">{{ status.detail }}</p></div></div></div>
          </section>
          <section class="space-y-3 rounded-[24px] border border-dashed border-border/80 bg-card p-4"><div class="flex items-center gap-3"><span class="flex size-10 items-center justify-center rounded-2xl bg-amber-500/10 text-amber-700"><AlertTriangle class="size-4" /></span><div><h3 class="text-sm font-black">智能停滞告警</h3><p class="mt-1 text-[11px] text-muted-foreground">这是系统内部运营规则，不是燕文官方状态结论。</p></div></div><div class="space-y-2 text-xs leading-5 text-muted-foreground"><div class="rounded-2xl border border-amber-500/20 bg-amber-500/5 p-3"><p class="font-bold text-amber-700 dark:text-amber-300">清关停滞 · ALERT</p><p class="mt-1">S303 或 IC50 超过 48 小时没有后续节点时，提示联系报关行。</p></div><div class="rounded-2xl border border-rose-500/20 bg-rose-500/5 p-3"><p class="font-bold text-rose-700 dark:text-rose-300">在途超时 · CRITICAL</p><p class="mt-1">LH20 后超过 7 个工作日没有目的国扫描时，生成客服异常工单。</p></div></div></section>
        </div>
      </section>
</template>
