<template>
  <section class="rounded-2xl border bg-muted/30 p-4">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div class="flex items-start gap-3">
        <div class="flex size-8 items-center justify-center rounded-lg border bg-background">
          <LockKeyhole class="size-3.5 text-orange-500" />
        </div>
        <div>
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">{{ t('payment.gatewayConnection') }}</p>
          <h3 class="mt-1 text-sm font-black tracking-tight text-foreground">{{ t('payment.enterCredentials') }}</h3>
          <p class="mt-1 text-xs leading-relaxed text-muted-foreground">
            {{ t('payment.credentialDescription') }}
          </p>
        </div>
      </div>
      <span
        class="rounded-full border px-2.5 py-1 text-[11px] font-black"
 :class="secretStoreConfigured ? 'border-emerald-500/20 bg-emerald-500/10 text-emerald-700 dark:text-emerald-200': 'border-rose-500/20 bg-rose-500/10 text-rose-700 dark:text-rose-200'"
      >
        {{ secretStoreConfigured ? t('payment.masterKeyReady') : t('payment.missingMasterKey') }}
      </span>
    </div>

    <div v-if="!selectedGateway" class="mt-4 rounded-xl border border-dashed p-5 text-sm text-muted-foreground">
      {{ t('payment.selectGatewayFirst') }}
    </div>

    <div v-else-if="!canEdit" class="mt-4 rounded-xl border border-dashed p-5 text-sm text-muted-foreground">
      {{ t('payment.noEditPermission') }}
    </div>

    <form v-else class="mt-4 space-y-4" @submit.prevent="saveConfig">
      <div v-if="!secretStoreConfigured" class="rounded-xl border border-rose-500/20 bg-rose-500/10 p-3">
        <p class="text-sm font-black text-rose-700 dark:text-rose-100">{{ t('payment.canFillCannotSave') }}</p>
        <p class="mt-1 text-xs leading-relaxed text-rose-700/80 dark:text-rose-100/75">
          {{ t('payment.masterKeyDescription') }}
        </p>
        <div class="mt-3 flex flex-wrap gap-2">
          <Button type="button" size="sm" variant="outline" @click="generateMasterKey">
            <KeyRound class="size-3.5" />
            {{ t('payment.generateMasterKey') }}
          </Button>
        </div>

        <div v-if="generatedMasterKey" class="mt-3 rounded-xl border border-rose-500/20 bg-background/70 p-3">
          <div class="flex flex-wrap items-center justify-between gap-2">
            <p class="text-xs font-black text-foreground">{{ t('payment.generatedKey') }}</p>
            <p class="text-[11px] text-muted-foreground">{{ t('payment.keyBrowserOnly') }}</p>
          </div>
          <div class="mt-2 flex min-w-0 items-center gap-2">
            <Input
              :model-value="generatedMasterKey"
              :type="masterKeyVisible ? 'text' : 'password'"
              readonly
              class="min-w-0 flex-1 font-mono text-[11px]"
            />
            <Button
              type="button"
              variant="outline"
              size="icon"
              class="size-9 flex-none"
              :aria-label="masterKeyVisible ? '隐藏 MASTER KEY' : '显示 MASTER KEY'"
              @click="masterKeyVisible = !masterKeyVisible"
            >
              <EyeOff v-if="masterKeyVisible" class="size-3.5" />
              <Eye v-else class="size-3.5" />
            </Button>
            <Button
              type="button"
              variant="outline"
              size="icon"
              class="size-9 flex-none"
              :aria-label="t('payment.copyMasterKey')"
              @click="copyGeneratedMasterKey"
            >
              <Copy class="size-3.5" />
            </Button>
          </div>
          <p class="mt-2 text-[11px] leading-relaxed text-muted-foreground">
            {{ t('payment.envLine') }}
          </p>
          <code class="mt-1 block break-all rounded-lg border bg-muted/50 p-2 font-mono text-[11px] text-foreground">
            PAYMENT_CONFIG_MASTER_KEY={{ generatedMasterKey }}
          </code>
          <p class="mt-2 text-[11px] leading-relaxed text-muted-foreground">
            {{ t('payment.restartBackend') }}
          </p>
        </div>
      </div>

      <div
        v-if="selectedGateway === 'paypal'"
        class="rounded-xl border border-sky-500/20 bg-sky-500/10 p-3 text-xs leading-relaxed text-sky-900 dark:text-sky-100"
      >
          {{ t('payment.paypalHostedCheckout') }}
      </div>
      <div class="grid gap-4 md:grid-cols-2">
        <AdminFormField
          :label="t('payment.integrationMode')"
          :description="t('payment.integrationModeDescription')"
        >
          <div class="flex h-10 items-center justify-between rounded-md border border-input bg-muted/40 px-3 text-sm">
            <span class="font-black text-foreground">
              {{ form.environment === 'sandbox' ? t('payment.sandbox') : t('payment.production') }}
            </span>
            <span class="text-xs text-muted-foreground">
              {{ form.environment === 'sandbox' ? t('payment.testCredentials') : t('payment.liveCredentials') }}
            </span>
          </div>
        </AdminFormField>

        <div class="rounded-xl border bg-background/70 p-3">
          <p class="text-[11px] font-black uppercase tracking-widest text-muted-foreground/60">{{ t('payment.credentialSource') }}</p>
          <p class="mt-1 text-sm font-black text-foreground">
            {{ credentialSourceLabel(status?.runtime_source) }}
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
          <select
            v-else-if="field.key === 'three_ds_mode'"
            v-model="form.credentials[field.key]"
            class="h-10 w-full rounded-md border border-input bg-background px-3 text-sm"
          >
             <option value="automatic">自动判断（automatic）：可能不触发 3DS，由 Stripe / 风控决定</option>
             <option value="any">要求 3DS（any）：可能无感完成，也可能进入银行挑战</option>
             <option value="challenge">更强挑战（challenge）：请求银行进入挑战路径，最终交互仍由发卡行决定</option>
           </select>
          <div v-if="field.key === 'three_ds_mode'" class="mt-3 grid gap-2 sm:grid-cols-3">
            <div
              v-for="mode in threeDSModeCards"
              :key="mode.value"
              class="border p-3"
 :class="normalizedThreeDSMode === mode.value ? 'border-primary/50 bg-primary/5': 'border-border/70'"
            >
              <div class="flex items-start justify-between gap-2">
                <p class="text-xs font-black">{{ mode.label }}</p>
                <code class="font-mono text-[10px] text-muted-foreground">{{ mode.value }}</code>
              </div>
              <p class="mt-1 text-[11px] leading-5 text-muted-foreground">{{ mode.description }}</p>
            </div>
          </div>
          <div v-if="field.key === 'three_ds_mode'" class="mt-2 border border-sky-500/20 bg-sky-500/5 p-3 text-[11px] leading-5 text-sky-900 dark:text-sky-100">
            <p class="font-black">当前选择的实际含义：{{ selectedThreeDSModeMeta.label }}</p>
            <p class="mt-1">{{ selectedThreeDSModeMeta.outcome }}</p>
            <p class="mt-1">这是 PaymentIntent 的基础请求值；风控自适应升级、30 天组合风险和人工保护只能把认证强度提高到 any 或 challenge，不会把高风险支付降级。</p>
          </div>
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

      <div v-if="form.environment === 'production'" class="rounded-xl border border-amber-500/20 bg-amber-500/10 p-3">
        <AdminFormField
          :label="t('payment.productionConfirmation')"
          :description="t('payment.productionConfirmationDescription')"
        >
          <Input
            v-model.trim="productionConfirmation"
            class="font-mono"
            autocomplete="off"
            placeholder="PRODUCTION"
          />
        </AdminFormField>
      </div>

      <div v-if="status?.admin_config_configured" class="rounded-xl border bg-background/70 p-3">
        <AdminFormField
          :label="t('payment.clearConfigConfirmation')"
          :description="`${t('payment.clearConfigConfirmation')} ${expectedDeleteConfirmation}`"
        >
          <Input
            v-model.trim="deleteConfirmation"
            class="font-mono"
            autocomplete="off"
            :placeholder="expectedDeleteConfirmation"
          />
        </AdminFormField>
      </div>

      <div class="flex flex-wrap items-center justify-between gap-3">
        <p class="text-xs text-muted-foreground">
          {{ t('payment.configuredFields', { fields: configuredFieldText }) }}
        </p>
        <div class="flex flex-wrap gap-2">
          <Button
            type="button"
            variant="outline"
            :disabled="clearing || saving || !canClearConfig"
            @click="clearConfig"
          >
            <LoaderCircle v-if="clearing" class="size-3.5 animate-spin" />
            <Trash2 v-else class="size-3.5" />
            {{ t('payment.clearEncryptedConfig') }}
          </Button>
          <Button type="submit" :disabled="saving || !secretStoreConfigured">
            <LoaderCircle v-if="saving" class="size-3.5 animate-spin" />
            <Save v-else class="size-3.5" />
            {{ saving ? t('common.saving') : secretStoreConfigured ? t('payment.saveEncryptedConfig') : t('payment.waitForMasterKey') }}
          </Button>
        </div>
      </div>
    </form>
  </section>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import { Copy, Eye, EyeOff, KeyRound, LoaderCircle, LockKeyhole, Save, Trash2 } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import axios from '@/utils/axios'
