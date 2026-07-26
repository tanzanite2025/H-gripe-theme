<template>
  <div class="space-y-4">
    <AdminPageHeader title="系统设置" description="管理站点、邮件、搜索、社交与支付配置">
      <template #actions>
        <Button
          v-if="hasPermission('settings:edit') && activeTab !== 'public_chat'"
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
        v-model:active-tab="activeTab"
        v-model:show-smtp-password="showSmtpPassword"
        v-model:show-payment-secrets="showPaymentSecrets"
        :site-settings="siteSettings"
        :email-settings="emailSettings"
        :seo-settings="seoSettings"
        :social-settings="socialSettings"
        :payment-settings="paymentSettings"
        :payment-runtime="paymentRuntime"
        :loading-payment-runtime="loadingPaymentRuntime"
        :social-fields="socialFields"
        :loading-public-chat-agents="loadingPublicChatAgents"
        :loading-public-chat-agent-candidates="loadingPublicChatAgentCandidates"
        :public-chat-agents-summary="publicChatAgentsSummary"
        :public-chat-agents="publicChatAgents"
        :public-chat-agent-warnings="publicChatAgentWarnings"
        :can-edit="hasPermission('settings:edit')"
        @open-agent-dialog="openPublicChatAgentDialog"
        @refresh-public-chat="fetchPublicChatAgents"
        @refresh-payment-runtime="fetchPaymentRuntime"
      />
    </div>

    <PublicChatAgentDialog
      v-model:open="publicChatAgentDialogOpen"
      :form="publicChatAgentForm"
      :candidates="publicChatAgentCandidates"
      :selected-candidate="selectedPublicChatAgentCandidate"
      :loading-candidates="loadingPublicChatAgentCandidates"
      :saving="publicChatAgentSaving"
      @save="savePublicChatAgent"
    />
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import { LoaderCircle, Save } from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import PublicChatAgentDialog from '@/components/admin/settings/PublicChatAgentDialog.vue'
import SettingsTabsPanel from '@/components/admin/settings/SettingsTabsPanel.vue'
import { Button } from '@/components/ui/button'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

const authStore = useAuthStore()
const activeTab = ref('site')
const saving = ref(false)
const loadingSettings = ref(false)
const showSmtpPassword = ref(false)
const showPaymentSecrets = ref(false)
const loadedGroups = new Set()

