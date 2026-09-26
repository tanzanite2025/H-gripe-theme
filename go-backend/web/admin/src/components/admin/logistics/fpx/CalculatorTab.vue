<script setup lang="ts">
import { computed, ref } from 'vue'
import { AlertTriangle, Calculator, Package, Truck } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'


const calculatorPresets = [
  { label: '标准 700c 轮组单箱', length: 83, width: 18, height: 68 },
  { label: '双轮组并包箱', length: 83, width: 32, height: 68 },
  { label: '一体把车架大箱', length: 105, width: 22, height: 60 },
]
const calculatorForm = ref({
  destination: '',
  length: '83',
  width: '18',
  height: '68',
  actualWeight: '',
  volumeDivisor: '6000',
})
const calculatorQuoted = ref(false)
const calculatorVolumetricWeight = computed(() => {
  const length = Number(calculatorForm.value.length)
  const width = Number(calculatorForm.value.width)
  const height = Number(calculatorForm.value.height)
  const divisor = Number(calculatorForm.value.volumeDivisor)
  if (![length, width, height, divisor].every((value) => Number.isFinite(value) && value > 0)) return 0
  return (length * width * height) / divisor
})
const calculatorActualWeight = computed(() => Math.max(0, Number(calculatorForm.value.actualWeight) || 0))
const calculatorBillingWeight = computed(() => Math.max(calculatorActualWeight.value, calculatorVolumetricWeight.value))
const calculatorCanQuote = computed(() =>
  calculatorForm.value.destination.trim().length > 0
  && calculatorActualWeight.value > 0
  && calculatorVolumetricWeight.value > 0,
)
const applyCalculatorPreset = (preset: typeof calculatorPresets[number]) => {
  calculatorForm.value.length = String(preset.length)
  calculatorForm.value.width = String(preset.width)
  calculatorForm.value.height = String(preset.height)
  calculatorQuoted.value = false
}
const requestCalculatorQuote = () => {
  if (!calculatorCanQuote.value) return
  calculatorQuoted.value = true
}
</script>