import { useAdminI18n } from '@/i18n'
import type { PaymentGatewayCredentialField, PaymentGatewayRuntimeStatus } from '@/modules/settings/types'

const props = withDefaults(defineProps<{
  selectedGateway?: string
  status?: PaymentGatewayRuntimeStatus | null
  canEdit?: boolean
  secretStoreConfigured?: boolean
}>(), {
  selectedGateway: '',
  status: null,
  canEdit: false,
  secretStoreConfigured: false,
})

const emit = defineEmits<{
  (event: 'saved'): void
}>()
const { t } = useAdminI18n()

const fieldDefinitions: Record<string, PaymentGatewayCredentialField[]> = {
  stripe: [
    { key: 'api_key', label: 'API Key / Secret Key', placeholder: 'sk_live_...', description: 'Stripe 后端 Secret Key。' },
    { key: 'publishable_key', label: 'Publishable Key', placeholder: 'pk_live_...', description: '仅返回浏览器初始化 Stripe.js，不能填写 Secret Key。' },
    { key: 'webhook_secret', label: 'Webhook Secret', placeholder: 'whsec_...', description: 'Stripe webhook endpoint secret。' },
     { key: 'three_ds_mode', label: 'Stripe 基础 3DS 认证策略', placeholder: 'automatic', description: '这是每笔 PaymentIntent 的起点，不是最终结果。自适应风控和人工保护只能把认证强度提升到 any 或 challenge，不会降级。' },
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
    { key: 'api_v3_key', label: 'API v3 Key', placeholder: 'APIv3 密钥', description: '用于回调 resource 解密。' },
    { key: 'platform_certificate', label: 'Platform Certificate', placeholder: '微信支付平台证书 PEM；与平台公钥二选一', description: '用于 API v3 回调验签。填写平台证书时可不填平台公钥。', multiline: true },
    { key: 'platform_public_key', label: 'Platform Public Key', placeholder: '微信支付平台公钥 PEM；与平台证书二选一', description: '用于 API v3 回调验签。填写平台公钥时必须同时填写 Platform Public Key ID。', multiline: true },
    { key: 'platform_public_key_id', label: 'Platform Public Key ID', placeholder: 'PUB_KEY_ID_...', description: '仅在使用微信支付平台公钥时必填。' },
  ],
}

