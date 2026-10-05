<script setup lang="ts">
import { computed, ref } from 'vue'
import { AlertTriangle, CheckCircle2, ClipboardCheck } from '@lucide/vue'
import yanwenLogisticsAdminApi from '@/api/yanwenLogisticsAdminApi'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useAuthStore } from '@/stores/auth'

type CustomsMode = 'korea' | 'us' | 'declaration'
type YanwenEnvironment = 'fat' | 'production'

const authStore = useAuthStore()
const canShip = computed(() => authStore.hasPermission('logistics:yanwen:ship'))
const customsMode = ref<CustomsMode>('korea')
const customsForm = ref({
  environment: 'fat' as YanwenEnvironment,
  pccc: '',
  recipientName: '',
  phone: '',
  postalCode: '',
  address: '',
  city: '',
  state: '',
})
const customsValidating = ref(false)
const customsValidated = ref(false)
const customsOfficialPassed = ref(false)
const customsStatusOnly = ref(false)
const customsValidationMessage = ref('')
const customsValidationTone = ref<'neutral' | 'warning' | 'critical' | 'success'>('neutral')
const customsNormalizedAddress = ref<{
  address: string
  city: string
  state: string
  zipCode4: string
  zipCode5: string
} | null>(null)

const customsValidationClass = computed(() => {
  if (customsValidationTone.value === 'success') return 'border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300'
  if (customsValidationTone.value === 'critical') return 'border-rose-500/30 bg-rose-500/10 text-rose-700 dark:text-rose-300'
  return 'border-amber-500/30 bg-amber-500/10 text-amber-700 dark:text-amber-300'
})

function getCustomsVerificationErrorMessage(cause: unknown, fallback: string): string {
  if (typeof cause === 'object' && cause !== null) {
    const candidate = cause as {
      message?: unknown
      response?: { data?: { message?: unknown; error?: unknown } }
    }
    const apiMessage = candidate.response?.data?.message ?? candidate.response?.data?.error
    if (typeof apiMessage === 'string' && apiMessage.trim()) return apiMessage
    if (typeof candidate.message === 'string' && candidate.message.trim()) return candidate.message
  }
  return fallback
}

function clearCustomsVerificationResult() {
  customsValidated.value = false
  customsOfficialPassed.value = false
  customsStatusOnly.value = false
  customsValidationMessage.value = ''
  customsValidationTone.value = 'neutral'
  customsNormalizedAddress.value = null
}

function selectCustomsMode(mode: CustomsMode) {
  customsMode.value = mode
  clearCustomsVerificationResult()
}

async function verifySelectedYanwenCustomsPreflight() {
  customsValidating.value = true
  clearCustomsVerificationResult()
  try {
    if (customsMode.value === 'declaration') {
      customsStatusOnly.value = true
      customsValidationTone.value = 'neutral'
      customsValidationMessage.value = '燕文没有独立的 IOSS / EORI 校验接口；真实建单时按官方 express.order.create 报文字段映射。此状态不代表燕文已校验税号。'
      return
    }
    if (!canShip.value) {
      customsValidationTone.value = 'critical'
      customsValidationMessage.value = '需要 logistics:yanwen:ship 权限才能调用燕文官方关务接口。'
      return
    }
    const form = customsForm.value
    if (customsMode.value === 'korea' && (!form.recipientName.trim() || !form.phone.trim() || !form.pccc.trim())) {
      customsValidationTone.value = 'critical'
      customsValidationMessage.value = '请填写收件人姓名、联系方式和韩国个人通关码。'
      return
    }
    if (customsMode.value === 'korea' && !/^\d{5}$/.test(form.postalCode.trim())) {
      customsValidationTone.value = 'critical'
      customsValidationMessage.value = '燕文官方要求韩国收件人邮编为 5 位数字。'
      return
    }
    if (customsMode.value === 'us' && (!form.address.trim() || !form.city.trim() || !form.state.trim() || !form.postalCode.trim())) {
      customsValidationTone.value = 'critical'
      customsValidationMessage.value = '请填写完整的美国地址、城市、州和邮编。'
      return
    }
    if (customsMode.value === 'korea') {
      const result = await yanwenLogisticsAdminApi.verifyYanwenKoreaPersonalCustomsClearanceCode({
        environment: form.environment,
        recipient_name: form.recipientName.trim(),
        phone: form.phone.trim(),
        tax_number: form.pccc.trim(),
        zip_code: form.postalCode.trim(),
      })
      customsOfficialPassed.value = result.official_passed === true
      if (customsOfficialPassed.value) {
        customsValidationTone.value = 'success'
        customsValidationMessage.value = result.message?.trim() || '燕文官方已返回 code=0，韩国 PCCC 校验通过。'
      } else {
        customsValidationTone.value = 'critical'
        customsValidationMessage.value = result.message?.trim() || '燕文官方未确认韩国 PCCC 通过。'
      }
      return
    }
    const result = await yanwenLogisticsAdminApi.verifyYanwenUnitedStatesAddress({
      environment: form.environment,
      address: form.address.trim(),
      zip_code: form.postalCode.trim(),
      city: form.city.trim(),
      state: form.state.trim(),
    })
    customsOfficialPassed.value = result.official_passed === true
    if (customsOfficialPassed.value) {
      customsValidationTone.value = 'success'
      customsValidationMessage.value = result.message?.trim() || '燕文官方已返回 code=0，美国地址校验通过。'
      customsNormalizedAddress.value = {
        address: result.normalized_address,
        city: result.normalized_city,
        state: result.normalized_state,
        zipCode4: result.normalized_zip_code4,
        zipCode5: result.normalized_zip_code5,
      }
    } else {
      customsValidationTone.value = 'critical'
      customsValidationMessage.value = result.message?.trim() || '燕文官方未确认美国地址通过。'
    }
  } catch (cause) {
    customsValidationTone.value = 'critical'
    customsValidationMessage.value = getCustomsVerificationErrorMessage(cause, customsMode.value === 'korea'
      ? '燕文韩国 PCCC 官方校验失败，未产生通过结论。'
      : '燕文美国地址官方校验失败，未产生通过结论。')
  } finally {
    customsValidating.value = false
    customsValidated.value = true
  }
}

