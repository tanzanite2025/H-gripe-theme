<template>
  <div class="space-y-4">
    <AdminPageHeader :title="pageTitle" :description="pageDescription">
      <template #actions>
        <Button
          v-if="hasPermission('settings:edit') && !selfSavingTabs.has(activeTab)"
          :disabled="saving || loadingSettings"
          @click="saveSettings"
        >
          <LoaderCircle v-if="saving" class="size-4 animate-spin" />
          <Save v-else class="size-4" />
          {{ saving ? '保存中' : '保存设置' }}
        </Button>
      </template>
    </AdminPageHeader>

    <div class="relative min-h-96">
      <div v-if="loadingSettings" class="absolute inset-0 z-10 flex items-center justify-center bg-background/75">
        <LoaderCircle class="size-5 animate-spin text-primary" aria-label="正在加载设置" />
      </div>

      <SettingsTabsPanel
        :active-tab="activeTab"
        v-model:show-smtp-password="showSmtpPassword"
        v-model:show-payment-secrets="showPaymentSecrets"
        :site-settings="siteSettings"
        :email-settings="emailSettings"
        :social-settings="socialSettings"
        :payment-settings="paymentSettings"
        :api-settings="apiSettings"
        :primary-pricing-currency="primaryPricingCurrency"
        :commercial-crawler-protection="commercialCrawlerProtection"
        :loading-commercial-crawler-protection="loadingCommercialCrawlerProtection"
        :uploading-site-logo="uploadingSiteLogo"
        :payment-runtime="paymentRuntime"
        :loading-payment-runtime="loadingPaymentRuntime"
        :syncing-exchange-rates="syncingExchangeRates"
        :saving-api-settings="saving && activeTab === 'api'"
        :social-fields="socialFields"
        :loading-public-chat-agents="loadingPublicChatAgents"
        :loading-public-chat-groups="loadingPublicChatGroups"
        :loading-public-chat-agent-candidates="loadingPublicChatAgentCandidates"
        :public-chat-agents-summary="publicChatAgentsSummary"
        :public-chat-agents="publicChatAgents"
        :public-chat-groups="publicChatGroups"
        :public-chat-agent-warnings="publicChatAgentWarnings"
        :can-edit="hasPermission('settings:edit')"
        @open-agent-dialog="openPublicChatAgentDialog"
        @open-group-dialog="openPublicChatGroupDialog"
        @edit-group="editPublicChatGroup"
        @delete-group="deletePublicChatGroup"
        @refresh-public-chat="refreshPublicChat"
        @refresh-payment-runtime="fetchPaymentRuntime"
        @sync-exchange-rates="syncExchangeRates"
        @currency-policy-saved="handleCurrencyPolicySaved"
        @refresh-commercial-crawler-protection="fetchCommercialCrawlerProtection"
        @upload-site-logo="uploadSiteLogo"
      />
    </div>

    <PublicChatAgentDialog
      v-model:open="publicChatAgentDialogOpen"
      :form="publicChatAgentForm"
      :candidates="publicChatAgentCandidates"
      :selected-candidate="selectedPublicChatAgentCandidate"
      :groups="publicChatGroups.filter((group) => group.status === 'active')"
      :loading-candidates="loadingPublicChatAgentCandidates"
      :saving="publicChatAgentSaving"
      @save="savePublicChatAgent"
    />

    <PublicChatGroupDialog
      v-model:open="publicChatGroupDialogOpen"
      :form="publicChatGroupForm"
      :saving="publicChatGroupSaving"
      @save="savePublicChatGroup"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import { LoaderCircle, Save } from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import PublicChatAgentDialog from '@/components/admin/settings/PublicChatAgentDialog.vue'
import PublicChatGroupDialog from '@/components/admin/settings/PublicChatGroupDialog.vue'
import SettingsTabsPanel from '@/components/admin/settings/SettingsTabsPanel.vue'
import { Button } from '@/components/ui/button'
import { useRouteTab } from '@/composables/useRouteTab'
import mediaApi from '@/api/media'
import { assetAccessURL } from '@/lib/mediaPresentation'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