const form = reactive<{
  environment: string
  credentials: Record<string, string>
}>({
  environment: 'sandbox',
  credentials: {},
})
const saving = ref(false)
const clearing = ref(false)
const productionConfirmation = ref('')
const deleteConfirmation = ref('')
const generatedMasterKey = ref('')
const masterKeyVisible = ref(false)

const providerFields = computed(() => fieldDefinitions[props.selectedGateway] || [])
const expectedDeleteConfirmation = computed(() => `DELETE ${String(props.selectedGateway || '').toUpperCase()}`)
const canClearConfig = computed(() => (
  props.status?.admin_config_configured === true &&
  deleteConfirmation.value === expectedDeleteConfirmation.value
))
const configuredFieldText = computed(() => {
  const fields = props.status?.configured_fields || []
  return fields.length ? fields.join(', ') : t('common.none')
})
const threeDSModeCards = [
  {
    value: 'automatic',
    label: '自动判断',
    description: 'Stripe 根据 SCA / 风险决定是否需要认证。',
    outcome: '可能直接完成、无感认证，或被运行时风控升级。',
  },
  {
    value: 'any',
    label: '要求 3DS',
    description: '请求在支持的情况下使用 3DS。',
    outcome: '可以无感完成，也可能进入银行 challenge，不保证弹窗。',
  },
  {
    value: 'challenge',
    label: '更强挑战',
    description: '请求发卡行进入更强的挑战路径。',
    outcome: '更可能出现银行挑战，但最终仍由发卡行决定。',
  },
]
const normalizedThreeDSMode = computed(() => {
  const value = String(form.credentials.three_ds_mode || '').trim().toLowerCase()
  return ['any', 'challenge'].includes(value) ? value : 'automatic'
})
const selectedThreeDSModeMeta = computed(() => (
  threeDSModeCards.find((mode) => mode.value === normalizedThreeDSMode.value) || threeDSModeCards[0]
))

