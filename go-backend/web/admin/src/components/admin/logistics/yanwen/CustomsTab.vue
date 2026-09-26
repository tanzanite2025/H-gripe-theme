<script setup lang="ts">
import { computed, ref } from 'vue'
import { Activity, AlertTriangle, Calculator, CheckCircle2, ClipboardCheck, Download, Globe2, PackageCheck, Plus, Power, Printer, RadioTower, RefreshCw, Search, Send, Settings2, Trash2, Truck } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'


const customsMode = ref<'korea' | 'us' | 'declaration'>('korea')
const customsForm = ref({
  pccc: '',
  recipientName: '',
  address: '',
  city: '',
  state: '',
  postcode: '',
  ioss: '',
  ukVat: '',
})
const customsValidating = ref(false)
const customsValidated = ref(false)
const customsValidationMessage = ref('')
const customsValidationTone = ref<'neutral' | 'warning' | 'critical'>('neutral')
const validateCustoms = async () => {
  customsValidating.value = true
  customsValidated.value = false
  customsValidationMessage.value = ''
  customsValidationTone.value = 'neutral'
  const form = customsForm.value
  if (customsMode.value === 'korea') {
    const pcccValid = /^P[A-Z0-9]{12}$/i.test(form.pccc.trim())
    if (!pcccValid || form.recipientName.trim().length === 0) {
      customsValidationTone.value = 'critical'
      customsValidationMessage.value = '本地格式校验未通过：韩国 PCCC 必须为 P 开头的 13 位编码，并填写匹配的收件人姓名。'
    }
  } else if (customsMode.value === 'us') {
    const postcodeValid = /^\d{5}(?:-\d{4})?$/.test(form.postcode.trim())
    if (!postcodeValid || [form.address, form.city, form.state].some((value) => value.trim().length === 0)) {
      customsValidationTone.value = 'critical'
      customsValidationMessage.value = '本地格式校验未通过：请填写完整美国地址、州、城市和 5 位或 9 位 ZIP Code。'
    }
  } else if (form.ioss.trim().length === 0 && form.ukVat.trim().length === 0) {
    customsValidationTone.value = 'warning'
    customsValidationMessage.value = '请至少录入欧盟 IOSS 或英国 VAT 税号，系统才能在运单报文中进行税号映射。'
  }
  if (!customsValidationMessage.value) {
    customsValidationTone.value = 'warning'
    customsValidationMessage.value = '本地格式校验已通过，但真实燕文关务接口尚未接入，未产生实名、地址或税号有效结论。'
  }
  await Promise.resolve()
  customsValidating.value = false
  customsValidated.value = true
}
const clearCustoms = () => {
  customsForm.value = { pccc: '', recipientName: '', address: '', city: '', state: '', postcode: '', ioss: '', ukVat: '' }
  customsValidated.value = false
  customsValidationMessage.value = ''
  customsValidationTone.value = 'neutral'
}
</script>

