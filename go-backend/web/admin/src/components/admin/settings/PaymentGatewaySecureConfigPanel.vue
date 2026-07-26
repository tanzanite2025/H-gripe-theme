<template>
  <section class="rounded-2xl border bg-card/75 p-4 shadow-sm">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div class="flex items-start gap-3">
        <div class="flex size-9 items-center justify-center rounded-full border bg-background">
          <LockKeyhole class="size-4 text-orange-500" />
        </div>
        <div>
          <p class="text-xs font-black uppercase tracking-widest text-muted-foreground/60">Encrypted Secret Store</p>
          <h3 class="mt-1 text-base font-black text-foreground">安全凭据写入</h3>
          <p class="mt-1 text-xs leading-relaxed text-muted-foreground">
            只写入，不回显。留空字段会保留现有加密值。
          </p>
        </div>
      </div>
      <span
        class="rounded-full border px-2.5 py-1 text-[11px] font-black"
        :class="secretStoreConfigured ? 'border-emerald-500/20 bg-emerald-500/10 text-emerald-700 dark:text-emerald-200' : 'border-rose-500/20 bg-rose-500/10 text-rose-700 dark:text-rose-200'"
      >
        {{ secretStoreConfigured ? 'MASTER KEY READY' : '缺 MASTER KEY' }}
      </span>
    </div>

    <div v-if="!selectedGateway" class="mt-4 rounded-xl border border-dashed p-5 text-sm text-muted-foreground">
      先选择一个支付服务商。
    </div>

    <div v-else-if="!canEdit" class="mt-4 rounded-xl border border-dashed p-5 text-sm text-muted-foreground">
      当前账号没有 settings:edit 权限，无法写入支付凭据。
    </div>

    <div v-else-if="!secretStoreConfigured" class="mt-4 rounded-xl border border-rose-500/20 bg-rose-500/10 p-3">
      <p class="text-sm font-black text-rose-700 dark:text-rose-100">需要先配置服务端 master key</p>
      <p class="mt-1 text-xs leading-relaxed text-rose-700/80 dark:text-rose-100/75">
        在后端环境变量设置 <span class="font-mono font-black">PAYMENT_CONFIG_MASTER_KEY</span> 后，后台才允许保存生产支付密钥。
      </p>
    </div>

    <form v-else class="mt-4 space-y-4" @submit.prevent="saveConfig">
      <div class="grid gap-4 md:grid-cols-2">
        <AdminFormField label="运行环境">
          <select
            v-model="form.environment"
            class="h-10 w-full rounded-md border border-input bg-background px-3 text-sm font-medium outline-none transition-colors focus-visible:ring-2 focus-visible:ring-ring"
          >
            <option value="sandbox">Sandbox / Test</option>
            <option value="production">Production / Live</option>
          </select>
        </AdminFormField>

        <div class="rounded-xl border bg-background/70 p-3">
          <p class="text-[11px] font-black uppercase tracking-widest text-muted-foreground/60">当前来源</p>
          <p class="mt-1 text-sm font-black text-foreground">
            {{ status?.runtime_source || 'environment' }}
          </p>
        </div>

        <AdminFormField
          v-for="field in providerFields"
          :key="field.key"
          :label="field.label"
          :description="field.description"
          class="md:col-span-2"
        >
          <Textarea
            v-if="field.multiline"
            v-model="form.credentials[field.key]"
            class="min-h-24 font-mono"
            autocomplete="new-password"
            :placeholder="field.placeholder"
          />
          <Input
            v-else
            v-model="form.credentials[field.key]"
            type="password"
            class="font-mono"
            autocomplete="new-password"
            :placeholder="field.placeholder"
          />
        </AdminFormField>
      </div>

      <div class="flex flex-wrap items-center justify-between gap-3">
        <p class="text-xs text-muted-foreground">
          已配置字段：{{ configuredFieldText }}
        </p>
        <div class="flex flex-wrap gap-2">
          <Button
            type="button"
            variant="outline"
            :disabled="clearing || saving || !status?.admin_config_configured"
            @click="clearConfig"
          >
            <LoaderCircle v-if="clearing" class="size-3.5 animate-spin" />
            <Trash2 v-else class="size-3.5" />
            清空加密配置
          </Button>
          <Button type="submit" :disabled="saving">
            <LoaderCircle v-if="saving" class="size-3.5 animate-spin" />
            <Save v-else class="size-3.5" />
            {{ saving ? '保存中' : '保存加密配置' }}
          </Button>
        </div>
      </div>
    </form>
  </section>