const generateMasterKey = (): void => {
  if (!globalThis.crypto?.getRandomValues) {
    toast.error(t('payment.secureRandomUnsupported'))
    return
  }

  const bytes = new Uint8Array(32)
  globalThis.crypto.getRandomValues(bytes)
  generatedMasterKey.value = Array.from(bytes, (byte) => byte.toString(16).padStart(2, '0')).join('')
  masterKeyVisible.value = true
  toast.success(t('payment.masterKeyGenerated'))
}

const copyGeneratedMasterKey = async (): Promise<void> => {
  if (!generatedMasterKey.value || !navigator.clipboard) {
    toast.error(t('payment.copyMasterKeyUnsupported'))
    return
  }

  try {
    await navigator.clipboard.writeText(`PAYMENT_CONFIG_MASTER_KEY=${generatedMasterKey.value}`)
    toast.success(t('payment.masterKeyCopied'))
  } catch (error) {
    console.error('Failed to copy payment config master key:', error)
    toast.error(t('payment.copyFailed'))
  }
}

const resetCredentialInputs = (): void => {
  const next: Record<string, string> = {}
  providerFields.value.forEach((field) => {
    next[field.key] = ''
  })
  form.credentials = next
}

const credentialPayload = (): Record<string, string> => {
  const payload: Record<string, string> = {}
  providerFields.value.forEach((field) => {
    const value = String(form.credentials[field.key] || '').trim()
    if (value) payload[field.key] = value
  })
  return payload
}

const saveConfig = async () => {
  if (!props.selectedGateway) return
  if (!props.secretStoreConfigured) {
    toast.error(t('payment.masterKeyRequired'))
    return
  }
  if (form.environment === 'production' && productionConfirmation.value !== 'PRODUCTION') {
    toast.error(t('payment.saveProductionConfirmation'))
    return
  }
  const credentials = credentialPayload()
  if (!Object.keys(credentials).length && !props.status?.admin_config_configured) {
    toast.error(t('payment.atLeastOneCredential'))
    return
  }
  if (
    props.status?.admin_config_configured &&
    props.status.environment &&
    props.status.environment !== form.environment &&
    !Object.keys(credentials).length
  ) {
    toast.error(t('payment.switchCredentialMode'))
    return
  }

  saving.value = true
  try {
    await axios.put(`/api/admin/settings/payment-gateways/${props.selectedGateway}`, {
      environment: form.environment,
      credentials,
      confirmation: form.environment === 'production' ? productionConfirmation.value : '',
    })
    toast.success(t('payment.encryptedConfigSaved'))
    productionConfirmation.value = ''
    resetCredentialInputs()
    emit('saved')
  } catch (error) {
    console.error('Failed to save payment gateway config:', error)
  } finally {
    saving.value = false
  }
}

const clearConfig = async () => {
  if (!props.selectedGateway) return
  if (!canClearConfig.value) {
    toast.error(t('payment.clearConfirmationRequired', { confirmation: expectedDeleteConfirmation.value }))
    return
  }
  clearing.value = true
  try {
    await axios.delete(`/api/admin/settings/payment-gateways/${props.selectedGateway}`, {
      data: { confirmation: deleteConfirmation.value },
    })
    toast.success(t('payment.encryptedConfigCleared'))
    deleteConfirmation.value = ''
    resetCredentialInputs()
    emit('saved')
  } catch (error) {
    console.error('Failed to clear payment gateway config:', error)
  } finally {
    clearing.value = false
  }
}

watch(() => props.selectedGateway, () => {
  form.environment = props.status?.environment || 'production'
  productionConfirmation.value = ''
  deleteConfirmation.value = ''
  resetCredentialInputs()
}, { immediate: true })

watch(() => props.status?.environment, (environment) => {
  if (environment) form.environment = environment
})

const credentialSourceLabel = (source?: string): string => {
  if (source === 'admin-encrypted') return t('payment.adminEncrypted')
  if (source === 'environment') return t('payment.serverEnvironment')
  return source || t('payment.credentialNotConfigured')
}
</script>
