<script setup lang="ts">
import { computed, ref } from 'vue'
import { Activity, AlertTriangle, Calculator, CheckCircle2, ClipboardCheck, Download, Globe2, PackageCheck, Plus, Power, Printer, RadioTower, RefreshCw, Search, Send, Settings2, Trash2, Truck } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'


const calculatorForm = ref({
  warehouse: '',
  destination: '',
  cargoType: '1',
  weight: '',
  length: '',
  width: '',
  height: '',
  postcode: '',
})
const calculatorRun = ref(false)
const calculatorError = ref('')
const calculatorDimensionsValid = computed(() => {
  const length = Number(calculatorForm.value.length)
  const width = Number(calculatorForm.value.width)
  const height = Number(calculatorForm.value.height)
  return [length, width, height].every((value) => Number.isFinite(value) && value > 0)
})
const calculatorActualWeight = computed(() => Math.max(0, Number(calculatorForm.value.weight) || 0) / 1000)
const calculatorCanRun = computed(() =>
  calculatorForm.value.warehouse.trim().length > 0
  && calculatorForm.value.destination.trim().length > 0
  && calculatorActualWeight.value > 0
  && calculatorDimensionsValid.value,
)
const rateCandidates: Array<{ code: string; name: string; validity: string }> = []
const runRateCalculator = () => {
  if (!calculatorCanRun.value) {
    calculatorError.value = '请填写交货仓、目的国、实重和完整三维后再执行试算。'
    calculatorRun.value = false
    return
  }
  calculatorError.value = ''
  calculatorRun.value = true
}
const resetRateCalculator = () => {
  calculatorForm.value = { warehouse: '', destination: '', cargoType: '1', weight: '', length: '', width: '', height: '', postcode: '' }
  calculatorRun.value = false
  calculatorError.value = ''
}
</script>