</template>

<script setup>
import { computed, reactive, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import { LoaderCircle, LockKeyhole, Save, Trash2 } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import axios from '@/utils/axios'

const props = defineProps({
  selectedGateway: { type: String, default: '' },
  status: { type: Object, default: null },
  canEdit: { type: Boolean, default: false },
  secretStoreConfigured: { type: Boolean, default: false },
})

const emit = defineEmits(['saved'])

const fieldDefinitions = {
  stripe: [
    { key: 'api_key', label: 'API Key / Secret Key', placeholder: 'sk_live_...', description: 'Stripe 后端 Secret Key。' },
    { key: 'webhook_secret', label: 'Webhook Secret', placeholder: 'whsec_...', description: 'Stripe webhook endpoint secret。' },
  ],
  paypal: [
    { key: 'client_id', label: 'Client ID', placeholder: 'PayPal client id' },
    { key: 'secret', label: 'Secret', placeholder: 'PayPal secret' },
    { key: 'webhook_id', label: 'Webhook ID', placeholder: 'Webhook id from PayPal developer portal' },
  ],
  alipay: [
    { key: 'app_id', label: 'App ID', placeholder: 'Alipay app id' },
    { key: 'private_key', label: 'Private Key', placeholder: 'Alipay merchant private key', multiline: true },
    { key: 'public_key', label: 'Alipay Public Key', placeholder: 'Alipay public key for notification verify', multiline: true },
  ],
  wechat: [
    { key: 'mch_id', label: 'Merchant ID', placeholder: '微信支付商户号' },
    { key: 'app_id', label: 'App ID', placeholder: '微信 App ID' },
    { key: 'private_key_path', label: 'Private Key Path', placeholder: '服务端私钥文件路径' },
    { key: 'merchant_serial', label: 'Merchant Serial', placeholder: '商户证书序列号' },
    { key: 'api_v3_key', label: 'API v3 Key', placeholder: 'APIv3 密钥' },
  ],
}

const form = reactive({
  environment: 'sandbox',
  credentials: {},
})
const saving = ref(false)
const clearing = ref(false)

const providerFields = computed(() => fieldDefinitions[props.selectedGateway] || [])
const configuredFieldText = computed(() => {
  const fields = props.status?.configured_fields || []
  return fields.length ? fields.join(', ') : '暂无'
})

const resetCredentialInputs = () => {
  const next = {}
  providerFields.value.forEach((field) => {
    next[field.key] = ''
  })
  form.credentials = next
}

const credentialPayload = () => {
  const payload = {}
  providerFields.value.forEach((field) => {
    const value = String(form.credentials[field.key] || '').trim()
    if (value) payload[field.key] = value
  })
  return payload
}

const saveConfig = async () => {
  if (!props.selectedGateway) return
  const credentials = credentialPayload()
  if (!Object.keys(credentials).length && !props.status?.admin_config_configured) {
    toast.error('至少填写一个支付凭据字段')
    return
  }

  saving.value = true
  try {
    await axios.put(`/api/admin/settings/payment-gateways/${props.selectedGateway}`, {
      environment: form.environment,
      credentials,
    })
    toast.success('已保存加密支付配置')
    resetCredentialInputs()
    emit('saved')
  } catch (error) {
    console.error('Failed to save payment gateway config:', error)
  } finally {
    saving.value = false
  }
}

const clearConfig = async () => {
  if (!props.selectedGateway || !window.confirm('清空该支付服务商的加密配置？')) return
  clearing.value = true
  try {
    await axios.delete(`/api/admin/settings/payment-gateways/${props.selectedGateway}`)
    toast.success('已清空加密支付配置')
    resetCredentialInputs()
    emit('saved')
  } catch (error) {
    console.error('Failed to clear payment gateway config:', error)
  } finally {
    clearing.value = false
  }
}

watch(() => props.selectedGateway, () => {
  form.environment = props.status?.environment || 'sandbox'
  resetCredentialInputs()
}, { immediate: true })

watch(() => props.status?.environment, (environment) => {
  if (environment) form.environment = environment
})
</script>
