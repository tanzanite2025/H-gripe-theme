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
        :class="secretStoreConfigured ? 'border-emerald-500/20 bg-emerald-500/10 text-emerald-700 dark:text-emerald-200' : 'border-rose-500/20 bg-rose-500/10 text-rose-700 dark:text-rose-200'"
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
import type { PaymentGatewayCredentialField, PaymentGatewayRuntimeStatus } from './settingsTypes'

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
    { key: 'three_ds_mode', label: '3DS Mode', placeholder: 'automatic / any / challenge', description: '默认 automatic；需要强制 3DS 时填写 any，高风险可填写 challenge。' },
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
