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
          {{ saving ? t('common.saving') : t('settings.saveSettings') }}
        </Button>
      </template>
    </AdminPageHeader>

    <div class="relative min-h-96">
      <div v-if="loadingSettings" class="absolute inset-0 z-10 flex items-center justify-center bg-background/75">
        <LoaderCircle class="size-5 animate-spin text-primary" :aria-label="t('settings.loadingSettings')" />
      </div>

      <SettingsTabsPanel
        :active-tab="activeTab"
        v-model:show-smtp-password="showSmtpPassword"
        :site-settings="siteSettings"
        :email-settings="emailSettings"
        :api-settings="apiSettings"
        :commercial-crawler-protection="commercialCrawlerProtection"
        :loading-commercial-crawler-protection="loadingCommercialCrawlerProtection"
        :uploading-site-logo="uploadingSiteLogo"
        :uploading-site-favicon="uploadingSiteFavicon"
        :saving-api-settings="saving && activeTab === 'api'"
        :loading-public-chat-agents="loadingPublicChatAgents"
        :loading-public-chat-groups="loadingPublicChatGroups"
        :loading-public-chat-agent-candidates="loadingPublicChatAgentCandidates"
        :public-chat-agents-summary="publicChatAgentsSummary"
        :public-chat-agents="publicChatAgents"
        :public-chat-groups="publicChatGroups"
        :public-chat-agent-warnings="publicChatAgentWarnings"
        :refund-return-policy="refundReturnPolicy"
        :refund-return-policy-locale="refundReturnPolicyLocale"
        :refund-return-policy-fallback="refundReturnPolicyFallback"
        :loading-refund-return-policy="loadingRefundReturnPolicy"
        :saving-refund-return-policy="savingRefundReturnPolicy"
        :uploading-refund-return-section="uploadingRefundReturnSection"
        :can-edit="canEditActiveTab"
        @open-agent-dialog="openPublicChatAgentDialog"
        @open-group-dialog="openPublicChatGroupDialog"
        @edit-group="editPublicChatGroup"
        @delete-group="deletePublicChatGroup"
        @refresh-public-chat="refreshPublicChat"
        @refresh-commercial-crawler-protection="fetchCommercialCrawlerProtection"
        @upload-site-logo="uploadSiteLogo"
        @clear-site-logo="clearSiteLogo"
        @upload-site-favicon="uploadSiteFavicon"
        @refund-return-locale-change="changeRefundReturnPolicyLocale"
        @save-refund-return-policy="saveRefundReturnPolicy"
        @upload-refund-return-image="uploadRefundReturnImage"
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
      @avatar-uploaded="handlePublicChatAgentAvatarUploaded"
      @avatar-removed="handlePublicChatAgentAvatarRemoved"
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
import { useAdminI18n } from '@/i18n'
import refundReturnPolicyApi from '@/api/refundReturnPolicy'
import type {
  RefundReturnPolicy,
  RefundReturnPolicyEditor,
  RefundReturnPolicyEditorSection,
  RefundReturnPolicySection,
} from '@/api/refundReturnPolicy'
import mediaApi from '@/api/media'
import { assetAccessURL } from '@/lib/mediaPresentation'
import { validateUploadFile } from '@/lib/uploadSpecs'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'
import type { ApiManagementSettings } from '@/modules/settings/types'