<template>
      <section class="space-y-5 rounded-[28px] border border-dashed border-border/80 bg-card p-5 shadow-sm sm:p-6">
        <div>
          <p class="text-[10px] font-black uppercase tracking-[0.18em] text-cyan-600">Bulky Rate Calculator</p>
          <h2 class="mt-1 text-lg font-black tracking-tight">大件运费试算与材积选优</h2>
          <p class="mt-1 max-w-3xl text-xs leading-5 text-muted-foreground">按实重与材积重取大计算计费重，并为接入 4PX ds.xms.estimated_cost.get 后的直发报价预留结果区。</p>
        </div>

        <div class="space-y-2">
          <p class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">箱型智能预设</p>
          <div class="flex flex-wrap gap-2">
            <Button v-for="preset in calculatorPresets" :key="preset.label" variant="outline" size="sm" @click="applyCalculatorPreset(preset)">
              <Package class="size-3.5" />
              {{ preset.label }}（{{ preset.length }} × {{ preset.width }} × {{ preset.height }}）
            </Button>
          </div>
        </div>

        <div class="grid gap-4 rounded-[24px] border border-dashed border-border/80 bg-muted/20 p-4 sm:grid-cols-2 xl:grid-cols-6">
          <label class="space-y-1.5 sm:col-span-2">
            <span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">目的国家 / 地区</span>
            <Input v-model="calculatorForm.destination" placeholder="例如 US、DE、GB" @input="calculatorQuoted = false" />
          </label>
          <label class="space-y-1.5"><span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">长度 cm</span><Input v-model="calculatorForm.length" inputmode="decimal" @input="calculatorQuoted = false" /></label>
          <label class="space-y-1.5"><span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">宽度 cm</span><Input v-model="calculatorForm.width" inputmode="decimal" @input="calculatorQuoted = false" /></label>
          <label class="space-y-1.5"><span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">高度 cm</span><Input v-model="calculatorForm.height" inputmode="decimal" @input="calculatorQuoted = false" /></label>
          <label class="space-y-1.5"><span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">实重 kg（提交时换算为 g）</span><Input v-model="calculatorForm.actualWeight" inputmode="decimal" placeholder="例如 12.5" @input="calculatorQuoted = false" /></label>
          <label class="space-y-1.5"><span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">材积除数（预校验）</span><Input v-model="calculatorForm.volumeDivisor" inputmode="numeric" @input="calculatorQuoted = false" /></label>
          <div class="flex items-end sm:col-span-2 xl:col-span-1"><Button class="w-full" :disabled="!calculatorCanQuote" @click="requestCalculatorQuote"><Calculator class="size-3.5" />开始试算</Button></div>
        </div>

        <div class="grid gap-3 sm:grid-cols-3">
          <article class="rounded-[22px] border border-dashed border-border/80 bg-muted/20 p-4"><p class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">实重</p><p class="mt-2 font-mono text-2xl font-black">{{ calculatorActualWeight.toFixed(2) }} kg</p></article>
          <article class="rounded-[22px] border border-dashed border-border/80 bg-muted/20 p-4"><p class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">材积重</p><p class="mt-2 font-mono text-2xl font-black">{{ calculatorVolumetricWeight.toFixed(2) }} kg</p><p class="mt-1 text-[11px] text-muted-foreground">长 × 宽 × 高 ÷ {{ calculatorForm.volumeDivisor || '6000' }}</p></article>
          <article class="rounded-[22px] border border-cyan-500/30 bg-cyan-500/10 p-4"><p class="text-[10px] font-black uppercase tracking-wider text-cyan-700 dark:text-cyan-300">计费重</p><p class="mt-2 font-mono text-2xl font-black">{{ calculatorBillingWeight.toFixed(2) }} kg</p><p class="mt-1 text-[11px] text-cyan-700/80 dark:text-cyan-300/80">实重与材积重取大</p></article>
        </div>
      </section>

      <section class="overflow-hidden rounded-[28px] border border-dashed border-border/80 bg-card shadow-sm">
        <div class="border-b border-dashed border-border/80 px-5 py-4 sm:px-6"><h2 class="text-base font-black">直发报价结果</h2><p class="mt-1 text-xs text-muted-foreground">4PX 直发真实价格与时效将在费率接口接通后返回。</p></div>
        <div class="grid gap-3 p-5 sm:p-6">
          <article class="rounded-[24px] border border-dashed border-border/80 p-5">
            <div class="flex items-center justify-between gap-3"><div><p class="text-[10px] font-black uppercase tracking-wider text-orange-600">Direct Bulky</p><h3 class="mt-1 font-black">国内大件专线直发</h3></div><Truck class="size-5 text-muted-foreground" /></div>
            <dl class="mt-5 grid grid-cols-2 gap-3 text-xs"><div class="rounded-2xl bg-muted/40 p-3"><dt class="text-muted-foreground">计费重</dt><dd class="mt-1 font-mono font-black">{{ calculatorBillingWeight.toFixed(2) }} kg</dd></div><div class="rounded-2xl bg-muted/40 p-3"><dt class="text-muted-foreground">参考时效</dt><dd class="mt-1 font-black">7–10 天</dd></div><div class="col-span-2 rounded-2xl bg-muted/40 p-3"><dt class="text-muted-foreground">4PX 实时运费</dt><dd class="mt-1 font-black">{{ calculatorQuoted ? '接口尚未接入，无法报价' : '填写信息后开始试算' }}</dd></div></dl>
          </article>
        </div>
        <div v-if="calculatorQuoted" class="mx-5 mb-5 flex items-start gap-2 rounded-2xl border border-amber-500/30 bg-amber-500/10 px-4 py-3 text-xs text-amber-800 sm:mx-6 sm:mb-6 dark:text-amber-200"><AlertTriangle class="mt-0.5 size-4 shrink-0" /><p>本地计算仅用于核对实重、材积重和计费重，没有生成虚假运费。接入 4PX ds.xms.estimated_cost.get 后再展示真实直发报价。</p></div>
      </section>
</template>
