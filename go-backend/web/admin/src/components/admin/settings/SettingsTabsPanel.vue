<template>
  <Tabs :model-value="activeTab" class="gap-5" @update:model-value="emit('update:activeTab', $event)">
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
                @click="emit('update:showSmtpPassword', !showSmtpPassword)"
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
                @click="emit('update:showPaymentSecrets', !showPaymentSecrets)"
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
      <PublicChatSettingsPanel
        :loading-agents="loadingPublicChatAgents"
        :loading-candidates="loadingPublicChatAgentCandidates"
        :summary="publicChatAgentsSummary"
        :agents="publicChatAgents"
        :warnings="publicChatAgentWarnings"
        :can-edit="canEdit"
        @open-agent-dialog="emit('open-agent-dialog')"
        @refresh="emit('refresh-public-chat')"
      />
    </TabsContent>
  </Tabs>
</template>

<script setup>
import { defineComponent, h } from 'vue'
import {
  CreditCard,
  Eye,
  EyeOff,
  Globe2,
  Headset,
  Mail,
  SearchCheck,
  Share2,
} from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import PublicChatSettingsPanel from '@/components/admin/settings/PublicChatSettingsPanel.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Textarea } from '@/components/ui/textarea'

defineProps({
  activeTab: { type: String, default: 'site' },
  siteSettings: { type: Object, required: true },
  emailSettings: { type: Object, required: true },
  seoSettings: { type: Object, required: true },
  socialSettings: { type: Object, required: true },
  paymentSettings: { type: Object, required: true },
  socialFields: { type: Array, default: () => [] },
  showSmtpPassword: { type: Boolean, default: false },
  showPaymentSecrets: { type: Boolean, default: false },
  loadingPublicChatAgents: { type: Boolean, default: false },
  loadingPublicChatAgentCandidates: { type: Boolean, default: false },
  publicChatAgentsSummary: { type: Object, default: () => ({}) },
  publicChatAgents: { type: Array, default: () => [] },
  publicChatAgentWarnings: { type: Array, default: () => [] },
  canEdit: { type: Boolean, default: false },
})

const emit = defineEmits([
  'update:activeTab',
  'update:showSmtpPassword',
  'update:showPaymentSecrets',
  'open-agent-dialog',
  'refresh-public-chat',
])

const SettingsSection = defineComponent({
  props: {
    title: { type: String, required: true },
    description: { type: String, default: '' }
  },
  setup(props, { slots }) {
    return () => h('section', { class: 'grid max-w-5xl gap-5 lg:grid-cols-[190px_minmax(0,1fr)]' }, [
      h('div', {}, [
        h('h2', { class: 'text-sm font-black tracking-tighter italic uppercase text-foreground' }, props.title),
        props.description ? h('p', { class: 'mt-1 text-[9px] font-black uppercase tracking-widest text-muted-foreground/60' }, props.description) : null
      ]),
      h('div', { class: 'min-w-0' }, slots.default?.())
    ])
  }
})
</script>