const authStore = useAuthStore()
const DAILY_API_REFRESH_MINUTES = 1440
const EXCHANGE_RATE_PROVIDER = 'ExchangeRate-API'
const DEFAULT_PRICING_CURRENCY = 'USD'
const EXCHANGE_RATE_ENDPOINT = 'https://v6.exchangerate-api.com/v6/{apiKey}/latest/{base}'
const activeTab = useRouteTab({
  defaultValue: 'site',
  values: ['site', 'email', 'social', 'currency', 'markets', 'payment', 'api', 'commercial_crawler', 'public_chat'],
  routes: {
    site: 'SettingsSite',
    email: 'SettingsEmail',
    social: 'SettingsSocial',
    currency: ['SettingsCurrency', 'PaymentCurrency'],
    markets: 'SettingsMarkets',
    payment: 'PaymentSettings',
    api: 'SettingsApi',
    commercial_crawler: 'SettingsCommercialCrawler',
    public_chat: 'SupportPublicChat',
  },
})

type SettingValueType = 'string' | 'number' | 'boolean'

interface SettingFieldDefinition {
  type: SettingValueType
  public: boolean
  description: string
}

interface SettingsGroupDefinition {
  target: Record<string, any>
  fields: Record<string, SettingFieldDefinition>
}

const pageTitle = computed(() => (
  ['currency', 'payment'].includes(activeTab.value)
    ? activeTab.value === 'currency' ? '系统设置' : '支付管理'
    : activeTab.value === 'public_chat'
      ? 'Public Chat'
      : '系统设置'
))
const pageDescription = computed(() => {
  if (activeTab.value === 'currency') return '管理主基准币种、次展示币种和汇率缓存联动'
  if (activeTab.value === 'markets') return '管理国家、市场、语言与展示币种解析规则'
  if (activeTab.value === 'payment') return '管理支付网关、支付方式与运行状态'
  if (activeTab.value === 'public_chat') return '管理公开客服、客服组与前台聊天配置'
  return '管理站点、邮件、社交、API、安全与客服配置'
})
const saving = ref(false)
const loadingSettings = ref(false)
const uploadingSiteLogo = ref(false)
const showSmtpPassword = ref(false)
const showPaymentSecrets = ref(false)
const loadedGroups = new Set()
const selfSavingTabs = new Set(['currency', 'markets', 'api', 'commercial_crawler', 'public_chat'])

const siteSettings = reactive({
  site_name: '',
  brand_title: '',
  site_description: '',
  site_logo: '',
  contact_email: '',
  contact_phone: '',
  copyright_holder: '',
  copyright_notice: '',
  copyright_url: '',
  admin_brand_name: '',
  admin_brand_initial: '',
  admin_panel_label: '',
  admin_login_title: '',
  admin_footer_text: '',
  admin_html_title: ''
})
const emailSettings = reactive({ smtp_host: '', smtp_port: 587, smtp_username: '', smtp_password: '', from_email: '', from_name: '' })
const socialSettings = reactive({ facebook: '', twitter: '', instagram: '', linkedin: '', youtube: '', wechat: '' })
const paymentSettings = reactive({ gateway: 'stripe', test_mode: true })
const apiSettings = reactive({
  exchange_rate_enabled: false,
  exchange_rate_provider: EXCHANGE_RATE_PROVIDER,
  exchange_rate_endpoint: EXCHANGE_RATE_ENDPOINT,
  exchange_rate_query_template: '',
  exchange_rate_refresh_minutes: DAILY_API_REFRESH_MINUTES,
  exchange_rate_api_key: '',
  time_api_enabled: false,
  time_api_provider: '',
  time_api_endpoint: '',
  time_api_query_template: 'timezone={timezone}',
  time_api_default_timezone: 'Asia/Shanghai',
  time_api_refresh_minutes: DAILY_API_REFRESH_MINUTES,
  time_api_key_ref: ''
})
const paymentRuntime = ref(null)
const loadingPaymentRuntime = ref(false)
const syncingExchangeRates = ref(false)
const primaryPricingCurrency = ref(DEFAULT_PRICING_CURRENCY)
const currencyPolicyLoaded = ref(false)
const commercialCrawlerProtection = ref(null)
const loadingCommercialCrawlerProtection = ref(false)