const siteSettings = reactive({
  site_name: '',
  brand_title: '',
  site_description: '',
  site_url: '',
  site_logo: '',
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
const seoSettings = reactive({ meta_title: '', meta_description: '', meta_keywords: '', google_analytics: '', google_tag_manager: '' })
const socialSettings = reactive({ facebook: '', twitter: '', instagram: '', linkedin: '', youtube: '', wechat: '' })
const paymentSettings = reactive({ gateway: '', test_mode: true })
const paymentRuntime = ref(null)
const loadingPaymentRuntime = ref(false)

const loadingPublicChatAgents = ref(false)
const publicChatAgentsOverview = ref(null)
const publicChatAgentsSummary = computed(() => publicChatAgentsOverview.value?.summary || {})
const publicChatAgents = computed(() => publicChatAgentsOverview.value?.agents || [])
const publicChatAgentWarnings = computed(() => publicChatAgentsOverview.value?.warnings || [])
const loadingPublicChatAgentCandidates = ref(false)
const publicChatAgentCandidates = ref([])
const publicChatAgentDialogOpen = ref(false)
const publicChatAgentSaving = ref(false)
const publicChatAgentForm = reactive({
  user_id: '',
  agent_id: '',
  name: '',
  email: '',
  avatar: '',
  whatsapp: '',
  status: 'active'
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

const groupDefinitions = {
  site: {
    target: siteSettings,
    fields: {
      site_name: { type: 'string', public: true, description: 'Site name' },
      brand_title: { type: 'string', public: true, description: 'Top gradient brand title' },
      site_description: { type: 'string', public: true, description: 'Site description' },
      site_url: { type: 'string', public: true, description: 'Public site URL' },
      site_logo: { type: 'string', public: true, description: 'Site logo URL' },
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
  seo: {
    target: seoSettings,
    fields: {
      meta_title: { type: 'string', public: true, description: 'Default meta title' },
      meta_description: { type: 'string', public: true, description: 'Default meta description' },
      meta_keywords: { type: 'string', public: true, description: 'Default meta keywords' },
      google_analytics: { type: 'string', public: true, description: 'Google Analytics ID' },
      google_tag_manager: { type: 'string', public: true, description: 'Google Tag Manager ID' }
    }
  },
  social: {
    target: socialSettings,
    fields: Object.fromEntries(socialFields.map((field) => [field.key, { type: 'string', public: true, description: field.label }]))
  },
  payment: {
    target: paymentSettings,
    fields: {
      gateway: { type: 'string', public: false, description: 'Payment gateway' },
      test_mode: { type: 'boolean', public: false, description: 'Payment test mode' }
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

const fetchSettings = async (group, force = false) => {
  const definition = groupDefinitions[group]
  if (!definition || (!force && loadedGroups.has(group))) return
  loadingSettings.value = true
  try {
    const response = await axios.get(`/api/admin/settings/${group}`, { params: { locale: 'en' } })
    const settings = Array.isArray(response.data.settings) ? response.data.settings : []
    const prefixed = settings.filter((setting) => setting.key.startsWith(`${group}_`))
    const canonical = settings.filter((setting) => !setting.key.startsWith(`${group}_`))
    ;[...prefixed, ...canonical].forEach((setting) => {
      const key = settingKey(setting, group, definition.fields)
      if (key in definition.target) {
        definition.target[key] = coerceSettingValue(setting.value, definition.fields[key]?.type || setting.type)
      }
    })
    loadedGroups.add(group)
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
    status: 'active'
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
}

const openPublicChatAgentDialog = async () => {
  resetPublicChatAgentForm()
  publicChatAgentDialogOpen.value = true
  await fetchPublicChatAgentCandidates()
}

const savePublicChatAgent = async () => {
  if (!publicChatAgentForm.user_id) {
    toast.error('请选择一个后台用户')
    return
  }

  publicChatAgentSaving.value = true
  try {
    const response = await axios.post('/api/admin/settings/public-chat-agents', {
      user_id: Number(publicChatAgentForm.user_id),
      agent_id: publicChatAgentForm.agent_id.trim(),
      name: publicChatAgentForm.name.trim(),
      email: publicChatAgentForm.email.trim(),
      avatar: publicChatAgentForm.avatar.trim(),
      whatsapp: publicChatAgentForm.whatsapp.trim(),
      status: publicChatAgentForm.status
    })
    toast.success(response.data?.created ? '已添加 Public Chat 客服 Profile' : '已更新 Public Chat 客服 Profile')
    publicChatAgentDialogOpen.value = false
    await fetchPublicChatAgents()
  } catch (error) {
    console.error('Failed to save Public Chat agent profile:', error)
  } finally {
    publicChatAgentSaving.value = false
  }
}

const saveSettings = async () => {
  const group = activeTab.value
  const definition = groupDefinitions[group]
  if (!definition) return
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
  } catch (error) {
    console.error('Failed to save settings:', error)
  } finally {
    saving.value = false
  }
}

watch(() => publicChatAgentForm.user_id, (userID) => {
  if (!userID) return
  applyPublicChatCandidateDefaults(selectedPublicChatAgentCandidate.value)
})

watch(activeTab, (tab) => {
  if (tab === 'public_chat') fetchPublicChatAgents()
  else {
    fetchSettings(tab)
    if (tab === 'payment') fetchPaymentRuntime()
  }
})

onMounted(() => fetchSettings('site'))
</script>