const authStore = useAuthStore()
const { t } = useAdminI18n()
const DISABLED_TIME_API_REFRESH_MINUTES = 0
const CUSTOMS_LOOKUP_US_HTS_ENDPOINT = 'https://hts.usitc.gov/reststop/search'
const CUSTOMS_LOOKUP_UK_TRADE_TARIFF_ENDPOINT = 'https://www.trade-tariff.service.gov.uk/api/v2/commodities'
const activeTab = useRouteTab({
  defaultValue: 'site',
  values: ['site', 'email', 'markets', 'api', 'commercial_crawler', 'public_chat', 'refund_return'],
  routes: {
    site: 'SettingsSite',
    email: 'SettingsEmail',
    markets: 'SettingsMarkets',
    api: 'SettingsApi',
    commercial_crawler: 'SettingsCommercialCrawler',
    public_chat: 'SupportPublicChat',
    refund_return: 'ContentRefundReturn',
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

const pageTitle = computed(() => {
  if (activeTab.value === 'public_chat') return t('settings.publicChatTitle')
  if (activeTab.value === 'refund_return') return '退货退款'
  return t('settings.systemTitle')
})
const pageDescription = computed(() => {
  if (activeTab.value === 'markets') return t('settings.marketsDescription')
  if (activeTab.value === 'public_chat') return t('settings.publicChatDescription')
  if (activeTab.value === 'refund_return') return '集中维护前台退货退款政策，支持图片说明并可在页面或弹窗复用。'
  return t('settings.systemDescription')
})
const saving = ref(false)
const loadingSettings = ref(false)
const uploadingSiteLogo = ref(false)
const uploadingSiteFavicon = ref(false)
const showSmtpPassword = ref(false)
const loadedGroups = new Set()
const selfSavingTabs = new Set(['markets', 'api', 'commercial_crawler', 'public_chat', 'refund_return'])

const siteSettings = reactive({
  site_name: '',
  site_description: '',
  site_logo: '',
  site_favicon: '',
  contact_email: '',
  contact_phone: '',
  admin_brand_name: '',
  admin_brand_initial: '',
  admin_panel_label: '',
  admin_login_title: '',
  admin_footer_text: '',
  admin_html_title: ''
})
const emailSettings = reactive({ smtp_host: '', smtp_port: 587, smtp_username: '', smtp_password: '', from_email: '', from_name: '' })
const apiSettings = reactive<ApiManagementSettings>({
  time_api_enabled: false,
  time_api_provider: 'built-in',
  time_api_endpoint: '',
  time_api_query_template: '',
  time_api_default_timezone: 'Asia/Shanghai',
  time_api_refresh_minutes: DISABLED_TIME_API_REFRESH_MINUTES,
  time_api_key_ref: '',
  customs_lookup_us_hts_enabled: true,
  customs_lookup_us_hts_endpoint: CUSTOMS_LOOKUP_US_HTS_ENDPOINT,
  customs_lookup_us_hts_api_key: '',
  customs_lookup_us_hts_api_key_header: 'X-API-Key',
  customs_lookup_uk_trade_tariff_enabled: true,
  customs_lookup_uk_trade_tariff_endpoint: CUSTOMS_LOOKUP_UK_TRADE_TARIFF_ENDPOINT,
  customs_lookup_uk_trade_tariff_api_key: '',
  customs_lookup_uk_trade_tariff_api_key_header: 'X-API-Key'
})
const commercialCrawlerProtection = ref(null)
const loadingCommercialCrawlerProtection = ref(false)
const refundReturnPolicyLocale = ref('en')
const refundReturnPolicyFallback = ref(false)
const loadingRefundReturnPolicy = ref(false)
const savingRefundReturnPolicy = ref(false)
const uploadingRefundReturnSection = ref<number | null>(null)

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

const createRefundReturnPolicyForm = (): RefundReturnPolicyEditor => ({
  title: '',
  intro: '',
  sections: [],
  contact_label: '',
  contact_url: '/company/contact',
  updated_at: '',
})

const refundReturnPolicy = reactive<RefundReturnPolicyEditor>(createRefundReturnPolicyForm())

const applyTimezoneDefaults = (clearExternalConfig = false) => {
  apiSettings.time_api_default_timezone = String(apiSettings.time_api_default_timezone || '').trim() || 'Asia/Shanghai'
  if (!clearExternalConfig) return
  apiSettings.time_api_enabled = false
  apiSettings.time_api_provider = 'built-in'
  apiSettings.time_api_endpoint = ''
  apiSettings.time_api_query_template = ''
  apiSettings.time_api_refresh_minutes = DISABLED_TIME_API_REFRESH_MINUTES
  apiSettings.time_api_key_ref = ''
}

const groupDefinitions: Record<string, SettingsGroupDefinition> = {
  site: {
    target: siteSettings,
    fields: {
      site_name: { type: 'string', public: true, description: 'Site name' },
      site_description: { type: 'string', public: true, description: 'Site description' },
      site_logo: { type: 'string', public: true, description: 'Site logo URL' },
      site_favicon: { type: 'string', public: true, description: 'Browser favicon URL' },
      contact_email: { type: 'string', public: true, description: 'Contact email' },
      contact_phone: { type: 'string', public: true, description: 'Contact phone' },
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
  api: {
    target: apiSettings,
    fields: {
      time_api_enabled: { type: 'boolean', public: false, description: 'External Time API disabled; timezone is built in' },
      time_api_provider: { type: 'string', public: false, description: 'Timezone source' },
      time_api_endpoint: { type: 'string', public: false, description: 'External Time API endpoint disabled' },
      time_api_query_template: { type: 'string', public: false, description: 'External Time API query template disabled' },
      time_api_default_timezone: { type: 'string', public: false, description: 'Default business timezone' },
      time_api_refresh_minutes: { type: 'number', public: false, description: 'External Time API refresh interval disabled' },
      time_api_key_ref: { type: 'string', public: false, description: 'External Time API key reference disabled' },
      customs_lookup_us_hts_enabled: { type: 'boolean', public: false, description: 'US HTS customs lookup enabled' },
      customs_lookup_us_hts_endpoint: { type: 'string', public: false, description: 'US HTS customs lookup endpoint' },
      customs_lookup_us_hts_api_key: { type: 'string', public: false, description: 'US HTS customs lookup API key' },
      customs_lookup_us_hts_api_key_header: { type: 'string', public: false, description: 'US HTS customs lookup API key header' },
      customs_lookup_uk_trade_tariff_enabled: { type: 'boolean', public: false, description: 'UK Trade Tariff customs lookup enabled' },
      customs_lookup_uk_trade_tariff_endpoint: { type: 'string', public: false, description: 'UK Trade Tariff customs lookup endpoint' },
      customs_lookup_uk_trade_tariff_api_key: { type: 'string', public: false, description: 'UK Trade Tariff customs lookup API key' },
      customs_lookup_uk_trade_tariff_api_key_header: { type: 'string', public: false, description: 'UK Trade Tariff customs lookup API key header' }
    }
  }
}

const hasPermission = (permission) => authStore.hasPermission(permission)
const canEditActiveTab = computed(() => (
  activeTab.value === 'refund_return'
    ? hasPermission('content:edit')
    : hasPermission('settings:edit')
))
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

const applyFetchedSetting = (setting, group, definition) => {
  const key = settingKey(setting, group, definition.fields)
  if (key in definition.target) {
    definition.target[key] = coerceSettingValue(setting.value, definition.fields[key]?.type || setting.type)
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
    if (group === 'api') {
      applyTimezoneDefaults()
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

const normalizePolicySection = (section: Partial<RefundReturnPolicySection>, index: number): RefundReturnPolicyEditorSection => ({
  id: String(section?.id || `section-${index + 1}`).trim(),
  title: String(section?.title || '').trim(),
  body: String(section?.body || '').trim(),
  bullets: Array.isArray(section?.bullets) ? section.bullets.map((item) => String(item || '').trim()).filter(Boolean) : [],
  bulletsText: Array.isArray(section?.bullets) ? section.bullets.map((item) => String(item || '').trim()).filter(Boolean).join('\n') : '',
  image: {
    url: String(section?.image?.url || '').trim(),
    alt: String(section?.image?.alt || '').trim(),
    caption: String(section?.image?.caption || '').trim(),
  },
})

const applyRefundReturnPolicy = (policy: Partial<RefundReturnPolicy> = {}) => {
  Object.assign(refundReturnPolicy, {
    title: String(policy.title || '').trim(),
    intro: String(policy.intro || '').trim(),
    sections: Array.isArray(policy.sections)
      ? policy.sections.map((section, index) => normalizePolicySection(section, index))
      : [],
    contact_label: String(policy.contact_label || '').trim(),
    contact_url: String(policy.contact_url || '/company/contact').trim(),
    updated_at: String(policy.updated_at || '').trim(),
  })
}

const fetchRefundReturnPolicy = async () => {
  loadingRefundReturnPolicy.value = true
  try {
    const response = await refundReturnPolicyApi.get(refundReturnPolicyLocale.value)
    applyRefundReturnPolicy(response.policy || {})
    refundReturnPolicyFallback.value = Boolean(response.fallback)
  } catch (error) {
    console.error('Failed to fetch refund return policy:', error)
    toast.error('退货退款内容加载失败')
  } finally {
    loadingRefundReturnPolicy.value = false
  }
}

const normalizePolicyForSave = (): RefundReturnPolicy => ({
  title: refundReturnPolicy.title.trim(),
  intro: refundReturnPolicy.intro.trim(),
  sections: refundReturnPolicy.sections.map((section, index) => ({
    id: section.id.trim() || `section-${index + 1}`,
    title: section.title.trim(),
    body: section.body.trim(),
    bullets: String(section.bulletsText || '')
      .split(/\r?\n/)
      .map((item) => item.trim())
      .filter(Boolean),
    image: section.image.url.trim()
      ? {
          url: section.image.url.trim(),
          alt: section.image.alt.trim(),
          caption: section.image.caption.trim(),
        }
      : undefined,
  })),
  contact_label: refundReturnPolicy.contact_label.trim(),
  contact_url: refundReturnPolicy.contact_url.trim(),
  updated_at: refundReturnPolicy.updated_at,
})

const saveRefundReturnPolicy = async () => {
  if (!hasPermission('content:edit')) return
  savingRefundReturnPolicy.value = true
  try {
    const response = await refundReturnPolicyApi.update(refundReturnPolicyLocale.value, normalizePolicyForSave())
    applyRefundReturnPolicy(response.policy || {})
    refundReturnPolicyFallback.value = Boolean(response.fallback)
    toast.success('退货退款内容已保存')
  } catch (error) {
    console.error('Failed to save refund return policy:', error)
    toast.error(error?.response?.data?.error || '退货退款内容保存失败')
  } finally {
    savingRefundReturnPolicy.value = false
  }
}

const changeRefundReturnPolicyLocale = async (locale) => {
  const nextLocale = String(locale || '').trim()
  if (!nextLocale || nextLocale === refundReturnPolicyLocale.value) return
  refundReturnPolicyLocale.value = nextLocale
  await fetchRefundReturnPolicy()
}

const uploadRefundReturnImage = async ({ index, file }) => {
  if (!hasPermission('content:edit')) return
  if (!file || !refundReturnPolicy.sections[index]) return
  const validation = await validateUploadFile(file, 'refund_return_image')
  if (!validation.ok) {
    toast.error(validation.error || '退货退款图片不符合上传规范')
    return
  }
  if (validation.warning) toast.warning(validation.warning)
  uploadingRefundReturnSection.value = index
  try {
    const formData = new FormData()
    formData.append('file', file)
    formData.append('media_type', 'image')
    formData.append('image_purpose', 'refund_return_image')
    formData.append('alt', refundReturnPolicy.sections[index].image.alt || refundReturnPolicy.sections[index].title || 'Refund return policy image')
    const asset = await refundReturnPolicyApi.uploadImage(formData)
    const imageURL = String(assetAccessURL(asset) || asset?.url || '').trim()
    if (!imageURL) {
      toast.error('上传成功但没有返回图片地址')
      return
    }
    refundReturnPolicy.sections[index].image.url = imageURL
    if (!refundReturnPolicy.sections[index].image.alt && asset.alt) {
      refundReturnPolicy.sections[index].image.alt = String(asset.alt)
    }
    toast.success('图片已上传，保存后前台生效')
  } catch (error) {
    console.error('Failed to upload refund return image:', error)
    toast.error(error?.response?.data?.error || '退货退款图片上传失败')
  } finally {
    uploadingRefundReturnSection.value = null
  }
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
  if (publicChatAgentForm.status === 'active' && !publicEmail && !publicWhatsApp) {
    toast.error('公开客服启用前至少填写邮箱或 WhatsApp 其中一个，前台会展示可用的联系方式')
    return
  }

  publicChatAgentSaving.value = true
  try {
    const response = await axios.post('/api/admin/settings/public-chat-agents', {
      user_id: Number(publicChatAgentForm.user_id),
      agent_id: publicChatAgentForm.agent_id.trim(),
      name: publicChatAgentForm.name.trim(),
      email: publicEmail,
      whatsapp: publicWhatsApp,
      status: publicChatAgentForm.status,
      online_status: publicChatAgentForm.online_status,
      group_ids: Array.isArray(publicChatAgentForm.group_ids)
        ? publicChatAgentForm.group_ids.map((id) => Number(id)).filter(Boolean)
        : []
    })
    const savedAvatar = String(response.data?.agent?.avatar || publicChatAgentForm.avatar || '').trim()
    publicChatAgentForm.avatar = savedAvatar
    toast.success(response.data?.created ? 'Profile 已保存，现在可以上传头像' : 'Public Chat 客服 Profile 已保存')
    await Promise.all([fetchPublicChatAgents(), fetchPublicChatAgentCandidates()])
  } catch (error) {
    console.error('Failed to save Public Chat agent profile:', error)
     toast.error(error?.response?.data?.error || 'Public Chat 客服 Profile 保存失败')
  } finally {
    publicChatAgentSaving.value = false
  }
}

const handlePublicChatAgentAvatarUploaded = async (avatar) => {
  publicChatAgentForm.avatar = String(avatar || '').trim()
  await Promise.all([fetchPublicChatAgents(), fetchPublicChatAgentCandidates()])
}

const handlePublicChatAgentAvatarRemoved = async () => {
  publicChatAgentForm.avatar = ''
  await Promise.all([fetchPublicChatAgents(), fetchPublicChatAgentCandidates()])
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
  const validation = await validateUploadFile(file, 'site_logo')
  if (!validation.ok) {
    toast.error(validation.error || '站点 Logo 不符合上传规范')
    return
  }

  uploadingSiteLogo.value = true
  try {
    const formData = new FormData()
    formData.append('file', file)
    const response = await axios.post('/api/admin/settings/site-logo', formData)
    const logo = response.data?.data?.logo || response.data?.logo || {}
    const logoURL = String(logo?.url || logo?.access_url || '').trim()
    if (!logoURL) {
      toast.error('上传成功但没有返回 Logo 地址')
      return
    }
    siteSettings.site_logo = logoURL
    toast.success('Logo 已上传并替换当前站点 Logo')
  } catch (error) {
    console.error('Failed to upload site logo:', error)
    toast.error('Logo 上传失败，请检查文件类型和大小')
  } finally {
    uploadingSiteLogo.value = false
  }
}

const clearSiteLogo = async () => {
  if (!siteSettings.site_logo || uploadingSiteLogo.value) return

  uploadingSiteLogo.value = true
  try {
    await axios.delete('/api/admin/settings/site-logo')
    siteSettings.site_logo = ''
    toast.success('站点 Logo 已删除')
  } catch (error) {
    console.error('Failed to delete site logo:', error)
    toast.error(error?.response?.data?.error || '站点 Logo 删除失败')
  } finally {
    uploadingSiteLogo.value = false
  }
}

const uploadSiteFavicon = async (file) => {
  if (!file) return
  const validation = await validateUploadFile(file, 'site_favicon')
  if (!validation.ok) {
    toast.error(validation.error || 'Favicon 不符合上传规范')
    return
  }
  if (validation.warning) toast.warning(validation.warning)

  uploadingSiteFavicon.value = true
  try {
      const formData = new FormData()
      formData.append('file', file)
      formData.append('media_type', 'image')
      formData.append('image_purpose', 'site_favicon')
    const asset = await mediaApi.uploadAsset(formData)
    const faviconURL = String(assetAccessURL(asset) || asset?.url || '').trim()
    if (!faviconURL) {
      toast.error('上传成功但没有返回 Favicon 地址')
      return
    }
    siteSettings.site_favicon = faviconURL
    toast.success('Favicon 已上传，保存设置后前台生效')
  } catch (error) {
    console.error('Failed to upload site favicon:', error)
    toast.error('Favicon 上传失败，请检查文件类型和大小')
  } finally {
    uploadingSiteFavicon.value = false
  }
}

const saveSettings = async () => {
  const group = activeTab.value
  const definition = groupDefinitions[group]
  if (!definition) return
  if (group === 'api') {
    applyTimezoneDefaults(true)
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
    await fetchSettings(group, true)
    return true
  } catch (error) {
    console.error('Failed to save settings:', error)
    return false
  } finally {
    saving.value = false
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
  else if (tab === 'refund_return') {
    fetchRefundReturnPolicy()
  }
  else {
    fetchSettings(tab)
  }
}, { immediate: true })
</script>