function clearCustomsForm() {
  customsForm.value = {
    environment: 'fat',
    pccc: '',
    recipientName: '',
    phone: '',
    postalCode: '',
    address: '',
    city: '',
    state: '',
  }
  clearCustomsVerificationResult()
}
</script>

<template>
  <section class="space-y-4 rounded-[28px] border border-dashed border-border/80 bg-card p-5 shadow-sm sm:p-6">
    <div>
      <p class="text-[10px] font-black uppercase tracking-[0.18em] text-emerald-600">Customs Compliance &amp; Pre-Validation</p>
      <h2 class="mt-1 text-lg font-black tracking-tight">关务合规与前置校验</h2>
      <p class="mt-1 max-w-3xl text-xs leading-5 text-muted-foreground">韩国 PCCC 和美国地址调用燕文官方校验；IOSS、EORI 仅在真实建单时按燕文官方报文字段映射，不把本地格式结果当成官方通过。</p>
    </div>

    <div class="grid gap-4 xl:grid-cols-[minmax(280px,0.95fr)_minmax(0,1.05fr)]">
      <section class="space-y-4 rounded-[24px] border border-dashed border-border/80 bg-muted/20 p-4">
        <div class="flex items-center justify-between gap-3"><h3 class="text-sm font-black">前置校验项目</h3><ClipboardCheck class="size-4 text-muted-foreground" aria-hidden="true" /></div>
        <div class="grid grid-cols-3 gap-2 rounded-2xl bg-muted p-1" role="group" aria-label="关务校验项目">
          <button v-for="item in [{ key: 'korea', label: '韩国 PCCC' }, { key: 'us', label: '美国地址' }, { key: 'declaration', label: '建单税号映射' }]" :key="item.key" type="button" :class="customsMode === item.key ? 'bg-background text-foreground shadow-sm' : 'text-muted-foreground'" class="min-h-9 rounded-xl px-2 py-1 text-[11px] font-black" @click="selectCustomsMode(item.key as CustomsMode)">{{ item.label }}</button>
        </div>
        <template v-if="customsMode === 'korea'">
          <label class="space-y-1.5"><span class="text-xs font-bold">验证环境</span><select v-model="customsForm.environment" class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"><option value="fat">FAT 测试环境</option><option value="production">PRD 生产环境</option></select></label>
          <label class="space-y-1.5"><span class="text-xs font-bold">个人通关码 PCCC（燕文 taxNumber）</span><Input v-model="customsForm.pccc" placeholder="填写收件人的 PCCC" autocomplete="off" /></label>
          <label class="space-y-1.5"><span class="text-xs font-bold">收件人姓名</span><Input v-model="customsForm.recipientName" placeholder="按燕文官方实名资料填写" autocomplete="name" /></label>
          <div class="grid grid-cols-2 gap-2"><label class="space-y-1.5"><span class="text-[11px] font-bold">收件人联系方式</span><Input v-model="customsForm.phone" placeholder="Phone" autocomplete="tel" /></label><label class="space-y-1.5"><span class="text-[11px] font-bold">5 位邮编</span><Input v-model="customsForm.postalCode" placeholder="例如 03603" inputmode="numeric" autocomplete="postal-code" /></label></div>
        </template>
        <template v-else-if="customsMode === 'us'">
          <label class="space-y-1.5"><span class="text-xs font-bold">验证环境</span><select v-model="customsForm.environment" class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"><option value="fat">FAT 测试环境</option><option value="production">PRD 生产环境</option></select></label>
          <div class="rounded-2xl border border-emerald-500/30 bg-emerald-500/5 px-4 py-3 text-xs leading-5 text-emerald-700 dark:text-emerald-300">燕文官方会返回标准化街道、城市、州以及 4 位和 5 位邮编。</div>
          <label class="space-y-1.5"><span class="text-xs font-bold">街道地址</span><Input v-model="customsForm.address" placeholder="Street address" /></label>
          <div class="grid grid-cols-2 gap-2"><label class="space-y-1.5"><span class="text-[11px] font-bold">城市</span><Input v-model="customsForm.city" placeholder="City" /></label><label class="space-y-1.5"><span class="text-[11px] font-bold">州</span><Input v-model="customsForm.state" placeholder="State" /></label></div>
          <label class="space-y-1.5"><span class="text-xs font-bold">ZIP Code</span><Input v-model="customsForm.postalCode" placeholder="按收件人地址填写" /></label>
        </template>
        <template v-else>
          <div class="rounded-2xl border border-emerald-500/30 bg-emerald-500/5 px-4 py-3 text-xs leading-5 text-emerald-700 dark:text-emerald-300">燕文官方没有独立的 IOSS / EORI 校验方法。创建真实运单时，在同一份 `express.order.create` 报文中按官方字段发送：`receiverInfo.taxNumber`、`parcelInfo.ioss` 和 `importCustomsInfo.eori`。</div>
          <p class="text-xs leading-5 text-muted-foreground">请在“专线运单”TAB 的“创建真实运单”表单填写这些可选字段。此处只说明燕文域的接口边界，不保存通用税号，也不连接其他物流域。</p>
        </template>
        <div class="flex flex-wrap gap-2"><Button :disabled="customsValidating || (customsMode !== 'declaration' && !canShip)" :title="customsMode !== 'declaration' && !canShip ? '需要 logistics:yanwen:ship 权限' : undefined" @click="verifySelectedYanwenCustomsPreflight"><ClipboardCheck class="size-3.5" />{{ customsMode === 'declaration' ? '查看接入状态' : '调用燕文官方校验' }}</Button><Button variant="ghost" size="sm" @click="clearCustomsForm">清空</Button></div>
         <div v-if="customsValidated" :class="customsValidationClass" class="rounded-2xl border px-4 py-3 text-xs leading-5"><span v-if="customsStatusOnly" class="inline-flex items-center gap-1 font-bold"><ClipboardCheck class="size-3.5" />燕文报文映射已接入</span><span v-else-if="customsOfficialPassed" class="inline-flex items-center gap-1 font-bold"><CheckCircle2 class="size-3.5" />官方已通过</span><span v-else class="inline-flex items-center gap-1 font-bold"><AlertTriangle class="size-3.5" />未获得官方通过</span><p class="mt-1">{{ customsValidationMessage }}</p><div v-if="customsOfficialPassed && customsMode === 'us' && customsNormalizedAddress" class="mt-3 space-y-1 rounded-xl border border-emerald-500/20 bg-background/60 p-3"><p class="font-bold">燕文标准化地址</p><p>{{ customsNormalizedAddress.address }}</p><p>{{ customsNormalizedAddress.city }}, {{ customsNormalizedAddress.state }} {{ customsNormalizedAddress.zipCode5 }}-{{ customsNormalizedAddress.zipCode4 }}</p></div></div>
      </section>

      <section class="space-y-3 rounded-[24px] border border-dashed border-border/80 bg-card p-4"><div class="flex items-center gap-3"><span class="flex size-10 items-center justify-center rounded-2xl bg-emerald-500/10 text-emerald-600"><CheckCircle2 class="size-4" /></span><div><h3 class="text-sm font-black">关务接入状态</h3><p class="mt-1 text-[11px] text-muted-foreground">只有燕文官方明确返回通过，页面才会显示“官方已通过”。</p></div></div><div class="space-y-2 text-xs leading-5"><div class="rounded-2xl border border-emerald-500/30 bg-emerald-500/5 p-3"><p class="font-bold">1 · 韩国 PCCC（已接入）</p><p class="mt-1 text-muted-foreground">后端代理 `common.verify.kr.pccc`，提交姓名、电话、taxNumber 和 5 位 zipCode；浏览器不直连燕文。</p></div><div class="rounded-2xl border border-emerald-500/30 bg-emerald-500/5 p-3"><p class="font-bold">2 · 美国地址（已接入）</p><p class="mt-1 text-muted-foreground">后端代理 `common.verify.us.address`，返回标准化地址、城市、州、zipCode4 和 zipCode5；浏览器不直连燕文。</p></div><div class="rounded-2xl border border-emerald-500/30 bg-emerald-500/5 p-3"><p class="font-bold">3 · IOSS / EORI 建单映射（已接入）</p><p class="mt-1 text-muted-foreground">真实 `express.order.create` 请求按官方契约发送 `receiverInfo.taxNumber`、`parcelInfo.ioss` 和 `importCustomsInfo.eori`；燕文没有独立税号校验接口。</p></div></div></section>
    </div>
  </section>
</template>
