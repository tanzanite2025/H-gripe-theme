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

      <Tabs v-model="activeTab" class="gap-5">
        <TabsList variant="line" class="h-10 w-full justify-start overflow-x-auto rounded-none border-b bg-transparent p-0">
          <TabsTrigger value="site" class="h-9 flex-none px-3"><Globe2 class="size-4" />站点</TabsTrigger>
          <TabsTrigger value="email" class="h-9 flex-none px-3"><Mail class="size-4" />邮件</TabsTrigger>
          <TabsTrigger value="seo" class="h-9 flex-none px-3"><SearchCheck class="size-4" />SEO</TabsTrigger>
          <TabsTrigger value="social" class="h-9 flex-none px-3"><Share2 class="size-4" />社交媒体</TabsTrigger>
          <TabsTrigger value="payment" class="h-9 flex-none px-3"><CreditCard class="size-4" />支付</TabsTrigger>
          <TabsTrigger value="public_chat" class="h-9 flex-none px-3"><Headset class="size-4" />Public Chat</TabsTrigger>
        </TabsList>

        <TabsContent value="site">
          <SettingsSection title="站点资料" description="前台使用的品牌和联系信息。">
            <div class="grid gap-4 md:grid-cols-2">
              <AdminFormField label="站点名称">
                <Input v-model="siteSettings.site_name" />
              </AdminFormField>
              <AdminFormField label="联系邮箱">
                <Input v-model="siteSettings.contact_email" type="email" />
              </AdminFormField>
              <AdminFormField label="联系电话">
                <Input v-model="siteSettings.contact_phone" type="tel" />
              </AdminFormField>
              <AdminFormField label="站点 Logo">
                <Input v-model="siteSettings.site_logo" placeholder="Logo URL" />
              </AdminFormField>
              <AdminFormField label="站点描述" class="md:col-span-2">
                <Textarea v-model="siteSettings.site_description" class="min-h-24" />
              </AdminFormField>
              <div v-if="siteSettings.site_logo" class="flex h-28 items-center justify-center overflow-hidden rounded-lg border bg-muted md:col-span-2">
                <img :src="siteSettings.site_logo" alt="站点 Logo 预览" class="max-h-20 max-w-[80%] object-contain" />
              </div>
            </div>
          </SettingsSection>
        </TabsContent>

        <TabsContent value="email">
          <SettingsSection title="SMTP 配置" description="用于系统通知与业务邮件发送。">
            <div class="grid gap-4 md:grid-cols-2">
              <AdminFormField label="SMTP 主机">
                <Input v-model="emailSettings.smtp_host" placeholder="smtp.example.com" />
              </AdminFormField>
              <AdminFormField label="SMTP 端口">
                <Input v-model.number="emailSettings.smtp_port" type="number" min="1" max="65535" />
              </AdminFormField>
              <AdminFormField label="SMTP 用户名">
                <Input v-model="emailSettings.smtp_username" autocomplete="off" />
              </AdminFormField>
              <AdminFormField label="SMTP 密码">
                <div class="relative">
                  <Input
                    v-model="emailSettings.smtp_password"
                    :type="showSmtpPassword ? 'text' : 'password'"
                    class="pr-10"
                    autocomplete="new-password"
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    class="absolute right-0 top-0"
                    :aria-label="showSmtpPassword ? '隐藏 SMTP 密码' : '显示 SMTP 密码'"
                    @click="showSmtpPassword = !showSmtpPassword"
                  >
                    <EyeOff v-if="showSmtpPassword" class="size-4" />
                    <Eye v-else class="size-4" />
                  </Button>
                </div>
              </AdminFormField>
              <AdminFormField label="发件人邮箱">
                <Input v-model="emailSettings.from_email" type="email" />
              </AdminFormField>
              <AdminFormField label="发件人名称">
                <Input v-model="emailSettings.from_name" />
              </AdminFormField>
            </div>
          </SettingsSection>
        </TabsContent>

        <TabsContent value="seo">
          <SettingsSection title="默认搜索信息" description="未单独配置页面 SEO 时使用的默认值。">
            <div class="grid gap-4 md:grid-cols-2">
              <AdminFormField label="Meta 标题" class="md:col-span-2">
                <Input v-model="seoSettings.meta_title" />
              </AdminFormField>
              <AdminFormField label="Meta 描述" class="md:col-span-2">
                <Textarea v-model="seoSettings.meta_description" class="min-h-24" />
              </AdminFormField>
              <AdminFormField label="Meta 关键词" class="md:col-span-2">
                <Input v-model="seoSettings.meta_keywords" placeholder="用逗号分隔" />
              </AdminFormField>
              <AdminFormField label="Google Analytics">
                <Input v-model="seoSettings.google_analytics" placeholder="GA 跟踪 ID" />
              </AdminFormField>
              <AdminFormField label="Google Tag Manager">
                <Input v-model="seoSettings.google_tag_manager" placeholder="GTM ID" />
              </AdminFormField>
            </div>
          </SettingsSection>
        </TabsContent>

        <TabsContent value="social">
          <SettingsSection title="社交媒体" description="前台展示的官方账号与页面链接。">
            <div class="grid gap-4 md:grid-cols-2">
              <AdminFormField v-for="field in socialFields" :key="field.key" :label="field.label">
                <Input v-model="socialSettings[field.key]" type="url" :placeholder="field.placeholder" />
              </AdminFormField>
            </div>
          </SettingsSection>
        </TabsContent>

        <TabsContent value="payment">
          <SettingsSection title="支付网关" description="支付凭据仅供后端使用，不公开到前台。">
            <div class="grid gap-4 md:grid-cols-2">
              <AdminFormField label="支付网关">
                <Select v-model="paymentSettings.gateway">
                  <SelectTrigger class="w-full"><SelectValue placeholder="请选择支付网关" /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="stripe">Stripe</SelectItem>
                    <SelectItem value="paypal">PayPal</SelectItem>
                    <SelectItem value="alipay">支付宝</SelectItem>
                    <SelectItem value="wechat">微信支付</SelectItem>
                  </SelectContent>
                </Select>
              </AdminFormField>
              <div class="flex items-center justify-between gap-3 rounded-lg border px-3 py-2.5">
                <div>
                  <span class="text-xs font-medium">测试模式</span>
                  <p class="mt-0.5 text-xs text-muted-foreground">启用后使用网关测试环境。</p>
                </div>
                <Switch v-model="paymentSettings.test_mode" aria-label="支付测试模式" />
              </div>
              <AdminFormField label="API Key" class="md:col-span-2">
                <div class="relative">
                  <Input v-model="paymentSettings.api_key" :type="showPaymentSecrets ? 'text' : 'password'" class="pr-10 font-mono" autocomplete="off" />
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    class="absolute right-0 top-0"
                    :aria-label="showPaymentSecrets ? '隐藏支付凭据' : '显示支付凭据'"
                    @click="showPaymentSecrets = !showPaymentSecrets"
                  >
                    <EyeOff v-if="showPaymentSecrets" class="size-4" />
                    <Eye v-else class="size-4" />
                  </Button>
                </div>
              </AdminFormField>
              <AdminFormField label="API Secret" class="md:col-span-2">
                <Input v-model="paymentSettings.api_secret" :type="showPaymentSecrets ? 'text' : 'password'" class="font-mono" autocomplete="off" />
              </AdminFormField>
            </div>
          </SettingsSection>
        </TabsContent>

        <TabsContent value="public_chat" class="space-y-4">
          <Alert>
            <Info class="size-4" />
            <AlertTitle>公开客服状态</AlertTitle>
            <AlertDescription>公开客服需绑定活跃用户，且角色为 admin、manager 或 support。</AlertDescription>
          </Alert>

          <div class="flex justify-end">
            <Button variant="outline" size="sm" :disabled="loadingPublicChatAgents" @click="fetchPublicChatAgents">
              <RefreshCw :class="['size-3.5', { 'animate-spin': loadingPublicChatAgents }]" />
              刷新概览
            </Button>
          </div>

          <section class="grid gap-3 sm:grid-cols-2" aria-label="Public Chat 客服统计">
            <div class="rounded-lg border bg-card p-4 shadow-xs">
              <span class="text-xs font-medium text-muted-foreground">Profile 总数</span>
              <strong class="mt-2 block text-2xl font-semibold tabular-nums">{{ publicChatAgentsSummary.profile_count || 0 }}</strong>
            </div>
            <div class="rounded-lg border bg-card p-4 shadow-xs">
              <span class="text-xs font-medium text-muted-foreground">公开客服数</span>
              <strong class="mt-2 block text-2xl font-semibold tabular-nums">{{ publicChatAgentsSummary.exposed_agents || 0 }}</strong>
            </div>
          </section>

          <Alert v-for="warning in publicChatAgentWarnings" :key="warning" class="border-amber-200 bg-amber-50 text-amber-900">
            <TriangleAlert class="size-4" />
            <AlertTitle>配置提醒</AlertTitle>
            <AlertDescription>{{ warning }}</AlertDescription>
          </Alert>

          <AdminTablePanel :loading="loadingPublicChatAgents">
            <Table class="min-w-[1280px]">
              <TableHeader>
                <TableRow>
                  <TableHead class="w-16">ID</TableHead>
                  <TableHead>客服</TableHead>
                  <TableHead class="w-32">Agent ID</TableHead>
                  <TableHead class="w-20">User ID</TableHead>
                  <TableHead class="w-36">原始角色</TableHead>
                  <TableHead class="w-28">Go 角色</TableHead>
                  <TableHead class="w-28">用户状态</TableHead>
                  <TableHead class="w-28">Profile</TableHead>
                  <TableHead class="w-24">在线状态</TableHead>
                  <TableHead class="w-24">公开</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableEmpty v-if="publicChatAgents.length === 0" :colspan="10">
                  <div class="flex flex-col items-center text-muted-foreground">
                    <Headset class="mb-2 size-7 opacity-55" />
                    <span class="text-xs">暂无 Public Chat 客服 Profile</span>
                  </div>
                </TableEmpty>
                <TableRow v-for="agent in publicChatAgents" :key="agent.id">
                  <TableCell class="font-mono text-xs text-muted-foreground">{{ agent.id }}</TableCell>
                  <TableCell>
                    <div class="flex items-center gap-2.5">
                      <Avatar class="size-8">
                        <AvatarImage v-if="agent.avatar" :src="agent.avatar" :alt="agent.display_name" />
                        <AvatarFallback>{{ agentInitials(agent) }}</AvatarFallback>
                      </Avatar>
                      <div class="min-w-0">
                        <span class="block truncate font-medium">{{ agent.display_name || agent.username || '-' }}</span>
                        <span class="block truncate text-xs text-muted-foreground">{{ agent.email || agent.whatsapp || '-' }}</span>
                      </div>
                    </div>
                  </TableCell>
                  <TableCell class="font-mono text-xs">{{ agent.agent_id || '-' }}</TableCell>
                  <TableCell class="font-mono text-xs">{{ agent.user_id || '-' }}</TableCell>
                  <TableCell>{{ agent.raw_role || '-' }}</TableCell>
                  <TableCell>{{ agent.normalized_role || '-' }}</TableCell>
                  <TableCell><AdminStatusBadge :tone="agent.user_status === 'active' ? 'green' : 'gray'">{{ agent.user_status || '-' }}</AdminStatusBadge></TableCell>
                  <TableCell><AdminStatusBadge :tone="agent.profile_status === 'active' ? 'green' : 'gray'">{{ agent.profile_status || '-' }}</AdminStatusBadge></TableCell>
                  <TableCell><AdminStatusBadge :tone="agent.online_status === 'online' ? 'green' : 'gray'">{{ agent.online_status || '-' }}</AdminStatusBadge></TableCell>
                  <TableCell><AdminStatusBadge :tone="agent.exposed ? 'green' : 'coral'">{{ agent.exposed ? '是' : '否' }}</AdminStatusBadge></TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </AdminTablePanel>
        </TabsContent>
      </Tabs>
    </div>
  </div>