<template>
      <section class="space-y-4 rounded-[28px] border border-dashed border-border/80 bg-card p-5 shadow-sm sm:p-6">
        <div>
          <p class="text-[10px] font-black uppercase tracking-[0.18em] text-emerald-600">Customs Compliance &amp; Pre-Validation</p>
          <h2 class="mt-1 text-lg font-black tracking-tight">关务合规与前置校验</h2>
          <p class="mt-1 max-w-3xl text-xs leading-5 text-muted-foreground">在创建燕文运单前校验韩国通关码、美国地址和欧盟/英国税号。格式校验与官方实名、邮政、税号接口结果分开显示。</p>
        </div>

        <div class="grid gap-4 xl:grid-cols-[minmax(280px,0.95fr)_minmax(0,1.05fr)]">
          <section class="space-y-4 rounded-[24px] border border-dashed border-border/80 bg-muted/20 p-4">
            <div class="flex items-center justify-between gap-3"><h3 class="text-sm font-black">前置校验项目</h3><ClipboardCheck class="size-4 text-muted-foreground" aria-hidden="true" /></div>
            <div class="grid grid-cols-3 gap-2 rounded-2xl bg-muted p-1" role="group" aria-label="关务校验项目"><button v-for="item in [{ key: 'korea', label: '韩国 PCCC' }, { key: 'us', label: '美国地址' }, { key: 'declaration', label: '税号映射' }]" :key="item.key" type="button" :class="customsMode === item.key ? 'bg-background text-foreground shadow-sm' : 'text-muted-foreground'" class="min-h-9 rounded-xl px-2 py-1 text-[11px] font-black" @click="customsMode = item.key as typeof customsMode">{{ item.label }}</button></div>
            <template v-if="customsMode === 'korea'">
              <label class="space-y-1.5"><span class="text-xs font-bold">个人通关码 PCCC</span><Input v-model="customsForm.pccc" placeholder="P 开头的 13 位编码" autocomplete="off" /></label>
              <label class="space-y-1.5"><span class="text-xs font-bold">韩文 / 英文收件人姓名</span><Input v-model="customsForm.recipientName" placeholder="必须与 PCCC 实名一致" /></label>
            </template>
            <template v-else-if="customsMode === 'us'">
              <label class="space-y-1.5"><span class="text-xs font-bold">街道地址</span><Input v-model="customsForm.address" placeholder="Street address" /></label>
              <div class="grid grid-cols-2 gap-2"><label class="space-y-1.5"><span class="text-[11px] font-bold">城市</span><Input v-model="customsForm.city" placeholder="City" /></label><label class="space-y-1.5"><span class="text-[11px] font-bold">州</span><Input v-model="customsForm.state" placeholder="State" /></label></div>
              <label class="space-y-1.5"><span class="text-xs font-bold">ZIP Code</span><Input v-model="customsForm.postcode" placeholder="5 位或 9 位邮编" /></label>
            </template>
            <template v-else>
              <label class="space-y-1.5"><span class="text-xs font-bold">欧盟 IOSS 税号</span><Input v-model="customsForm.ioss" placeholder="用于欧盟订单税号映射" /></label>
              <label class="space-y-1.5"><span class="text-xs font-bold">英国 VAT 税号</span><Input v-model="customsForm.ukVat" placeholder="用于英国订单税号映射" /></label>
            </template>
            <div class="flex flex-wrap gap-2"><Button :disabled="customsValidating" @click="validateCustoms"><ClipboardCheck class="size-3.5" />执行前置校验</Button><Button variant="ghost" size="sm" @click="clearCustoms">清空</Button></div>
            <div v-if="customsValidated" :class="customsValidationTone === 'critical' ? 'border-rose-500/30 bg-rose-500/10 text-rose-700 dark:text-rose-300' : 'border-amber-500/30 bg-amber-500/10 text-amber-700 dark:text-amber-300'" class="rounded-2xl border px-4 py-3 text-xs leading-5">{{ customsValidationMessage }}</div>
          </section>

          <section class="space-y-3 rounded-[24px] border border-dashed border-border/80 bg-card p-4"><div class="flex items-center gap-3"><span class="flex size-10 items-center justify-center rounded-2xl bg-emerald-500/10 text-emerald-600"><CheckCircle2 class="size-4" /></span><div><h3 class="text-sm font-black">三道防线</h3><p class="mt-1 text-[11px] text-muted-foreground">仅当真实接口响应明确通过后，才允许继续推单。</p></div></div><div class="space-y-2 text-xs leading-5"><div class="rounded-2xl border border-border/70 bg-muted/30 p-3"><p class="font-bold">1 · 韩国 PCCC 强校验</p><p class="mt-1 text-muted-foreground">调用 common.verify.kr.pccc，提交收件人姓名、电话、税号和 5 位邮编。</p></div><div class="rounded-2xl border border-border/70 bg-muted/30 p-3"><p class="font-bold">2 · 美国地址规范</p><p class="mt-1 text-muted-foreground">调用 common.verify.us.address，返回标准化地址、城市、州和 ZIP 拆分结果。</p></div><div class="rounded-2xl border border-border/70 bg-muted/30 p-3"><p class="font-bold">3 · 税号与申报映射</p><p class="mt-1 text-muted-foreground">IOSS、EORI 等字段只按订单目的国和官方报文字段映射，不把本地格式校验当作官方通过。</p></div></div></section>
        </div>
      </section>
</template>