<template>
      <section class="space-y-4 rounded-[28px] border border-dashed border-border/80 bg-card p-5 shadow-sm sm:p-6">
        <div>
          <p class="text-[10px] font-black uppercase tracking-[0.18em] text-emerald-600">Rate Calculator &amp; Routing</p>
          <h2 class="mt-1 text-lg font-black tracking-tight">实时运价与路由选优</h2>
          <p class="mt-1 max-w-3xl text-xs leading-5 text-muted-foreground">录入交货仓、目的地、货品属性和包裹尺寸，先校验发货参数，再由 calc.list 返回官方渠道报价与时效。当前不伪造 RMB 运价。</p>
        </div>

        <div class="grid gap-4 xl:grid-cols-[minmax(280px,0.9fr)_minmax(0,1.5fr)]">
          <section class="space-y-4 rounded-[24px] border border-dashed border-border/80 bg-muted/20 p-4">
            <div class="flex items-center justify-between gap-3"><h3 class="text-sm font-black">测算参数</h3><Calculator class="size-4 text-muted-foreground" aria-hidden="true" /></div>
            <label class="space-y-1.5"><span class="text-xs font-bold">出发交货仓</span><Input v-model="calculatorForm.warehouse" placeholder="仓库代码 / 名称（待主数据同步）" /></label>
            <label class="space-y-1.5"><span class="text-xs font-bold">目的国家 / 地区</span><Input v-model="calculatorForm.destination" placeholder="US / 美国" /></label>
            <label class="space-y-1.5"><span class="text-xs font-bold">货品属性</span><select v-model="calculatorForm.cargoType" class="h-9 w-full rounded-md border border-border bg-background px-3 text-sm"><option value="1">普货（代码 1）</option><option value="2">特货 / 带电（代码 2）</option><option value="3">敏感货（代码 3）</option><option value="4">特品（代码 4）</option></select></label>
            <label class="space-y-1.5"><span class="text-xs font-bold">实重（g）</span><Input v-model="calculatorForm.weight" inputmode="decimal" placeholder="例如 850" /></label>
            <div class="grid grid-cols-3 gap-2"><label class="space-y-1.5"><span class="text-[11px] font-bold">长（cm）</span><Input v-model="calculatorForm.length" inputmode="decimal" placeholder="20" /></label><label class="space-y-1.5"><span class="text-[11px] font-bold">宽（cm）</span><Input v-model="calculatorForm.width" inputmode="decimal" placeholder="15" /></label><label class="space-y-1.5"><span class="text-[11px] font-bold">高（cm）</span><Input v-model="calculatorForm.height" inputmode="decimal" placeholder="10" /></label></div>
            <label class="space-y-1.5"><span class="text-xs font-bold">目的国邮编（可选）</span><Input v-model="calculatorForm.postcode" placeholder="用于偏远附加费校验" /></label>
            <div class="flex flex-wrap gap-2"><Button class="rounded-full" :disabled="!calculatorCanRun" @click="runRateCalculator"><Calculator class="size-3.5" />立即执行运价测算</Button><Button variant="ghost" size="sm" @click="resetRateCalculator">清空</Button></div>
            <p v-if="calculatorError" class="rounded-2xl border border-rose-500/30 bg-rose-500/10 px-3 py-2 text-xs leading-5 text-rose-700 dark:text-rose-300">{{ calculatorError }}</p>
          </section>

          <section class="space-y-4 rounded-[24px] border border-dashed border-border/80 bg-card p-4">
            <div class="flex items-center justify-between gap-3"><div><p class="text-[10px] font-black uppercase tracking-[0.16em] text-muted-foreground/70">Channel Comparison Matrix</p><h3 class="mt-1 text-sm font-black">渠道对比结果</h3></div><span class="rounded-full bg-muted px-2.5 py-1 text-[10px] font-bold text-muted-foreground">{{ calculatorRun ? '已计算计费重量' : '等待测算' }}</span></div>
            <div v-if="rateCandidates.length > 0" class="grid gap-3 sm:grid-cols-2">
              <article v-for="candidate in rateCandidates" :key="candidate.code" class="rounded-2xl border border-dashed border-border/80 bg-muted/20 p-4">
                <div class="flex items-start justify-between gap-2"><div><p class="text-sm font-black">{{ candidate.name }}</p><p class="mt-1 font-mono text-[10px] text-muted-foreground">{{ candidate.code }}</p></div><span class="rounded-full bg-muted px-2 py-1 text-[10px] font-bold text-muted-foreground">待接口</span></div>
                <dl class="mt-4 space-y-2 text-xs"><div class="flex justify-between gap-3"><dt class="text-muted-foreground">计费重量</dt><dd class="font-mono font-bold">—</dd></div><div class="flex justify-between gap-3"><dt class="text-muted-foreground">预估总运费</dt><dd class="font-mono font-bold">—</dd></div><div class="flex justify-between gap-3"><dt class="text-muted-foreground">预计时效</dt><dd class="font-bold">{{ candidate.validity }}</dd></div></dl>
                <p class="mt-3 text-[11px] leading-5 text-muted-foreground">报价字段由 calc.list 官方响应提供。</p>
              </article>
            </div>
            <div v-else class="flex min-h-48 flex-col items-center justify-center gap-3 rounded-2xl border border-dashed border-border/80 bg-muted/30 px-6 py-10 text-center"><span class="flex size-11 items-center justify-center rounded-2xl bg-muted text-muted-foreground"><Calculator class="size-5" /></span><div><p class="text-sm font-black">暂无官方报价结果</p><p class="mt-1 max-w-lg text-xs leading-5 text-muted-foreground">calc.list 需要燕文销售开通权限，并由后端按 cityId、countryId、productAttributes、productTypeList、weight 和尺寸参数调用。当前不硬编码产品 ID、材积系数或金额。</p></div></div>
            <div class="flex items-start gap-2 rounded-2xl border border-border/70 bg-muted/50 px-4 py-3 text-xs leading-5 text-muted-foreground"><Activity class="mt-0.5 size-4 shrink-0" aria-hidden="true" /><p>试算只在燕文域内运行。未接入官方报价接口前，系统不会用示例金额替代真实报价。</p></div>
          </section>
        </div>
      </section>
</template>
