<script setup lang="ts">
import { computed, ref } from 'vue'
import { Activity, AlertTriangle, Calculator, CheckCircle2, ClipboardCheck, Download, Globe2, PackageCheck, Plus, Power, Printer, RadioTower, RefreshCw, Search, Send, Settings2, Trash2, Truck } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'


const overviewMetrics = [
  { label: '今日创建运单', value: '—', detail: '等待 express.order.create 数据接入', icon: Send },
  { label: '待打印面单', value: '—', detail: '等待面单生成与打印队列接入', icon: PackageCheck },
  { label: '揽收在途包裹', value: '—', detail: '等待揽收与轨迹轮询接入', icon: Truck },
  { label: '清关异常预警', value: '—', detail: '等待 S303 / IC51 / IC70 状态接入', icon: AlertTriangle },
  { label: '本月预估运费', value: '—', detail: '等待 calc.list 运价数据接入', icon: Activity },
]

const destinationFlows: Array<{ code: string; name: string; service: string }> = []

const gatewayChecking = ref(false)
const gatewayChecked = ref(false)
const gatewayCheckMessage = ref('尚未执行自检')
const checkGatewayHealth = async () => {
  gatewayChecking.value = true
  gatewayChecked.value = false
  gatewayCheckMessage.value = '正在准备燕文网关自检...'
  await Promise.resolve()
  gatewayChecking.value = false
  gatewayChecked.value = true
  gatewayCheckMessage.value = '真实网关接口尚未接入，未产生成功结论'
}
</script>

<template>
      <section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-5" aria-label="燕文运营指标">
        <article
          v-for="metric in overviewMetrics"
          :key="metric.label"
          class="group relative min-h-36 overflow-hidden rounded-[24px] border border-dashed border-border/80 bg-card p-4"
        >
          <div class="pointer-events-none absolute inset-0 bg-gradient-to-br from-emerald-500/5 via-transparent to-transparent" />
          <div class="relative flex h-full flex-col justify-between gap-4">
            <div class="flex items-center justify-between gap-2">
              <p class="text-[10px] font-black uppercase tracking-wider text-muted-foreground/80">{{ metric.label }}</p>
              <component :is="metric.icon" class="size-4 text-muted-foreground/70" aria-hidden="true" />
            </div>
            <div>
              <p class="font-mono text-3xl font-black tracking-tight text-foreground">{{ metric.value }}</p>
              <p class="mt-1 text-[11px] leading-4 text-muted-foreground">{{ metric.detail }}</p>
            </div>
          </div>
        </article>
      </section>

      <section class="space-y-4 rounded-[28px] border border-dashed border-border/80 bg-muted/5 p-5 sm:p-6">
        <div class="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <p class="text-[10px] font-black uppercase tracking-[0.18em] text-muted-foreground/70">Yanwen Logistics Operations</p>
            <h2 class="mt-1 text-lg font-black tracking-tight">燕文跨境专线全链路监控</h2>
            <p class="mt-1 text-xs text-muted-foreground">运单吞吐、面单时效、在途清关与接口健康度将在燕文 API 接入后展示。</p>
          </div>
          <span class="inline-flex w-fit items-center gap-2 rounded-full border border-amber-500/30 bg-amber-500/10 px-3 py-1.5 text-[11px] font-bold text-amber-700 dark:text-amber-300">
            <span class="size-1.5 rounded-full bg-amber-500" />
            接口待接入
          </span>
        </div>

        <div class="grid gap-4 xl:grid-cols-2">
          <article class="rounded-[24px] border border-dashed border-border/80 bg-card p-4">
            <div class="flex items-center justify-between gap-3">
              <div>
                <p class="text-[10px] font-black uppercase tracking-[0.16em] text-muted-foreground/70">Destination Mix</p>
                <h3 class="mt-1 text-sm font-black">专线国家流向 Top 5</h3>
              </div>
              <Globe2 class="size-4 text-muted-foreground" aria-hidden="true" />
            </div>
            <div v-if="destinationFlows.length > 0" class="mt-4 divide-y divide-dashed divide-border/70">
              <div v-for="destination in destinationFlows" :key="destination.code" class="flex items-center gap-3 py-3 first:pt-0 last:pb-0">
                <span class="flex size-9 shrink-0 items-center justify-center rounded-xl bg-muted font-mono text-[10px] font-black">{{ destination.code }}</span>
                <div class="min-w-0 flex-1"><p class="truncate text-sm font-bold">{{ destination.name }}</p><p class="mt-0.5 text-[11px] text-muted-foreground">{{ destination.service }}</p></div>
                <span class="rounded-full bg-muted px-2.5 py-1 text-[10px] font-bold text-muted-foreground">待同步</span>
              </div>
            </div>
            <div v-else class="mt-4 rounded-2xl border border-dashed border-border/80 bg-muted/30 px-4 py-8 text-center text-xs leading-5 text-muted-foreground">通达国家和产品主数据尚未同步，不展示示例流向。</div>
          </article>

          <article class="rounded-[24px] border border-dashed border-border/80 bg-card p-4">
            <div class="flex items-center justify-between gap-3">
              <div>
                <p class="text-[10px] font-black uppercase tracking-[0.16em] text-muted-foreground/70">API Gateway Health</p>
                <h3 class="mt-1 text-sm font-black">燕文网关连通性自检</h3>
              </div>
              <Button variant="outline" size="sm" :disabled="gatewayChecking" @click="checkGatewayHealth">
                <RefreshCw :class="['size-3.5', { 'animate-spin': gatewayChecking }]" />
                自检
              </Button>
            </div>
            <div class="mt-4 space-y-3">
              <div class="flex items-center gap-3 rounded-2xl border border-dashed border-border/80 bg-muted/30 p-3">
                <span :class="gatewayChecked ? 'bg-amber-500/10 text-amber-600' : 'bg-muted text-muted-foreground'" class="flex size-10 items-center justify-center rounded-2xl"><CheckCircle2 class="size-4" /></span>
                <div><p class="text-sm font-black">{{ gatewayChecked ? '未形成健康结论' : '尚未执行自检' }}</p><p class="mt-1 text-[11px] leading-5 text-muted-foreground">{{ gatewayCheckMessage }}</p></div>
              </div>
              <div class="grid gap-2 sm:grid-cols-3">
                <div class="rounded-2xl bg-muted/50 p-3"><p class="text-[10px] font-bold text-muted-foreground">运行环境</p><p class="mt-1 text-xs font-black">待配置</p></div>
                <div class="rounded-2xl bg-muted/50 p-3"><p class="text-[10px] font-bold text-muted-foreground">平均延迟</p><p class="mt-1 font-mono text-xs font-black">—</p></div>
                <div class="rounded-2xl bg-muted/50 p-3"><p class="text-[10px] font-bold text-muted-foreground">今日调用</p><p class="mt-1 font-mono text-xs font-black">—</p></div>
              </div>
            </div>
          </article>
        </div>

        <div class="flex items-start gap-2 rounded-2xl border border-border/70 bg-card/70 px-4 py-3 text-xs leading-5 text-muted-foreground">
          <AlertTriangle class="mt-0.5 size-4 shrink-0" aria-hidden="true" />
          <p>当前大盘只展示燕文域结构与待接入状态，不使用模拟订单、运价、清关或接口健康数据。完成网关配置后，再接入运单、面单、轨迹和关务汇总。</p>
        </div>
      </section>
</template>