const loadingPublicChatAgents = ref(false)
const publicChatAgentsOverview = ref(null)
const publicChatAgentsSummary = computed(() => publicChatAgentsOverview.value?.summary || {})
const publicChatAgents = computed(() => publicChatAgentsOverview.value?.agents || [])
const publicChatAgentWarnings = computed(() => publicChatAgentsOverview.value?.warnings || [])
const loadingPublicChatGroups = ref(false)
const publicChatGroups = ref([])
const loadingPublicChatAgentCandidates = ref(false)
const publicChatAgentCandidates = ref([])
const publicChatAgentDialogOpen = ref(false)
const publicChatAgentSaving = ref(false)
const publicChatGroupDialogOpen = ref(false)
const publicChatGroupSaving = ref(false)
const publicChatAgentForm = reactive({
  user_id: '',
  agent_id: '',
  name: '',
  email: '',
  avatar: '',
  whatsapp: '',
  status: 'active',
  online_status: 'offline',
  group_ids: []
})
const publicChatGroupForm = reactive({
  id: 0,
  code: '',
  name: '',
  description: '',
  status: 'active',
  sort_order: 0
})
const selectedPublicChatAgentCandidate = computed(() =>
  publicChatAgentCandidates.value.find((candidate) => String(candidate.user_id) === String(publicChatAgentForm.user_id))
)

const socialFields = [
  { key: 'facebook', label: 'Facebook', placeholder: 'Facebook 页面 URL' },
  { key: 'twitter', label: 'Twitter / X', placeholder: '账号 URL' },
  { key: 'instagram', label: 'Instagram', placeholder: '账号 URL' },
  { key: 'linkedin', label: 'LinkedIn', placeholder: '页面 URL' },
  { key: 'youtube', label: 'YouTube', placeholder: '频道 URL' },
  { key: 'wechat', label: '微信', placeholder: '二维码 URL' }
]

const normalizeCurrencyCode = (currency) => String(currency || '').trim().toUpperCase()
const validCurrencyCodeOrDefault = (currency) => {
  const normalized = normalizeCurrencyCode(currency)
  return /^[A-Z]{3}$/.test(normalized) ? normalized : DEFAULT_PRICING_CURRENCY
}

const fetchCurrencyPolicyForSettings = async (force = false) => {
  if (!force && currencyPolicyLoaded.value) return
  try {
    const response = await axios.get('/api/admin/settings/currency-policy')
    const policy = response.data?.policy || {}
    primaryPricingCurrency.value = validCurrencyCodeOrDefault(policy.primary_currency)
    currencyPolicyLoaded.value = true
  } catch (error) {
    console.error('Failed to fetch currency policy:', error)
    primaryPricingCurrency.value = DEFAULT_PRICING_CURRENCY
  }
}

const handleCurrencyPolicySaved = async (policy = null) => {
  currencyPolicyLoaded.value = false
  if (policy) {
    primaryPricingCurrency.value = validCurrencyCodeOrDefault(policy.primary_currency)
    currencyPolicyLoaded.value = true
  } else {
    await fetchCurrencyPolicyForSettings(true)
  }
}

const applyExchangeRatePreset = () => {
  apiSettings.exchange_rate_provider = EXCHANGE_RATE_PROVIDER
  apiSettings.exchange_rate_endpoint = EXCHANGE_RATE_ENDPOINT
  apiSettings.exchange_rate_query_template = ''
  apiSettings.exchange_rate_refresh_minutes = DAILY_API_REFRESH_MINUTES
}

const applyAPIRefreshDefaults = () => {
  applyExchangeRatePreset()
  apiSettings.time_api_refresh_minutes = DAILY_API_REFRESH_MINUTES
}