</template>

<script setup>
import { computed, defineComponent, h, onMounted, reactive, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import {
  CreditCard,
  Eye,
  EyeOff,
  Globe2,
  Headset,
  Info,
  LoaderCircle,
  Mail,
  RefreshCw,
  Save,
  SearchCheck,
  Share2,
  TriangleAlert
} from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Textarea } from '@/components/ui/textarea'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

const SettingsSection = defineComponent({
  props: {
    title: { type: String, required: true },
    description: { type: String, default: '' }
  },
  setup(props, { slots }) {
    return () => h('section', { class: 'grid max-w-5xl gap-5 lg:grid-cols-[190px_minmax(0,1fr)]' }, [
      h('div', {}, [
        h('h2', { class: 'text-sm font-semibold' }, props.title),
        props.description ? h('p', { class: 'mt-1 text-xs leading-5 text-muted-foreground' }, props.description) : null
      ]),
      h('div', { class: 'min-w-0' }, slots.default?.())
    ])
  }
})

const authStore = useAuthStore()
const activeTab = ref('site')
const saving = ref(false)
const loadingSettings = ref(false)
const showSmtpPassword = ref(false)
const showPaymentSecrets = ref(false)
const loadedGroups = new Set()

const siteSettings = reactive({ site_name: '', site_description: '', site_logo: '', contact_email: '', contact_phone: '' })
const emailSettings = reactive({ smtp_host: '', smtp_port: 587, smtp_username: '', smtp_password: '', from_email: '', from_name: '' })
const seoSettings = reactive({ meta_title: '', meta_description: '', meta_keywords: '', google_analytics: '', google_tag_manager: '' })
const socialSettings = reactive({ facebook: '', twitter: '', instagram: '', linkedin: '', youtube: '', wechat: '' })
const paymentSettings = reactive({ gateway: '', api_key: '', api_secret: '', test_mode: true })

