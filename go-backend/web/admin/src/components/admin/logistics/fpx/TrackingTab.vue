<script setup lang="ts">
import { computed, ref } from 'vue'
import { AlertTriangle, CheckCircle2, Download, RadioTower, RefreshCw, Search } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'


const trackingQuery = ref('')
const trackingQueried = ref(false)
const trackingRefreshing = ref(false)
const trackingQueryError = ref('')
const trackingMilestones = ref<Array<{
  code: string
  label: string
  detail: string
  occurredAt: string
}>>([])
const trackingCanQuery = computed(() => trackingQuery.value.trim().length >= 4)
const trackingQueryLabel = computed(() => trackingQuery.value.trim() || '运单号 / 订单号')
const refreshTracking = async () => {
  if (!trackingCanQuery.value) {
    trackingQueryError.value = '请输入至少 4 位 4PX deliveryOrderNo；商城订单号需先完成委托映射'
    trackingQueried.value = false
    return
  }
  trackingRefreshing.value = true
  trackingQueryError.value = ''
  trackingQueried.value = true
  trackingMilestones.value = []
  await Promise.resolve()
  trackingRefreshing.value = false
}
const clearTracking = () => {
  trackingQuery.value = ''
  trackingQueried.value = false
  trackingQueryError.value = ''
  trackingMilestones.value = []
}
</script>