const groupDefinitions: Record<string, SettingsGroupDefinition> = {
  site: {
    target: siteSettings,
    fields: {
      brand_title: { type: 'string', public: true, description: 'Public brand title' },
      site_description: { type: 'string', public: true, description: 'Site description' },
      site_logo: { type: 'string', public: true, description: 'Site logo URL' },
      contact_email: { type: 'string', public: true, description: 'Contact email' },
      contact_phone: { type: 'string', public: true, description: 'Contact phone' },
      copyright_holder: { type: 'string', public: false, description: 'Copyright holder for image evidence' },
      copyright_notice: { type: 'string', public: false, description: 'Copyright notice for image evidence' },
      copyright_url: { type: 'string', public: false, description: 'Copyright policy URL for image evidence' },
      admin_brand_name: { type: 'string', public: true, description: 'Admin brand name' },
      admin_brand_initial: { type: 'string', public: true, description: 'Admin brand initial' },
      admin_panel_label: { type: 'string', public: true, description: 'Admin panel label' },
      admin_login_title: { type: 'string', public: true, description: 'Admin login title' },
      admin_footer_text: { type: 'string', public: true, description: 'Admin login footer text' },
      admin_html_title: { type: 'string', public: true, description: 'Admin browser title' }
    }
  },
  email: {
    target: emailSettings,
    fields: {
      smtp_host: { type: 'string', public: false, description: 'SMTP server host' },
      smtp_port: { type: 'number', public: false, description: 'SMTP server port' },
      smtp_username: { type: 'string', public: false, description: 'SMTP username' },
      smtp_password: { type: 'string', public: false, description: 'SMTP password' },
      from_email: { type: 'string', public: false, description: 'Sender email' },
      from_name: { type: 'string', public: false, description: 'Sender name' }
    }
  },
  social: {
    target: socialSettings,
    fields: Object.fromEntries(socialFields.map((field) => [
      field.key,
      { type: 'string' as const, public: true, description: field.label }
    ])) as Record<string, SettingFieldDefinition>
  },
  payment: {
    target: paymentSettings,
    fields: {
      gateway: { type: 'string', public: false, description: 'Payment gateway' },
      test_mode: { type: 'boolean', public: false, description: 'Payment test mode' }
    }
  },
  api: {
    target: apiSettings,
    fields: {
      exchange_rate_enabled: { type: 'boolean', public: false, description: 'Exchange rate API enabled' },
      exchange_rate_provider: { type: 'string', public: false, description: 'Exchange rate API provider' },
      exchange_rate_endpoint: { type: 'string', public: false, description: 'Exchange rate API endpoint' },
      exchange_rate_query_template: { type: 'string', public: false, description: 'Exchange rate API query template' },
      exchange_rate_refresh_minutes: { type: 'number', public: false, description: 'Exchange rate refresh interval in minutes' },
      exchange_rate_api_key: { type: 'string', public: false, description: 'ExchangeRate-API key' },
      time_api_enabled: { type: 'boolean', public: false, description: 'Time API enabled' },
      time_api_provider: { type: 'string', public: false, description: 'Time API provider' },
      time_api_endpoint: { type: 'string', public: false, description: 'Time API endpoint' },
      time_api_query_template: { type: 'string', public: false, description: 'Time API query template' },
      time_api_default_timezone: { type: 'string', public: false, description: 'Time API default timezone' },
      time_api_refresh_minutes: { type: 'number', public: false, description: 'Time API refresh interval in minutes' },
      time_api_key_ref: { type: 'string', public: false, description: 'Time API key reference' }
    }
  }
}

const hasPermission = (permission) => authStore.hasPermission(permission)
const coerceSettingValue = (value, type) => {
  if (type === 'number') {
    const parsed = Number(value)
    return Number.isFinite(parsed) ? parsed : 0
  }
  if (type === 'boolean') return value === true || value === 'true' || value === '1'
  return value ?? ''
}
const settingKey = (setting, group, fields) => {
  if (setting.key in fields) return setting.key
  return setting.key.startsWith(`${group}_`) ? setting.key.slice(group.length + 1) : setting.key
}

const legacySiteSettingKeys = new Set(['site_name'])

const applyFetchedSetting = (setting, group, definition) => {
  const key = settingKey(setting, group, definition.fields)
  if (key in definition.target) {
    definition.target[key] = coerceSettingValue(setting.value, definition.fields[key]?.type || setting.type)
    return
  }
  if (group === 'site' && legacySiteSettingKeys.has(key)) {
    definition.target[key] = coerceSettingValue(setting.value, setting.type || 'string')
  }
}