const loadingPublicChatAgents = ref(false)
const publicChatAgentsOverview = ref(null)
const publicChatAgentsSummary = computed(() => publicChatAgentsOverview.value?.summary || {})
const publicChatAgents = computed(() => publicChatAgentsOverview.value?.agents || [])
const publicChatAgentWarnings = computed(() => publicChatAgentsOverview.value?.warnings || [])

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
      site_description: { type: 'string', public: true, description: 'Site description' },
      site_logo: { type: 'string', public: true, description: 'Site logo URL' },
      contact_email: { type: 'string', public: true, description: 'Contact email' },
      contact_phone: { type: 'string', public: true, description: 'Contact phone' }
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
      api_key: { type: 'string', public: false, description: 'Payment API key' },
      api_secret: { type: 'string', public: false, description: 'Payment API secret' },
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
const settingKey = (setting, group) => setting.key.startsWith(`${group}_`) ? setting.key.slice(group.length + 1) : setting.key

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
      const key = settingKey(setting, group)
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
    await fetchSettings(group, true)
  } catch (error) {
    console.error('Failed to save settings:', error)
  } finally {
    saving.value = false
  }
}

const agentInitials = (agent) => {
  const name = agent.display_name || agent.username || agent.email || '?'
  return name.slice(0, 2).toUpperCase()
}

watch(activeTab, (tab) => {
  if (tab === 'public_chat') fetchPublicChatAgents()
  else fetchSettings(tab)
})

onMounted(() => fetchSettings('site'))
</script>