<template>
      <section class="space-y-5 rounded-[28px] border border-dashed border-border/80 bg-card p-5 shadow-sm sm:p-6">
        <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div>
            <p class="text-[10px] font-black uppercase tracking-[0.18em] text-indigo-600">Milestones Tracking & POD Evidence</p>
            <h2 class="mt-1 text-lg font-black tracking-tight">全球在途追踪与授权证明</h2>
            <p class="mt-1 max-w-3xl text-xs leading-5 text-muted-foreground">仅在 4PX 域内查询大件全节点轨迹。签收照片、签名原图和签收单只有在账号与产品获得证明接口授权后，才能关联到订单凭证包。</p>
          </div>
          <span class="inline-flex w-fit items-center gap-2 rounded-full border border-indigo-500/30 bg-indigo-500/10 px-3 py-1.5 text-[11px] font-bold text-indigo-700 dark:text-indigo-300">
            <RadioTower class="size-3.5" />
            仅 4PX 域内
          </span>
        </div>

        <form class="grid gap-3 rounded-[24px] border border-dashed border-border/80 bg-muted/20 p-4 md:grid-cols-[minmax(0,1fr)_auto_auto]" @submit.prevent="refreshTracking">
          <label class="space-y-1.5">
            <span class="flex items-center gap-1.5 text-[10px] font-black uppercase tracking-wider text-muted-foreground"><Search class="size-3" /> 运单查询</span>
            <Input v-model="trackingQuery" placeholder="输入 4PX deliveryOrderNo；商城单号需先映射" @input="trackingQueryError = ''" />
          </label>
          <Button type="submit" class="self-end" :disabled="trackingRefreshing">
            <RefreshCw :class="['size-3.5', { 'animate-spin': trackingRefreshing }]" />
            查询轨迹
          </Button>
          <Button type="button" variant="outline" class="self-end" @click="clearTracking">清空</Button>
        </form>
        <p v-if="trackingQueryError" class="rounded-2xl border border-destructive/30 bg-destructive/5 px-4 py-3 text-xs text-destructive">{{ trackingQueryError }}</p>
      </section>

      <section class="grid gap-4 xl:grid-cols-[minmax(0,1.2fr)_minmax(0,0.8fr)]">
        <article class="rounded-[28px] border border-dashed border-border/80 bg-card p-5 shadow-sm sm:p-6">
          <div class="flex items-start justify-between gap-3">
            <div>
              <p class="text-[10px] font-black uppercase tracking-[0.18em] text-muted-foreground">Tracking Timeline</p>
              <h2 class="mt-1 text-base font-black">轨迹里程碑</h2>
            </div>
            <span v-if="trackingQueried" class="rounded-full bg-muted px-2.5 py-1 text-[10px] font-bold text-muted-foreground">{{ trackingQueryLabel }}</span>
          </div>

          <div v-if="!trackingQueried" class="mt-5 flex min-h-64 flex-col items-center justify-center gap-3 rounded-[24px] border border-dashed border-border/80 bg-muted/20 px-6 py-10 text-center">
            <span class="flex size-12 items-center justify-center rounded-2xl bg-background text-muted-foreground shadow-sm"><RadioTower class="size-5" /></span>
            <div>
              <p class="text-sm font-black">输入运单号开始追踪</p>
              <p class="mt-1 max-w-md text-xs leading-5 text-muted-foreground">查询会调用 4PX tr.order.tracking.get；接口要求 deliveryOrderNo，商城订单号必须先由委托查询完成映射。是否包含签收节点取决于产品和承运商回传。</p>
            </div>
          </div>
          <div v-else-if="trackingMilestones.length === 0" class="mt-5 flex min-h-64 flex-col items-center justify-center gap-3 rounded-[24px] border border-dashed border-amber-500/30 bg-amber-500/5 px-6 py-10 text-center">
            <span class="flex size-12 items-center justify-center rounded-2xl bg-amber-500/10 text-amber-700 dark:text-amber-300"><AlertTriangle class="size-5" /></span>
            <div>
              <p class="text-sm font-black">暂无 4PX 轨迹数据</p>
              <p class="mt-1 max-w-md text-xs leading-5 text-muted-foreground">当前查询已执行，但 4PX 轨迹接口尚未接入或没有返回节点，不展示模拟运输状态。</p>
            </div>
          </div>
          <ol v-else class="mt-5 space-y-4">
            <li v-for="milestone in trackingMilestones" :key="milestone.code" class="flex gap-3">
              <span class="mt-0.5 flex size-8 shrink-0 items-center justify-center rounded-full bg-indigo-500/10 text-indigo-600"><CheckCircle2 class="size-4" /></span>
              <div><p class="text-sm font-bold">{{ milestone.label }}</p><p class="mt-1 text-xs text-muted-foreground">{{ milestone.detail }}</p><p class="mt-1 font-mono text-[10px] text-muted-foreground">{{ milestone.occurredAt }}</p></div>
            </li>
          </ol>
        </article>

        <article class="rounded-[28px] border border-dashed border-border/80 bg-card p-5 shadow-sm sm:p-6">
          <div class="flex items-start justify-between gap-3">
            <div>
              <p class="text-[10px] font-black uppercase tracking-[0.18em] text-muted-foreground">Proof Of Delivery</p>
              <h2 class="mt-1 text-base font-black">签收存证链</h2>
            </div>
            <Button variant="outline" size="sm" disabled title="需要 4PX 证明接口授权并完成轨迹查询">
              <Download class="size-3.5" />
              导出举证包
            </Button>
          </div>

          <div class="mt-5 space-y-3">
            <div class="rounded-2xl border border-dashed border-border/80 bg-muted/20 p-4">
              <div class="flex items-center justify-between gap-3"><p class="text-xs font-black">末端照片</p><span class="rounded-full bg-muted px-2.5 py-1 text-[10px] font-bold text-muted-foreground">待同步</span></div>
              <p class="mt-2 text-[11px] leading-5 text-muted-foreground">买家门口照片、包裹外箱和派送位置图片需要 4PX 账号/产品级证明接口授权后才能关联至订单证据包。</p>
            </div>
            <div class="rounded-2xl border border-dashed border-border/80 bg-muted/20 p-4">
              <div class="flex items-center justify-between gap-3"><p class="text-xs font-black">签名原图 / 签收单</p><span class="rounded-full bg-muted px-2.5 py-1 text-[10px] font-bold text-muted-foreground">待同步</span></div>
              <p class="mt-2 text-[11px] leading-5 text-muted-foreground">原图 Base64 与签收人信息仅在 `ds.xms.certificate.sign.query` 获得授权并返回后保存为订单凭证附件。</p>
            </div>
            <div class="flex items-start gap-2 rounded-2xl border border-indigo-500/30 bg-indigo-500/10 px-4 py-3 text-xs leading-5 text-indigo-800 dark:text-indigo-200">
              <CheckCircle2 class="mt-0.5 size-4 shrink-0" />
              <p>证据包设计与现有履约证据域对齐，发生 PayPal 申诉或 Chargeback 时，再由授权用户导出完整快照。</p>
            </div>
          </div>
        </article>
      </section>
</template>