const normalizeSiteBrandSettings = () => {
  if (!siteSettings.brand_title.trim() && siteSettings.site_name.trim()) {
    siteSettings.brand_title = siteSettings.site_name
  }
}

const fetchSettings = async (group, force = false) => {
  const definition = groupDefinitions[group]
  if (!definition || (!force && loadedGroups.has(group))) return
  loadingSettings.value = true
  try {
    const response = await axios.get(`/api/admin/settings/${group}`, { params: { locale: 'en' } })
    const settings = Array.isArray(response.data.settings) ? response.data.settings : []
    const prefixed = settings.filter((setting) => setting.key.startsWith(`${group}_`))
    const canonical = settings.filter((setting) => !setting.key.startsWith(`${group}_`))
    ;[...prefixed, ...canonical].forEach((setting) => applyFetchedSetting(setting, group, definition))
    loadedGroups.add(group)
    if (group === 'site') normalizeSiteBrandSettings()
    if (group === 'payment' && !String(paymentSettings.gateway || '').trim()) {
      paymentSettings.gateway = 'stripe'
    }
    if (group === 'api') {
      applyAPIRefreshDefaults()
    }
  } catch (error) {
    console.error(`Failed to fetch ${group} settings:`, error)
  } finally {
    loadingSettings.value = false
  }
}

const fetchPublicChatAgents = async () => {
	loadingPublicChatAgents.value = true
  try {
    const response = await axios.get('/api/admin/settings/public-chat-agents')
    publicChatAgentsOverview.value = response.data || null
  } catch (error) {
    console.error('Failed to fetch Public Chat agents:', error)
  } finally {
    loadingPublicChatAgents.value = false
	}
}

const fetchPublicChatGroups = async () => {
  loadingPublicChatGroups.value = true
  try {
    const response = await axios.get('/api/admin/settings/public-chat-groups')
    publicChatGroups.value = Array.isArray(response.data?.groups) ? response.data.groups : []
  } catch (error) {
    console.error('Failed to fetch Public Chat groups:', error)
    publicChatGroups.value = []
  } finally {
    loadingPublicChatGroups.value = false
  }
}

const refreshPublicChat = async () => {
  await Promise.all([fetchPublicChatAgents(), fetchPublicChatGroups()])
}

const fetchPaymentRuntime = async () => {
  loadingPaymentRuntime.value = true
  try {
    const response = await axios.get('/api/admin/settings/payment-runtime')
    paymentRuntime.value = response.data?.data || response.data || null
  } catch (error) {
    console.error('Failed to fetch payment runtime:', error)
    paymentRuntime.value = null
  } finally {
    loadingPaymentRuntime.value = false
  }
}

const fetchCommercialCrawlerProtection = async () => {
  loadingCommercialCrawlerProtection.value = true
  try {
    const response = await axios.get('/api/admin/settings/commercial-crawler-protection')
    commercialCrawlerProtection.value = response.data || null
  } catch (error) {
    console.error('Failed to fetch commercial crawler protection:', error)
    commercialCrawlerProtection.value = null
  } finally {
    loadingCommercialCrawlerProtection.value = false
  }
}

const deleteLegacyPaymentSecrets = async () => {
  await Promise.allSettled(
    ['api_key', 'api_secret'].map((key) =>
      axios.delete(`/api/admin/settings/${key}`, { params: { locale: 'en' } })
    )
  )
}

const fetchPublicChatAgentCandidates = async () => {
  loadingPublicChatAgentCandidates.value = true
  try {
    const response = await axios.get('/api/admin/settings/public-chat-agent-candidates')
    publicChatAgentCandidates.value = Array.isArray(response.data?.candidates) ? response.data.candidates : []
  } catch (error) {
    console.error('Failed to fetch Public Chat agent candidates:', error)
  } finally {
    loadingPublicChatAgentCandidates.value = false
  }
}

const resetPublicChatAgentForm = () => {
  Object.assign(publicChatAgentForm, {
    user_id: '',
    agent_id: '',
    name: '',
    email: '',
    avatar: '',
    whatsapp: '',
    status: 'active',
    online_status: 'offline',
    group_ids: []
  })
}

const applyPublicChatCandidateDefaults = (candidate) => {
  if (!candidate) return
  publicChatAgentForm.agent_id = candidate.agent_id || `user-${candidate.user_id}`
  publicChatAgentForm.name = candidate.profile_name || candidate.display_name || candidate.username || ''
  publicChatAgentForm.email = candidate.profile_email || candidate.email || ''
  publicChatAgentForm.avatar = candidate.profile_avatar || ''
  publicChatAgentForm.whatsapp = candidate.profile_whatsapp || ''
  publicChatAgentForm.status = candidate.profile_status || 'active'
  publicChatAgentForm.online_status = candidate.profile_online_status || 'offline'
  publicChatAgentForm.group_ids = Array.isArray(candidate.profile_group_ids)
    ? candidate.profile_group_ids.map((id) => Number(id)).filter(Boolean)
    : []
}

const openPublicChatAgentDialog = async () => {
  resetPublicChatAgentForm()
  publicChatAgentDialogOpen.value = true
  await Promise.all([fetchPublicChatAgentCandidates(), fetchPublicChatGroups()])
}

const savePublicChatAgent = async () => {
  if (!publicChatAgentForm.user_id) {
    toast.error('请选择一个后台用户')
    return
  }

   const publicEmail = publicChatAgentForm.email.trim() || selectedPublicChatAgentCandidate.value?.email?.trim() || ''
   const publicWhatsApp = publicChatAgentForm.whatsapp.trim()
   if (publicChatAgentForm.status === 'active' && (!publicEmail || !publicWhatsApp)) {
     toast.error('公开客服启用前必须填写公开邮箱和 WhatsApp，前台聊天头像选择弹层会使用这两个联系方式')
     return
   }

  publicChatAgentSaving.value = true
  try {
    const response = await axios.post('/api/admin/settings/public-chat-agents', {
      user_id: Number(publicChatAgentForm.user_id),
      agent_id: publicChatAgentForm.agent_id.trim(),
      name: publicChatAgentForm.name.trim(),
      email: publicEmail,
      avatar: publicChatAgentForm.avatar.trim(),
      whatsapp: publicWhatsApp,
      status: publicChatAgentForm.status,
      online_status: publicChatAgentForm.online_status,
      group_ids: Array.isArray(publicChatAgentForm.group_ids)
        ? publicChatAgentForm.group_ids.map((id) => Number(id)).filter(Boolean)
        : []
    })
    toast.success(response.data?.created ? '已添加 Public Chat 客服 Profile' : '已更新 Public Chat 客服 Profile')
    publicChatAgentDialogOpen.value = false
    await fetchPublicChatAgents()
  } catch (error) {
    console.error('Failed to save Public Chat agent profile:', error)
     toast.error(error?.response?.data?.error || 'Public Chat 客服 Profile 保存失败')
  } finally {
    publicChatAgentSaving.value = false
  }
}

const resetPublicChatGroupForm = () => {
  Object.assign(publicChatGroupForm, {
    id: 0,
    code: '',
    name: '',
    description: '',
    status: 'active',
    sort_order: 0
  })
}

const openPublicChatGroupDialog = () => {
  resetPublicChatGroupForm()
  publicChatGroupDialogOpen.value = true
}

const editPublicChatGroup = (group) => {
  Object.assign(publicChatGroupForm, {
    id: Number(group?.id || 0),
    code: group?.code || '',
    name: group?.name || '',
    description: group?.description || '',
    status: group?.status || 'active',
    sort_order: Number(group?.sort_order || 0)
  })
  publicChatGroupDialogOpen.value = true
}

const savePublicChatGroup = async () => {
  if (!publicChatGroupForm.name.trim()) {
    toast.error('请输入客服组名称')
    return
  }
  publicChatGroupSaving.value = true
  try {
    const payload = {
      code: publicChatGroupForm.code.trim(),
      name: publicChatGroupForm.name.trim(),
      description: publicChatGroupForm.description.trim(),
      status: publicChatGroupForm.status,
      sort_order: Number(publicChatGroupForm.sort_order) || 0
    }
    const response = publicChatGroupForm.id
      ? await axios.put(`/api/admin/settings/public-chat-groups/${publicChatGroupForm.id}`, payload)
      : await axios.post('/api/admin/settings/public-chat-groups', payload)
    toast.success(response.data?.created ? '客服组已创建' : '客服组已保存')
    publicChatGroupDialogOpen.value = false
    await Promise.all([fetchPublicChatGroups(), fetchPublicChatAgents()])
  } catch (error) {
    toast.error(error?.response?.data?.error || '客服组保存失败')
  } finally {
    publicChatGroupSaving.value = false
  }
}

const deletePublicChatGroup = async (group) => {
  if (!group?.id || !window.confirm(`确定删除客服组“${group.name}”吗？已绑定客服会变为未分组。`)) return
  try {
    await axios.delete(`/api/admin/settings/public-chat-groups/${group.id}`)
    toast.success('客服组已删除')
    await Promise.all([fetchPublicChatGroups(), fetchPublicChatAgents()])
  } catch (error) {
    toast.error(error?.response?.data?.error || '客服组删除失败')
  }
}

const uploadSiteLogo = async (file) => {
  if (!file) return
  if (!file.type?.startsWith('image/')) {
    toast.error('Logo 只能上传图片文件')
    return
  }

  uploadingSiteLogo.value = true
  try {
    const formData = new FormData()
    formData.append('file', file)
    formData.append('media_type', 'image')
    const asset = await mediaApi.uploadAsset(formData)
    const logoURL = String(assetAccessURL(asset) || asset?.url || '').trim()
    if (!logoURL) {
      toast.error('上传成功但没有返回 Logo 地址')
      return
    }
    siteSettings.site_logo = logoURL
    toast.success('Logo 已上传，保存设置后前台生效')
  } catch (error) {
    console.error('Failed to upload site logo:', error)
    toast.error('Logo 上传失败，请检查文件类型和大小')
  } finally {
    uploadingSiteLogo.value = false
  }
}

const saveSettings = async () => {
  const group = activeTab.value
  const definition = groupDefinitions[group]
  if (!definition) return
  if (group === 'site') normalizeSiteBrandSettings()
  if (group === 'api') {
    applyAPIRefreshDefaults()
  }
  const settings = Object.entries(definition.fields).map(([key, metadata]) => ({
    key,
    value: String(definition.target[key] ?? ''),
    type: metadata.type,
    group,
    locale: 'en',
    is_public: metadata.public,
    description: metadata.description
  }))
  saving.value = true
  try {
    const response = await axios.post('/api/admin/settings/batch', { settings })
    toast.success(`已保存 ${response.data.count ?? settings.length} 项设置`)
    loadedGroups.delete(group)
    if (group === 'payment') await deleteLegacyPaymentSecrets()
    await fetchSettings(group, true)
    if (group === 'payment') await fetchPaymentRuntime()
    return true
  } catch (error) {
    console.error('Failed to save settings:', error)
    return false
  } finally {
    saving.value = false
  }
}

const syncExchangeRates = async () => {
  if (!hasPermission('settings:edit')) return
  syncingExchangeRates.value = true
  try {
    const response = await axios.post('/api/admin/settings/exchange-rates/sync')
    const rates = response.data?.data?.rates || response.data?.rates || []
    toast.success(`汇率缓存已同步：${rates.length} 个币种`)
    loadedGroups.delete('api')
    await fetchSettings('api', true)
  } catch (error) {
    console.error('Failed to sync exchange rates:', error)
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '汇率同步失败')
  } finally {
    syncingExchangeRates.value = false
  }
}

watch(() => publicChatAgentForm.user_id, (userID) => {
  if (!userID) return
  applyPublicChatCandidateDefaults(selectedPublicChatAgentCandidate.value)
})

watch(activeTab, (tab) => {
  if (tab === 'public_chat') {
    fetchPublicChatAgents()
    fetchPublicChatGroups()
  }
  else if (tab === 'commercial_crawler') {
    fetchCommercialCrawlerProtection()
  }
  else if (tab === 'api') {
    fetchCurrencyPolicyForSettings().finally(() => fetchSettings(tab))
  }
  else {
    fetchSettings(tab)
    if (tab === 'payment') fetchPaymentRuntime()
  }
}, { immediate: true })
</script>
